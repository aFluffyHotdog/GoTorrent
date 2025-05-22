package torrentFile

import (
	"fmt"
	"os"
	"testing"
)

func TestOpen(t *testing.T) {
	data, err := os.Open("panor.torrent")
	if err != nil {
		fmt.Println("failed to open file")
	}

	benCodeObj, err := Open(data)
	if err != nil {
		fmt.Println("failed to convert into bencode object")
	}
	testTorrentFile, err := benCodeObj.ToTorrentFile()
	if err != nil {
		fmt.Println("failed to convert into TorrentFile object")
	}
	fmt.Println(testTorrentFile.Announce)

	//TODO: add proper testing checks

	peerID := [20]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20}
	const port uint16 = 6882

	peers, err := testTorrentFile.requestPeers(peerID, port)

	if err != nil {
		fmt.Println("failed to request peers")
		fmt.Println(err)
	}
	fmt.Println("Peers Length: ", len(peers))
}
