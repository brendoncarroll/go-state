package posixfs

import (
	"errors"
	"io"
	gofs "io/fs"
	"os"
)

const Sep = "/"

const (
	O_RDONLY = os.O_RDONLY
	O_WRONLY = os.O_WRONLY
	O_EXCL   = os.O_EXCL
	O_APPEND = os.O_APPEND
	O_CREATE = os.O_CREATE
	O_TRUNC  = os.O_TRUNC
)

var (
	ErrExist    = gofs.ErrExist
	ErrNotExist = gofs.ErrNotExist
	ErrClosed   = gofs.ErrClosed
)

func IsErrNotExist(err error) bool {
	return errors.Is(err, ErrNotExist)
}

func IsErrExist(err error) bool {
	return errors.Is(err, ErrExist)
}

type FileMode = gofs.FileMode

type DirEnt struct {
	Name string
	Mode FileMode
}

type FileInfo = gofs.FileInfo

type File interface {
	Stat() (FileInfo, error)
	Sync() error
	io.Closer
	io.Reader
	io.Writer
	io.Seeker

	ReadDir(n int) ([]DirEnt, error)
}

type FS interface {
	OpenFile(p string, flag int, perm os.FileMode) (File, error)
	Mkdir(p string, perm os.FileMode) error
	Rmdir(p string) error
	Remove(p string) error
	Rename(oldPath, newPath string) error
	Stat(p string) (FileInfo, error)
	Symlink(target, link string) error
}
