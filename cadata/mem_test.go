package cadata_test

import (
	"testing"

	"go.brendoncarroll.net/state/cadata"
	"go.brendoncarroll.net/state/cadata/storetest"
)

func TestMemStore(t *testing.T) {
	storetest.TestStore(t, func(t testing.TB) cadata.Store {
		return cadata.NewMem(cadata.DefaultHash, cadata.DefaultMaxSize)
	})
}
