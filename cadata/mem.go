package cadata

import (
	"bytes"
	"context"
	"io"

	"go.brendoncarroll.net/state/kv"
)

var _ Store = &MemStore{}

type MemStore struct {
	hash    HashFunc
	maxSize int
	s       *kv.MemStore[ID, []byte]
}

func NewMem(hf HashFunc, maxSize int) *MemStore {
	return &MemStore{
		maxSize: maxSize,
		hash:    hf,
		s: kv.NewMemStore[ID, []byte](func(a, b ID) int {
			return bytes.Compare(a[:], b[:])
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
	var data []byte
	if err := s.s.Get(ctx, id, &data); err != nil {
		return 0, err
	}
	if len(buf) < len(data) {
		return 0, io.ErrShortBuffer
	}
	return copy(buf, data), nil
}

func (s *MemStore) List(ctx context.Context, span Span, ids []ID) (n int, err error) {
	return s.s.List(ctx, span, ids)
}

func (s *MemStore) Delete(ctx context.Context, id ID) error {
	return s.s.Delete(ctx, id)
}

func (s *MemStore) Exists(ctx context.Context, id ID) (bool, error) {
	return s.s.Exists(ctx, id)
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
	return 0, ErrNotFound{Key: id}
}

func (s Void) Exists(ctx context.Context, id ID) (bool, error) {
	return false, nil
}

func (s Void) List(ctx context.Context, span Span, ids []ID) (int, error) {
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
