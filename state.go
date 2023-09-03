package state

import (
	"bytes"
	"errors"
	"fmt"
)

type ErrNotFound[K any] struct {
	Key K
}

func (e ErrNotFound[K]) Error() string {
	return fmt.Sprintf("no entry found for %v", e.Key)
}

func IsErrNotFound[K any](err error) bool {
	return errors.As(err, &ErrNotFound[K]{})
}

// ByteSpan represents a lexicographically sorted span of []byte.
// The span is [Begin, End), meaning Begin is included, and End is excluded.
// End is ignored if set to nil, and the span is assumed to have no upper bound.
// When Begin is the empty []byte, it implies no lower bound, since it is inclusive.
type ByteSpan struct {
	Begin []byte
	End   []byte
}

// String implements fmt.Stringer
func (s ByteSpan) String() string {
	return fmt.Sprintf("[%v, %v)", s.Begin, s.End)
}

// Contains returns true if x is in the Span
func (s ByteSpan) Contains(x []byte) bool {
	return !(s.AllGt(x) || s.AllLt(x))
}

// AllLt returns true if every key in the span is less than x
func (s ByteSpan) AllLt(x []byte) bool {
	return s.End != nil && bytes.Compare(s.End, x) <= 0
}

// AllGt returns true if every key in the span is greater than x
func (s ByteSpan) AllGt(x []byte) bool {
	return bytes.Compare(s.Begin, x) > 0
}

func (s ByteSpan) ToSpan() Span[[]byte] {
	span := TotalSpan[[]byte]()
	if s.Begin != nil {
		span = span.WithLowerIncl(s.Begin)
	}
	if s.End != nil {
		span = span.WithUpperExcl(s.End)
	}
	return span
}
