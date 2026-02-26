package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type OllamaProvider struct {
	url    string
	model  string
	client *http.Client
}

func NewOllama(url, model string) *OllamaProvider {
	return &OllamaProvider{
		url:   url,
		model: model,
		client: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

type ollamaRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
	Stream bool   `json:"stream"`
}

type ollamaResponse struct {
	Response string `json:"response"`
}

func (o *OllamaProvider) Review(
	ctx context.Context,
	r ReviewRequest,
) (ReviewResponse, error) {

	reqBody := ollamaRequest{
		Model:  o.model,
		Prompt: buildPrompt(r),
		Stream: false,
	}

	b, err := json.Marshal(reqBody)
	if err != nil {
		return ReviewResponse{}, fmt.Errorf("marshal ollama request: %w", err)
	}

	req, err := http.NewRequestWithContext(
		ctx,
		"POST",
		o.url+"/api/generate",
		bytes.NewBuffer(b),
	)
	if err != nil {
		return ReviewResponse{}, fmt.Errorf("build ollama request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := o.client.Do(req)
	if err != nil {
		return ReviewResponse{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		msg, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return ReviewResponse{}, fmt.Errorf("ollama status %d: %s", resp.StatusCode, string(msg))
	}

	var out ollamaResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return ReviewResponse{}, fmt.Errorf("decode ollama response: %w", err)
	}

	usage := estimateUsage(reqBody.Prompt, out.Response)

	return ReviewResponse{
		Content:  out.Response,
		Provider: "ollama",
		Model:    o.model,
		Usage:    usage,
	}, nil
}

func estimateUsage(prompt, completion string) Usage {
	promptTokens := estimateTokens(prompt)
	completionTokens := estimateTokens(completion)
	return Usage{
		PromptTokens:     promptTokens,
		CompletionTokens: completionTokens,
		TotalTokens:      promptTokens + completionTokens,
	}
}

func estimateTokens(s string) int {
	// Simple fallback estimate: ~4 chars/token for English-like text.
	if len(s) == 0 {
		return 0
	}
	n := len(s) / 4
	if n == 0 {
		return 1
	}
	return n
}
