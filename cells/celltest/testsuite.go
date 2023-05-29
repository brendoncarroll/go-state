package celltest

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/brendoncarroll/go-state/cells"
)

type Cell = cells.Cell

func CellTestSuite(t *testing.T, factory func(t testing.TB) Cell) {
	ctx := context.TODO()
	t.Run("InitEmpty", func(t *testing.T) {
		t.Parallel()
		c := factory(t)
		buf := make([]byte, c.MaxSize())
		n, err := c.Read(ctx, buf)
		require.NoError(t, err)
		assert.Equal(t, n, 0)
	})
	t.Run("CAS", func(t *testing.T) {
		t.Parallel()
		c := factory(t)
		buf := make([]byte, c.MaxSize())
		n, err := c.Read(ctx, buf)
		require.NoError(t, err)
		const N = 10
		for i := 0; i < N; i++ {
			prev := buf[:n]
			next := []byte(fmt.Sprint("test data ", i))
			var success bool
			success, n, err = c.CAS(ctx, buf, prev, next)
			require.NoError(t, err)
			require.True(t, success)
		}
	})
	t.Run("TooLarge", func(t *testing.T) {
		t.Parallel()
		c := factory(t)
		buf := make([]byte, c.MaxSize())
		next := make([]byte, c.MaxSize()+1)
		success, _, err := c.CAS(ctx, buf, nil, next)
		assert.False(t, success)
		assert.Error(t, err)
	})
}
