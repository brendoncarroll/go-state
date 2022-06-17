package cadata

import (
	mrand "math/rand"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBase64Alphabet(t *testing.T) {
	a := Base64Alphabet
	for i := 1; i < len(a); i++ {
		require.Greater(t, a[i], a[i-1], "%s, %s", string(a[i]), string(a[i-1]))
	}
}

func TestBase64Marshal(t *testing.T) {
	const N = 100
	var buf [32]byte
	for i := 0; i < N; i++ {
		mrand.Read(buf[:])
		expected := IDFromBytes(buf[:])
		data, err := expected.MarshalBase64()
		require.NoError(t, err)
		actual := ID{}
		require.NoError(t, actual.UnmarshalBase64(data))
		require.Equal(t, expected, actual)
	}
}
