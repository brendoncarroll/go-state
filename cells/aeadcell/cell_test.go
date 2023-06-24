package aeadcell

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"testing"

	"github.com/brendoncarroll/go-state/cells"
	"github.com/brendoncarroll/go-state/cells/celltest"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/chacha20poly1305"
)

const maxSize = 1 << 16

func TestAEAD(t *testing.T) {
	t.Run("ChaCha20Poly1305", func(t *testing.T) {
		celltest.TestCell(t, func(testing.TB) cells.Cell[[]byte] {
			mc := cells.NewMemBytes(maxSize)
			aead, err := chacha20poly1305.NewX(testSecret(t))
			require.NoError(t, err)
			return New(mc, aead)
		})
	})
	t.Run("AES256-GCM", func(t *testing.T) {
		celltest.TestCell(t, func(testing.TB) cells.Cell[[]byte] {
			mc := cells.NewMemBytes(maxSize)
			ciph, err := aes.NewCipher(testSecret(t))
			require.NoError(t, err)
			aead, err := cipher.NewGCMWithNonceSize(ciph, 24)
			require.NoError(t, err)
			return New(mc, aead)
		})
	})
}

func testSecret(t *testing.T) []byte {
	secret := make([]byte, 32)
	_, err := rand.Read(secret)
	require.NoError(t, err)
	return secret
}
