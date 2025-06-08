package torrentFile

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/anacrolix/torrent/bencode"
)

type Peer struct {
	addr net.IP
	port uint16
}

// format peer as an address:port string
func (p Peer) String() string {
	return fmt.Sprintf("%s:%d", p.addr.String(), p.port)
}

type bencodeTrackerResp struct {
	Interval int    `bencode:"interval"`
	Peers    string `bencode:"peers"`
}

// With the torrent informationi we have, generate a valid tracker URL
func (t *TorrentFile) buildTrackerURL(peerID [20]byte, port uint16) (string, error) {
	base, err := url.Parse(t.Announce)
	if err != nil {
		return "", err
	}

	fmt.Printf("converted hash: %v \n", hex.EncodeToString(t.InfoHash[:]))
	params := base.Query()
	params.Set("info_hash", string(t.InfoHash[:]))
	params.Set("peer_id", string(peerID[:]))
	params.Set("port", strconv.Itoa(int(port)))
	params.Set("uploaded", "0")
	params.Set("downloaded", "0")
	params.Set("left", strconv.Itoa(t.Length))
	params.Set("compact", "1")

	base.RawQuery = params.Encode()
	return base.String(), nil
}

// Sends a GET request to the generated tracker URL, returns timeout + a list of peers
func (t *TorrentFile) RequestPeers(peerID [20]byte, port uint16) ([]Peer, error) {
	url, err := t.buildTrackerURL(peerID, port)
	if err != nil {
		return nil, err
	}

	// Create a new HTTP client with a 15 second timeout
	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		fmt.Println("GET request failed")
		fmt.Println(err.Error())
		return nil, err
	}

	defer resp.Body.Close()

	// Decode the response body into a bencodeTrackerResp struct
	trackerResp := bencodeTrackerResp{}
	d := bencode.NewDecoder(resp.Body)
	err = d.Decode(&trackerResp)

	if err != nil {
		fmt.Println("failed to unmarshal response from tracker")
		return nil, err
	}

	fmt.Printf("Unmarshal check: interval= %d, peers= %d bytes\n", trackerResp.Interval, len(trackerResp.Peers))

	peers := []Peer{}
	// Parse peers, each peer is 6 bytes
	for i := 0; i < len(trackerResp.Peers); i += 6 {
		peers = append(peers, Peer{
			// first 4 is the IP Address
			addr: net.IP(trackerResp.Peers[i : i+4]),
			// Last 2 is the port
			port: binary.BigEndian.Uint16([]byte(trackerResp.Peers[i+4 : i+6])), // conver str -> byte -> uint16
		})
	}

	return peers, nil
}
