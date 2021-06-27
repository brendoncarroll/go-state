package cadata_test

import (
	"testing"

	"github.com/brendoncarroll/go-state/cadata"
	"github.com/brendoncarroll/go-state/cadata/storetest"
)

func TestMemStore(t *testing.T) {
	storetest.TestStore(t, func(t testing.TB) cadata.Store {
		return cadata.NewMem(cadata.DefaultMaxSize)
	})
}
