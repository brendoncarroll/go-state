package state

import (
	"context"
	"errors"

	"golang.org/x/sync/errgroup"
)

var (
	ErrNotFound = errors.New("no entry found")
)

// Entry is a key value mapping.
type Entry[K, V any] struct {
	Key   K
	Value V
}

// Poster has the Post method
type Poster[K, V any] interface {
	// Post creates an entry holding v and returns a key for it.
	Post(ctx context.Context, v V) (K, error)
}

// Putter has the Put method
type Putter[K, V any] interface {
	// Put replaces data at key with value, or creates a mapping from k -> v
	// if it does not exist.
	Put(ctx context.Context, k K, v V) error
}

// Deleter has the Delete method
type Deleter[K any] interface {
	// Delete removes the data at k
	Delete(ctx context.Context, k K) error
}

// Updater has the Update method
type Updater[K, V any] interface {
	// Update applies an update to a key serially with other calls to update
	Update(ctx context.Context, k K, fn func(v *V) (*V, error)) error
}

// Lister has the List method.
type Lister[K any] interface {
	// List copies keys from the store into the slice entries
	// and returns the number copied.
	// List signals the end of the list by returning (0, nil)
	// List may fill ks with fewer than len(ks), but will always return > 0, unless it is the end.
	// List will only return keys which are contained by in the span.
	List(ctx context.Context, span Span[K], ks []K) (int, error)
}

// Exister has the Exists method
type Exister[K any] interface {
	// Exists returns true if the store contains an entry for k and false otherwise.
	Exists(ctx context.Context, k K) (bool, error)
}

// Getter has the Get method
type Getter[K, V any] interface {
	Get(ctx context.Context, k K) (V, error)
}

// KVStore is an ordered Key-Value Store
type KVStore[K, V any] interface {
	Getter[K, V]
	Putter[K, V]
	Deleter[K]
	Lister[K]
}

// ReadOnlyKVStore is a KVStore which does not allow mutating data.
type ReadOnlyKVStore[K, V any] interface {
	Getter[K, V]
	Lister[K]
}

// KVStoreTx is a transactional key value store
type KVStoreTx[K, V any] interface {
	// View calls fn with a ReadOnly view of the store
	View(ctx context.Context, fn func(tx ReadOnlyKVStore[K, V]) error) error
	// Modify calls fn with a mutable view of the store.
	// Any operations on the store are either all applied if fn and Modify both return nil,
	// or all reverted if either fn or View returns a non-nil error.
	Modify(ctx context.Context, fn func(tx KVStore[K, V]) error) error
}

// ForEach calls fn with all the keys in x constrained by gteq, and lt if they exist.
// `fn` may be called in another go rountine during the execution of ForEachSpan.
// `fn` will not be called after ForEachSpan returns.
func ForEach[K any](ctx context.Context, x Lister[K], span Span[K], fn func(K) error) error {
	const batchSize = 16
	const chanSize = 32

	eg, ctx := errgroup.WithContext(ctx)
	ch := make(chan K, chanSize)
	eg.Go(func() error {
		defer close(ch)
		buf := make([]K, batchSize)
		for i := 0; ; i++ {
			n, err := x.List(ctx, span, buf)
			if err != nil {
				return err
			}
			items := buf[:n]
			if len(items) == 0 {
				return nil
			}
			for _, k := range items {
				select {
				case <-ctx.Done():
					return ctx.Err()
				case ch <- k:
				}
			}
			span = span.WithLowerExcl(items[len(items)-1])
		}
	})
	eg.Go(func() error {
		for ent := range ch {
			if err := fn(ent); err != nil {
				return err
			}
		}
		return nil
	})
	return eg.Wait()
}

// Exists checks if k exists in x.
// Exists will check if s also implements Exister and use the Exists method.
// If not it will call ExistsUsingList
func Exists[K any](ctx context.Context, s Lister[K], k K) (bool, error) {
	if exister, ok := s.(Exister[K]); ok {
		return exister.Exists(ctx, k)
	}
	return ExistsUsingList(ctx, s, k)
}

// ExistsUsingList implements Exists in terms of List
func ExistsUsingList[K any](ctx context.Context, s Lister[K], k K) (bool, error) {
	span := PointSpan(k)
	ks := [1]K{}
	n, err := s.List(ctx, span, ks[:])
	if err != nil {
		return false, err
	}
	return n > 0, nil
}
