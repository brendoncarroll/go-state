package celltest

import (
	"testing"

	"github.com/brendoncarroll/go-state/cells"
)

func TestMemCell(t *testing.T) {
	const maxSize = 1 << 16
	CellTestSuite(t, func(t testing.TB) cells.Cell {
		return cells.NewMem(maxSize)
	})
}
