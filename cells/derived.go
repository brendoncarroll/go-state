package cells

import (
	"context"
	"sync"
)

var _ Cell[struct{}] = &Derived[struct{}, struct{}]{}

type Derived[X, Y any] struct {
	p DerivedParams[X, Y]

	mu       sync.Mutex
	valid    bool
	currentX X
	currentY Y
}

type DerivedParams[X, Y any] struct {
	Inner   Cell[X]
	Forward func(ctx context.Context, dst *Y, src X) error
	Inverse func(ctx context.Context, dst *X, src Y) error
	Eq      func(Y, Y) bool
	Copy    func(*Y, Y)
}

func NewDerived[X, Y any](params DerivedParams[X, Y]) *Derived[X, Y] {
	return &Derived[X, Y]{
		p: params,
	}
}

func (c *Derived[X, Y]) Load(ctx context.Context, dst *Y) error {
	var x X
	if err := c.p.Inner.Load(ctx, &x); err != nil {
		return err
	}
	if err := c.forward(ctx, dst, x); err != nil {
		return err
	}
	return nil
}

func (c *Derived[X, Y]) CAS(ctx context.Context, actual *Y, prev, next Y) (bool, error) {
	// First get the previous
	var prevX X
	c.mu.Lock()
	if !c.valid || !c.p.Eq(c.currentY, prev) {
		c.mu.Unlock()
		return false, c.Load(ctx, actual)
	} else {
		c.p.Inner.Copy(&prevX, c.currentX)
		c.mu.Unlock()
	}
	var actualX, nextX X
	if err := c.p.Inverse(ctx, &nextX, next); err != nil {
		return false, err
	}
	swapped, err := c.p.Inner.CAS(ctx, &actualX, prevX, nextX)
	if err != nil {
		return false, err
	}
	return swapped, c.forward(ctx, actual, actualX)
}

// forward applies the Forward function, but will skip computation if the input is unchanged.
func (c *Derived[X, Y]) forward(ctx context.Context, dst *Y, src X) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if !c.valid || !c.p.Inner.Equals(c.currentX, src) {
		c.valid = false
		c.p.Inner.Copy(&c.currentX, src)
		if err := c.p.Forward(ctx, &c.currentY, c.currentX); err != nil {
			return err
		}
		c.valid = true
	}
	c.p.Copy(dst, c.currentY)
	return nil
}

func (c *Derived[X, Y]) Equals(a, b Y) bool {
	return c.p.Eq(a, b)
}

func (c *Derived[X, Y]) Copy(dst *Y, src Y) {
	c.p.Copy(dst, src)
}
