package peers

import (
	"GoTorrent/torrentFile"
	"net"
	"time"
)

// A Client is a TCP connection with a peer
type Client struct {
	Conn   net.Conn
	Choked bool
	//Bitfield bitfield.Bitfield
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

	return nil, nil
}

func CompleteHandshake(c *Client) (Handshake, error) {
	// send handshake to peer
	// read response from peer
	// check if response is valid
	// return error if not valid
	return nil
}
