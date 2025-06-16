package peers

import (
	"GoTorrent/torrentFile"
	"fmt"
	"log"
	"time"
)

// MaxBlockSize is the largest number of bytes a request can ask for
const MaxBlockSize = 16384
const maxBackLog = 5

type Torrent struct {
	Peers       []torrentFile.Peer
	PeerID      [20]byte
	InfoHash    [20]byte
	PieceHashes [][20]byte
	PieceLength int
	Length      int
	Name        string
}

type pieceWork struct {
	index  int
	hash   [20]byte
	length int
}

type pieceResult struct {
	index int
	buf   []byte
}

type pieceProgress struct {
	index      int     // keeps track of which piece we are trying to download
	client     *Client // client object that maintains connections
	buf        []byte
	downloaded int // the amount of pieces downloaded
	requested  int // the number of bytes requested??
	backlog    int // the amount of requests sent??
}

// FSM for handling the messages received
func (state *pieceProgress) processMessage() error {
	// Read from peer connection
	msg, err := state.client.Read()
	if err != nil {
		return err
	}

	if msg == nil {
		return nil
	}

	switch msg.ID {
	case MsgUnchoke:
		state.client.Choked = false
	case MsgChoke:
		state.client.Choked = true
	case MsgHave:
		// parse HAVE and set bitfield accordingly
		index, err := ParseHave(msg)
		if err != nil {
			return err
		}
		state.client.Bitfield.SetPiece(index)
	case MsgPiece:
		// parses Piece and copy piece to buf
		ret, err := ParsePiece(msg, state.buf, state.index)
		if err != nil {
			return err
		}
		state.downloaded += ret
		state.backlog--
	}
	return nil
}

func (t *Torrent) CalculatePieceSize(index int) int {
	return 0
}

// TODO: use a separate queue to write pieces into disk
// Big boi function to run when you want to download a torrent
func (t *Torrent) Download() {
	maxPiecesBuffered := 24
	fmt.Println("Starting download for ", t.Name)
	// init each queues
	workQ := make(chan *pieceWork, len(t.PieceHashes))   // downloads each piece
	resultQ := make(chan *pieceResult)                   // checks the hashes
	writeQ := make(chan *pieceResult, maxPiecesBuffered) // writes the pieces to disk
	// Send each piece into the work queue
	for index, hash := range t.PieceHashes {
		length := t.CalculatePieceSize(index)
		workQ <- &pieceWork{index, hash, length}
	}

	// start go routines
	for _, peer := range t.Peers {
		go t.startDownloadWorker(peer, workQ, resultQ)
	}

	go func() {
		for res := range writeQ {
			// TODO: calculate where to write to file, and then write it
			fmt.Println(res.index)
		}

	}()

}

func (t *Torrent) startDownloadWorker(peer torrentFile.Peer, workQ chan *pieceWork, resultQ chan *pieceResult) {

	// create client
	c, err := NewClient(peer, t.PeerID, t.InfoHash)
	// perform handshake
	_, err = c.CompleteHandshake()
	if err != nil {
		fmt.Println("Failed to complete handshake with peer:", peer)
		return
	}

	// Send unchoke and interested
	c.SendUnchoke()
	c.SendInterested()

	// iterate over work queue
	for pieceWork := range workQ {
		if !c.Bitfield.HasPiece(pieceWork.index) { // If peer doesn't have the piece
			workQ <- pieceWork // put it back in the queue
			continue
		}

		// Download the piece
		buf, err := attemptDownloadPiece(c, pieceWork)
		if err != nil {
			log.Println("Exiting", err)
			workQ <- pieceWork // Put piece back on the queue
			return
		}

	}

	// download that piece

	// verify the integrity of downloaded piece

	// send it through the result channel

}

func attemptDownloadPiece(c *Client, pieceWork *pieceWork) ([]byte, error) {
	state := pieceProgress{
		index:  pieceWork.index,
		client: c,
		buf:    make([]byte, pieceWork.length),
	}

	// set 30 second deadline to finish downloading
	c.Conn.SetDeadline(time.Now().Add(30 * time.Second))
	defer c.Conn.SetDeadline(time.Time{}) // Disable the deadline

	for state.downloaded < pieceWork.length {
		// If client is not choked we send it requests
		if !state.client.Choked {
			for state.backlog < maxBackLog && state.requested < pieceWork.length {
				blockSize := MaxBlockSize
				// TODO: Figure out the edge case the OG code had
				err := c.SendRequest(state.index, state.requested, blockSize)
				if err != nil {
					return nil, err
				}
				// TODO: figure out what backlog and requested keeps track of
				state.backlog++
				state.requested += blockSize
			}
		}
		err := state.processMessage()
		if err != nil {
			return nil, err
		}
	}
	return state.buf, nil
}
