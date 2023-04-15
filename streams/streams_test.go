package streams

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSlice(t *testing.T) {
	ctx := context.Background()
	it := NewSlice([]int{0, 1, 2, 3, 4}, nil)

	var dst int
	for i := 0; i < 5; i++ {
		require.NoError(t, it.Next(ctx, &dst))
		require.Equal(t, i, dst)
	}
	for i := 0; i < 3; i++ {
		require.ErrorIs(t, it.Next(ctx, &dst), EOS())
	}
}
