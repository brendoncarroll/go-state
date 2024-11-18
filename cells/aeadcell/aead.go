package aeadcell

import (
	"context"
	"crypto/cipher"
	"crypto/rand"
	"fmt"
	"io"

	"go.brendoncarroll.net/state/cells"
)

const MinNonceSize = 20

type Cell struct {
	inner cells.BytesCell
	aead  cipher.AEAD
	cells.Cell[[]byte]
}

func (c *Cell) MaxSize() int {
	overhead := c.aead.NonceSize() + c.aead.Overhead()
	return c.inner.MaxSize() - overhead
}

// New creates a new cell using AEAD using secret
// and then calls NewAEAD with it.
func New(inner cells.BytesCell, aead cipher.AEAD) cells.BytesCell {
	if aead.NonceSize() < MinNonceSize {
		panic("AEAD's nonce size is too small to use a random nonce")
	}
	overhead := aead.NonceSize() + aead.Overhead()
	forward := func(ctx context.Context, dst *[]byte, src []byte) error {
		if len(src) == 0 {
			cells.CopyBytes(dst, nil)
			return nil
		}
		if len(src) < overhead {
			return fmt.Errorf("too short (len=%d) to be AEAD nonce + ciphertext, min=%d", len(src), overhead)
		}
		nonce := src[:aead.NonceSize()]
		ctext := src[aead.NonceSize():]
		var err error
		*dst, err = aead.Open((*dst)[:0], nonce, ctext, nil)
		return err
	}
	inverse := func(ctx context.Context, dst *[]byte, src []byte) error {
		*dst = (*dst)[:0]
		*dst = append(*dst, make([]byte, aead.NonceSize())...)
		nonce := (*dst)[:aead.NonceSize()]
		if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
			return err
		}
		*dst = aead.Seal(*dst, nonce, src, nil)
		return nil
	}
	d := cells.NewDerived[[]byte, []byte](cells.DerivedParams[[]byte, []byte]{
		Inner:   inner,
		Forward: forward,
		Inverse: inverse,
		Copy:    cells.CopyBytes,
		Equals:  cells.EqualBytes,
	})
	return &Cell{inner: inner, aead: aead, Cell: d}
}
