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

type Reader interface {
	Read(ctx context.Context, id ID, buf []byte) (int, error)
}

type Poster interface {
	Post(ctx context.Context, data []byte) (ID, error)
}

type Adder interface {
	Add(ctx context.Context, id ID) error
}

type Deleter interface {
	Delete(ctx context.Context, id ID) error
}

type Lister interface {
	List(ctx context.Context, prefix []byte, ids []ID) (int, error)
}

type Set interface {
	Exists(ctx context.Context, id ID) (bool, error)
	Lister
}

type Store interface {
	Poster
	Reader
	Deleter
	Set

	Hash(data []byte) ID
	MaxSize() int
}

var (
	ErrNotFound = errors.New("no data found with that ID")
	ErrTooLarge = errors.New("data is too large for store")
	ErrTooMany  = errors.New("too many blobs to list")
)

func IsNotFound(err error) bool {
	return err == ErrNotFound
}

func IsTooLarge(err error) bool {
	return err == ErrTooLarge
}

func IsTooMany(err error) bool {
	return err == ErrTooMany
}
