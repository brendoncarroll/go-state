package cells_test

import (
	"context"
	"strconv"
	"testing"

	"go.brendoncarroll.net/state/cells"
	"go.brendoncarroll.net/state/cells/celltest"
)

func TestDerived(t *testing.T) {
	celltest.TestCell(t, func(t testing.TB) cells.Cell[uint64] {
		c1 := cells.NewMem[string](cells.DefaultEquals[string], cells.DefaultCopy[string])
		c2 := cells.NewDerived(cells.DerivedParams[string, uint64]{
			Inner: c1,
			Forward: func(ctx context.Context, dst *uint64, src string) (err error) {
				if src == "" {
					*dst = 0
					return nil
				}
				*dst, err = strconv.ParseUint(src, 10, 64)
				return err
			},
			Inverse: func(ctx context.Context, dst *string, src uint64) error {
				*dst = strconv.FormatUint(src, 10)
				return nil
			},
			Equals: cells.DefaultEquals[uint64],
			Copy:   cells.DefaultCopy[uint64],
		})
		return c2
	})
}
