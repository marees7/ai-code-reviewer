package github

import (
	"context"
	"encoding/json"
)

func (h *WebhookHandler) handlePullRequest(payload []byte) {

	var event PullRequestEvent
	_ = json.Unmarshal(payload, &event)

	if event.Action != "opened" && event.Action != "synchronize" {
		return
	}

	client := NewClient(h.cfg, h.logger)

	diff, err := client.GetPRDiff(
		context.Background(),
		event.Repository.FullName,
		event.PullRequest.Number,
	)

	if err != nil {
		h.logger.Error("failed diff", "error", err)
		return
	}

	h.logger.Info("diff fetched",
		"size", len(diff),
	)
}
