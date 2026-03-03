package github

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

const (
	prActionOpened      = "opened"
	prActionSynchronize = "synchronize"
	botLoginToken       = "bot"
	enqueueTimeout      = 3 * time.Second
	tenantFallback      = "default"
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
		botLoginToken,
	) {
		h.logger.Info("bot pr ignored",
			"user", event.PullRequest.User.Login,
		)
		return
	}

	// Only specific actions
	if event.Action != prActionOpened &&
		event.Action != prActionSynchronize {
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
		enqueueTimeout,
	)
	defer cancel()

	err := h.queue.Enqueue(
		ctx,
		resolveTenant(event),
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

func resolveTenant(event PullRequestEvent) string {
	if event.Installation.ID > 0 {
		return fmt.Sprintf("gh-installation:%d", event.Installation.ID)
	}
	repo := strings.TrimSpace(event.Repository.FullName)
	parts := strings.SplitN(repo, "/", 2)
	if len(parts) == 2 && strings.TrimSpace(parts[0]) != "" {
		return strings.TrimSpace(parts[0])
	}
	return tenantFallback
}
