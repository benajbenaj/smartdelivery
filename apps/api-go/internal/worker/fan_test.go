package worker

import (
	"context"
	"testing"
)

func TestFanOutSplitsJobsAcrossWorkers(t *testing.T) {
	jobs := make(chan int, 3)
	jobs <- 1
	jobs <- 2
	jobs <- 3
	close(jobs)

	outputs := FanOut(context.Background(), 2, jobs, func(_ context.Context, job int) (int, bool) {
		return job * 2, true
	})

	results := FanIn(context.Background(), 3, outputs...)

	seen := map[int]bool{}
	for result := range results {
		seen[result] = true
	}

	for _, expected := range []int{2, 4, 6} {
		if !seen[expected] {
			t.Fatalf("missing result %d from %v", expected, seen)
		}
	}
}

func TestFanInClosesAfterInputsClose(t *testing.T) {
	first := make(chan int, 1)
	second := make(chan int, 1)
	first <- 1
	second <- 2
	close(first)
	close(second)

	results := FanIn(context.Background(), 2, first, second)

	count := 0
	for range results {
		count++
	}
	if count != 2 {
		t.Fatalf("merged result count = %d, want 2", count)
	}
}
