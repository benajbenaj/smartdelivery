package worker

import (
	"context"
	"sync"
)

type FanOutFunc[Job any, Result any] func(ctx context.Context, job Job) (Result, bool)

func FanOut[Job any, Result any](ctx context.Context, workerCount int, jobs <-chan Job, handle FanOutFunc[Job, Result]) []<-chan Result {
	if workerCount <= 0 {
		workerCount = 1
	}

	outputs := make([]<-chan Result, 0, workerCount)
	for range workerCount {
		output := make(chan Result)
		go func() {
			defer close(output)
			for job := range jobs {
				if ctx.Err() != nil {
					return
				}

				result, ok := handle(ctx, job)
				if !ok {
					continue
				}

				select {
				case output <- result:
				case <-ctx.Done():
					return
				}
			}
		}()
		outputs = append(outputs, output)
	}

	return outputs
}

func FanIn[T any](ctx context.Context, bufferSize int, inputs ...<-chan T) <-chan T {
	if bufferSize < 0 {
		bufferSize = 0
	}

	output := make(chan T, bufferSize)

	var wg sync.WaitGroup
	wg.Add(len(inputs))
	for _, input := range inputs {
		go func(input <-chan T) {
			defer wg.Done()
			for value := range input {
				select {
				case output <- value:
				case <-ctx.Done():
					return
				}
			}
		}(input)
	}

	go func() {
		wg.Wait()
		close(output)
	}()

	return output
}
