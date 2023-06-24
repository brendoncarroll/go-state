package cells

import (
	"context"
	"errors"
)

// Cell is a compare-and-swap cell
type Cell[T any] interface {
	// CAS sets the contents of the cell to next, IFF prev equals the cell contents.
	// returns whether or not the swap was successful, the actual value in the cell, or an error
	// if err != nil then success must be false.
	// the swap failing is not considered an error.
	CAS(ctx context.Context, actual *T, prev, next T) (success bool, err error)

	// Load retrieves the contents of the cell, and writes them to buf.
	// If err != nil the data returned is invalid.
	// If err == nil, n will be the number of bytes written
	Load(ctx context.Context, dst *T) error

	// Equals is the equality function used by the cell.
	Equals(a, b T) bool

	// Copy is the function used to copy values in and out of the cell.
	Copy(dst *T, src T)
}

func DefaultEquals[T comparable](a, b T) bool {
	return a == b
}

func DefaultCopy[T any](dst *T, src T) {
	*dst = src
}

// Apply attempts to do a CAS on the cell by applying fn to the current value to get the next value.
func Apply[T any](ctx context.Context, cell Cell[T], maxAttempts int, fn func(T) (T, error)) error {
	var (
		actual  T
		success bool
	)
	if err := cell.Load(ctx, &actual); err != nil {
		return err
	}
	for i := 0; i < maxAttempts; i++ {
		prev := actual
		next, err := fn(prev)
		if err != nil {
			return err
		}
		success, err = cell.CAS(ctx, &actual, prev, next)
		if err != nil {
			return err
		}
		if success {
			return nil
		}
	}
	return ErrCASMaxAttempts{}
}

func Load[T any](ctx context.Context, cell Cell[T]) (ret T, _ error) {
	return ret, cell.Load(ctx, &ret)
}

type ErrCASMaxAttempts struct{}

func IsErrCASMaxAttempts(err error) bool {
	return errors.Is(err, ErrCASMaxAttempts{})
}

func (e ErrCASMaxAttempts) Error() string {
	return "cell CAS attempts maxed out"
}
