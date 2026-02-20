package github

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type CommentService struct {
	token string
	http  *http.Client
}

func NewCommentService(token string) *CommentService {
	return &CommentService{
		token: token,
		http: &http.Client{
			Timeout: 15 * time.Second,
		},
	}
}

func (c *CommentService) CreateComment(
	ctx context.Context,
	repo string,
	pr int,
	body string,
) error {

	url := fmt.Sprintf(
		"https://api.github.com/repos/%s/issues/%d/comments",
		repo,
		pr,
	)

	payload := map[string]string{
		"body": body,
	}

	b, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal comment payload: %w", err)
	}

	req, err := http.NewRequestWithContext(
		ctx,
		"POST",
		url,
		bytes.NewReader(b),
	)
	if err != nil {
		return fmt.Errorf("build comment request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("User-Agent", "ai-code-reviewer")

	res, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode >= 300 {
		msg, _ := io.ReadAll(io.LimitReader(res.Body, 4096))
		return fmt.Errorf("github status %d: %s", res.StatusCode, string(msg))
	}

	return nil
}
