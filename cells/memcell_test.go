package cells_test

import (
	"testing"

	"github.com/brendoncarroll/go-state/cells"
	"github.com/brendoncarroll/go-state/cells/celltest"
)

func TestMemCell(t *testing.T) {
	const maxSize = 1 << 16
	celltest.CellTestSuite(t, func(t testing.TB) cells.Cell {
		return cells.NewMem(maxSize)
	})
}
