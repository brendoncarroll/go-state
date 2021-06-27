package cadata

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
)

const IDSize = 32

// ID identifies a particular piece of data
type ID [IDSize]byte

func IDFromBytes(x []byte) ID {
	id := ID{}
	copy(id[:], x)
	return id
}

func ZeroID() ID { return ID{} }

func (id ID) String() string {
	return base64.RawURLEncoding.EncodeToString(id[:])
}

func (id *ID) UnmarshalB64(data []byte) error {
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
