package github

import (
	"context"
	"encoding/json"
	"strings"
	"time"
)

func (h *WebhookHandler) handlePullRequest(payload []byte) {

	var event PullRequestEvent

	if err := json.Unmarshal(payload, &event); err != nil {
		h.logger.Error("failed to parse pr event",
			"error", err,
		)
		return
	}

	// ─────────────────────────────────────
	// 1. FILTERS
	// ─────────────────────────────────────

	// Ignore draft PRs
	if event.PullRequest.Draft {
		h.logger.Info("draft pr ignored",
			"repo", event.Repository.FullName,
			"pr", event.PullRequest.Number,
		)
		return
	}

	// Ignore bots
	if strings.Contains(
		strings.ToLower(event.PullRequest.User.Login),
		"bot",
	) {
		h.logger.Info("bot pr ignored",
			"user", event.PullRequest.User.Login,
		)
		return
	}

	// Only specific actions
	if event.Action != "opened" &&
		event.Action != "synchronize" {
		h.logger.Info("action ignored",
			"action", event.Action,
		)
		return
	}

	// ─────────────────────────────────────
	// 2. ENQUEUE JOB (FAST PATH)
	// ─────────────────────────────────────

	ctx, cancel := context.WithTimeout(
		context.Background(),
		3*time.Second,
	)
	defer cancel()

	err := h.queue.Enqueue(
		ctx,
		event.Repository.FullName,
		event.PullRequest.Number,
	)

	if err != nil {
		h.logger.Error("failed to enqueue job",
			"error", err,
			"repo", event.Repository.FullName,
			"pr", event.PullRequest.Number,
		)
		return
	}

	h.logger.Info("pr job queued",
		"repo", event.Repository.FullName,
		"pr", event.PullRequest.Number,
		"action", event.Action,
	)
}
