package peers

import (
	"GoTorrent/torrentFile"
	"fmt"
)

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

func (t *Torrent) CalculatePieceSize(index int) int {
	return 0
}

// TODO: use a separate queue to write pieces into disk
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
	//c, err := NewClient(peer, t.PeerID, t.InfoHash)
	// perform handshake
	//_, err := c.CompleteHandshake()
	//send unchoke and interested

	// iterate over work queue

	// check peer's bitfield
	// If peer doesn't have that piece, put it back in q

	// download that piece

	// verify the integrity of downloaded piece

	// send it through the result channel

}
