package ai

import "context"

type ReviewRequest struct {
	File    string
	Content string
}

//go:generate mockery --name Provider --output ../mocks --with-expecter
type Provider interface {
	Review(ctx context.Context, r ReviewRequest) (string, error)
}
