package torrentFile

import (
	"crypto/sha1"
	"errors"
	"fmt"
	"io"

	"github.com/anacrolix/torrent/bencode"
)

type bencodeTorrent struct {
	InfoBytes Bytes `bencode:"info"` // BEP 3
	AlsoInfo  torrentInfo
	Announce  string `bencode:"announce"` // BEP 3
}

type torrentInfo struct {
	Pieces      string `bencode:"pieces"`
	PieceLength int    `bencode:"piece length"`
	Length      int    `bencode:"length"`
	Name        string `bencode:"name"`
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
	bto := &bencodeTorrent{}
	d := bencode.NewDecoder(r)
	err := d.Decode(&bto)
	if err != nil {
		fmt.Println("failed to unmarshal bencode")
		fmt.Println(err)
		return nil, err
	}

	var tempInfo torrentInfo
	err = bencode.Unmarshal(bto.InfoBytes, &tempInfo)
	if err != nil {
		fmt.Println("failed to unmarshal torrentinfo")
		fmt.Println(err)
		return nil, err
	}
	bto.AlsoInfo = tempInfo
	return bto, nil
}

func (i *bencodeTorrent) createInfoHash() ([20]byte, error) {
	h := sha1.Sum(i.InfoBytes) // turn the bytes in info into a 20 bit hash
	return h, nil
}

func (i *bencodeTorrent) createPiecesHash() ([][20]byte, error) {
	hashLen := 20
	buf := []byte(i.AlsoInfo.Pieces)

	// check if Pieces contains the correct no. of bytes
	if len(buf)%hashLen != 0 {
		return nil, errors.New("pieces has a malformed length")
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
	newInfoHash, err := bto.createInfoHash()
	if err != nil {
		return TorrentFile{}, err
	}

	// create the pieces hash
	piecesHash, err := bto.createPiecesHash()
	if err != nil {
		return TorrentFile{}, err
	}

	t := TorrentFile{
		Announce:    bto.Announce,
		InfoHash:    newInfoHash,
		PieceHashes: piecesHash,
		PieceLength: bto.AlsoInfo.PieceLength,
		Length:      bto.AlsoInfo.Length,
		Name:        bto.AlsoInfo.Name,
	}

	return t, nil
}
