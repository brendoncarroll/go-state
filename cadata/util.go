package cadata

import (
	"context"
	"errors"
	"runtime"

	"golang.org/x/sync/errgroup"
)

// ForEach calls fn once with every ID in s
func ForEach(ctx context.Context, s Lister, fn func(ID) error) error {
	return forEach(ctx, s, []byte{}, fn)
}

func forEach(ctx context.Context, s Lister, prefix []byte, fn func(ID) error) error {
	ids := make([]ID, 1<<10)
	n, err := s.List(ctx, prefix, ids)
	switch {
	case err == ErrTooMany:
		for i := 0; i < 256; i++ {
			prefix2 := append(prefix, byte(i))
			if err := forEach(ctx, s, prefix2, fn); err != nil {
				return err
			}
		}
		return nil
	case err != nil:
		return err
	default:
		for _, id := range ids[:n] {
			if err := fn(id); err != nil {
				return err
			}
		}
		return nil
	}
}

// Copy copies the data referenced by id from src to dst.
func Copy(ctx context.Context, dst Poster, src Reader, id ID) error {
	if adder, ok := dst.(Adder); ok {
		if err := adder.Add(ctx, id); err != ErrNotFound {
			return err
		}
	}
	buf := make([]byte, DefaultMaxSize)
	n, err := src.Read(ctx, id, buf)
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
