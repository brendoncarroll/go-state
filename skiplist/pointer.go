package skiplist

import (
	"sync/atomic"
	"unsafe"
)

// A Pointer is an atomic pointer of type *T. The zero value is a nil *T.
type Pointer[T any] struct {
	v unsafe.Pointer
}

// Load atomically loads and returns the value stored in x.
func (x *Pointer[T]) Load() *T { return (*T)(atomic.LoadPointer(&x.v)) }

// Store atomically stores val into x.
func (x *Pointer[T]) Store(val *T) { atomic.StorePointer(&x.v, unsafe.Pointer(val)) }

// Swap atomically stores new into x and returns the previous value.
func (x *Pointer[T]) Swap(new *T) (old *T) {
	return (*T)(atomic.SwapPointer(&x.v, unsafe.Pointer(new)))
}

// CompareAndSwap executes the compare-and-swap operation for x.
func (x *Pointer[T]) CompareAndSwap(old, new *T) (swapped bool) {
	return atomic.CompareAndSwapPointer(&x.v, unsafe.Pointer(old), unsafe.Pointer(new))
}
