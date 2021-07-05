package cryptocell

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/hex"
	"testing"

	"github.com/brendoncarroll/go-state/cells"
	"github.com/brendoncarroll/go-state/cells/celltest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/chacha20poly1305"
)

const maxSize = 1 << 16

func TestAEAD(t *testing.T) {
	t.Run("ChaCha20Poly1305", func(t *testing.T) {
		celltest.CellTestSuite(t, func(testing.TB) cells.Cell {
			mc := cells.NewMem(maxSize)
			aead, err := chacha20poly1305.NewX(testSecret(t))
			require.NoError(t, err)
			return NewAEAD(mc, aead)
		})
	})
	t.Run("AES256-GCM", func(t *testing.T) {
		celltest.CellTestSuite(t, func(testing.TB) cells.Cell {
			mc := cells.NewMem(maxSize)
			ciph, err := aes.NewCipher(testSecret(t))
			require.NoError(t, err)
			aead, err := cipher.NewGCMWithNonceSize(ciph, 24)
			require.NoError(t, err)
			return NewAEAD(mc, aead)
		})
	})
}

func TestSecretBox(t *testing.T) {
	celltest.CellTestSuite(t, func(testing.TB) cells.Cell {
		mc := cells.NewMem(maxSize)
		return NewSecretBox(mc, testSecret(t))
	})
}

func TestSecretBoxEncryptDecrypt(t *testing.T) {
	secret := make([]byte, 32)

	ptext := []byte("hello world")
	ctext := encrypt(ptext, secret)
	t.Log(hex.Dump(ctext))

	ptext2, err := decrypt(ctext, secret)
	require.Nil(t, err)
	t.Log(string(ptext2))

	ctextTamper := append([]byte{}, ctext...)
	ctextTamper[0] ^= 1
	_, err = decrypt(ctextTamper, secret)
	assert.NotNil(t, err)
}

func testSecret(t *testing.T) []byte {
	secret := make([]byte, 32)
	_, err := rand.Read(secret)
	require.NoError(t, err)
	return secret
}
