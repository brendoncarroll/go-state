package state

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMemStore(t *testing.T) {
	ctx := context.Background()
	s := NewMemKVStore[string, int](func(a, b string) bool { return a < b })
	s.Put(ctx, "a", 1)
	s.Put(ctx, "b", 2)
	s.Put(ctx, "c", 3)
	s.Put(ctx, "b", 4)
	s.Put(ctx, "d", 5)
	s.Delete(ctx, "d")

	v, err := s.Get(ctx, "b")
	require.NoError(t, err)
	require.Equal(t, 4, v)

	var ks []string
	err = ForEach[string](ctx, s, func(k string) error {
		ks = append(ks, k)
		return nil
	})
	require.NoError(t, err)
	require.Equal(t, []string{"a", "b", "c"}, ks)
}
