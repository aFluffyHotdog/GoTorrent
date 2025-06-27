package main

import (
	p2p "GoTorrent/peers"
	"GoTorrent/torrentFile"
	"crypto/rand"
	"fmt"
	"os"
)

func GeneratePeerID() [20]byte {
	var peerID [20]byte
	copy(peerID[:], []byte("-GT0001-"))
	_, err := rand.Read(peerID[8:])
	if err != nil {
		panic(err)
	}
	return peerID
}

func main() {
	data, err := os.Open("e2e/debian-12.11.0-amd64-netinst.iso.torrent")
	if err != nil {
		fmt.Println("failed to open file")
	}
	temp, err := torrentFile.Open(data)
	if err != nil {
		fmt.Println(err)
		return
	}

	t, err := temp.ToTorrentFile()
	if err != nil {
		fmt.Println(err)
		return
	}

	const Port uint16 = 6881
	peerID := GeneratePeerID()

	peers, err := t.RequestPeers(peerID, Port)
	if err != nil {
		fmt.Println(err)
		return
	}
	torrent := p2p.Torrent{
		Peers:       peers,
		PeerID:      peerID,
		InfoHash:    t.InfoHash,
		PieceHashes: t.PieceHashes,
		PieceLength: t.PieceLength,
		Length:      t.Length,
		Name:        t.Name,
	}
	_, err = torrent.Download()

}
