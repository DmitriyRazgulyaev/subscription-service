package resilience

import (
	"context"
	"time"
)

func Timeout[T any](
	ctx context.Context,
	operation func(context.Context) (T, error),
	timeout time.Duration,
) (T, error) {

	ctxWithTimeout, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	result := make(chan struct {
		val T
		err error
	}, 1)

	go func() {
		v, err := operation(ctxWithTimeout)
		result <- struct {
			val T
			err error
		}{v, err}
	}()

	select {
	case r := <-result:
		return r.val, r.err
	case <-ctxWithTimeout.Done():
		var zero T
		return zero, ctxWithTimeout.Err()
	}
}
