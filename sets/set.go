package state

import (
	"context"

	"go.brendoncarroll.net/state/kv"
)

// Adder has the Add method
type Adder[K comparable] interface {
	Add(ctx context.Context, k K) error
}

// Set represents a set of elements
type Set[K comparable] interface {
	kv.Lister[K]
	Adder[K]
	kv.Deleter[K]
}

type memSet[K comparable] struct {
	*kv.MemStore[K, struct{}]
}

// NewMemSet returns a Set using memory for storage.
func NewMemSet[K comparable](cmp func(a, b K) int) Set[K] {
	return memSet[K]{
		MemStore: kv.NewMemStore[K, struct{}](cmp),
	}
}

func (ms memSet[K]) Add(ctx context.Context, k K) error {
	return ms.MemStore.Put(ctx, k, struct{}{})
}
