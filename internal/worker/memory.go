package worker

import (
	"context"
)

type MemoryQueue struct {
	ch chan Job
}

func NewMemoryQueue(size int) *MemoryQueue {
	return &MemoryQueue{
		ch: make(chan Job, size),
	}
}

func (m *MemoryQueue) Push(ctx context.Context, j Job) error {
	select {
	case m.ch <- j:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (m *MemoryQueue) Pop(ctx context.Context) (Job, error) {
	select {
	case j := <-m.ch:
		return j, nil
	case <-ctx.Done():
		return Job{}, ctx.Err()
	}
}
