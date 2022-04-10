package state

import (
	"context"
	"errors"
	"sort"
	"sync"

	"golang.org/x/exp/maps"
	"golang.org/x/sync/errgroup"
)

var (
	EndOfList   = errors.New("end of list")
	ErrNotFound = errors.New("no entry found")
)

// Entry is a key value mapping.
type Entry[K comparable, V any] struct {
	Key   K
	Value V
}

// Poster has the Post method
type Poster[K comparable, V any] interface {
	// Post creates an entry holding v and returns a key for it.
	Post(ctx context.Context, v V) (K, error)
}

// Putter has the Put method
type Putter[K comparable, V any] interface {
	// Put replaces data at key with value, or creates a mapping from k -> v
	// if it does not exist.
	Put(ctx context.Context, k K, v V) error
}

// Deleter has the Delete method
type Deleter[K comparable] interface {
	// Delete removes the data at k
	Delete(ctx context.Context, k K) error
}

// Updater has the Update method
type Updater[K comparable, V any] interface {
	// Update applies an update to a key serially with other calls to update
	Update(ctx context.Context, k K, fn func(v *V) (*V, error)) error
}

// Lister has the List method.
type Lister[K comparable] interface {
	// List copies keys from the store into the slice entries
	// and returns the number copied.
	List(ctx context.Context, first K, ks []K) (int, error)
}

// Getter has the Get method
type Getter[K comparable, V any] interface {
	Get(ctx context.Context, k K) (V, error)
}

// KVStore is an ordered Key-Value Store
type KVStore[K comparable, V any] interface {
	Getter[K, V]
	Putter[K, V]
	Deleter[K]
	Lister[K]
}

// ReadOnlyKVStore is a KVStore which does not allow mutating data.
type ReadOnlyKVStore[K comparable, V any] interface {
	Getter[K, V]
	Lister[K]
}

// KVStoreTx is a transactional key value store
type KVStoreTx[K comparable, V any] interface {
	// View calls fn with a ReadOnly view of the store
	// Any operations on the store are either all applied if fn and View both return nil,
	// or all reverted if either fn or View returns a non-nil error.
	View(ctx context.Context, fn func(tx ReadOnlyKVStore[K, V]) error) error
	// Modify
	Modify(ctx context.Context, fn func(tx KVStore[K, V]) error) error
}

// ForEach calls fn for each k in the listable KVStore x.
// ForEach calls ForEachSpan internally with an all-including Span.
func ForEach[K comparable](ctx context.Context, x Lister[K], fn func(K) error) error {
	return ForEachSpan(ctx, x, Span[K]{}, fn)
}

// ForEachSpan calls fn with all the keys in x contained in span.
// `fn` may be called in another go rountine during the execution of ForEachSpan.
// `fn` will not be called after ForEachSpan returns.
func ForEachSpan[K comparable](ctx context.Context, x Lister[K], span Span[K], fn func(K) error) error {
	const batchSize = 16
	const chanSize = 32

	eg, ctx := errgroup.WithContext(ctx)
	ch := make(chan K, chanSize)
	eg.Go(func() error {
		defer close(ch)
		buf := make([]K, batchSize)
		first := span.Begin
		for i := 0; ; i++ {
			n, err := x.List(ctx, first, buf)
			if err != nil && !errors.Is(err, EndOfList) {
				return err
			}
			for _, k := range buf[:n] {
				if i > 0 && k == first {
					continue
				}
				select {
				case <-ctx.Done():
					return ctx.Err()
				case ch <- k:
				}
			}
			if errors.Is(err, EndOfList) {
				return nil
			}
			if n == 0 {
				return errors.New("List returned 0 without io.EOF")
			}
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

type MemKVStore[K comparable, V any] struct {
	lt func(a, b K) bool
	mu sync.RWMutex
	m  map[K]V
}

func NewMemKVStore[K comparable, V any](lessThan func(a, b K) bool) KVStore[K, V] {
	return &MemKVStore[K, V]{
		lt: lessThan,
		m:  make(map[K]V),
	}
}

func (s *MemKVStore[K, V]) Put(ctx context.Context, k K, v V) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.m[k] = v
	return nil
}

func (s *MemKVStore[K, V]) Delete(ctx context.Context, k K) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.m, k)
	return nil
}

func (s *MemKVStore[K, V]) Get(ctx context.Context, k K) (V, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	v, exists := s.m[k]
	if !exists {
		return v, ErrNotFound
	}
	return v, nil
}

func (s *MemKVStore[K, V]) List(ctx context.Context, first K, buf []K) (n int, _ error) {
	s.mu.RLock()
	keys := maps.Keys(s.m)
	s.mu.RUnlock()
	sort.Slice(keys, func(i, j int) bool {
		return s.lt(keys[i], keys[j])
	})
	for i := range keys {
		if n == len(buf) {
			break
		}
		if s.lt(keys[i], first) {
			continue
		}
		buf[n] = keys[i]
		n++
	}
	if n < len(buf) {
		return n, EndOfList
	}
	return n, nil
}
