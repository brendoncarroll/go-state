package posixfs

import (
	"bytes"
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPutFile(t *testing.T) {
	ctx := context.Background()
	x := NewTestFS(t)
	testPath := "test-path"

	err := PutFile(ctx, x, testPath, 0o644, bytes.NewBufferString("test-data"))
	require.NoError(t, err)
	finfo, err := x.Stat(testPath)
	require.NoError(t, err)
	assert.False(t, finfo.IsDir())
}

func TestDeleteFile(t *testing.T) {
	ctx := context.Background()
	x := NewTestFS(t)
	testPath := "test-path"

	err := PutFile(ctx, x, testPath, 0o644, bytes.NewBufferString("test-data"))
	require.NoError(t, err)
	// should be idempotent
	for i := 0; i < 3; i++ {
		err = DeleteFile(ctx, x, testPath)
		require.NoError(t, err)
	}
	_, err = x.Stat(testPath)
	require.True(t, errors.Is(err, ErrNotExist))
}
