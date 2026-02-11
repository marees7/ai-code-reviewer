package worker

import "ai-code-reviewer/internal/config"

func NewQueue(cfg *config.Config) Queue {

	if cfg.QueueType == "redis" {
		return NewRedisQueue(
			cfg.RedisAddr,
			"ai_reviewer_jobs",
		)
	}

	return NewMemoryQueue(100)
}
