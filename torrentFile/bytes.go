package torrentFile

import (
	"errors"
	"fmt"
)

type Bytes []byte

func (me *Bytes) UnmarshalBencode(b []byte) error {
	*me = append([]byte(nil), b...)
	return nil
}

func (me Bytes) MarshalBencode() ([]byte, error) {
	if len(me) == 0 {
		return nil, errors.New("marshalled Bytes should not be zero-length")
	}
	return me, nil
}

func (me Bytes) GoString() string {
	return fmt.Sprintf("bencode.Bytes(%q)", []byte(me))
}
