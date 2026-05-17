package audit

import (
	"context"
	"fmt"
	"log/slog"
	"sync/atomic"
	"time"

	"smartdelivery/apps/api-go/internal/model"
)

const DefaultBufferSize = 100

const (
	EventDeliveryRuleCreated       = "delivery_rule.created"
	EventDeliveryRuleStatusUpdated = "delivery_rule.status_updated"
)

type OverflowBehavior string

const (
	OverflowBlock OverflowBehavior = "block"
	OverflowDrop  OverflowBehavior = "drop"
)

type Event struct {
	ShopID         uint
	DeliveryRuleID *uint
	Type           string
	Message        string
}

type Store interface {
	CreateAuditLog(ctx context.Context, log *model.AuditLog) error
}

type Publisher interface {
	Publish(ctx context.Context, event Event) error
}

type Config struct {
	BufferSize       int
	OverflowBehavior OverflowBehavior
	Logger           *slog.Logger
}

type Worker struct {
	store    Store
	events   chan Event
	overflow OverflowBehavior
	logger   *slog.Logger
	metrics  Metrics
}

type MetricsSnapshot struct {
	Published uint64
	Dropped   uint64
	Blocked   uint64
}

type Metrics struct {
	published atomic.Uint64
	dropped   atomic.Uint64
	blocked   atomic.Uint64
}

func NewWorker(store Store, configs ...Config) *Worker {
	cfg := Config{
		BufferSize:       DefaultBufferSize,
		OverflowBehavior: OverflowBlock,
		Logger:           slog.Default(),
	}
	if len(configs) > 0 {
		cfg = configs[0].withDefaults()
	}

	return &Worker{
		store:    store,
		events:   make(chan Event, cfg.BufferSize),
		overflow: cfg.OverflowBehavior,
		logger:   cfg.Logger,
	}
}

func (worker *Worker) Publish(ctx context.Context, event Event) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	select {
	case worker.events <- event:
		worker.metrics.published.Add(1)
		return nil
	default:
	}

	switch worker.overflow {
	case OverflowDrop:
		worker.metrics.dropped.Add(1)
		worker.logger.Warn("audit event dropped", "event", event.Type, "shop_id", event.ShopID)
		return nil
	default:
		worker.metrics.blocked.Add(1)
		worker.logger.Warn("audit event publish blocked", "event", event.Type, "shop_id", event.ShopID)
	}

	select {
	case worker.events <- event:
		worker.metrics.published.Add(1)
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (worker *Worker) Metrics() MetricsSnapshot {
	return MetricsSnapshot{
		Published: worker.metrics.published.Load(),
		Dropped:   worker.metrics.dropped.Load(),
		Blocked:   worker.metrics.blocked.Load(),
	}
}

func (worker *Worker) Run(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case event := <-worker.events:
			if err := worker.store.CreateAuditLog(ctx, event.toModel()); err != nil {
				return fmt.Errorf("write audit log: %w", err)
			}
		}
	}
}

func (cfg Config) withDefaults() Config {
	if cfg.BufferSize < 0 {
		cfg.BufferSize = 0
	}
	if cfg.OverflowBehavior == "" {
		cfg.OverflowBehavior = OverflowBlock
	}
	if cfg.Logger == nil {
		cfg.Logger = slog.Default()
	}
	return cfg
}

func (event Event) toModel() *model.AuditLog {
	return &model.AuditLog{
		ShopID:         event.ShopID,
		DeliveryRuleID: event.DeliveryRuleID,
		Event:          event.Type,
		Message:        event.Message,
		CreatedAt:      time.Now().UTC(),
	}
}
