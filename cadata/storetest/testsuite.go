package storetest

import (
	"context"
	mrand "math/rand"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/brendoncarroll/go-state/cadata"
)

type (
	Store = cadata.Store
	ID    = cadata.ID
)

func TestStore(t *testing.T, newStore func(t testing.TB) Store) {
	t.Run("ExistsPostGet", func(t *testing.T) {
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
	t.Run("List", func(t *testing.T) {
		s := newStore(t)
		buf := make([]byte, 1024)
		for i := 0; i < 100; i++ {
			readRandom(i, buf)
			post(t, s, buf)
		}
		ids := list(t, s)
		require.Len(t, ids, 100)
	})
	t.Run("RecallOne", func(t *testing.T) {
		s := newStore(t)
		testData := []byte("test string goes here")
		id := post(t, s, testData)
		actual := get(t, s, id)
		require.Equal(t, testData, actual)
	})
	t.Run("MaxSize", func(t *testing.T) {
		s := newStore(t)
		data := make([]byte, s.MaxSize())
		post(t, s, data)
		dataTooBig := make([]byte, s.MaxSize()+1)
		ctx := context.Background()
		_, err := s.Post(ctx, dataTooBig)
		require.ErrorIs(t, err, cadata.ErrTooLarge)
	})
}

func get(t *testing.T, s Store, id ID) []byte {
	ctx := context.Background()
	data, err := cadata.GetBytes(ctx, s, id)
	require.NoError(t, err)
	return data
}

func post(t *testing.T, s Store, data []byte) ID {
	ctx := context.Background()
	id, err := s.Post(ctx, data)
	require.NoError(t, err)
	return id
}

func exists(t *testing.T, s Store, id ID) bool {
	ctx := context.Background()
	yes, err := cadata.Exists(ctx, s, id)
	require.NoError(t, err)
	return yes
}

func list(t *testing.T, s Store) (ret []cadata.ID) {
	ctx := context.Background()
	err := cadata.ForEach(ctx, s, cadata.Span{}, func(id ID) error {
		ret = append(ret, id)
		return nil
	})
	require.NoError(t, err)
	return ret
}

func readRandom(i int, buf []byte) {
	rng := mrand.New(mrand.NewSource(int64(i)))
	rng.Read(buf)
}
