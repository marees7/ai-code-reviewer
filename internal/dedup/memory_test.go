package dedup

import (
	"context"
	"testing"
	"time"
)

func TestMemory_EvictsWhenMaxEntriesExceeded(t *testing.T) {
	store := NewMemory()
	store.maxEntries = 2
	store.ttl = time.Hour

	ctx := context.Background()
	_ = store.Mark(ctx, "k1")
	_ = store.Mark(ctx, "k2")
	_ = store.Mark(ctx, "k3")

	if store.Seen(ctx, "k1") {
		t.Fatalf("expected oldest key to be evicted")
	}
	if !store.Seen(ctx, "k2") || !store.Seen(ctx, "k3") {
		t.Fatalf("expected newer keys to stay in cache")
	}
}

func TestMemory_ExpiresKeysByTTL(t *testing.T) {
	store := NewMemory()
	store.ttl = 5 * time.Millisecond

	ctx := context.Background()
	_ = store.Mark(ctx, "expiring")
	time.Sleep(10 * time.Millisecond)

	if store.Seen(ctx, "expiring") {
		t.Fatalf("expected key to expire after ttl")
	}
}
