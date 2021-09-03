package posixfs

import "github.com/pkg/errors"

type filtered struct {
	x         FS
	predicate func(string) bool
}

// NewFiltered returns a filesystem with some paths filtered.
func NewFiltered(x FS, predicate func(string) bool) FS {
	return filtered{
		x:         x,
		predicate: predicate,
	}
}

func (fs filtered) OpenFile(p string, flag int, perm FileMode) (File, error) {
	if err := fs.checkPath(p); err != nil {
		return nil, err
	}
	return fs.x.OpenFile(p, flag, perm)
}

func (fs filtered) Mkdir(p string, perm FileMode) error {
	if err := fs.checkPath(p); err != nil {
		return err
	}
	return fs.x.Mkdir(p, perm)
}

func (fs filtered) Rmdir(p string) error {
	if err := fs.checkPath(p); err != nil {
		return err
	}
	return fs.x.Rmdir(p)
}

func (fs filtered) Remove(p string) error {
	if err := fs.checkPath(p); err != nil {
		return err
	}
	return fs.x.Remove(p)
}

func (fs filtered) Rename(prev, next string) error {
	if err := fs.checkPath(prev); err != nil {
		return err
	}
	if err := fs.checkPath(next); err != nil {
		return err
	}
	return fs.x.Rename(prev, next)
}

func (fs filtered) Stat(p string) (FileInfo, error) {
	if err := fs.checkPath(p); err != nil {
		return nil, err
	}
	return fs.x.Stat(p)
}

func (fs filtered) checkPath(p string) error {
	if fs.predicate(p) {
		return nil
	}
	return errors.Errorf("path %s has been filtered", p)
}