package cells

import (
	"bytes"
	"context"
	"io"
	"sync"
)

type MemCell struct {
	mu      sync.RWMutex
	value   []byte
	maxSize int
}

func NewMem(maxSize int) Cell {
	return &MemCell{
		maxSize: maxSize,
	}
}

func (c *MemCell) Read(ctx context.Context, buf []byte) (int, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return copy(buf, c.value), nil
}

func (c *MemCell) CAS(ctx context.Context, actual, prev, next []byte) (bool, int, error) {
	if len(next) > c.MaxSize() {
		return false, 0, ErrTooLarge{}
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	var success bool
	if bytes.Equal(prev, c.value) {
		c.value = append(c.value[:0], next...)
		success = true
	}
	if len(actual) < len(c.value) {
		return success, 0, io.ErrShortBuffer
	}
	n := copy(actual, c.value)
	return success, n, nil
}

func (c *MemCell) MaxSize() int {
	return c.maxSize
}
