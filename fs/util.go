package fs

import (
	"context"
	"io"
	gofs "io/fs"
	"io/ioutil"
	"os"
	"path"
	"testing"
	"time"

	"github.com/pkg/errors"
)

func NewTestFS(t testing.TB) FS {
	return NewDirFS(t.TempDir())
}

// ReadDir returns all the children of the directory at p
func ReadDir(fs FS, p string) ([]DirEnt, error) {
	f, err := fs.OpenFile(p, 0, gofs.FileMode(os.O_RDONLY))
	if err != nil {
		return nil, err
	}
	dir, ok := f.(Directory)
	if !ok {
		return nil, errors.Errorf("non-directory at %q", p)
	}
	return dir.ReadDir(0)
}

// PutFile opens the file in fs at path p, truncates it, and writes from r until io.EOF
func PutFile(ctx context.Context, fs FS, p string, perm os.FileMode, r io.Reader) error {
	f, err := fs.OpenFile(p, O_TRUNC|O_WRONLY|O_CREATE, perm)
	if err != nil {
		return err
	}
	defer f.Close()
	rf, ok := f.(RegularFile)
	if !ok {
		return errors.Errorf("cannot write to non-regular file at %q", p)
	}
	if err := maybeSetReadDeadline(ctx, rf); err != nil {
		return err
	}
	if _, err := io.Copy(rf, r); err != nil {
		return err
	}
	return f.Close()
}

func AppendFile(ctx context.Context, fs FS, p string, perm os.FileMode, data []byte) error {
	f, err := fs.OpenFile(p, O_WRONLY|O_APPEND|O_CREATE, perm)
	if err != nil {
		return err
	}
	defer f.Close()
	rf, ok := f.(RegularFile)
	if !ok {
		return errors.Errorf("cannot append to non-regular file at %q", p)
	}
	if err := maybeSetWriteDeadline(ctx, rf); err != nil {
		return err
	}
	if _, err := rf.Write(data); err != nil {
		return err
	}
	return f.Close()
}

// ReadFile reads an entire file into memory and returns it.
func ReadFile(ctx context.Context, fs FS, p string) ([]byte, error) {
	f, err := fs.OpenFile(p, O_RDONLY, 0)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	rf, ok := f.(RegularFile)
	if !ok {
		return nil, errors.Errorf("cannot read from non-regular file at %q", p)
	}
	if err := maybeSetWriteDeadline(ctx, rf); err != nil {
		return nil, err
	}
	return ioutil.ReadAll(rf)
}

// DeleteFile is an idempotent delete operation.
// It calls remove, but does not error if the path is already gone
func DeleteFile(ctx context.Context, fs FS, p string) error {
	err := fs.Remove(p)
	if errors.Is(err, ErrNotExist) {
		return nil
	}
	return err
}

func maybeSetWriteDeadline(ctx context.Context, f File) error {
	type writeDeadline interface {
		SetWriteDeadlin(t time.Time) error
	}
	deadline, ok := ctx.Deadline()
	if !ok {
		return nil
	}
	x, ok := f.(writeDeadline)
	if !ok {
		return nil
	}
	return x.SetWriteDeadlin(deadline)
}

func maybeSetReadDeadline(ctx context.Context, f File) error {
	type readDeadline interface {
		SetReadDeadline(t time.Time) error
	}
	deadline, ok := ctx.Deadline()
	if !ok {
		return nil
	}
	x, ok := f.(readDeadline)
	if !ok {
		return nil
	}
	return x.SetReadDeadline(deadline)
}

// WalkLeaves walks fsx starting at path p, and calls fn for every non-dir file encountered.
// The first argument to fn will be the path of the file.
// The second argument to fn will be its DirEnt in its immediate parent in the walk.
func WalkLeaves(x FS, p string, fn func(string, DirEnt) error) error {
	dirEnts, err := ReadDir(x, p)
	if err != nil {
		return err
	}
	for _, dirEnt := range dirEnts {
		p2 := path.Join(p, dirEnt.Name)
		if dirEnt.Mode.IsDir() {
			if err := WalkLeaves(x, p2, fn); err != nil {
				return err
			}
		} else {
			if err := fn(p2, dirEnt); err != nil {
				return err
			}
		}
	}
	return err
}
