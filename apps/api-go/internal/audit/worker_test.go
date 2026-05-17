package audit

import (
	"context"
	"errors"
	"testing"
	"time"

	"smartdelivery/apps/api-go/internal/model"
)

func TestWorkerWritesAuditLog(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	store := &fakeStore{created: make(chan *model.AuditLog, 1)}
	worker := NewWorker(store)

	go func() {
		if err := worker.Run(ctx); err != nil && !errors.Is(err, context.Canceled) {
			t.Errorf("Run() error = %v", err)
		}
	}()

	ruleID := uint(42)
	if err := worker.Publish(ctx, Event{
		ShopID:         7,
		DeliveryRuleID: &ruleID,
		Type:           EventDeliveryRuleCreated,
		Message:        "delivery rule created",
	}); err != nil {
		t.Fatalf("Publish() error = %v", err)
	}

	select {
	case log := <-store.created:
		if log.ShopID != 7 {
			t.Fatalf("ShopID = %d, want 7", log.ShopID)
		}
		if log.DeliveryRuleID == nil || *log.DeliveryRuleID != ruleID {
			t.Fatalf("DeliveryRuleID = %v, want %d", log.DeliveryRuleID, ruleID)
		}
		if log.Event != EventDeliveryRuleCreated {
			t.Fatalf("Event = %q, want %q", log.Event, EventDeliveryRuleCreated)
		}
		if log.Message != "delivery rule created" {
			t.Fatalf("Message = %q, want delivery rule created", log.Message)
		}
		if log.CreatedAt.IsZero() {
			t.Fatalf("CreatedAt is zero, want timestamp")
		}
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for audit log")
	}
}

func TestPublishUsesUnbufferedBackpressure(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	store := &fakeStore{created: make(chan *model.AuditLog, 1)}
	worker := NewWorker(store)
	published := make(chan error, 1)

	go func() {
		published <- worker.Publish(ctx, Event{
			ShopID:  1,
			Type:    EventDeliveryRuleCreated,
			Message: "delivery rule created",
		})
	}()

	select {
	case err := <-published:
		t.Fatalf("Publish() completed before worker received event: %v", err)
	case <-time.After(20 * time.Millisecond):
	}

	go func() {
		if err := worker.Run(ctx); err != nil && !errors.Is(err, context.Canceled) {
			t.Errorf("Run() error = %v", err)
		}
	}()

	select {
	case err := <-published:
		if err != nil {
			t.Fatalf("Publish() error = %v", err)
		}
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for publish")
	}
}

func TestWorkerStopsWhenContextIsCanceled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	worker := NewWorker(&fakeStore{created: make(chan *model.AuditLog, 1)})

	err := worker.Run(ctx)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("Run() error = %v, want context.Canceled", err)
	}
}

func TestPublishReturnsContextErrorWhenNoWorkerReceives(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	worker := NewWorker(&fakeStore{created: make(chan *model.AuditLog, 1)})

	err := worker.Publish(ctx, Event{
		ShopID:  1,
		Type:    EventDeliveryRuleCreated,
		Message: "delivery rule created",
	})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("Publish() error = %v, want context.Canceled", err)
	}
}

type fakeStore struct {
	created chan *model.AuditLog
}

func (store *fakeStore) CreateAuditLog(_ context.Context, log *model.AuditLog) error {
	store.created <- log
	return nil
}
