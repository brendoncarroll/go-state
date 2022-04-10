package state

import (
	"bytes"
	"fmt"
)

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

// Span represents a sorted span of values of type T.
// Elements of the span will all be >= Begin and < End.
//
// When End is the zero value of type T, the Span should be considered to have no upper bound.
// Constructing an empty Span is much less useful than a total span.
// An empty Span can be constructed with a non-zero Begin and End where Begin == End.
// Spans where End < Begin are considered invalid.
type Span[T comparable] struct {
	Begin T
	End   T
}
