package state

import "bytes"

// ByteSpan represents a lexicographically sorted span of []byte.
// The span is [Begin, End), meaning Begin is included, and End is excluded.
// End is ignored if set to nil, and the span is assumed to have no upper bound.
// When Begin is the empty []byte, it implies no lower bound, since it is inclusive.
type ByteSpan struct {
	Begin []byte
	End   []byte
}

// Contains returns true if x is in the Span
func (r ByteSpan) Contains(x []byte) bool {
	return !(r.AllGt(x) || r.AllLt(x))
}

// AllLt returns true if every key in the span is less than x
func (r ByteSpan) AllLt(x []byte) bool {
	return r.End != nil && bytes.Compare(r.End, x) <= 0
}

// AllGt returns true if every key in the span is greater than x
func (r ByteSpan) AllGt(x []byte) bool {
	return bytes.Compare(r.Begin, x) > 0
}
