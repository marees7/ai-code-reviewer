package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
)

type OllamaProvider struct {
	url   string
	model string
}

func NewOllama(url, model string) *OllamaProvider {
	return &OllamaProvider{
		url:   url,
		model: model,
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

	b, _ := json.Marshal(reqBody)

	req, _ := http.NewRequestWithContext(
		ctx,
		"POST",
		o.url+"/api/generate",
		bytes.NewBuffer(b),
	)

	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var out ollamaResponse
	json.NewDecoder(resp.Body).Decode(&out)

	return out.Response, nil
}
