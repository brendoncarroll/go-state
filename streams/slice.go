package streams

import "context"

type Slice[T any] struct {
	xs  []T
	pos int
	cp  func(dst *T, src T)
}

func NewSlice[T any](xs []T, cp func(*T, T)) *Slice[T] {
	if cp == nil {
		cp = func(dst *T, src T) { *dst = src }
	}
	return &Slice[T]{xs: xs, pos: 0, cp: cp}
}

func (it *Slice[T]) Next(ctx context.Context, dst *T) error {
	if it.pos >= len(it.xs) {
		return EOS()
	}
	it.cp(dst, it.xs[it.pos])
	it.pos++
	return nil
}

func (it *Slice[T]) Reset() {
	it.pos = 0
}
