package cadata

import (
	"bytes"
	"context"
	"errors"
	"io"

	"github.com/brendoncarroll/go-state"
)

var _ Store = &MemStore{}

type MemStore struct {
	hash    HashFunc
	maxSize int
	s       *state.MemKVStore[ID, []byte]
}

func NewMem(hf HashFunc, maxSize int) *MemStore {
	return &MemStore{
		maxSize: maxSize,
		hash:    hf,
		s: state.NewMemKVStore[ID, []byte](func(a, b ID) bool {
			return bytes.Compare(a[:], b[:]) < 0
		}),
	}
}

func (s *MemStore) Post(ctx context.Context, data []byte) (ID, error) {
	if len(data) > s.MaxSize() {
		return ID{}, ErrTooLarge
	}
	data = append([]byte{}, data...)
	id := s.hash(data)
	if err := s.s.Put(ctx, id, data); err != nil {
		return ID{}, err
	}
	return id, nil
}

func (s *MemStore) Get(ctx context.Context, id ID, buf []byte) (int, error) {
	data, err := s.s.Get(ctx, id)
	if err != nil {
		if errors.Is(err, state.ErrNotFound) {
			err = ErrNotFound
		}
		return 0, err
	}
	if len(buf) < len(data) {
		return 0, io.ErrShortBuffer
	}
	return copy(buf, data), nil
}

func (s *MemStore) List(ctx context.Context, first ID, ids []ID) (n int, err error) {
	return s.s.List(ctx, first, ids)
}

func (s *MemStore) Delete(ctx context.Context, id ID) error {
	return s.s.Delete(ctx, id)
}

func (s *MemStore) Exists(ctx context.Context, id ID) (bool, error) {
	return state.Exists[ID](ctx, s.s, id)
}

func (s *MemStore) Len() (count int) {
	return s.s.Len()
}

func (s *MemStore) Hash(x []byte) ID {
	return s.hash(x)
}

func (s *MemStore) MaxSize() int {
	return s.maxSize
}

var _ Store = Void{}

type Void struct {
	hf      HashFunc
	maxSize int
}

func NewVoid(hf HashFunc, maxSize int) Void {
	return Void{
		hf:      hf,
		maxSize: maxSize,
	}
}

func (s Void) Post(ctx context.Context, data []byte) (ID, error) {
	return s.Hash(data), nil
}

func (s Void) Get(ctx context.Context, id ID, buf []byte) (int, error) {
	return 0, ErrNotFound
}

func (s Void) Exists(ctx context.Context, id ID) (bool, error) {
	return false, nil
}

func (s Void) List(ctx context.Context, first ID, ids []ID) (int, error) {
	return 0, nil
}

func (s Void) Delete(ctx context.Context, id ID) error {
	return nil
}

func (s Void) MaxSize() int {
	return s.maxSize
}

func (s Void) Hash(x []byte) ID {
	return s.hf(x)
}
