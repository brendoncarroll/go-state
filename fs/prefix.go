package fs

import (
	"os"
	"path"
)

type prefixed struct {
	x      FS
	prefix string
}

func NewPrefixed(x FS, prefix string) FS {
	return prefixed{
		x:      x,
		prefix: prefix,
	}
}

func (fs prefixed) OpenFile(p string, flag int, perm os.FileMode) (File, error) {
	p = path.Join(fs.prefix, p)
	return fs.x.OpenFile(p, flag, perm)
}

func (fs prefixed) Mkdir(p string, perm FileMode) error {
	p = path.Join(fs.prefix, p)
	return fs.x.Mkdir(p, perm)
}

func (fs prefixed) Rmdir(p string) error {
	p = path.Join(fs.prefix, p)
	return fs.x.Rmdir(p)
}

func (fs prefixed) Remove(p string) error {
	p = path.Join(fs.prefix, p)
	return fs.x.Remove(p)
}

func (fs prefixed) Rename(oldPath, newPath string) error {
	oldPath = path.Join(fs.prefix, oldPath)
	newPath = path.Join(fs.prefix, newPath)
	return fs.x.Rename(oldPath, newPath)
}

func (fs prefixed) Stat(p string) (FileInfo, error) {
	p = path.Join(fs.prefix, p)
	return fs.x.Stat(p)
}
