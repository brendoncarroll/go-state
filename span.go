package state

// Span is a specification for iteration through values of type T.
// It includes begin and end bounds, which can be inclusive or exclusive of the bounding value.
// The empty Span is equivalent to TotalSpan() and includes all elements of type T.
type Span[T any] struct {
	lower, upper T
	mode         int32
}

const (
	spanModeHasLower = 1 << iota
	spanModeIncludesLower
	spanModeHasUpper
	spanModeIncludesUpper
)

// TotalSpan returns a Span[T] which contains all elements of T
func TotalSpan[T any]() Span[T] {
	return Span[T]{}
}

// WithLowerIncl returns a copy of s with an inclusive lower bound
func (s Span[T]) WithLowerIncl(x T) Span[T] {
	s.lower = x
	s.mode |= spanModeHasLower
	s.mode |= spanModeIncludesLower
	return s
}

// WithLower returns a copy of s with an exclusive lower bound.
func (s Span[T]) WithLowerExcl(x T) Span[T] {
	s.lower = x
	s.mode |= spanModeHasLower
	s.mode &= ^spanModeIncludesLower
	return s
}

// WithUpperIncl returns a copy of s with an inclusive upper bound
func (s Span[T]) WithUpperIncl(x T) Span[T] {
	s.upper = x
	s.mode |= spanModeHasUpper
	s.mode |= spanModeIncludesUpper
	return s
}

// WithLowerExcl returns a copy of s with an exclusive lower bound
func (s Span[T]) WithUpperExcl(x T) Span[T] {
	s.upper = x
	s.mode |= spanModeHasUpper
	s.mode &= ^spanModeIncludesUpper
	return s
}

// WithoutLower returns a copy of s with no lower bound.
func (s Span[T]) WithoutLower(x T) Span[T] {
	s.mode &= ^spanModeHasLower
	s.mode &= ^spanModeIncludesLower
	return s
}

// WithoutUpper returns a copy of s with no upper bound.
func (s Span[T]) WithoutUpper(x T) Span[T] {
	s.mode &= ^spanModeHasUpper
	s.mode &= ^spanModeIncludesUpper
	return s
}

// LowerBound returns the Span's lower bound and true if the span has a lower bound.
func (s Span[T]) LowerBound() (T, bool) {
	ok := s.mode&spanModeHasLower > 0
	return s.lower, ok
}

// UpperBound returns the Span's upper bound and true if the span has an upper bound.
func (s Span[T]) UpperBound() (T, bool) {
	ok := s.mode&spanModeHasUpper > 0
	return s.upper, ok
}

func (s Span[T]) IsDesc() bool {
	// TODO: descending Span's not yet supported.
	return false
}

// IncludesUpper returns true if the upper bound is inclusive
func (s Span[T]) IncludesUpper() bool {
	return s.mode&spanModeIncludesUpper > 0
}

// IncludesLower returns true if the lower bound is inclusive
func (s Span[T]) IncludesLower() bool {
	return s.mode&spanModeIncludesLower > 0
}

// Compare determines if an element T is below the Span, in the Span or above the Span.
// The output should be interpretted as s - x.  Similar to bytes.Compare and strings.Compare.
// If the span > x then 1.
// If the span < x then -1.
// If the span contains x then 0.
func (s Span[T]) Compare(x T, cmp func(a, b T) int) int {
	if lower, ok := s.LowerBound(); ok {
		c := cmp(lower, x)
		if c > 0 {
			return 1
		} else if c == 0 && !s.IncludesLower() {
			return 1
		}
	}
	if upper, ok := s.UpperBound(); ok {
		c := cmp(upper, x)
		if c < 0 {
			return -1
		} else if c == 0 && !s.IncludesUpper() {
			return -1
		}
	}
	return 0
}

// Contains returns true if the span contains x.
func (s Span[T]) Contains(x T, cmp func(a, b T) int) bool {
	return s.Compare(x, cmp) == 0
}
