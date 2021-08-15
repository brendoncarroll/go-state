package cadata

import (
	"bytes"
	"context"
	"io"
	"sync"
)

var _ Store = &MemStore{}

type MemStore struct {
	hash    HashFunc
	maxSize int

	m sync.Map
}

func NewMem(maxSize int) *MemStore {
	return &MemStore{
		maxSize: maxSize,
		hash:    DefaultHash,
	}
}

func (s *MemStore) Post(ctx context.Context, data []byte) (ID, error) {
	data = append([]byte{}, data...)
	id := s.hash(data)
	s.m.Store(id, data)
	return id, nil
}

func (s *MemStore) Get(ctx context.Context, id ID, buf []byte) (int, error) {
	v, exists := s.m.Load(id)
	if !exists {
		return 0, ErrNotFound
	}
	data := v.([]byte)
	if len(buf) < len(data) {
		return 0, io.ErrShortBuffer
	}
	return copy(buf, data), nil
}

func (s *MemStore) List(ctx context.Context, first []byte, ids []ID) (n int, err error) {
	if len(ids) == 0 {
		return 0, nil
	}
	s.m.Range(func(k, v interface{}) bool {
		id := k.(ID)
		if bytes.Compare(id[:], first) < 0 {
			return true
		}
		ids[n] = id
		n++
		return len(ids) < n
	})
	if n == 0 {
		err = ErrEndOfList
	}
	return n, err
}

func (s *MemStore) Delete(ctx context.Context, id ID) error {
	s.m.Delete(id)
	return nil
}

func (s *MemStore) Exists(ctx context.Context, id ID) (bool, error) {
	_, ok := s.m.Load(id)
	return ok, nil
}

func (s *MemStore) Len() (count int) {
	s.m.Range(func(k, v interface{}) bool {
		count++
		return true
	})
	return count
}

func (s *MemStore) Hash(x []byte) ID {
	return s.hash(x)
}

func (s *MemStore) MaxSize() int {
	return s.maxSize
}

var _ Store = Void{}

type Void struct{}

func (s Void) Post(ctx context.Context, data []byte) (ID, error) {
	return DefaultHash(data), nil
}

func (s Void) Get(ctx context.Context, id ID, buf []byte) (int, error) {
	return 0, ErrNotFound
}

func (s Void) Exists(ctx context.Context, id ID) (bool, error) {
	return false, nil
}

func (s Void) List(ctx context.Context, prefix []byte, ids []ID) (int, error) {
	return 0, nil
}

func (s Void) Delete(ctx context.Context, id ID) error {
	return nil
}

func (s Void) MaxSize() int {
	return DefaultMaxSize
}

func (s Void) Hash(x []byte) ID {
	return DefaultHash(x)
}
