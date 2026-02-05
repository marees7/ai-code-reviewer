package github

import (
	"sync"
	"time"
)

type tokenCache struct {
	mu    sync.Mutex
	token string
	exp   time.Time
}

func (t *tokenCache) Get() (string, bool) {
	t.mu.Lock()
	defer t.mu.Unlock()

	if time.Now().Before(t.exp) {
		return t.token, true
	}
	return "", false
}

func (t *tokenCache) Set(token string, ttl time.Duration) {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.token = token
	t.exp = time.Now().Add(ttl)
}
