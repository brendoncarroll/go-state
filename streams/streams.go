package streams

import (
	"context"
	"errors"
)

// EndOfStream is returned by Next and Seek to indicate that the stream has no more elements.
type EndOfStream struct{}

func (EndOfStream) Error() string {
	return "end of stream"
}

// EOS returns a new EndOfStream error
func EOS() error {
	return EndOfStream{}
}

// IsEOS returns true if the error is an EndOfStream error
func IsEOS(err error) bool {
	return errors.Is(err, EndOfStream{})
}

type Iterator[T any] interface {
	// Next advances the iterator and reads the next element into dst
	Next(ctx context.Context, dst *T) error
}

// Peekable is an Iterator which also has the Peek method
type Peekable[T any] interface {
	Iterator[T]

	// Peek shows the next element of the Iterator without changing the state of the Iterator
	Peek(ctx context.Context, dst *T) error
}

// Seeker contains the Seek method
type Seeker[T any] interface {
	// Seek ensures that all future elements of the iterator will be >= gteq
	Seek(ctx context.Context, gteq T) error
}

// ForEach calls fn for each element of it.
// fn must not retain dst, between calls.
func ForEach[T any](ctx context.Context, it Iterator[T], fn func(T) error) error {
	var dst T
	for {
		if err := it.Next(ctx, &dst); err != nil {
			if IsEOS(err) {
				break
			}
			return err
		}
		if err := fn(dst); err != nil {
			return err
		}
	}
	var zero T
	dst = zero
	return nil
}

// Read copies elements from the iterator into buf.
// Read returns EOS when the iterator is empty.
func Read[T any](ctx context.Context, it Iterator[T], buf []T) (int, error) {
	for i := range buf {
		if err := it.Next(ctx, &buf[i]); err != nil {
			return i, err
		}
	}
	return len(buf), nil
}

// Collect is used to collect all of the items from an Iterator.
// If more than max elements are emitted, then Collect will return an error.
func Collect[T any](ctx context.Context, it Iterator[T], max int) (ret []T, _ error) {
	for {
		var dst T
		if err := it.Next(ctx, &dst); err != nil {
			if IsEOS(err) {
				break
			}
			return ret, err
		}
		if len(ret) < max {
			ret = append(ret, dst)
		} else {
			return ret, errors.New("streams: too many elements to collect")
		}
	}
	return ret, nil
}
