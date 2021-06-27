package cells

import (
	"context"
	"errors"
)

// Cell is a compare-and-swap cell
type Cell interface {
	// CAS sets the contents of the cell to next, IFF prev equals the cell contents.
	// returns whether or not the swap was successful, the actual value in the cell, or an error
	// if err != nil then success must be false.
	// the swap failing is not considered an error.
	CAS(ctx context.Context, actual, prev, next []byte) (success bool, n int, err error)

	// Read retrieves the contents of the cell, and writes them to buf.
	// If err != nil the data returned is invalid.
	// If err == nil, n will be the number of bytes written
	Read(ctx context.Context, buf []byte) (n int, err error)

	// MaxSize returns the maximum amount of data that can be stored in this cell
	MaxSize() int
}

// Apply attempts to do a CAS on the cell by applying fn to the current value to get the next value.
func Apply(ctx context.Context, cell Cell, fn func([]byte) ([]byte, error)) error {
	const maxAttempts = 10

	var (
		buf     = make([]byte, cell.MaxSize())
		success bool
	)
	n, err := cell.Read(ctx, buf)
	if err != nil {
		return err
	}
	for i := 0; i < maxAttempts; i++ {
		prev := buf[:n]
		next, err := fn(prev)
		if err != nil {
			return err
		}
		success, n, err = cell.CAS(ctx, buf, prev, next)
		if err != nil {
			return err
		}
		if success {
			return nil
		}
	}
	return errors.New("cell CAS attempts maxed out")
}

// GetBytes is a convenience function that allocates memory, fills it with the contents of cell and returns it
func GetBytes(ctx context.Context, cell Cell) ([]byte, error) {
	buf := make([]byte, cell.MaxSize())
	n, err := cell.Read(ctx, buf)
	if err != nil {
		return nil, err
	}
	return buf[:n], nil
}

type ErrTooLarge struct{}

func IsErrTooLarge(err error) bool {
	return errors.Is(err, ErrTooLarge{})
}

func (e ErrTooLarge) Error() string {
	return "data too large for cell"
}
