package cadata

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"

	"database/sql/driver"

	"go.brendoncarroll.net/state"
)

var _ driver.Value = ID{}

const (
	IDSize = 32
	// Base64Alphabet is used when encoding IDs as base64 strings.
	// It is a URL and filepath safe encoding, which maintains ordering.
	Base64Alphabet = "-0123456789" + "ABCDEFGHIJKLMNOPQRSTUVWXYZ" + "_" + "abcdefghijklmnopqrstuvwxyz"
)

// ID identifies a particular piece of data
type ID [IDSize]byte

func IDFromBytes(x []byte) ID {
	id := ID{}
	copy(id[:], x)
	return id
}

var enc = base64.NewEncoding(Base64Alphabet).WithPadding(base64.NoPadding)

func (id ID) String() string {
	return enc.EncodeToString(id[:])
}

// MarshalBase64 encodes ID using Base64Alphabet
func (id ID) MarshalBase64() ([]byte, error) {
	buf := make([]byte, enc.EncodedLen(len(id)))
	enc.Encode(buf, id[:])
	return buf, nil
}

// UnmarshalBase64 decodes data into the ID using Base64Alphabet
func (id *ID) UnmarshalBase64(data []byte) error {
	n, err := enc.Decode(id[:], data)
	if err != nil {
		return err
	}
	if n != IDSize {
		return errors.New("base64 string is too short")
	}
	return nil
}

func (a ID) Equals(b ID) bool {
	return a.Compare(b) == 0
}

func (a ID) Compare(b ID) int {
	return bytes.Compare(a[:], b[:])
}

func (id ID) IsZero() bool {
	return id == (ID{})
}

func (id ID) MarshalJSON() ([]byte, error) {
	s := enc.EncodeToString(id[:])
	return json.Marshal(s)
}

func (id *ID) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	_, err := enc.Decode(id[:], []byte(s))
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

// Successor returns the ID immediately after this ID
func (id ID) Successor() ID {
	for i := len(id) - 1; i >= 0; i-- {
		id[i]++
		if id[i] != 0 {
			break
		}
	}
	return id
}

// Span is a Span of ID's
type Span = state.Span[ID]

// BeginFromSpan returns the ID which begins the span.
// It will be lteq every other ID in the Span.
func BeginFromSpan(x Span) ID {
	lower, ok := x.LowerBound()
	if !ok {
		return ID{}
	}
	if x.IncludesLower() {
		return lower
	}
	return lower.Successor()
}

// EndFromSpan returns the cadata.ID which is greater than ever ID in the span, and true.
// Or it returns the zero ID and false if no such ID exists.
func EndFromSpan(x Span) (ID, bool) {
	upper, ok := x.UpperBound()
	if !ok {
		return ID{}, false
	}
	if !x.IncludesUpper() {
		return upper, true
	}
	suc := upper.Successor()
	return suc, suc.IsZero()
}
