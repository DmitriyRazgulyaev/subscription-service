package resilience

import (
	"context"
	"errors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"math"
	"math/rand"
	"time"
)

var ErrAllAttemptsFailed = errors.New("operation failed after all attempts")

type Action int

const (
	Succeed Action = iota
	Retry
	Fail
)

func exponentialBackoff(attemptNum int, min, max time.Duration) time.Duration {
	factor := 2.0
	random := rand.New(rand.NewSource(time.Now().UnixNano()))
	delay := time.Duration(math.Pow(factor, float64(attemptNum)) * float64(min))
	jitter := time.Duration(random.Float64() * float64(min) * float64(attemptNum))

	delay = delay + jitter
	if delay > max {
		delay = max
	}
	return delay
}

func RetryOperation[T any](
	ctx context.Context,
	operation func(ctx context.Context) (T, error),
	maxRetries int,
	baseDelay time.Duration,
	maxDelay time.Duration,
) (T, error) {

	for currentAttempt := 0; currentAttempt < maxRetries; currentAttempt++ {
		resp, err := operation(ctx)
		switch retryPolicy(err) {
		case Succeed:
			return resp, err
		case Retry:
			delay := exponentialBackoff(currentAttempt, baseDelay, maxDelay)
			timeout := time.After(delay)
			if err := sleep(ctx, timeout); err != nil {
				return resp, err
			}
		case Fail:
			return resp, err
		}
	}
	var zero T
	return zero, ErrAllAttemptsFailed
}

func sleep(ctx context.Context, t <-chan time.Time) error {
	select {
	case <-t:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func retryPolicy(err error) Action {
	if err == nil {
		return Succeed
	}
	code := status.Code(err)
	if code == codes.InvalidArgument {
		return Fail
	}
	if code == codes.AlreadyExists {
		return Fail
	}
	if code == codes.NotFound {
		return Fail
	}
	if code == codes.Unavailable {
		return Retry
	}
	if code == codes.DeadlineExceeded || code == codes.Canceled {
		return Retry
	}
	if code == codes.Internal {
		return Retry
	}
	return Retry
}
