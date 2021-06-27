package cryptocell

import (
	"crypto/rand"
	"encoding/hex"
	"testing"

	"github.com/brendoncarroll/go-state/cells"
	"github.com/brendoncarroll/go-state/cells/celltest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const maxSize = 1 << 16

// TODO: add memcell for tests
func TestSecretBox(t *testing.T) {
	celltest.CellTestSuite(t, func(testing.TB) cells.Cell {
		mc := cells.NewMem(maxSize)
		return NewSecretBox(mc, testSecret(t))
	})
}

func TestEncryptDecrypt(t *testing.T) {
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
