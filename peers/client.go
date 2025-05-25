package peers

import (
	"GoTorrent/torrentFile"
	"bytes"
	"fmt"
	"net"
	"time"
)

// A Client is a TCP connection with a peer
type Client struct {
	Conn     net.Conn
	Choked   bool
	Bitfield Bitfield
	peer     torrentFile.Peer
	infoHash [20]byte
	peerID   [20]byte
}

// Completes a handshake with a peer, returns a new client
func NewClient(peer torrentFile.Peer, peerID, infoHash [20]byte) (*Client, error) {
	// create a new connection to peer with 3 second time out
	conn, err := net.DialTimeout("tcp", peer.String(), 3*time.Second)
	if err != nil {
		return nil, err
	}

	Client := &Client{
		Conn:     conn,
		Choked:   false,
		peer:     peer,
		infoHash: infoHash,
		peerID:   peerID,
	}
	return Client, nil
}

func CompleteHandshake(c *Client) (Handshake, error) {
	// send handshake to peer
	handShake := NewHandshake(c.infoHash, c.peerID)
	_, err := c.Conn.Write(handShake.Encode())
	if err != nil {
		return Handshake{}, err
	}
	// read response from peer
	buf := make([]byte, 68) // buffer for handshake response
	n, err := c.Conn.Read(buf)
	if n != 68 {
		fmt.Println("erroralformed handshake response")
		return Handshake{}, err
	}
	recvHandshake, err := ReadHandshake(bytes.NewReader(buf))
	// compare infoHashes
	if !bytes.Equal(recvHandshake.InfoHash[:], c.infoHash[:]) {
		return Handshake{}, err
	}
	return *recvHandshake, nil
}
