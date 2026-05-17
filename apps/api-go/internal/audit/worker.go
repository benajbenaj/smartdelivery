package audit

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"smartdelivery/apps/api-go/internal/model"
	workerpkg "smartdelivery/apps/api-go/internal/worker"
)

const (
	DefaultBufferSize   = 100
	DefaultDrainTimeout = workerpkg.DefaultDrainTimeout
)

const (
	EventDeliveryRuleCreated       = "delivery_rule.created"
	EventDeliveryRuleStatusUpdated = "delivery_rule.status_updated"
)

type OverflowBehavior = workerpkg.OverflowBehavior

const (
	OverflowBlock = workerpkg.OverflowBlock
	OverflowDrop  = workerpkg.OverflowDrop
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
	DrainTimeout     time.Duration
}

type Worker struct {
	queue *workerpkg.QueueWorker[Event]
}

type MetricsSnapshot = workerpkg.QueueMetricsSnapshot

func NewWorker(store Store, configs ...Config) *Worker {
	cfg := Config{
		BufferSize:       DefaultBufferSize,
		OverflowBehavior: OverflowBlock,
		Logger:           slog.Default(),
		DrainTimeout:     DefaultDrainTimeout,
	}
	if len(configs) > 0 {
		cfg = configs[0].withDefaults()
	}

	return &Worker{
		queue: workerpkg.NewQueueWorker(func(ctx context.Context, event Event) error {
			if err := store.CreateAuditLog(ctx, event.toModel()); err != nil {
				return fmt.Errorf("write audit log: %w", err)
			}
			return nil
		}, workerpkg.QueueWorkerConfig[Event]{
			BufferSize:       cfg.BufferSize,
			OverflowBehavior: cfg.OverflowBehavior,
			Logger:           cfg.Logger,
			DrainTimeout:     cfg.DrainTimeout,
			LogName:          "audit event",
			LogAttrs: func(event Event) []any {
				return []any{"event", event.Type, "shop_id", event.ShopID}
			},
		}),
	}
}

func (worker *Worker) Publish(ctx context.Context, event Event) error {
	return worker.queue.Publish(ctx, event)
}

func (worker *Worker) Metrics() MetricsSnapshot {
	return worker.queue.Metrics()
}

func (worker *Worker) Run(ctx context.Context) error {
	return worker.queue.Run(ctx)
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
	if cfg.DrainTimeout == 0 {
		cfg.DrainTimeout = DefaultDrainTimeout
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
