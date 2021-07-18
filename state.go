package state

import "bytes"

// ByteRange represents a lexicographically sorted range of []byte.
// The range is [First, Last), meaning First is included, and Last is excluded.
// Last is ignored if set to nil, and the range is assumed to have no upper bound.
// When First is the empty []byte, it implies no lower bound, since it is inclusive.
type ByteRange struct {
	First []byte
	Last  []byte
}

// Contains returns true if x is in the Range
func (r ByteRange) Contains(x []byte) bool {
	return !(r.AllGt(x) || r.AllLt(x))
}

// AllLt returns true if every key in the range is less than x
func (r ByteRange) AllLt(x []byte) bool {
	return r.Last != nil && bytes.Compare(r.Last, x) <= 0
}

// AllGt returns true if every key in the range is greater than or equal to x
func (r ByteRange) AllGt(x []byte) bool {
	return bytes.Compare(r.First, x) > 0
}
