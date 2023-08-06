package cadata

import (
	"context"
	"errors"
	"runtime"

	"github.com/brendoncarroll/go-state/kv"
	"golang.org/x/sync/errgroup"
)

// ForEachSpan calls fn with every ID in the range
func ForEach(ctx context.Context, s Lister, span Span, fn func(ID) error) error {
	return kv.ForEach[ID](ctx, s, span, fn)
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
	CopyAllFrom(ctx context.Context, src Getter) error
}

// CopyAll copies all the data from src to dst
func CopyAll(ctx context.Context, dst Poster, src GetLister) error {
	if caf, ok := dst.(CopyAllFrom); ok {
		return caf.CopyAllFrom(ctx, src)
	}
	return CopyAllBasic(ctx, dst, src)
}

func CopyAllBasic(ctx context.Context, dst Poster, src GetLister) error {
	numWorkers := runtime.GOMAXPROCS(0)
	ch := make(chan ID)
	eg, ctx := errgroup.WithContext(ctx)
	eg.Go(func() error {
		return ForEach(ctx, src, Span{}, func(id ID) error {
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
	return ForEach(ctx, s, Span{}, func(id ID) error {
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

// Exists returns (true, nil) if x has ID
// If x implements Exister, then x.Exists is called
// If x implements Lister, then x.List is called
func Exists(ctx context.Context, x Lister, id ID) (bool, error) {
	if exister, ok := x.(Exister); ok {
		return exister.Exists(ctx, id)
	}
	return kv.ExistsUsingList[ID](ctx, x, id)
}
