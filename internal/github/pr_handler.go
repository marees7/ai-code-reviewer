package github

import (
	"encoding/json"
)

func (h *WebhookHandler) handlePullRequest(payload []byte) {

	var event PullRequestEvent

	if err := json.Unmarshal(payload, &event); err != nil {
		h.logger.Error("failed to parse pr event", "error", err)
		return
	}

	// We only care about opened / synchronize
	if event.Action != "opened" && event.Action != "synchronize" {
		h.logger.Info("pr action ignored",
			"action", event.Action,
		)
		return
	}

	h.logger.Info("pull request received",
		"repo", event.Repository.FullName,
		"number", event.PullRequest.Number,
		"title", event.PullRequest.Title,
	)
}
