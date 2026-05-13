package httpx

import (
	"context"
	"time"
)

type Limiter interface {
	Wait(ctx context.Context) error
}

type tokenBucket struct {
	tokens chan struct{}
}

func NewTokenBucket(rps int) Limiter {
	if rps <= 0 {
		return nil
	}
	b := &tokenBucket{tokens: make(chan struct{}, rps)}
	go func() {
		t := time.NewTicker(time.Second / time.Duration(rps))
		defer t.Stop()
		for range t.C {
			select {
			case b.tokens <- struct{}{}:
			default:
			}
		}
	}()
	return b
}

func (b *tokenBucket) Wait(ctx context.Context) error {
	select {
	case <-b.tokens:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}
