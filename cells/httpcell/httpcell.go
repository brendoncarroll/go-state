package httpcell

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/brendoncarroll/go-state/cells"
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

var _ cells.Cell = &Cell{}

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

func (c *Cell) Read(ctx context.Context, buf []byte) (int, error) {
	req := c.newRequest(ctx, http.MethodGet, c.spec.URL, nil)

	resp, err := c.hc.Do(req)
	if err != nil {
		return 0, err
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			log.Println(err)
		}
	}()
	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("bad response %v", resp.Status)
	}
	return readResponse(buf, resp.Body)
}

func (c *Cell) CAS(ctx context.Context, actual, cur, next []byte) (bool, int, error) {
	if len(next) > c.MaxSize() {
		return false, 0, cells.ErrTooLarge{}
	}
	curHash := sha3.Sum256(cur)
	curHashb64 := base64.URLEncoding.EncodeToString(curHash[:])

	req := c.newRequest(ctx, http.MethodPut, c.spec.URL, bytes.NewBuffer(next))
	req.Header.Set(CurrentHeader, curHashb64)

	resp, err := c.hc.Do(req)
	if err != nil {
		return false, 0, err
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			log.Println(err)
		}
	}()
	n, err := readResponse(actual, resp.Body)
	if err != nil {
		return false, 0, err
	}
	success := bytes.Equal(next, actual[:n])
	return success, n, nil
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

func readResponse(dst []byte, r io.Reader) (int, error) {
	var n int
	for {
		n2, err := r.Read(dst)
		if err != nil && err != io.EOF {
			return n, err
		}
		n += n2
		if err == io.EOF {
			break
		}
	}
	return n, nil
}
