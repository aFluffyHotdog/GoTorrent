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

	// perform handshake
	_, err = Client.CompleteHandshake()
	if err != nil {
		err = fmt.Errorf("failed to complete handshake with peer: %s", peer)
		return nil, err
	}

	bitfield, err := Client.receiveBitField()
	if err != nil {
		return nil, err
	}
	Client.Bitfield = bitfield

	fmt.Println("Completed handshake with peer: ", peer)
	return Client, nil
}

func (c *Client) CompleteHandshake() (Handshake, error) {
	// send handshake to peer
	c.Conn.SetDeadline(time.Now().Add(3 * time.Second))
	defer c.Conn.SetDeadline(time.Time{}) // Disable the deadline

	handShake := NewHandshake(c.infoHash, c.peerID)
	// fmt.Println("Handshake Pstr: ", handShake.Pstr)
	// fmt.Println("Handshake infohash: ", hex.EncodeToString(handShake.InfoHash[:]))
	// fmt.Println("Handshake PeerID: ", string(handShake.PeerID[:]))
	_, err := c.Conn.Write(handShake.Encode())
	if err != nil {
		return Handshake{}, err
	}

	// set a read deadline before reading handshake response
	c.Conn.SetReadDeadline(time.Now().Add(3 * time.Second))
	defer c.Conn.SetReadDeadline(time.Time{}) // Disable the read deadline

	// read response from peer
	buf := make([]byte, 68) // buffer for handshake response
	n, err := c.Conn.Read(buf)
	if n != 68 {
		// fmt.Println("malformed handshake response. buffer len: ", n)
		s := string(buf)
		fmt.Println(s)
		return Handshake{}, err
	}
	recvHandshake, err := ReadHandshake(bytes.NewReader(buf))
	// compare infoHashes
	if !bytes.Equal(recvHandshake.InfoHash[:], c.infoHash[:]) {
		return Handshake{}, err
	}
	return *recvHandshake, nil
}

func (c *Client) receiveBitField() (Bitfield, error) {
	c.Conn.SetDeadline(time.Now().Add(5 * time.Second))
	defer c.Conn.SetDeadline(time.Time{})

	msg, err := Read(c.Conn)

	if err != nil {
		return nil, err
	}

	if msg == nil {
		err := fmt.Errorf("no message received")
		return nil, err
	}
	if msg.ID != MsgBitfield {
		err := fmt.Errorf("expected bitfield but got type: %d", msg.ID)
		return nil, err
	}

	return msg.Payload, nil

}

// reads a message from the client connection
func (c *Client) Read() (*Message, error) {
	msg, err := Read(c.Conn)
	return msg, err
}

// SendRequest sends a Request message to the peer
func (c *Client) SendRequest(index, begin, length int) error {
	req := FormatRequest(index, begin, length)
	_, err := c.Conn.Write(req.SerializeMessage())
	return err
}

// SendInterested sends an Interested message to the peer
func (c *Client) SendInterested() error {
	msg := Message{ID: MsgInterested}
	_, err := c.Conn.Write(msg.SerializeMessage())
	return err
}

// SendNotInterested sends a NotInterested message to the peer
func (c *Client) SendNotInterested() error {
	msg := Message{ID: MsgNotInterested}
	_, err := c.Conn.Write(msg.SerializeMessage())
	return err
}

// SendUnchoke sends an Unchoke message to the peer
func (c *Client) SendUnchoke() error {
	msg := Message{ID: MsgUnchoke}
	_, err := c.Conn.Write(msg.SerializeMessage())
	return err
}

// SendHave sends a Have message to the peer
func (c *Client) SendHave(index int) error {
	msg := FormatHave(index)
	_, err := c.Conn.Write(msg.SerializeMessage())
	return err
}
