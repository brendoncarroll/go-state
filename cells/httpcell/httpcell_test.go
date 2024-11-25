package httpcell

import (
	"context"
	"fmt"
	"math/rand"
	"testing"

	"go.brendoncarroll.net/state/cells"
	"go.brendoncarroll.net/state/cells/celltest"
)

func TestSuite(t *testing.T) {
	celltest.TestCell(t, func(t testing.TB) cells.Cell[[]byte] {
		ctx := context.TODO()
		ctx, cf := context.WithCancel(ctx)
		t.Cleanup(cf)
		const addr = "127.0.0.1:"
		server := NewServer()
		go server.Serve(ctx, addr)

		n := rand.Int()
		name := fmt.Sprint("cell-", n)
		server.newCell(name)

		u := server.URL() + name
		cell := New(Spec{URL: u})

		return cell
	})
}
