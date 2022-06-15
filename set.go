package state

import "context"

// Adder has the Add method
type Adder[K comparable] interface {
	Add(ctx context.Context, k K) error
}

// Set represents a set of elements
type Set[K comparable] interface {
	Lister[K]
	Adder[K]
	Deleter[K]
}

type memSet[K comparable] struct {
	*MemKVStore[K, struct{}]
}

// NewMemSet returns a Set using memory for storage.
func NewMemSet[K comparable](cmp func(a, b K) int) Set[K] {
	return memSet[K]{
		MemKVStore: NewMemKVStore[K, struct{}](cmp),
	}
}

func (ms memSet[K]) Add(ctx context.Context, k K) error {
	return ms.Put(ctx, k, struct{}{})
}
