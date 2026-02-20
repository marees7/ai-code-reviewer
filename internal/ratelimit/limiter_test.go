package ratelimit

import (
	"testing"
	"time"
)

func TestLimiter_PrunesExpiredEntries(t *testing.T) {
	l := New(1, 1)
	l.ttl = 5 * time.Millisecond

	first := l.Get("repo-a")
	if first == nil {
		t.Fatalf("expected limiter instance")
	}

	time.Sleep(10 * time.Millisecond)
	l.lastPruned = time.Now().Add(-2 * time.Minute)

	// Trigger prune and new allocation.
	second := l.Get("repo-b")
	if second == nil {
		t.Fatalf("expected limiter instance")
	}

	if _, ok := l.limiters["repo-a"]; ok {
		t.Fatalf("expected stale limiter to be pruned")
	}
}
