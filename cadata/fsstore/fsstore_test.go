package fsstore

import (
	"context"
	mrand "math/rand"
	"testing"

	"github.com/brendoncarroll/go-state/cadata"
	"github.com/brendoncarroll/go-state/cadata/storetest"
	"github.com/brendoncarroll/go-state/posixfs"
	"github.com/stretchr/testify/require"
)

func TestFSStore(t *testing.T) {
	storetest.TestStore(t, func(t testing.TB) cadata.Store {
		fsx := posixfs.NewTestFS(t)
		return New(fsx, cadata.DefaultHash, cadata.DefaultMaxSize)
	})
}

func BenchmarkStore(b *testing.B) {
	ctx := context.Background()
	rng := mrand.New(mrand.NewSource(0))
	b.Run("Post", func(b *testing.B) {
		b.ReportAllocs()
		b.StopTimer()
		fsx := posixfs.NewTestFS(b)
		s := New(fsx, cadata.DefaultHash, cadata.DefaultMaxSize)
		var buf [1 << 16]byte
		for i := 0; i < b.N; i++ {
			b.StopTimer()
			rng.Read(buf[:])
			b.StartTimer()
			_, err := s.Post(ctx, buf[:])
			require.NoError(b, err)
			b.SetBytes(int64(len(buf)))
		}
	})
}
