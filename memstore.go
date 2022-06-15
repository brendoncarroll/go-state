package state

import (
	"context"
	"sync"

	"golang.org/x/exp/slices"
)

type MemKVStore[K, V any] struct {
	cmp     func(a, b K) int
	mu      sync.RWMutex
	entries []Entry[K, V]
}

func NewMemKVStore[K, V any](cmp func(a, b K) int) *MemKVStore[K, V] {
	return &MemKVStore[K, V]{
		cmp:     cmp,
		entries: make([]Entry[K, V], 0),
	}
}

func (s *MemKVStore[K, V]) Put(ctx context.Context, k K, v V) error {
	return s.Modify(ctx, func(s KVStore[K, V]) error {
		return s.Put(ctx, k, v)
	})
}

func (s *MemKVStore[K, V]) Delete(ctx context.Context, k K) error {
	return s.Modify(ctx, func(s KVStore[K, V]) error {
		return s.Delete(ctx, k)
	})
}

func (s *MemKVStore[K, V]) Get(ctx context.Context, k K) (ret V, err error) {
	err = s.View(ctx, func(s ReadOnlyKVStore[K, V]) error {
		ret, err = s.Get(ctx, k)
		return err
	})
	return ret, err
}

func (s *MemKVStore[K, V]) List(ctx context.Context, span Span[K], buf []K) (n int, err error) {
	err = s.View(ctx, func(s ReadOnlyKVStore[K, V]) error {
		n, err = s.List(ctx, span, buf)
		return err
	})
	return n, err
}

func (s *MemKVStore[K, V]) Len() (count int) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.entries)
}

func (s *MemKVStore[K, V]) View(ctx context.Context, fn func(ReadOnlyKVStore[K, V]) error) error {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return fn(&memTxStore[K, V]{
		read: s.entries,
		cmp:  s.cmp,
	})
}

func (s *MemKVStore[K, V]) Modify(ctx context.Context, fn func(KVStore[K, V]) error) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	tx := &memTxStore[K, V]{
		read:    s.entries,
		puts:    make([]Entry[K, V], 0),
		deletes: make([]Entry[K, struct{}], 0),
		cmp:     s.cmp,
	}
	if err := fn(tx); err != nil {
		return err
	}
	for _, e := range tx.puts {
		s.entries = putEntry(s.entries, e.Key, e.Value, s.cmp)
	}
	for _, e := range tx.deletes {
		s.entries = deleteEntry(s.entries, e.Key, s.cmp)
	}
	return nil
}

type memTxStore[K, V any] struct {
	puts    []Entry[K, V]
	deletes []Entry[K, struct{}]
	read    []Entry[K, V]
	cmp     func(a, b K) int
}

func (s *memTxStore[K, V]) Put(ctx context.Context, k K, v V) error {
	s.puts = putEntry(s.puts, k, v, s.cmp)
	s.deletes = deleteEntry(s.deletes, k, s.cmp)
	return nil
}

func (s *memTxStore[K, V]) Delete(ctx context.Context, k K) error {
	s.deletes = putEntry(s.deletes, k, struct{}{}, s.cmp)
	s.puts = deleteEntry(s.puts, k, s.cmp)
	return nil
}

func (s *memTxStore[K, V]) Get(ctx context.Context, k K) (V, error) {
	var zero V
	if e, exists := getEntry(s.puts, k, s.cmp); exists {
		return e.Value, nil
	}
	if _, exists := getEntry(s.deletes, k, s.cmp); exists {
		return zero, ErrNotFound
	}
	e, exists := getEntry(s.read, k, s.cmp)
	if !exists {
		return zero, ErrNotFound
	}
	return e.Value, nil
}

func (s *memTxStore[K, V]) List(ctx context.Context, span Span[K], buf []K) (n int, _ error) {
	// TODO: implement this for transactions
	if s.puts != nil || s.deletes != nil {
		panic("List not yet implemented for write transactions")
	}
	for _, e := range s.read {
		if n == len(buf) {
			break
		}
		c := span.Compare(e.Key, s.cmp)
		if c > 0 {
			continue
		} else if c < 0 {
			break
		}
		buf[n] = e.Key
		n++
	}
	return n, nil
}

func putEntry[K, V any, S []Entry[K, V]](s S, k K, v V, cmp func(a, b K) int) S {
	x := Entry[K, V]{Key: k, Value: v}
	i, ok := slices.BinarySearchFunc(s, x, func(a, b Entry[K, V]) int {
		return cmp(a.Key, b.Key)
	})
	if ok {
		s[i] = x
	} else {
		s = append(s, x)
		for j := len(s) - 1; j > i; j-- {
			s[j-1], s[j] = s[j], s[j-1]
		}
	}
	return s
}

func getEntry[K, V any, S ~[]Entry[K, V]](s S, k K, cmp func(a, b K) int) (Entry[K, V], bool) {
	x := Entry[K, V]{Key: k}
	i, ok := slices.BinarySearchFunc(s, x, func(a, b Entry[K, V]) int {
		return cmp(a.Key, b.Key)
	})
	if ok {
		return s[i], true
	}
	var zero Entry[K, V]
	return zero, false
}

func deleteEntry[K, V any, S ~[]Entry[K, V]](s S, k K, cmp func(a, b K) int) S {
	x := Entry[K, V]{Key: k}
	i, ok := slices.BinarySearchFunc(s, x, func(a, b Entry[K, V]) int {
		return cmp(a.Key, b.Key)
	})
	if ok {
		s = slices.Delete(s, i, i+1)
	}
	return s
}
