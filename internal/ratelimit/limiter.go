package ratelimit

import (
	"sync"
	"time"

	"golang.org/x/time/rate"
)

type limiterEntry struct {
	limiter  *rate.Limiter
	lastUsed time.Time
}

type Limiter struct {
	mu         sync.Mutex
	limiters   map[string]*limiterEntry
	rps        rate.Limit
	burst      int
	ttl        time.Duration
	lastPruned time.Time
}

func New(rps int, burst int) *Limiter {
	return &Limiter{
		limiters: make(map[string]*limiterEntry),
		rps:      rate.Limit(rps),
		burst:    burst,
		ttl:      30 * time.Minute,
	}
}

func (l *Limiter) Get(repo string) *rate.Limiter {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := time.Now()
	l.pruneLocked(now)

	if entry, ok := l.limiters[repo]; ok {
		entry.lastUsed = now
		return entry.limiter
	}

	limiter := rate.NewLimiter(l.rps, l.burst)
	l.limiters[repo] = &limiterEntry{
		limiter:  limiter,
		lastUsed: now,
	}
	return limiter
}

func (l *Limiter) pruneLocked(now time.Time) {
	if !l.lastPruned.IsZero() && now.Sub(l.lastPruned) < time.Minute {
		return
	}

	for repo, entry := range l.limiters {
		if now.Sub(entry.lastUsed) > l.ttl {
			delete(l.limiters, repo)
		}
	}
	l.lastPruned = now
}
