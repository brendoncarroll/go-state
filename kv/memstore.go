package kv

import (
	"context"
	"sync"

	"github.com/google/btree"
	"go.brendoncarroll.net/state"
	"golang.org/x/exp/slices"
)

type MemStore[K, V any] struct {
	cmp  func(a, b K) int
	mu   sync.RWMutex
	tree btree.BTreeG[Entry[K, V]]
}

func NewMemStore[K, V any](cmp func(a, b K) int) *MemStore[K, V] {
	return &MemStore[K, V]{
		cmp: cmp,
		tree: *btree.NewG[Entry[K, V]](2, func(a, b Entry[K, V]) bool {
			return cmp(a.Key, b.Key) < 0
		}),
	}
}

func (s *MemStore[K, V]) Put(ctx context.Context, k K, v V) error {
	return s.Modify(ctx, func(s Store[K, V]) error {
		return s.Put(ctx, k, v)
	})
}

func (s *MemStore[K, V]) Delete(ctx context.Context, k K) error {
	return s.Modify(ctx, func(s Store[K, V]) error {
		return s.Delete(ctx, k)
	})
}

func (s *MemStore[K, V]) Exists(ctx context.Context, k K) (bool, error) {
	return ExistsUsingList[K](ctx, s, k)
}

func (s *MemStore[K, V]) Get(ctx context.Context, k K, dst *V) error {
	return s.View(ctx, func(s ReadOnlyStore[K, V]) error {
		return s.Get(ctx, k, dst)
	})
}

func (s *MemStore[K, V]) List(ctx context.Context, span state.Span[K], buf []K) (n int, err error) {
	err = s.View(ctx, func(s ReadOnlyStore[K, V]) error {
		n, err = s.List(ctx, span, buf)
		return err
	})
	return n, err
}

func (s *MemStore[K, V]) Len() (count int) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.tree.Len()
}

func (s *MemStore[K, V]) View(ctx context.Context, fn func(ReadOnlyStore[K, V]) error) error {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return fn(&memTxStore[K, V]{
		read: &s.tree,
		cmp:  s.cmp,
	})
}

func (s *MemStore[K, V]) Modify(ctx context.Context, fn func(Store[K, V]) error) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	tx := &memTxStore[K, V]{
		read:    &s.tree,
		puts:    make([]Entry[K, V], 0),
		deletes: make([]Entry[K, struct{}], 0),
		cmp:     s.cmp,
	}
	if err := fn(tx); err != nil {
		return err
	}
	for _, e := range tx.puts {
		s.tree.ReplaceOrInsert(e)
	}
	for _, e := range tx.deletes {
		s.tree.Delete(Entry[K, V]{Key: e.Key})
	}
	return nil
}

type memTxStore[K, V any] struct {
	puts    []Entry[K, V]
	deletes []Entry[K, struct{}]
	read    *btree.BTreeG[Entry[K, V]]
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

func (s *memTxStore[K, V]) Get(ctx context.Context, k K, dst *V) error {
	if e, exists := getEntry(s.puts, k, s.cmp); exists {
		*dst = e.Value
		return nil
	}
	if _, exists := getEntry(s.deletes, k, s.cmp); exists {
		return state.ErrNotFound[K]{Key: k}
	}
	e, exists := s.read.Get(Entry[K, V]{Key: k})
	if !exists {
		return state.ErrNotFound[K]{Key: k}
	}
	*dst = e.Value
	return nil
}

func (s *memTxStore[K, V]) Exists(ctx context.Context, k K) (bool, error) {
	return ExistsUsingList[K](ctx, s, k)
}

func (s *memTxStore[K, V]) List(ctx context.Context, span state.Span[K], buf []K) (n int, _ error) {
	// TODO: implement this for transactions
	if s.puts != nil || s.deletes != nil {
		panic("List not yet implemented for write transactions")
	}
	s.read.Ascend(func(e Entry[K, V]) bool {
		if n == len(buf) {
			return false
		}
		c := span.Compare(e.Key, s.cmp)
		if c > 0 {
			return true
		} else if c < 0 {
			return false
		}
		buf[n] = e.Key
		n++
		return true
	})
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
