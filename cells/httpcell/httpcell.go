package httpcell

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"log"
	"net/http"

	"go.brendoncarroll.net/state/cells"
	"go.brendoncarroll.net/stdctx/logctx"
	"go.uber.org/zap"
	"golang.org/x/crypto/sha3"
)

const (
	CurrentHeader = "X-Current"
	MaxSize       = 1 << 16
)

type Spec struct {
	URL     string
	Headers map[string]string
}

var _ cells.Cell[[]byte] = &Cell{}

type Cell struct {
	spec Spec
	hc   *http.Client
}

func New(spec Spec) *Cell {
	return &Cell{
		spec: spec,
		hc:   http.DefaultClient,
	}
}

func (c *Cell) URL() string {
	return c.spec.URL
}

func (c *Cell) Load(ctx context.Context, dst *[]byte) error {
	req := c.newRequest(ctx, http.MethodGet, c.spec.URL, nil)

	resp, err := c.hc.Do(req)
	if err != nil {
		return err
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			logctx.Error(ctx, "closing http response body", zap.Error(err))
		}
	}()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad response %v", resp.Status)
	}
	*dst, err = readResponse((*dst)[:0], resp.Body)
	return err
}

func (c *Cell) CAS(ctx context.Context, actual *[]byte, cur, next []byte) (bool, error) {
	if len(next) > c.MaxSize() {
		return false, cells.ErrTooLarge{}
	}
	curHash := sha3.Sum256(cur)
	curHashb64 := base64.URLEncoding.EncodeToString(curHash[:])

	req := c.newRequest(ctx, http.MethodPut, c.spec.URL, bytes.NewBuffer(next))
	req.Header.Set(CurrentHeader, curHashb64)

	resp, err := c.hc.Do(req)
	if err != nil {
		return false, err
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			log.Println(err)
		}
	}()
	*actual, err = readResponse((*actual)[:0], resp.Body)
	if err != nil {
		return false, err
	}
	success := bytes.Equal(next, *actual)
	return success, nil
}

func (c *Cell) Copy(dst *[]byte, src []byte) {
	cells.CopyBytes(dst, src)
}

func (c *Cell) Equals(a, b []byte) bool {
	return cells.EqualBytes(a, b)
}

func (c *Cell) MaxSize() int {
	return MaxSize
}

func (c *Cell) newRequest(ctx context.Context, method, u string, body io.Reader) *http.Request {
	req, err := http.NewRequest(method, u, body)
	if err != nil {
		panic(err)
	}
	req = req.WithContext(ctx)

	for k, v := range c.spec.Headers {
		req.Header.Set(k, v)
	}
	return req
}

// readResponse appends data from r to out, and returns it.
func readResponse(out []byte, r io.Reader) ([]byte, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}
	return append(out, data...), nil
}
