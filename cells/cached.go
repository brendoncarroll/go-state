package cells

import (
	"context"
	"sync"

	"golang.org/x/sync/semaphore"
)

var _ Cell[struct{}] = &Cached[struct{}]{}

type Cached[T any] struct {
	inner Cell[T]

	sem *semaphore.Weighted
	buf *T // buf is used as scratch space

	mu     sync.RWMutex
	cached *T
}

func NewCached[T any](x Cell[T]) *Cached[T] {
	return &Cached[T]{
		inner: x,

		sem: semaphore.NewWeighted(1),
		buf: new(T),
	}
}

func (c *Cached[T]) Load(ctx context.Context, dst *T) error {
	if c.loadFromCache(dst) {
		return nil
	}
	if err := c.sem.Acquire(ctx, 1); err != nil {
		return err
	}
	defer c.sem.Release(1)
	if c.loadFromCache(dst) {
		return nil
	}
	if err := c.inner.Load(ctx, c.buf); err != nil {
		return err
	}
	c.inner.Copy(dst, *c.buf)
	c.swapIntoCache()
	return nil
}

func (c *Cached[T]) CAS(ctx context.Context, actual *T, prev, next T) (bool, error) {
	if err := c.sem.Acquire(ctx, 1); err != nil {
		return false, err
	}
	defer c.sem.Release(1)
	swapped, err := c.inner.CAS(ctx, c.buf, prev, next)
	if err != nil {
		return false, err
	}
	c.inner.Copy(actual, *c.buf)
	c.swapIntoCache()
	return swapped, nil
}

func (c *Cached[T]) Copy(dst *T, src T) {
	c.inner.Copy(dst, src)
}

func (c *Cached[T]) Equals(a, b T) bool {
	return c.inner.Equals(a, b)
}

func (c *Cached[T]) Invalidate() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.cached = nil
}

func (c *Cached[T]) loadFromCache(dst *T) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.cached != nil {
		c.inner.Copy(dst, *c.cached)
		return true
	}
	return false
}

func (c *Cached[T]) swapIntoCache() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.cached, c.buf = c.buf, c.cached
	if c.buf == nil {
		c.buf = new(T)
	}
}
