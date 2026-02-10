package dedup

import "context"

type Store interface {
	Seen(ctx context.Context, key string) bool
	Mark(ctx context.Context, key string) error
}
