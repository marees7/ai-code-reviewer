package dedup

import (
	"context"
	"sync"
	"time"
)

type Memory struct {
	mu         sync.Mutex
	seen       map[string]time.Time
	order      []string
	maxEntries int
	ttl        time.Duration
}

func NewMemory() *Memory {
	return &Memory{
		seen:       make(map[string]time.Time),
		order:      make([]string, 0, 1024),
		maxEntries: 10000,
		ttl:        24 * time.Hour,
	}
}

func (m *Memory) Seen(ctx context.Context, key string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()

	exp, ok := m.seen[key]
	if !ok {
		return false
	}

	if time.Now().After(exp) {
		delete(m.seen, key)
		return false
	}

	return true
}

func (m *Memory) Mark(ctx context.Context, key string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	now := time.Now()
	m.seen[key] = now.Add(m.ttl)
	m.order = append(m.order, key)

	// Keep memory bounded by trimming old keys in insertion order.
	for len(m.seen) > m.maxEntries && len(m.order) > 0 {
		oldest := m.order[0]
		m.order = m.order[1:]

		exp, ok := m.seen[oldest]
		if !ok {
			continue
		}
		if now.After(exp) || len(m.seen) > m.maxEntries {
			delete(m.seen, oldest)
		}
	}

	return nil
}
