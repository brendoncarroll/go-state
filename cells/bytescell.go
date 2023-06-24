package cells

import (
	"bytes"
	"errors"
)

// BytesCell is a cell holding []byte
type BytesCell interface {
	Cell[[]byte]
	MaxSize() int
}

func CopyBytes(dst *[]byte, src []byte) {
	*dst = append((*dst)[:0], src...)
}

func EqualBytes(a, b []byte) bool {
	return bytes.Equal(a, b)
}

// ErrTooLarge is returned from BytesCells when the object is too large to store
type ErrTooLarge struct{}

func IsErrTooLarge(err error) bool {
	return errors.Is(err, ErrTooLarge{})
}

func (e ErrTooLarge) Error() string {
	return "data too large for cell"
}

type MemBytes struct {
	*MemCell[[]byte]
	maxSize int
}

func NewMemBytes(maxSize int) *MemBytes {
	return &MemBytes{
		MemCell: NewMem(EqualBytes, CopyBytes),
		maxSize: maxSize,
	}
}

func (mc *MemBytes) MaxSize() int {
	return mc.maxSize
}
