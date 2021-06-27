package cryptocell

import (
	"bytes"
	"context"
	"crypto/rand"
	"io"

	"github.com/brendoncarroll/go-state/cells"
	"github.com/pkg/errors"
	"golang.org/x/crypto/nacl/secretbox"
)

var _ cells.Cell = &SecretBoxCell{}

type SecretBoxCell struct {
	inner  cells.Cell
	secret []byte
}

func NewSecretBox(inner cells.Cell, secret []byte) *SecretBoxCell {
	return &SecretBoxCell{
		inner:  inner,
		secret: secret,
	}
}

func (c *SecretBoxCell) Read(ctx context.Context, buf []byte) (int, error) {
	data, _, err := c.get(ctx)
	if err != nil {
		return 0, err
	}
	if len(buf) < len(data) {
		return 0, io.ErrShortBuffer
	}
	return copy(buf, data), nil
}

func (c *SecretBoxCell) get(ctx context.Context) (data, ctext []byte, err error) {
	buf := make([]byte, c.inner.MaxSize())
	n, err := c.inner.Read(ctx, buf)
	if err != nil {
		return nil, nil, err
	}
	ctext = buf[:n]
	if len(ctext) == 0 {
		return nil, nil, nil
	}
	ptext, err := decrypt(ctext, c.secret)
	if err != nil {
		return nil, ctext, err
	}
	return ptext, ctext, nil
}

func (c *SecretBoxCell) CAS(ctx context.Context, actual, prev, next []byte) (bool, int, error) {
	data, ctext, err := c.get(ctx)
	if err != nil {
		return false, 0, err
	}
	n, err := copyToBuffer(actual, data)
	if err != nil {
		return false, 0, err
	}
	// if they have the wrong plaintext, they won't have the right ciphertext
	if !bytes.Equal(data, prev) {
		return false, n, nil
	}
	nextCtext := encrypt(next, c.secret)
	swapped, n, err := c.inner.CAS(ctx, actual, ctext, nextCtext)
	if err != nil {
		return false, n, err
	}
	if n > 0 {
		// decrypt and copy into actual
		// TODO: inplace decryption
		ptext, err := decrypt(actual[:n], c.secret)
		if err != nil {
			return false, 0, err
		}
		n = copy(actual, ptext)
	}
	return swapped, n, nil
}

func (c *SecretBoxCell) MaxSize() int {
	return c.inner.MaxSize() - secretbox.Overhead
}

func encrypt(ptext, secret []byte) []byte {
	nonce := [24]byte{}
	if _, err := rand.Read(nonce[:]); err != nil {
		panic(err)
	}
	s := [32]byte{}
	copy(s[:], secret)
	return secretbox.Seal(nonce[:], ptext, &nonce, &s)
}

func decrypt(ctext, secret []byte) ([]byte, error) {
	const nonceSize = 24
	if len(ctext) < nonceSize {
		return nil, errors.Errorf("secret box too short")
	}
	nonce := [nonceSize]byte{}
	copy(nonce[:], ctext[:nonceSize])
	s := [32]byte{}
	copy(s[:], secret)
	ptext, success := secretbox.Open([]byte{}, ctext[nonceSize:], &nonce, &s)
	if !success {
		return nil, errors.Errorf("secret box was invalid")
	}
	return ptext, nil
}

func copyToBuffer(dst, src []byte) (int, error) {
	if len(dst) < len(src) {
		return 0, io.ErrShortBuffer
	}
	return copy(dst, src), nil
}
