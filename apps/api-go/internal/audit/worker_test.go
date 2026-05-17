package audit

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"testing"
	"time"

	"smartdelivery/apps/api-go/internal/model"
)

func TestWorkerWritesAuditLog(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	store := &fakeStore{created: make(chan *model.AuditLog, 1)}
	worker := NewWorker(store, testWorkerConfig(1, OverflowBlock))

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

func TestPublishUsesBufferedChannel(t *testing.T) {
	worker := NewWorker(
		&fakeStore{created: make(chan *model.AuditLog, 1)},
		testWorkerConfig(1, OverflowBlock),
	)

	if err := worker.Publish(context.Background(), Event{
		ShopID:  1,
		Type:    EventDeliveryRuleCreated,
		Message: "delivery rule created",
	}); err != nil {
		t.Fatalf("Publish() error = %v", err)
	}

	metrics := worker.Metrics()
	if metrics.Published != 1 {
		t.Fatalf("Published = %d, want 1", metrics.Published)
	}
}

func TestPublishDropsWhenBufferIsFull(t *testing.T) {
	worker := NewWorker(
		&fakeStore{created: make(chan *model.AuditLog, 1)},
		testWorkerConfig(1, OverflowDrop),
	)

	if err := worker.Publish(context.Background(), Event{
		ShopID:  1,
		Type:    EventDeliveryRuleCreated,
		Message: "delivery rule created",
	}); err != nil {
		t.Fatalf("Publish() first error = %v", err)
	}
	if err := worker.Publish(context.Background(), Event{
		ShopID:  1,
		Type:    EventDeliveryRuleStatusUpdated,
		Message: "delivery rule status updated",
	}); err != nil {
		t.Fatalf("Publish() second error = %v", err)
	}

	metrics := worker.Metrics()
	if metrics.Published != 1 {
		t.Fatalf("Published = %d, want 1", metrics.Published)
	}
	if metrics.Dropped != 1 {
		t.Fatalf("Dropped = %d, want 1", metrics.Dropped)
	}
}

func TestPublishBlocksWhenBufferIsFull(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	store := &fakeStore{created: make(chan *model.AuditLog, 2)}
	worker := NewWorker(store, testWorkerConfig(1, OverflowBlock))

	if err := worker.Publish(ctx, Event{
		ShopID:  1,
		Type:    EventDeliveryRuleCreated,
		Message: "delivery rule created",
	}); err != nil {
		t.Fatalf("Publish() first error = %v", err)
	}

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
		t.Fatalf("Publish() completed before buffer had space: %v", err)
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

	metrics := worker.Metrics()
	if metrics.Blocked != 1 {
		t.Fatalf("Blocked = %d, want 1", metrics.Blocked)
	}
}

func TestWorkerStopsWhenContextIsCanceled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	worker := NewWorker(
		&fakeStore{created: make(chan *model.AuditLog, 1)},
		testWorkerConfig(1, OverflowBlock),
	)

	err := worker.Run(ctx)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("Run() error = %v, want context.Canceled", err)
	}
}

func TestPublishReturnsContextErrorWhenNoWorkerReceives(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	worker := NewWorker(
		&fakeStore{created: make(chan *model.AuditLog, 1)},
		testWorkerConfig(1, OverflowBlock),
	)

	if err := worker.Publish(context.Background(), Event{
		ShopID:  1,
		Type:    EventDeliveryRuleCreated,
		Message: "delivery rule created",
	}); err != nil {
		t.Fatalf("Publish() first error = %v", err)
	}

	cancel()

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

func testWorkerConfig(bufferSize int, overflow OverflowBehavior) Config {
	return Config{
		BufferSize:       bufferSize,
		OverflowBehavior: overflow,
		Logger:           slog.New(slog.NewTextHandler(io.Discard, nil)),
	}
}
