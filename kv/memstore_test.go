package kv

import (
	"context"
	"strings"
	"testing"

	"github.com/brendoncarroll/go-state"
	"github.com/stretchr/testify/require"
)

func TestMemStore(t *testing.T) {
	ctx := context.Background()
	s := NewMemStore[string, int](func(a, b string) int {
		return strings.Compare(a, b)
	})
	s.Put(ctx, "a", 1)
	s.Put(ctx, "b", 2)
	s.Put(ctx, "c", 3)
	s.Put(ctx, "b", 4)
	s.Put(ctx, "d", 5)
	s.Delete(ctx, "d")

	v, err := Get[string, int](ctx, s, "b")
	require.NoError(t, err)
	require.Equal(t, 4, v)

	var ks []string
	err = ForEach[string](ctx, s, state.TotalSpan[string](), func(k string) error {
		ks = append(ks, k)
		return nil
	})
	require.NoError(t, err)
	require.Equal(t, []string{"a", "b", "c"}, ks)
}
