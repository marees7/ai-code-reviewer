package budget

import (
	"strings"

	"ai-code-reviewer/internal/config"
)

func NewStore(cfg *config.Config) Store {
	if cfg == nil {
		return NewMemoryStore()
	}

	storeType := strings.ToLower(strings.TrimSpace(cfg.BudgetStore))
	if storeType == "redis" {
		addr := strings.TrimSpace(cfg.BudgetRedisAddr)
		if addr == "" {
			addr = cfg.RedisAddr
		}
		return NewRedisStore(addr)
	}

	return NewMemoryStore()
}
