package github

import (
	"context"
	"encoding/json"
)

func (h *WebhookHandler) handlePullRequest(payload []byte) {

	var event PullRequestEvent

	if err := json.Unmarshal(payload, &event); err != nil {
		h.logger.Error("failed to parse pr event", "error", err)
		return
	}

	// Only care about these actions
	if event.Action != "opened" && event.Action != "synchronize" {
		h.logger.Info("pr action ignored", "action", event.Action)
		return
	}

	// 🔥 DAY-3 FEATURE ENTRY POINT
	files, err := h.client.GetPRFiles(
		context.Background(),
		event.Repository.FullName,
		event.PullRequest.Number,
	)

	if err != nil {
		h.logger.Error("failed to fetch pr files",
			"error", err,
		)
		return
	}

	h.logger.Info("pr files fetched",
		"count", len(files),
		"repo", event.Repository.FullName,
		"pr", event.PullRequest.Number,
	)

	// Next days will send these files to:
	// → diff parser
	// → chunker
	// → AI reviewer
}
