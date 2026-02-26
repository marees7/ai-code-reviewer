package budget

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestGuardBlocksWhenPRLimitExceeded(t *testing.T) {
	store := NewMemoryStore()
	g := NewGuard(true, 100, 1.0, store)

	ctx := context.Background()
	now := time.Now().UTC()

	require.NoError(t, g.Record(ctx, "acme/repo", 1, 0.9, now))

	allowed, reason, err := g.Allow(ctx, "acme/repo", 1, 0.2, now)
	require.NoError(t, err)
	require.False(t, allowed)
	require.Contains(t, reason, "PR budget exceeded")
}

func TestGuardBlocksWhenDailyLimitExceeded(t *testing.T) {
	store := NewMemoryStore()
	g := NewGuard(true, 1.0, 10.0, store)

	ctx := context.Background()
	now := time.Now().UTC()

	require.NoError(t, g.Record(ctx, "acme/repo", 1, 0.95, now))

	allowed, reason, err := g.Allow(ctx, "other/repo", 2, 0.1, now)
	require.NoError(t, err)
	require.False(t, allowed)
	require.Contains(t, reason, "Daily budget exceeded")
}
