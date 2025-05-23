package peers

import (
	"fmt"
	"io"
)

type Handshake struct {
	Pstr     string
	InfoHash [20]byte
	PeerID   [20]byte
}

func NewHandshake(InfoHash, peerID [20]byte) *Handshake {
	return &Handshake{
		Pstr:     "BitTorrent protocol",
		InfoHash: InfoHash,
		PeerID:   peerID,
	}
}

// Serialize the handshake message
func (h *Handshake) Encode() []byte {
	buf := make([]byte, len(h.Pstr)+49)
	buf[0] = byte(len(h.Pstr))
	curr := 1
	curr += copy(buf[curr:], h.Pstr)
	curr += copy(buf[curr:], make([]byte, 8)) // 8 reserved bytes, set to 0x0
	curr += copy(buf[curr:], h.InfoHash[:])
	curr += copy(buf[curr:], h.PeerID[:])
	return buf
}

func ReadHandshake(r io.Reader) (*Handshake, error) {
	buf := make([]byte, 68)
	_, err := io.ReadFull(r, buf)
	if err != nil {
		return nil, err
	}

	// Check the pstr length
	pstrLen := int(buf[0])
	if pstrLen != 19 {
		return nil, fmt.Errorf("invalid pstr length: %d", pstrLen)
	}

	// Load pstr
	curr := 1
	pstr := string(buf[curr : curr+pstrLen])
	curr += pstrLen

	if pstr != "BitTorrent protocol" {
		return nil, fmt.Errorf("invalid pstr: %s", pstr)
	}

	// Skip reserved bytes
	curr += 8
	InfoHash := [20]byte{}
	ret := copy(InfoHash[:], buf[curr:curr+20])
	if ret != 20 {
		return nil, fmt.Errorf("invalid InfoHash length: %d", ret)
	}

	// Read PeerID
	curr += 20
	PeerID := [20]byte{}
	ret = copy(PeerID[:], buf[curr:curr+20])
	if ret != 20 {
		return nil, fmt.Errorf("invalid PeerID length: %d", ret)
	}

	return &Handshake{
		Pstr:     pstr,
		InfoHash: InfoHash,
		PeerID:   PeerID,
	}, nil

}
