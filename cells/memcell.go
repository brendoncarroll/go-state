package cells

import (
	"context"
	"sync"
)

type MemCell[T any] struct {
	eq func(T, T) bool
	cp func(*T, T)

	mu    sync.RWMutex
	value T
}

func NewMem[T any](eq func(T, T) bool, cp func(*T, T)) *MemCell[T] {
	return &MemCell[T]{
		eq: eq,
		cp: cp,
	}
}

func (c *MemCell[T]) Load(ctx context.Context, dst *T) error {
	c.mu.RLock()
	defer c.mu.RUnlock()
	c.cp(dst, c.value)
	return nil
}

func (c *MemCell[T]) CAS(ctx context.Context, actual *T, prev, next T) (bool, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	var success bool
	if c.eq(prev, c.value) {
		c.cp(&c.value, next)
		success = true
	}
	c.cp(actual, c.value)
	return success, nil
}

func (c *MemCell[T]) Copy(dst *T, src T) {
	c.cp(dst, src)
}

func (c *MemCell[T]) Equals(a, b T) bool {
	return c.eq(a, b)
}
