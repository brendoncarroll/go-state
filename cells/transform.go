package cells

import (
	"bytes"
	"context"
	"io"
)

// TransformFunc is called to transform the contents of a cell.
type TransformFunc = func(dst []byte, src []byte) (int, error)

// IdentityTransform is a trival TransformFunc which copies src to dst.
// If len(dst) < len(src) it returns io.ErrShortBuffer
func IdentityTransform(dst, src []byte) (int, error) {
	if len(dst) < len(src) {
		return 0, io.ErrShortBuffer
	}
	return copy(dst, src), nil
}

type Transform struct {
	inner    Cell
	upward   TransformFunc
	downward TransformFunc
	overhead int
}

func NewTransform(inner Cell, upward, downward TransformFunc, overhead int) Cell {
	return &Transform{
		inner:    inner,
		upward:   upward,
		downward: downward,
		overhead: overhead,
	}
}

func (tf *Transform) Read(ctx context.Context, buf []byte) (int, error) {
	buf2, relBuf := tf.acquire()
	defer relBuf()
	n, err := tf.inner.Read(ctx, buf2)
	if err != nil {
		return 0, err
	}
	return tf.upward(buf, buf2[:n])
}

func (tf *Transform) CAS(ctx context.Context, actual []byte, prev, next []byte) (bool, int, error) {
	actual2, relActual2 := tf.acquire()
	defer relActual2()

	a2n, err := tf.inner.Read(ctx, actual2[:])
	if err != nil {
		return false, 0, err
	}
	an, err := tf.upward(actual, actual2[:a2n])
	if err != nil {
		return false, 0, err
	}
	if !bytes.Equal(actual[:an], prev) {
		return false, an, nil
	}

	next2, relNext2 := tf.acquire()
	defer relNext2()
	n2n, err := tf.downward(next2, next)
	if err != nil {
		return false, 0, err
	}
	swapped, an2, err := tf.inner.CAS(ctx, actual2, actual2[:a2n], next2[:n2n])
	if err != nil {
		return false, 0, err
	}
	an, err = tf.upward(actual, actual2[:an2])
	if err != nil {
		return false, 0, err
	}
	return swapped, an, nil
}

func (tf *Transform) MaxSize() int {
	return tf.inner.MaxSize() - tf.overhead
}

func (tf *Transform) acquire() ([]byte, func()) {
	// TODO: buffer pool
	return make([]byte, tf.inner.MaxSize()), func() {}
}
