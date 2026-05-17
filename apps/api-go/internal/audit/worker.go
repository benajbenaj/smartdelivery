package audit

import (
	"context"
	"fmt"
	"time"

	"smartdelivery/apps/api-go/internal/model"
)

const (
	EventDeliveryRuleCreated       = "delivery_rule.created"
	EventDeliveryRuleStatusUpdated = "delivery_rule.status_updated"
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

type Worker struct {
	store  Store
	events chan Event
}

func NewWorker(store Store) *Worker {
	return &Worker{
		store:  store,
		events: make(chan Event),
	}
}

func (worker *Worker) Publish(ctx context.Context, event Event) error {
	select {
	case worker.events <- event:
		return nil
	case <-ctx.Done():
		return ctx.Err()
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

func (event Event) toModel() *model.AuditLog {
	return &model.AuditLog{
		ShopID:         event.ShopID,
		DeliveryRuleID: event.DeliveryRuleID,
		Event:          event.Type,
		Message:        event.Message,
		CreatedAt:      time.Now().UTC(),
	}
}
