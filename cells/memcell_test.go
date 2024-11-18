package cells_test

import (
	"testing"

	"go.brendoncarroll.net/state/cells"
	"go.brendoncarroll.net/state/cells/celltest"
)

func TestMemCell(t *testing.T) {
	const maxSize = 1 << 16
	celltest.TestCell[string](t, func(t testing.TB) cells.Cell[string] {
		return cells.NewMem(cells.DefaultEquals[string], cells.DefaultCopy[string])
	})
}
