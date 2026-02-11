package retry

import (
	"context"
	"time"
)

type Fn func() error

func Do(ctx context.Context, attempts int, wait time.Duration, fn Fn) error {

	var err error

	for i := 0; i < attempts; i++ {

		if ctx.Err() != nil {
			return ctx.Err()
		}

		err = fn()
		if err == nil {
			return nil
		}

		time.Sleep(wait)
		wait = wait * 2
	}

	return err
}
