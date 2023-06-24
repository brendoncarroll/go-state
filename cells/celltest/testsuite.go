package celltest

import (
	"context"
	"testing"

	fuzz "github.com/google/gofuzz"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/brendoncarroll/go-state/cells"
)

func TestCell[T any](t *testing.T, factory func(t testing.TB) cells.Cell[T]) {
	ctx := context.TODO()
	t.Run("InitEmpty", func(t *testing.T) {
		t.Parallel()
		c := factory(t)
		var buf T
		require.NoError(t, c.Load(ctx, &buf))
	})
	t.Run("CAS", func(t *testing.T) {
		t.Parallel()
		c := factory(t)
		var buf T
		require.NoError(t, c.Load(ctx, &buf))
		const N = 10
		fz := fuzz.New()
		for i := 0; i < N; i++ {
			prev := buf
			var next T
			fz.Fuzz(&next)
			t.Log("next:", next)
			success, err := c.CAS(ctx, &buf, prev, next)
			require.NoError(t, err)
			require.True(t, success)
		}
	})
}

func TestBytesCell(t *testing.T, factory func(t testing.TB) cells.BytesCell) {
	ctx := context.TODO()
	TestCell[[]byte](t, func(t testing.TB) cells.Cell[[]byte] { return factory(t) })

	t.Run("TooLarge", func(t *testing.T) {
		t.Parallel()
		c := factory(t)
		var buf []byte
		next := make([]byte, c.MaxSize()+1)
		success, err := c.CAS(ctx, &buf, nil, next)
		assert.False(t, success)
		assert.Error(t, err)
	})
}
