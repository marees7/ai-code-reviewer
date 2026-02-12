package ai

import "context"

type FallbackProvider struct {
	primary   Provider
	secondary Provider
}

func NewFallback(p1, p2 Provider) *FallbackProvider {
	return &FallbackProvider{
		primary:   p1,
		secondary: p2,
	}
}

func (f *FallbackProvider) Review(
	ctx context.Context,
	r ReviewRequest,
) (string, error) {

	resp, err := f.primary.Review(ctx, r)
	if err == nil {
		return resp, nil
	}

	return f.secondary.Review(ctx, r)
}
