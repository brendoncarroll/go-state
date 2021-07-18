package fs

import (
	"os"
	"path/filepath"
)

// NewDirFS creates an OS backed FS rooted at p
func NewDirFS(p string) FS {
	return NewPrefixed(NewOSFS(), p)
}

type osFS struct{}

// NewOSFS creates a new filesystem backed by the operating system
func NewOSFS() FS {
	return osFS{}
}

func (osFS) OpenFile(p string, flag int, perm os.FileMode) (File, error) {
	p = filepath.FromSlash(p)
	f, err := os.OpenFile(p, flag, perm)
	if err != nil {
		return nil, err
	}
	finfo, err := f.Stat()
	if err != nil {
		f.Close()
		return nil, err
	}
	if finfo.IsDir() {
		return osDir{f}, nil
	} else {
		return osFile{f}, nil
	}
}

func (osFS) Mkdir(p string, perm os.FileMode) error {
	p = filepath.FromSlash(p)
	return os.Mkdir(p, perm)
}

func (osFS) Rmdir(p string) error {
	p = filepath.FromSlash(p)
	return os.RemoveAll(p)
}

func (osFS) Stat(p string) (FileInfo, error) {
	p = filepath.FromSlash(p)
	return os.Stat(p)
}

func (osFS) Remove(p string) error {
	p = filepath.FromSlash(p)
	return os.Remove(p)
}

func (osFS) Rename(oldPath, newPath string) error {
	oldPath = filepath.FromSlash(oldPath)
	newPath = filepath.FromSlash(newPath)
	return os.Rename(oldPath, newPath)
}

var _ RegularFile = osFile{}

type osFile struct {
	f *os.File
}

func (f osFile) Read(p []byte) (int, error) {
	return f.f.Read(p)
}

func (f osFile) Write(p []byte) (int, error) {
	return f.f.Write(p)
}

func (f osFile) Close() error {
	return f.f.Close()
}

func (f osFile) Sync() error {
	return f.f.Sync()
}

func (f osFile) Stat() (FileInfo, error) {
	return f.f.Stat()
}

var _ Directory = osDir{}

type osDir struct {
	f *os.File
}

func (d osDir) ReadDir(n int) ([]DirEnt, error) {
	dirEnts, err := d.f.ReadDir(n)
	if err != nil {
		return nil, err
	}
	ents := make([]DirEnt, len(dirEnts))
	for i := range dirEnts {
		ents[i] = DirEnt{
			Mode: dirEnts[i].Type(),
			Name: dirEnts[i].Name(),
		}
	}
	return ents, nil
}

func (d osDir) Close() error {
	return d.f.Close()
}

func (d osDir) Sync() error {
	return d.f.Sync()
}

func (d osDir) Stat() (FileInfo, error) {
	return d.f.Stat()
}
