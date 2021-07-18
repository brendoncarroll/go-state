package fsstore

import (
	"bytes"
	"context"
	"encoding/base64"
	"io"
	"path"
	"path/filepath"
	"strings"

	"github.com/brendoncarroll/go-state/cadata"
	"github.com/brendoncarroll/go-state/fs"
	"github.com/pkg/errors"
)

var _ cadata.Store = FSStore{}

type FSStore struct {
	fs       fs.FS
	hashFunc cadata.HashFunc
	maxSize  int
}

func New(x fs.FS, hashFunc cadata.HashFunc, maxSize int) FSStore {
	return FSStore{
		fs:       x,
		hashFunc: hashFunc,
		maxSize:  maxSize,
	}
}

func (s FSStore) Post(ctx context.Context, data []byte) (cadata.ID, error) {
	id := s.hashFunc(data)
	staging := stagingPathForID(id)
	final := pathForID(id)
	if err := s.ensureDirForPath(staging); err != nil {
		return cadata.ID{}, nil
	}
	if err := s.ensureDirForPath(final); err != nil {
		return cadata.ID{}, nil
	}
	if err := atomicPutFile(ctx, s.fs, staging, final, 0o600, data); err != nil {
		return cadata.ID{}, err
	}
	return id, nil
}

func (s FSStore) Read(ctx context.Context, id cadata.ID, buf []byte) (int, error) {
	p := pathForID(id)
	f, err := s.fs.OpenFile(p, fs.O_RDONLY, 0)
	if err != nil {
		if fs.IsErrNotExist(err) {
			err = cadata.ErrNotFound
		}
		return 0, err
	}
	defer f.Close()
	file, ok := f.(fs.RegularFile)
	if !ok {
		return 0, errors.Errorf("unexpected directory at %v", p)
	}
	var n int
	for {
		n2, err := file.Read(buf)
		n += n2
		if err != nil {
			if err == io.EOF {
				err = nil
			}
			return n, err
		}
	}
}

func (s FSStore) Exists(ctx context.Context, id cadata.ID) (bool, error) {
	p := pathForID(id)
	finfo, err := s.fs.Stat(p)
	if errors.Is(err, fs.ErrNotExist) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	if finfo.IsDir() {
		return false, errors.Errorf("expected file, but found directory: %s", p)
	}
	return true, nil
}

func (s FSStore) Delete(ctx context.Context, id cadata.ID) error {
	p := pathForID(id)
	return fs.DeleteFile(ctx, s.fs, p)
}

func (s FSStore) List(ctx context.Context, first []byte, ids []cadata.ID) (int, error) {
	var n int
	err := fs.WalkLeaves(s.fs, "", func(p string, dirEnt fs.DirEnt) error {
		if strings.HasPrefix(p, "tmp/") {
			return nil
		}
		id, err := parsePath(p)
		if err != nil {
			return err
		}
		if bytes.Compare(id[:], first) < 0 {
			return nil
		}
		ids[n] = id
		n++
		return nil
	})
	if err != nil {
		return 0, err
	}
	if n == 0 {
		err = cadata.ErrEndOfList
	}
	return n, err
}

func (s FSStore) MaxSize() int {
	return s.maxSize
}

func (s FSStore) Hash(x []byte) cadata.ID {
	return s.hashFunc(x)
}

func (s FSStore) ensureDirForPath(p string) error {
	dirPath := path.Dir(p)
	return s.fs.Mkdir(dirPath, 0o755)
}

var enc = base64.RawURLEncoding

func pathForID(id cadata.ID) string {
	p := enc.EncodeToString(id[:])
	return path.Join(p[:1], p[1:])
}

func parsePath(p string) (cadata.ID, error) {
	p = strings.Trim(p, "/")
	parts := strings.SplitN(p, "/", 2)
	if len(parts) != 2 {
		return cadata.ID{}, errors.Errorf("could not parse path %q", p)
	}
	data, err := enc.DecodeString(parts[0] + parts[1])
	if err != nil {
		return cadata.ID{}, err
	}
	id := cadata.ID{}
	copy(id[:], data)
	return id, nil
}

func stagingPathForID(id cadata.ID) string {
	p := enc.EncodeToString(id[:])
	return filepath.Join("tmp", p)
}

func atomicPutFile(ctx context.Context, fsx fs.FS, staging, final string, mode fs.FileMode, buf []byte) error {
	if err := fs.PutFile(ctx, fsx, staging, mode, bytes.NewReader(buf)); err != nil {
		return err
	}
	return fsx.Rename(staging, final)
}
