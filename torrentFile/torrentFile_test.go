package torrentFile

import (
	"fmt"
	"os"
	"testing"
)

func TestOpen(t *testing.T) {
	data, err := os.Open("test.torrent")
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
	fmt.Println(testTorrentFile.Name)

	//TODO: add proper testing checks

	peerID := [20]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20}
	const port uint16 = 6882

	trackerURL, err := testTorrentFile.buildTrackerURL(peerID, port)
	if err != nil {
		fmt.Println("failed to build tracker URL")
	}
	fmt.Println(trackerURL)

	peers, err := testTorrentFile.requestPeers(peerID, port)

	if err != nil {
		fmt.Println("failed to request peers")
		t.Fail()
	}
	fmt.Println("peers:")
	fmt.Println(peers)
}
