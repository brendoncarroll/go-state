package state

import (
	"context"
	"sort"
	"sync"

	"golang.org/x/exp/maps"
)

type MemKVStore[K comparable, V any] struct {
	lt func(a, b K) bool
	mu sync.RWMutex
	m  map[K]V
}

func NewMemKVStore[K comparable, V any](lessThan func(a, b K) bool) *MemKVStore[K, V] {
	return &MemKVStore[K, V]{
		lt: lessThan,
		m:  make(map[K]V),
	}
}

func (s *MemKVStore[K, V]) Put(ctx context.Context, k K, v V) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return memTxStore[K, V]{read: s.m, puts: s.m}.Put(ctx, k, v)
}

func (s *MemKVStore[K, V]) Delete(ctx context.Context, k K) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return memTxStore[K, V]{read: s.m, puts: s.m}.Delete(ctx, k)
}

func (s *MemKVStore[K, V]) Get(ctx context.Context, k K) (V, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return memTxStore[K, V]{read: s.m}.Get(ctx, k)
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

func (s *MemKVStore[K, V]) Len() (count int) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.m)
}

func (s *MemKVStore[K, V]) View(ctx context.Context, fn func(ReadOnlyKVStore[K, V]) error) error {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return fn(memTxStore[K, V]{read: s.m})
}

func (s *MemKVStore[K, V]) Modify(ctx context.Context, fn func(KVStore[K, V]) error) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	tx := memTxStore[K, V]{
		read:    s.m,
		puts:    make(map[K]V),
		deletes: make(map[K]struct{}),
	}
	if err := fn(tx); err != nil {
		return err
	}
	for k, v := range tx.puts {
		s.m[k] = v
	}
	for k := range tx.deletes {
		delete(s.m, k)
	}
	return nil
}

type memTxStore[K comparable, V any] struct {
	puts    map[K]V
	deletes map[K]struct{}
	read    map[K]V
	lt      func(a, b K) bool
}

func (s memTxStore[K, V]) Put(ctx context.Context, k K, v V) error {
	s.puts[k] = v
	delete(s.deletes, k)
	return nil
}

func (s memTxStore[K, V]) Delete(ctx context.Context, k K) error {
	delete(s.puts, k)
	if s.deletes != nil {
		s.deletes[k] = struct{}{}
	}
	return nil
}

func (s memTxStore[K, V]) Get(ctx context.Context, k K) (V, error) {
	if v, exists := s.puts[k]; exists {
		return v, nil
	}
	if _, exists := s.deletes[k]; exists {
		var ret V
		return ret, ErrNotFound
	}
	v, exists := s.read[k]
	if !exists {
		return v, ErrNotFound
	}
	return v, nil
}

func (s memTxStore[K, V]) List(ctx context.Context, first K, buf []K) (n int, _ error) {
	// TODO: implement this for transactions
	if s.puts != nil || s.deletes != nil {
		panic("List not yet implemented for write transactions")
	}
	keys := maps.Keys(s.read)
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
