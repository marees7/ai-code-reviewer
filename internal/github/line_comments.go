package github

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type LineComment struct {
	Body string `json:"body"`
	Path string `json:"path"`
	Line int    `json:"line"`
	Side string `json:"side"` // RIGHT = new code
}

func (c *CommentService) CreateLineComment(
	ctx context.Context,
	repo string,
	pr int,
	l LineComment,
) error {

	url := fmt.Sprintf(
		"https://api.github.com/repos/%s/pulls/%d/comments",
		repo, pr,
	)

	b, err := json.Marshal(l)
	if err != nil {
		return fmt.Errorf("marshal line comment: %w", err)
	}

	req, err := http.NewRequestWithContext(
		ctx, "POST", url, bytes.NewReader(b),
	)
	if err != nil {
		return fmt.Errorf("build line comment request: %w", err)
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
		return fmt.Errorf("github %d: %s", res.StatusCode, string(msg))
	}

	return nil
}
