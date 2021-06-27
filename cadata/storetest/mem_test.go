package storetest

import (
	"testing"

	"github.com/brendoncarroll/go-state/cadata"
)

func TestMemStore(t *testing.T) {
	TestStore(t, func(t testing.TB) Store {
		return cadata.NewMem(cadata.DefaultMaxSize)
	})
}
