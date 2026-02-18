package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type OpenAI struct {
	Key   string
	Model string
}

func NewOpenAI(key, model string) *OpenAI {
	return &OpenAI{Key: key, Model: model}
}

func (o *OpenAI) Review(ctx context.Context, r ReviewRequest) (string, error) {

	prompt := BuildPrompt(r)

	body := map[string]any{
		"model": o.Model,
		"messages": []map[string]string{
			{"role": "system", "content": systemPrompt},
			{"role": "user", "content": prompt},
		},
	}

	b, _ := json.Marshal(body)

	req, _ := http.NewRequestWithContext(
		ctx,
		"POST",
		"https://api.openai.com/v1/chat/completions",
		bytes.NewReader(b),
	)

	req.Header.Set("Authorization", "Bearer "+o.Key)
	req.Header.Set("Content-Type", "application/json")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()

	if res.StatusCode >= 300 {
		b, _ := io.ReadAll(res.Body)
		return "", fmt.Errorf("openai status %d: %s", res.StatusCode, string(b))
	}

	var out struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}

	if err := json.NewDecoder(res.Body).Decode(&out); err != nil {
		return "", err
	}

	if len(out.Choices) == 0 {
		return "", fmt.Errorf("no response")
	}

	return out.Choices[0].Message.Content, nil
}
