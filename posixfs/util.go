package posixfs

import (
	"context"
	"fmt"
	"io"
	gofs "io/fs"
	"os"
	"path"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/brendoncarroll/go-state"
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
	defer f.Close()
	return f.ReadDir(0)
}

// PutFile opens the file in fs at path p, truncates it, and writes from r until io.EOF
func PutFile(ctx context.Context, fs FS, p string, perm os.FileMode, r io.Reader) error {
	f, err := fs.OpenFile(p, O_TRUNC|O_WRONLY|O_CREATE, perm)
	if err != nil {
		return err
	}
	defer f.Close()
	if err := maybeSetWriteDeadline(ctx, f); err != nil {
		return err
	}
	if _, err := io.Copy(f, r); err != nil {
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
	if err := maybeSetWriteDeadline(ctx, f); err != nil {
		return err
	}
	if _, err := f.Write(data); err != nil {
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
	if err := maybeSetReadDeadline(ctx, f); err != nil {
		return nil, err
	}
	return io.ReadAll(f)
}

// DeleteFile is an idempotent delete operation.
// It calls remove, but does not error if the path is already gone
func DeleteFile(ctx context.Context, fs FS, p string) error {
	err := fs.Remove(p)
	if IsErrNotExist(err) {
		return nil
	}
	return err
}

func maybeSetWriteDeadline(ctx context.Context, f File) error {
	type writeDeadline interface {
		SetWriteDeadline(t time.Time) error
	}
	deadline, ok := ctx.Deadline()
	if !ok {
		return nil
	}
	x, ok := f.(writeDeadline)
	if !ok {
		return nil
	}
	return x.SetWriteDeadline(deadline)
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
func WalkLeaves(ctx context.Context, x FS, p string, fn func(string, DirEnt) error) error {
	f, err := x.OpenFile(p, O_RDONLY, 0)
	if err != nil {
		return err
	}
	defer f.Close()
	finfo, err := f.Stat()
	if err != nil {
		return err
	}
	if !finfo.IsDir() {
		return fn(p, DirEnt{Name: finfo.Name(), Mode: finfo.Mode()})
	}
	dirEnts, err := f.ReadDir(0)
	if err != nil {
		return err
	}
	for _, dirEnt := range dirEnts {
		p2 := path.Join(p, dirEnt.Name)
		if dirEnt.Mode.IsDir() {
			if err := checkContext(ctx); err != nil {
				return err
			}
			if err := WalkLeaves(ctx, x, p2, fn); err != nil {
				return err
			}
		} else {
			if err := fn(p2, dirEnt); err != nil {
				return err
			}
		}
	}
	return nil
}

// WalkLeavesSpan walks the leaves in x which are contained in the span.
// WalkLeavesSpan emits paths in sorted order.
func WalkLeavesSpan(ctx context.Context, x FS, p string, span state.Span[string], fn func(string, DirEnt) error) error {
	f, err := x.OpenFile(p, O_RDONLY, 0)
	if err != nil {
		return err
	}
	defer f.Close()
	finfo, err := f.Stat()
	if err != nil {
		return err
	}
	if !finfo.IsDir() {
		if !span.Contains(p, strings.Compare) {
			return nil
		}
		return fn(p, DirEnt{Name: finfo.Name(), Mode: finfo.Mode()})
	}
	dirEnts, err := f.ReadDir(0)
	if err != nil {
		return err
	}
	sort.Slice(dirEnts, func(i, j int) bool {
		return dirEnts[i].Name < dirEnts[j].Name
	})
	for _, dirEnt := range dirEnts {
		p2 := path.Join(p, dirEnt.Name)
		if dirEnt.Mode.IsDir() {
			if err := checkContext(ctx); err != nil {
				return err
			}
			begin := p2 + "/"
			end := prefixEnd(begin)
			if span.Compare(begin, strings.Compare) < 0 || span.Compare(end, strings.Compare) > 0 {
				continue
			}
			if err := WalkLeavesSpan(ctx, x, p2, span, fn); err != nil {
				return err
			}
		} else {
			if !span.Contains(p2, strings.Compare) {
				continue
			}
			if err := fn(p2, dirEnt); err != nil {
				return err
			}
		}
	}
	return nil
}

func checkContext(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		return nil
	}
}

func cleanPath(x string) string {
	x = path.Clean(x)
	if x == "." || x == "/" {
		x = ""
	}
	return x
}

// MkdirAll creates all of the directories from the root of x, down to p
func MkdirAll(x FS, p string, perm FileMode) error {
	p = cleanPath(p)
	if p != "" {
		parent, _ := path.Split(p)
		if err := MkdirAll(x, parent, perm); err != nil {
			return err
		}
	}
	err := x.Mkdir(p, perm)
	if IsErrExist(err) {
		finfo, err := x.Stat(p)
		if err != nil {
			return err
		}
		if !finfo.IsDir() {
			return fmt.Errorf("non-dir at %v", p)
		}
		return nil
	}
	return err
}

func prefixEnd(prefix string) string {
	if len(prefix) == 0 {
		return ""
	}
	var end []byte
	for i := len(prefix) - 1; i >= 0; i-- {
		c := prefix[i]
		if c < 0xff {
			end = make([]byte, i+1)
			copy(end, prefix)
			end[i] = c + 1
			break
		}
	}
	return string(end)
}
