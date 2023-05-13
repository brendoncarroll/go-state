package streams

import "context"

// LoadChan loads a channel from an Iterator.
// If the context is cancelled, LoadChan returns that error.
// If it.Next errors other than EOS, LoadChan returns that error.
func LoadChan[T any](ctx context.Context, it Iterator[T], out chan<- T) error {
	for {
		var dst T
		if err := it.Next(ctx, &dst); err != nil {
			if IsEOS(err) {
				break
			}
			return err
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case out <- dst:
		}
	}
	return nil
}

type Chan[T any] <-chan T

func (c Chan[T]) Next(ctx context.Context, dst *T) error {
	var ok bool
	*dst, ok = <-c
	if !ok {
		return EOS()
	}
	return nil
}
