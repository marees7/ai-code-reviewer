package ai

import "context"

type ReviewRequest struct {
	File    string
	Content string
}

type Usage struct {
	PromptTokens     int
	CompletionTokens int
	TotalTokens      int
}

type ReviewResponse struct {
	Content  string
	Provider string
	Model    string
	Usage    Usage
}

//go:generate mockery --name Provider --output ../mocks --with-expecter
type Provider interface {
	Review(ctx context.Context, r ReviewRequest) (ReviewResponse, error)
}
