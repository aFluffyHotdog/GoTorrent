package e2e_test

import (
	"GoTorrent/peers"
	"GoTorrent/torrentFile"
	"crypto/rand"
	"fmt"
	"os"
	"testing"
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
func TestOpen(t *testing.T) {
	data, err := os.Open("debian-12.11.0-amd64-netinst.iso.torrent")
	if err != nil {
		fmt.Println("failed to open file")
	}

	// Test creating a torrentFile object
	benCodeObj, err := torrentFile.Open(data)
	if err != nil {
		fmt.Println("failed to convert into bencode obect")
	}
	testTorrentFile, err := benCodeObj.ToTorrentFile()
	if err != nil {
		fmt.Println("failed to convert into TorrentFile object")
	}
	fmt.Println(testTorrentFile.Announce)
	fmt.Println(testTorrentFile.Name)
	// Test requesting peers
	peerID := GeneratePeerID()
	const port uint16 = 6882

	peers_list, err := testTorrentFile.RequestPeers(peerID, port)

	if err != nil {
		fmt.Println(" Failed to request peers")
		fmt.Println(err)
	}
	fmt.Println("Peers Length: ", len(peers_list))

	clients := make([]peers.Client, len(peers_list))
	for i := 0; i < len(peers_list); i++ {
		tempClient, err := peers.NewClient(peers_list[i], peerID, testTorrentFile.InfoHash)
		if err != nil {
			fmt.Println("failed to create client for peer", peers_list[i])
			fmt.Println(err)
			continue
		}
		clients[i] = *tempClient
		handshake, err := tempClient.CompleteHandshake()
		if err != nil {
			fmt.Println("failed to complete handshake with peer", peers_list[i])
			fmt.Println(err)
			continue
		}
		fmt.Println("Handshake completed with peer:", peers_list[i])
		fmt.Println("Peer ID:", string(handshake.PeerID[:]))
	}
}
