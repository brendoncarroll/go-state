package fsstore

import (
	"testing"

	"github.com/brendoncarroll/go-state/cadata"
	"github.com/brendoncarroll/go-state/cadata/storetest"
	"github.com/brendoncarroll/go-state/posixfs"
)

func TestFSStore(t *testing.T) {
	storetest.TestStore(t, func(t testing.TB) cadata.Store {
		fsx := posixfs.NewTestFS(t)
		return New(fsx, cadata.DefaultHash, cadata.DefaultMaxSize)
	})
}
