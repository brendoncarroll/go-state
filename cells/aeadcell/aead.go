package aeadcell

import (
	"crypto/cipher"
	"crypto/rand"
	"fmt"
	"io"

	"github.com/brendoncarroll/go-state/cells"
)

const MinNonceSize = 20

// New creates a new cell using AEAD using secret
// and then calls NewAEAD with it.
func New(inner cells.Cell, aead cipher.AEAD) cells.Cell {
	if aead.NonceSize() < MinNonceSize {
		panic("AEAD's nonce size is too small to use a random nonce")
	}
	overhead := aead.NonceSize() + aead.Overhead()
	upward := func(dst, src []byte) (int, error) {
		if len(dst) < len(src)-overhead {
			return 0, io.ErrShortBuffer
		}
		if len(src) == 0 {
			return 0, nil
		}
		if len(src) < overhead {
			return 0, fmt.Errorf("too short (len=%d) to be AEAD nonce + ciphertext, min=%d", len(src), overhead)
		}
		nonce := src[:aead.NonceSize()]
		ctext := src[aead.NonceSize():]
		_, err := aead.Open(dst[:0], nonce, ctext, nil)
		return len(src) - overhead, err
	}
	downward := func(dst, src []byte) (int, error) {
		if len(dst) < len(src)+overhead {
			return 0, io.ErrShortBuffer
		}
		nonce := dst[:aead.NonceSize()]
		ctext := dst[aead.NonceSize():]
		if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
			return 0, err
		}
		aead.Seal(ctext[:0], nonce, src, nil)
		return len(src) + overhead, nil
	}
	return cells.NewTransform(inner, upward, downward, overhead)
}
