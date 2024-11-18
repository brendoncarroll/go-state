package fsstore

import (
	"context"
	mrand "math/rand"
	"testing"

	"github.com/stretchr/testify/require"
	"go.brendoncarroll.net/state/cadata"
	"go.brendoncarroll.net/state/cadata/storetest"
	"go.brendoncarroll.net/state/posixfs"
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
