package budget

import (
	"context"
	"fmt"
	"sync"
	"time"
)

type Store interface {
	AddSpend(ctx context.Context, repo string, pr int, usd float64, at time.Time) error
	GetPRSpend(ctx context.Context, repo string, pr int) (float64, error)
	GetDailySpend(ctx context.Context, day time.Time) (float64, error)
}

type Guard struct {
	enabled    bool
	dailyLimit float64
	prLimit    float64
	store      Store
}

func NewGuard(enabled bool, dailyLimit, prLimit float64, store Store) *Guard {
	return &Guard{
		enabled:    enabled,
		dailyLimit: dailyLimit,
		prLimit:    prLimit,
		store:      store,
	}
}

func (g *Guard) Enabled() bool {
	return g != nil && g.enabled
}

func (g *Guard) Allow(ctx context.Context, repo string, pr int, projectedCostUSD float64, now time.Time) (bool, string, error) {
	if g == nil || !g.enabled || g.store == nil {
		return true, "", nil
	}

	prSpend, err := g.store.GetPRSpend(ctx, repo, pr)
	if err != nil {
		return false, "", err
	}
	if g.prLimit > 0 && prSpend+projectedCostUSD > g.prLimit {
		return false, fmt.Sprintf("PR budget exceeded (limit=%.4f USD)", g.prLimit), nil
	}

	daySpend, err := g.store.GetDailySpend(ctx, now)
	if err != nil {
		return false, "", err
	}
	if g.dailyLimit > 0 && daySpend+projectedCostUSD > g.dailyLimit {
		return false, fmt.Sprintf("Daily budget exceeded (limit=%.4f USD)", g.dailyLimit), nil
	}

	return true, "", nil
}

func (g *Guard) Record(ctx context.Context, repo string, pr int, usd float64, now time.Time) error {
	if g == nil || !g.enabled || g.store == nil || usd <= 0 {
		return nil
	}
	return g.store.AddSpend(ctx, repo, pr, usd, now)
}

type MemoryStore struct {
	mu    sync.Mutex
	byPR  map[string]float64
	byDay map[string]float64
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		byPR:  make(map[string]float64),
		byDay: make(map[string]float64),
	}
}

func (m *MemoryStore) AddSpend(_ context.Context, repo string, pr int, usd float64, at time.Time) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.byPR[prKey(repo, pr)] += usd
	m.byDay[dayKey(at)] += usd
	return nil
}

func (m *MemoryStore) GetPRSpend(_ context.Context, repo string, pr int) (float64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.byPR[prKey(repo, pr)], nil
}

func (m *MemoryStore) GetDailySpend(_ context.Context, day time.Time) (float64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.byDay[dayKey(day)], nil
}

func prKey(repo string, pr int) string {
	return fmt.Sprintf("%s#%d", repo, pr)
}

func dayKey(t time.Time) string {
	return t.UTC().Format("2006-01-02")
}
