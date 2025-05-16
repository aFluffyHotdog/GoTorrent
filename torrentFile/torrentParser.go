package torrentFile

import (
	"bytes"
	"crypto/sha1"
	"errors"
	"io"

	"github.com/jackpal/bencode-go"
)

type bencodeInfo struct {
	Pieces      string `bencode:"pieces"`
	PieceLength int    `bencode:"piece length"`
	Length      int    `bencode:"length"`
	Name        string `bencode:"name"`
}

type bencodeTorrent struct {
	Announce string      `bencode:"announce"`
	Info     bencodeInfo `bencode:"info"`
}

// TorrentFile encodes the metadata from a .torrent file
type TorrentFile struct {
	Announce    string //the torrent URL
	InfoHash    [20]byte
	PieceHashes [][20]byte
	PieceLength int
	Length      int
	Name        string
}

// Open parses a torrent file
func Open(r io.Reader) (*bencodeTorrent, error) {
	bto := bencodeTorrent{}
	// we turn serial data into an object (also called unmarshalling)
	err := bencode.Unmarshal(r, &bto)
	if err != nil {
		return nil, err
	}
	return &bto, nil
}

func (i *bencodeInfo) createInfoHash() ([20]byte, error) {
	var infoBuffer bytes.Buffer
	err := bencode.Marshal(&infoBuffer, *i) // turn our object back into serial data
	if err != nil {
		return [20]byte{}, err
	}

	h := sha1.Sum(infoBuffer.Bytes()) // turn the bytes in info into a 20 bit hash
	return h, nil
}

func (i *bencodeInfo) createPiecesHash() ([][20]byte, error) {
	// TODO !!!
	hashLen := 20
	buf := []byte(i.Pieces)

	// check if Pieces contains the correct no. of bytes
	if len(buf)%hashLen != 0 {
		return nil, errors.New("Pieces has a malformed length")
	}

	numHashes := len(buf) / hashLen
	hashes := make([][20]byte, numHashes) // make is used for initializing slices, while new is like malloc
	for i := 0; i < numHashes; i++ {
		copy(hashes[i][:], buf[i*hashLen:(i+1)*hashLen]) // copy each hash slice onto the resulting array
	}
	return hashes, nil
}

func (bto *bencodeTorrent) ToTorrentFile() (TorrentFile, error) {

	// create the info hash
	newInfoHash, err := bto.Info.createInfoHash()
	if err != nil {
		return TorrentFile{}, err
	}

	// create the pieces hash
	piecesHash, err := bto.Info.createPiecesHash()
	if err != nil {
		return TorrentFile{}, err
	}

	t := TorrentFile{
		Announce:    bto.Announce,
		InfoHash:    newInfoHash,
		PieceHashes: piecesHash,
		PieceLength: bto.Info.PieceLength,
		Length:      bto.Info.Length,
		Name:        bto.Info.Name,
	}

	return t, nil
}
