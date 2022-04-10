package cadata

import (
	"context"
	"errors"
	"runtime"

	"github.com/brendoncarroll/go-state"
	"golang.org/x/sync/errgroup"
)

// ForEach calls fn once with every ID in s
func ForEach(ctx context.Context, s Lister, fn func(ID) error) error {
	return ForEachSpan(ctx, s, state.ByteSpan{}, fn)
}

// ForEachSpan calls fn with every ID in the range
func ForEachSpan(ctx context.Context, s Lister, span state.ByteSpan, fn func(ID) error) error {
	span2 := state.Span[ID]{Begin: IDFromBytes(span.Begin)}
	if span.End != nil {
		span2.End = IDFromBytes(span.End)
	}
	return state.ForEachSpan[ID](ctx, s, span2, fn)
}

// Copy copies the data referenced by id from src to dst.
func Copy(ctx context.Context, dst Poster, src Getter, id ID) error {
	if adder, ok := dst.(Adder); ok {
		if err := adder.Add(ctx, id); err != ErrNotFound {
			return err
		}
	}
	buf := make([]byte, src.MaxSize())
	n, err := src.Get(ctx, id, buf)
	if err != nil {
		return err
	}
	id2, err := dst.Post(ctx, buf[:n])
	if err != nil {
		return err
	}
	if !id.Equals(id2) {
		return errors.New("stores have different hash functions")
	}
	return nil
}

type CopyAllFrom interface {
	CopyAllFrom(ctx context.Context, src Store) error
}

// CopyAll copies all the data from src to dst
func CopyAll(ctx context.Context, dst, src Store) error {
	if caf, ok := dst.(CopyAllFrom); ok {
		return caf.CopyAllFrom(ctx, src)
	}
	return CopyAllBasic(ctx, dst, src)
}

func CopyAllBasic(ctx context.Context, dst, src Store) error {
	numWorkers := runtime.GOMAXPROCS(0)
	ch := make(chan ID)
	eg, ctx := errgroup.WithContext(ctx)
	eg.Go(func() error {
		return ForEach(ctx, src, func(id ID) error {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case ch <- id:
			}
			return nil
		})
	})
	for i := 0; i < numWorkers; i++ {
		eg.Go(func() error {
			for id := range ch {
				if err := Copy(ctx, dst, src, id); err != nil {
					return err
				}
			}
			return nil
		})
	}
	return eg.Wait()
}

// DeleteAll deletes all the data in s
func DeleteAll(ctx context.Context, s Store) error {
	return ForEach(ctx, s, func(id ID) error {
		return s.Delete(ctx, id)
	})
}

func GetF(ctx context.Context, s Getter, id ID, fn func([]byte) error) error {
	if getF, ok := s.(interface {
		GetF(context.Context, ID, func([]byte) error) error
	}); ok {
		return getF.GetF(ctx, id, fn)
	}
	data, err := GetBytes(ctx, s, id)
	if err != nil {
		return err
	}
	return fn(data)
}

func GetBytes(ctx context.Context, s Getter, id ID) ([]byte, error) {
	buf := make([]byte, s.MaxSize())
	n, err := s.Get(ctx, id, buf)
	return buf[:n], err
}

func Exists(ctx context.Context, x Lister, id ID) (bool, error) {
	return state.Exists[ID](ctx, x, id)
}
