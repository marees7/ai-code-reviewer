package github

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

type CommentService struct {
	token string
}

func NewCommentService(token string) *CommentService {
	return &CommentService{token: token}
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

	b, _ := json.Marshal(payload)

	req, _ := http.NewRequestWithContext(
		ctx,
		"POST",
		url,
		bytes.NewReader(b),
	)

	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/vnd.github+json")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode >= 300 {
		return fmt.Errorf("github status %d", res.StatusCode)
	}

	return nil
}
