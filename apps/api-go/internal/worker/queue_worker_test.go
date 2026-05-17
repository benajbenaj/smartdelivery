package worker

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"testing"
)

func TestQueueWorkerPublishesAndRuns(t *testing.T) {
	handled := make(chan int, 1)
	worker := NewQueueWorker(func(_ context.Context, item int) error {
		handled <- item
		return nil
	}, testQueueConfig[int](1, OverflowBlock))

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		if err := worker.Run(ctx); err != nil && !errors.Is(err, context.Canceled) {
			t.Errorf("Run() error = %v", err)
		}
	}()

	if err := worker.Publish(ctx, 7); err != nil {
		t.Fatalf("Publish() error = %v", err)
	}
	if got := <-handled; got != 7 {
		t.Fatalf("handled = %d, want 7", got)
	}
}

func TestQueueWorkerDropsWhenBufferIsFull(t *testing.T) {
	worker := NewQueueWorker(func(_ context.Context, _ int) error {
		return nil
	}, testQueueConfig[int](1, OverflowDrop))

	if err := worker.Publish(context.Background(), 1); err != nil {
		t.Fatalf("first Publish() error = %v", err)
	}
	if err := worker.Publish(context.Background(), 2); err != nil {
		t.Fatalf("second Publish() error = %v", err)
	}

	metrics := worker.Metrics()
	if metrics.Published != 1 {
		t.Fatalf("Published = %d, want 1", metrics.Published)
	}
	if metrics.Dropped != 1 {
		t.Fatalf("Dropped = %d, want 1", metrics.Dropped)
	}
}

func testQueueConfig[T any](bufferSize int, overflow OverflowBehavior) QueueWorkerConfig[T] {
	return QueueWorkerConfig[T]{
		BufferSize:       bufferSize,
		OverflowBehavior: overflow,
		Logger:           slog.New(slog.NewTextHandler(io.Discard, nil)),
		LogName:          "test item",
	}
}
