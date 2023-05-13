package streams

import (
	"context"
	"time"
)

type Batcher[T any] struct {
	inner Iterator[T]
	min   int
	dur   time.Duration
}

func NewBatcher[T any](inner Iterator[T], min int, dur time.Duration) *Batcher[T] {
	return &Batcher[T]{
		inner: inner,
		min:   min,
		dur:   dur,
	}
}

func (b *Batcher[T]) Next(ctx context.Context, dst *[]T) error {
	*dst = (*dst)[:0]
	start := time.Now()
	// TODO: need context to cancel long running calls to inner.Next
	for {
		var x T
		if err := b.inner.Next(ctx, &x); err != nil {
			if IsEOS(err) && len(*dst) > 0 {
				return nil
			}
			return err
		}
		*dst = append(*dst, x)
		if len(*dst) >= b.min {
			break
		}
		if now := time.Now(); now.Sub(start) >= b.dur {
			break
		}
	}
	return nil
}
