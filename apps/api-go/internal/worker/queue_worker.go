package worker

import (
	"context"
	"log/slog"
	"sync/atomic"
	"time"
)

const DefaultDrainTimeout = 5 * time.Second

type OverflowBehavior string

const (
	OverflowBlock OverflowBehavior = "block"
	OverflowDrop  OverflowBehavior = "drop"
)

type QueueWorkerConfig[T any] struct {
	BufferSize       int
	OverflowBehavior OverflowBehavior
	Logger           *slog.Logger
	DrainTimeout     time.Duration
	LogName          string
	LogAttrs         func(item T) []any
}

type QueueMetricsSnapshot struct {
	Published uint64
	Dropped   uint64
	Blocked   uint64
}

type QueueMetrics struct {
	published atomic.Uint64
	dropped   atomic.Uint64
	blocked   atomic.Uint64
}

type QueueWorker[T any] struct {
	handle       func(context.Context, T) error
	events       chan T
	overflow     OverflowBehavior
	logger       *slog.Logger
	drainTimeout time.Duration
	logName      string
	logAttrs     func(T) []any
	metrics      QueueMetrics
}

func NewQueueWorker[T any](handle func(context.Context, T) error, cfg QueueWorkerConfig[T]) *QueueWorker[T] {
	cfg = cfg.withDefaults()

	return &QueueWorker[T]{
		handle:       handle,
		events:       make(chan T, cfg.BufferSize),
		overflow:     cfg.OverflowBehavior,
		logger:       cfg.Logger,
		drainTimeout: cfg.DrainTimeout,
		logName:      cfg.LogName,
		logAttrs:     cfg.LogAttrs,
	}
}

func (worker *QueueWorker[T]) Publish(ctx context.Context, item T) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	select {
	case worker.events <- item:
		worker.metrics.published.Add(1)
		return nil
	default:
	}

	switch worker.overflow {
	case OverflowDrop:
		worker.metrics.dropped.Add(1)
		worker.logger.Warn(worker.logName+" dropped", worker.logAttributes(item)...)
		return nil
	default:
		worker.metrics.blocked.Add(1)
		worker.logger.Warn(worker.logName+" publish blocked", worker.logAttributes(item)...)
	}

	select {
	case worker.events <- item:
		worker.metrics.published.Add(1)
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (worker *QueueWorker[T]) Metrics() QueueMetricsSnapshot {
	return QueueMetricsSnapshot{
		Published: worker.metrics.published.Load(),
		Dropped:   worker.metrics.dropped.Load(),
		Blocked:   worker.metrics.blocked.Load(),
	}
}

func (worker *QueueWorker[T]) Run(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			return worker.shutdown(ctx)
		default:
		}

		select {
		case item := <-worker.events:
			if ctx.Err() != nil {
				return worker.shutdown(ctx, item)
			}
			if err := worker.handle(ctx, item); err != nil {
				return err
			}
		case <-ctx.Done():
			return worker.shutdown(ctx)
		}
	}
}

func (worker *QueueWorker[T]) shutdown(ctx context.Context, pending ...T) error {
	drainCtx, cancel := context.WithTimeout(context.Background(), worker.drainTimeout)
	defer cancel()

	for _, item := range pending {
		if err := worker.handle(drainCtx, item); err != nil {
			return err
		}
	}

	for {
		select {
		case item := <-worker.events:
			if err := worker.handle(drainCtx, item); err != nil {
				return err
			}
		default:
			return ctx.Err()
		}
	}
}

func (worker *QueueWorker[T]) logAttributes(item T) []any {
	if worker.logAttrs == nil {
		return nil
	}
	return worker.logAttrs(item)
}

func (cfg QueueWorkerConfig[T]) withDefaults() QueueWorkerConfig[T] {
	if cfg.BufferSize < 0 {
		cfg.BufferSize = 0
	}
	if cfg.OverflowBehavior == "" {
		cfg.OverflowBehavior = OverflowBlock
	}
	if cfg.Logger == nil {
		cfg.Logger = slog.Default()
	}
	if cfg.DrainTimeout == 0 {
		cfg.DrainTimeout = DefaultDrainTimeout
	}
	if cfg.LogName == "" {
		cfg.LogName = "worker item"
	}
	return cfg
}
