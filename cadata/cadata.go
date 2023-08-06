package cadata

import (
	"context"
	"errors"

	"lukechampine.com/blake3"
)

type HashFunc = func(data []byte) ID

func DefaultHash(x []byte) ID {
	return ID(blake3.Sum256(x))
}

const DefaultMaxSize = 1 << 20

// Getter defines the Get method
type Getter interface {
	// Get copies data identified by id into buf.
	// If not all the data can be copied, it returns io.ErrShortBuffer
	// Otherwise n will be the number of bytes copied into buf.
	Get(ctx context.Context, id ID, buf []byte) (int, error)
	Hash(x []byte) ID
	MaxSize() int
}

// Poster defines the Post method
type Poster interface {
	// Post will store data, and return an ID that can be used to retrieve it later
	Post(ctx context.Context, data []byte) (ID, error)
	MaxSize() int
	Hash(x []byte) ID
}

// Adder defines the Add method
type Adder interface {
	// Add adds data to the store by ID.
	// It will return ErrNotFound if the data cannot be added.
	Add(ctx context.Context, id ID) error
}

// Deleter defines the Delete method
type Deleter interface {
	// Delete removes the data identified by id from the store
	Delete(ctx context.Context, id ID) error
}

// Lister defines the List method
type Lister interface {
	// List reads IDs from the store, in asceding order into ids.
	// All the ids will be >= gteq, if it is not nil.
	// All the ids will be < lt if it is not nil.
	List(ctx context.Context, span Span, ids []ID) (int, error)
}

// Exister defines the Exists method
type Exister interface {
	// Exists returns true, nil if the ID exists, and false, nil if it does not.
	// The boolean should not be interpretted if err != nil
	Exists(ctx context.Context, id ID) (bool, error)
}

// GetLister combines the  Getter and Lister interfaces
type GetLister interface {
	Lister
	Getter
}

type GetPoster interface {
	Getter
	Poster
}

type PostExister interface {
	Poster
	Exister
}

type Set interface {
	Adder
	Deleter
	Exister
	Lister
}

type Store interface {
	Poster
	Getter
	Deleter
	Lister
	Exister
}

var (
	ErrNotFound = errors.New("no data found with that ID")
	ErrTooLarge = errors.New("data is too large for store")
	ErrBadData  = errors.New("data does not match ID")
)

func IsNotFound(err error) bool {
	return errors.Is(err, ErrNotFound)
}

func IsTooLarge(err error) bool {
	return errors.Is(err, ErrTooLarge)
}

// Check ensures that hf(data) == id and returns ErrBadData if it does not.
func Check(hf HashFunc, id ID, data []byte) error {
	id2 := hf(data)
	if id != id2 {
		return ErrBadData
	}
	return nil
}
