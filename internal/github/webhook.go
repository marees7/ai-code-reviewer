package github

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"net/http"

	"ai-code-reviewer/internal/config"
	"ai-code-reviewer/internal/observability"
)

type WebhookHandler struct {
	cfg    *config.Config
	logger *observability.Logger
}

func NewWebhookHandler(cfg *config.Config, logger *observability.Logger) *WebhookHandler {
	return &WebhookHandler{
		cfg:    cfg,
		logger: logger,
	}
}

func (h *WebhookHandler) Handle(w http.ResponseWriter, r *http.Request) {
	payload, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}

	// Verify signature
	if !h.verifySignature(r.Header.Get("X-Hub-Signature-256"), payload) {
		h.logger.Error("invalid github signature")
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	event := r.Header.Get("X-GitHub-Event")

	h.logger.Info("github event received",
		"event", event,
	)

	switch event {
	case "pull_request":
		h.handlePullRequest(payload)
	default:
		h.logger.Info("event ignored", "event", event)
	}

	w.WriteHeader(http.StatusOK)
}

func (h *WebhookHandler) verifySignature(signature string, body []byte) bool {
	if h.cfg.GithubSecret == "" {
		h.logger.Error("github secret not configured")
		return false
	}

	mac := hmac.New(sha256.New, []byte(h.cfg.GithubSecret))
	mac.Write(body)
	expected := "sha256=" + hex.EncodeToString(mac.Sum(nil))

	return hmac.Equal([]byte(expected), []byte(signature))
}
