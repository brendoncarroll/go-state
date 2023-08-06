package posixfs

import (
	"fmt"
	"os"
)

// ErrReadOnly is returns by ReadOnly when an operation would
// mutate the underlying filesystem
type ErrReadOnly struct {
	Op string
}

func (e ErrReadOnly) Error() string {
	return fmt.Sprintf("operation (%v) not allowed because the filesystem is read only", e.Op)
}

var _ FS = ReadOnly{}

// ReadOnly is a read only filesystem
type ReadOnly struct {
	inner FS
}

// NewReadOnly returns a version of x which is read only
func NewReadOnly(x FS) ReadOnly {
	return ReadOnly{inner: x}
}

func (fs ReadOnly) OpenFile(p string, flags int, mode FileMode) (File, error) {
	flags |= O_RDONLY
	return fs.inner.OpenFile(p, flags, mode)
}

func (fs ReadOnly) Mkdir(p string, perm os.FileMode) error {
	return ErrReadOnly{Op: "mkdir"}
}

func (fs ReadOnly) Rmdir(p string) error {
	return ErrReadOnly{Op: "rmdir"}
}

func (fs ReadOnly) Remove(p string) error {
	return ErrReadOnly{Op: "remove"}
}

func (fs ReadOnly) Rename(oldPath, newPath string) error {
	return ErrReadOnly{Op: "rename"}
}

func (fs ReadOnly) Stat(p string) (FileInfo, error) {
	return fs.inner.Stat(p)
}

func (fs ReadOnly) Symlink(oldp, newp string) error {
	return ErrReadOnly{Op: "symlink"}
}
