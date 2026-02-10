package github

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
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

	b, _ := json.Marshal(l)

	req, _ := http.NewRequestWithContext(
		ctx, "POST", url, bytes.NewReader(b),
	)

	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Content-Type", "application/json")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode >= 300 {
		return fmt.Errorf("github %d", res.StatusCode)
	}

	return nil
}
