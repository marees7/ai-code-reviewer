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
	queue  JobQueue
}

const (
	maxWebhookBodyBytes = 1 << 20 // 1 MiB
	headerGithubEvent   = "X-GitHub-Event"
	headerSignature256  = "X-Hub-Signature-256"
	eventPullRequest    = "pull_request"
)

func NewWebhookHandler(
	cfg *config.Config,
	logger *observability.Logger,
	queue JobQueue,
) *WebhookHandler {
	return &WebhookHandler{
		cfg:    cfg,
		logger: logger,
		queue:  queue,
	}
}

func (h *WebhookHandler) Handle(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, maxWebhookBodyBytes)

	payload, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}

	if !h.verifySignature(r.Header.Get(headerSignature256), payload) {
		h.logger.Error("invalid github signature")
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	event := r.Header.Get(headerGithubEvent)
	h.logger.Info("github event received", "event", event)

	switch event {
	case eventPullRequest:
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
