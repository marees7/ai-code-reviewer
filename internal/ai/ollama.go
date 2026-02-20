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
) (string, error) {

	reqBody := ollamaRequest{
		Model:  o.model,
		Prompt: buildPrompt(r),
		Stream: false,
	}

	b, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("marshal ollama request: %w", err)
	}

	req, err := http.NewRequestWithContext(
		ctx,
		"POST",
		o.url+"/api/generate",
		bytes.NewBuffer(b),
	)
	if err != nil {
		return "", fmt.Errorf("build ollama request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := o.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		msg, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return "", fmt.Errorf("ollama status %d: %s", resp.StatusCode, string(msg))
	}

	var out ollamaResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return "", fmt.Errorf("decode ollama response: %w", err)
	}

	return out.Response, nil
}
