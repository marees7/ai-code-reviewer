package dedup

import (
	"context"
	"sync"
)

type Memory struct {
	mu   sync.Mutex
	seen map[string]bool
}

func NewMemory() *Memory {
	return &Memory{
		seen: make(map[string]bool),
	}
}

func (m *Memory) Seen(ctx context.Context, key string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.seen[key]
}

func (m *Memory) Mark(ctx context.Context, key string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.seen[key] = true
	return nil
}
