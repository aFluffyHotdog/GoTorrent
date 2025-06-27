package peers

import (
	"GoTorrent/torrentFile"
	"bytes"
	"crypto/sha1"
	"fmt"
	"log"
	"runtime"
	"sync"
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

// calculate the index where the piece would begin and end in the entire file
func (t *Torrent) CalculateBoundsForPiece(index int) (begin int, end int) {
	begin = index * t.PieceLength
	end = begin + t.PieceLength
	if end > t.Length {
		end = t.Length
	}

	return begin, end
}

func (t *Torrent) CalculatePieceSize(index int) int {
	begin, end := t.CalculateBoundsForPiece(index)
	return end - begin
}

func (t *Torrent) startDownloadWorker(peer torrentFile.Peer, workQ chan *pieceWork, resultQ chan *pieceResult) {
	// create client
	c, err := NewClient(peer, t.PeerID, t.InfoHash)
	if err != nil {
		fmt.Println("Failed to create new client")
		fmt.Println(err)
		return
	}
	// perform handshake
	fmt.Println("Completeing handshake with peer: ", peer)
	_, err = c.CompleteHandshake()
	if err != nil {
		fmt.Println("Failed to complete handshake with peer:", peer)
		return
	}
	fmt.Println("Completed handshake with peer: ", peer)

	// Send unchoke and interested
	err = c.SendUnchoke()
	if err != nil {
		fmt.Println("sending unchoke failed")
		return
	}
	err = c.SendInterested()
	if err != nil {
		fmt.Println("sending interested failed")
		return
	}

	fmt.Println("finished sending unchoke and interested")
	// iterate over work queue
	for pieceWork := range workQ {
		if !c.Bitfield.HasPiece(pieceWork.index) { // If peer doesn't have the piece
			workQ <- pieceWork // put it back in the queue
			log.Println("client doesn't have piece ", pieceWork.index)
			continue
		}

		// Download the piece
		buf, err := attemptDownloadPiece(c, pieceWork)
		if err != nil {
			log.Println("Exiting", err)
			workQ <- pieceWork // Put piece back on the queue
			return
		}

		// verify the integrity of the piece
		err = checkIntegrity(pieceWork, buf)
		if err != nil {
			log.Printf("malformed message: %v", err)
			workQ <- pieceWork // Put piece back on the queue
			return
		}

		// tell our peers we now have this piece
		c.SendHave(pieceWork.index)
		// send the finished piece into the results queue
		resultQ <- &pieceResult{
			index: pieceWork.index,
			buf:   buf,
		}

	}

}

func checkIntegrity(pw *pieceWork, buf []byte) error {
	obtained_hash := sha1.Sum(buf)

	if !bytes.Equal(obtained_hash[:], pw.hash[:]) {
		return fmt.Errorf("index %d failed integrity check", pw.index)
	}

	return nil
}

func attemptDownloadPiece(c *Client, pieceWork *pieceWork) ([]byte, error) {
	state := pieceProgress{
		index:  pieceWork.index,
		client: c,
		buf:    make([]byte, pieceWork.length),
	}

	fmt.Println("Attempting to download piece no. ", state.index)
	// set 30 second deadline to finish downloading
	c.Conn.SetDeadline(time.Now().Add(30 * time.Second))
	defer c.Conn.SetDeadline(time.Time{}) // Disable the deadline

	for state.downloaded < pieceWork.length {
		// If client is not choked we send it requests
		if !state.client.Choked {
			for state.backlog < maxBackLog && state.requested < pieceWork.length {
				blockSize := MaxBlockSize

				// Last block might be shorter than the typical block
				if pieceWork.length-state.requested < blockSize {
					blockSize = pieceWork.length - state.requested
				}
				err := c.SendRequest(state.index, state.requested, blockSize)
				if err != nil {
					return nil, err
				}

				state.backlog++              // keeps track of how many requests we've sent
				state.requested += blockSize // keeps track of how many bytes we've requested
			}
		}
		err := state.processMessage()
		if err != nil {
			return nil, err
		}
	}

	fmt.Print("Piece no ", state.index, "downloaded!")
	return state.buf, nil
}

// TODO: use a separate queue to write pieces into disk
// Big boi function to run when you want to download a torrent
func (t *Torrent) Download() ([]byte, error) {
	// maxPiecesBuffered := 24
	fmt.Println("Starting download for ", t.Name)
	// init each queues
	workQ := make(chan *pieceWork, len(t.PieceHashes)) // downloads each piece
	resultQ := make(chan *pieceResult)                 // checks the hashes
	var wg sync.WaitGroup
	// writeQ := make(chan *pieceResult, maxPiecesBuffered) // writes the pieces to disk
	// Send each piece into the work queue
	for index, hash := range t.PieceHashes {
		length := t.CalculatePieceSize(index)
		workQ <- &pieceWork{index, hash, length}
	}
	fmt.Printf("workQ: %d/%d\n", len(workQ), cap(workQ))

	// start downlaod go routines
	for _, peer := range t.Peers {
		wg.Add(1)
		go func(peer torrentFile.Peer) {
			defer wg.Done()
			t.startDownloadWorker(peer, workQ, resultQ)
		}(peer)
	}

	// go func() {
	// 	for res := range writeQ {
	// 		// TODO: calculate where to write to file, and then write it
	// 		fmt.Println(res.index)
	// 	}

	// }()

	// naive implementation for now
	buf := make([]byte, t.Length)
	donePieces := 0
	for donePieces < len(t.PieceHashes) {

		// take piece from results queue then write it to buffer
		res := <-resultQ
		begin, end := t.CalculateBoundsForPiece(res.index)
		copy(buf[begin:end], res.buf)
		donePieces++

		// print out the progress
		percent := float64(donePieces) / float64(len(t.PieceHashes)) * 100
		numWorkers := runtime.NumGoroutine() - 1 // subtract 1 for main thread
		log.Printf("(%0.2f%%) Downloaded piece #%d from %d peers\n", percent, res.index, numWorkers)
	}

	close(workQ)
	wg.Wait() // Wait for all workers to finish

	return buf, nil
}
