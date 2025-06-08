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
	// create a new connection to peer with 5 second time out
	conn, err := net.DialTimeout("tcp", peer.String(), 5*time.Second)
	if err != nil {
		return nil, err
	}

	Client := &Client{
		Conn:     conn,
		Choked:   true,
		peer:     peer,
		infoHash: infoHash,
		peerID:   peerID,
	}
	return Client, nil
}

func (c *Client) CompleteHandshake() (Handshake, error) {
	// send handshake to peer
	c.Conn.SetDeadline(time.Now().Add(3 * time.Second))
	defer c.Conn.SetDeadline(time.Time{}) // Disable the deadline

	handShake := NewHandshake(c.infoHash, c.peerID)
	_, err := c.Conn.Write(handShake.Encode())
	if err != nil {
		return Handshake{}, err
	}
	// read response from peer
	buf := make([]byte, 68) // buffer for handshake response
	n, err := c.Conn.Read(buf)
	if n != 68 {
		fmt.Println("error malformed handshake response")
		return Handshake{}, err
	}
	recvHandshake, err := ReadHandshake(bytes.NewReader(buf))
	// compare infoHashes
	if !bytes.Equal(recvHandshake.InfoHash[:], c.infoHash[:]) {
		return Handshake{}, err
	}
	return *recvHandshake, nil
}

// TODO: Write function to receive bitfield

// // SendRequest sends a Request message to the peer
// func (c *Client) SendRequest(index, begin, length int) error {
// 	req := message.FormatRequest(index, begin, length)
// 	_, err := c.Conn.Write(req.Serialize())
// 	return err
// }

// // SendInterested sends an Interested message to the peer
// func (c *Client) SendInterested() error {
// 	msg := message.Message{ID: message.MsgInterested}
// 	_, err := c.Conn.Write(msg.Serialize())
// 	return err
// }

// // SendNotInterested sends a NotInterested message to the peer
// func (c *Client) SendNotInterested() error {
// 	msg := message.Message{ID: message.MsgNotInterested}
// 	_, err := c.Conn.Write(msg.Serialize())
// 	return err
// }

// // SendUnchoke sends an Unchoke message to the peer
// func (c *Client) SendUnchoke() error {
// 	msg := message.Message{ID: message.MsgUnchoke}
// 	_, err := c.Conn.Write(msg.Serialize())
// 	return err
// }

// // SendHave sends a Have message to the peer
// func (c *Client) SendHave(index int) error {
// 	msg := message.FormatHave(index)
// 	_, err := c.Conn.Write(msg.Serialize())
// 	return err
// }
