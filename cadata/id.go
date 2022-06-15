package cadata

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"

	"database/sql/driver"

	"github.com/brendoncarroll/go-state"
)

var _ driver.Value = ID{}

const IDSize = 32

// ID identifies a particular piece of data
type ID [IDSize]byte

func IDFromBytes(x []byte) ID {
	id := ID{}
	copy(id[:], x)
	return id
}

func (id ID) String() string {
	return base64.RawURLEncoding.EncodeToString(id[:])
}

func (id ID) MarshalBase64() ([]byte, error) {
	buf := make([]byte, base64.RawURLEncoding.EncodedLen(len(id)))
	base64.RawURLEncoding.Encode(buf, id[:])
	return buf, nil
}

func (id *ID) UnmarshalBase64(data []byte) error {
	n, err := base64.RawURLEncoding.Decode(id[:], data)
	if err != nil {
		return err
	}
	if n != IDSize {
		return errors.New("base64 string is too short")
	}
	return nil
}

func (a ID) Equals(b ID) bool {
	return a.Cmp(b) == 0
}

func (a ID) Cmp(b ID) int {
	return bytes.Compare(a[:], b[:])
}

func (id ID) IsZero() bool {
	return id == (ID{})
}

func (id ID) MarshalJSON() ([]byte, error) {
	s := base64.RawURLEncoding.EncodeToString(id[:])
	return json.Marshal(s)
}

func (id *ID) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	_, err := base64.RawURLEncoding.Decode(id[:], []byte(s))
	return err
}

func (id *ID) Scan(x interface{}) error {
	switch x := x.(type) {
	case []byte:
		if len(x) != 32 {
			return fmt.Errorf("wrong length for cadata.ID HAVE: %d WANT: %d", len(x), IDSize)
		}
		*id = IDFromBytes(x)
		return nil
	default:
		return fmt.Errorf("cannot scan type %T", x)
	}
}

func (id ID) Value() (driver.Value, error) {
	return id[:], nil
}

// Span is a Span of ID's
type Span = state.Span[ID]
