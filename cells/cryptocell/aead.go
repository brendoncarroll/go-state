package cryptocell

import (
	"bytes"
	"context"
	"crypto/cipher"
	"crypto/rand"

	"github.com/brendoncarroll/go-state/cells"
	"github.com/pkg/errors"
	"golang.org/x/crypto/chacha20poly1305"
)

type AEAD struct {
	inner cells.Cell
	aead  cipher.AEAD
}

// NewChaCha20Poly1305 creats a chaha20poly1305 aead using secret
// and then calls NewAEAD with it.
func NewChaCha20Poly1305(inner cells.Cell, secret []byte) cells.Cell {
	a, err := chacha20poly1305.NewX(secret)
	if err != nil {
		panic(err)
	}
	return &AEAD{
		inner: inner,
		aead:  a,
	}
}

// NewAEAD returns a cell which encrypts its contents using the provided AEAD construction
// If the nonce size is less than 20 then NewAEAD panics
func NewAEAD(inner cells.Cell, aead cipher.AEAD) *AEAD {
	if aead.NonceSize() < 20 {
		panic("nonce size too small for random nonces")
	}
	return &AEAD{
		inner: inner,
		aead:  aead,
	}
}

func (c *AEAD) Read(ctx context.Context, buf []byte) (int, error) {
	msg, err := cells.GetBytes(ctx, c.inner)
	if err != nil {
		return 0, err
	}
	return c.open(buf, msg)
}

func (c *AEAD) CAS(ctx context.Context, actual, prev, next []byte) (bool, int, error) {
	msg, err := cells.GetBytes(ctx, c.inner)
	if err != nil {
		return false, 0, err
	}
	n, err := c.open(actual, msg)
	if err != nil {
		return false, 0, err
	}
	if !bytes.Equal(actual[:n], prev) {
		return false, 0, nil
	}
	buf := make([]byte, c.inner.MaxSize())
	n, err = c.seal(buf, next)
	if err != nil {
		return false, 0, err
	}
	success, n, err := c.inner.CAS(ctx, buf, msg, buf[:n])
	if err != nil {
		return false, 0, err
	}
	n, err = c.open(actual, buf[:n])
	return success, n, err
}

func (c *AEAD) MaxSize() int {
	return c.inner.MaxSize() - c.overhead()
}

func (c *AEAD) open(dst []byte, msg []byte) (int, error) {
	if len(dst) < len(msg)-c.overhead() {
		return 0, errors.Errorf("dst too short")
	}
	if len(msg) == 0 {
		return 0, nil
	}
	if len(msg) < c.aead.Overhead() {
		return 0, errors.Errorf("message too short")
	}
	nonce := msg[:c.aead.NonceSize()]
	ctext := msg[c.aead.NonceSize():]
	dst, err := c.aead.Open(dst[:0], nonce, ctext, nil)
	return len(dst), err
}

func (c *AEAD) seal(dst []byte, ptext []byte) (int, error) {
	if len(dst) < len(ptext)+c.overhead() {
		return 0, errors.Errorf("dst too short")
	}
	nonce := make([]byte, c.aead.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return 0, err
	}
	n := copy(dst, nonce)
	dst = dst[n:]
	c.aead.Seal(dst[:0], nonce, ptext, nil)
	return len(ptext) + c.overhead(), nil
}

func (c *AEAD) overhead() int {
	return c.aead.Overhead() + c.aead.NonceSize()
}
