package storetest

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/brendoncarroll/go-state/cadata"
)

type (
	Store = cadata.Store
	ID    = cadata.ID
)

func TestStore(t *testing.T, newStore func(t testing.TB) Store) {
	t.Run("ExistsPostRead", func(t *testing.T) {
		s := newStore(t)
		testData := make([]byte, 1024)
		id1 := s.Hash(testData)
		require.False(t, exists(t, s, id1))
		id2 := post(t, s, testData)
		require.Equal(t, id1, id2)
		require.True(t, exists(t, s, id1))
		dataOut := get(t, s, id2)
		require.Equal(t, testData, dataOut)
	})
}

func get(t *testing.T, s Store, id ID) []byte {
	ctx := context.Background()
	buf := make([]byte, s.MaxSize())
	n, err := s.Read(ctx, id, buf)
	require.NoError(t, err)
	return buf[:n]
}

func post(t *testing.T, s Store, data []byte) ID {
	ctx := context.Background()
	id, err := s.Post(ctx, data)
	require.NoError(t, err)
	return id
}

func exists(t *testing.T, s Store, id ID) bool {
	ctx := context.Background()
	yes, err := s.Exists(ctx, id)
	require.NoError(t, err)
	return yes
}
