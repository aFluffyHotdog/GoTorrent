package torrentFile

import (
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/jackpal/bencode-go"
)

type Peer struct {
	addr net.IP
	port uint16
}

type bencodeTrackerResp struct {
	Interval int    `bencode:"interval"`
	Peers    []byte `bencode:"peers"`
}

func (t *TorrentFile) buildTrackerURL(peerID [20]byte, port uint16) (string, error) {
	base, err := url.Parse(t.Announce)
	if err != nil {
		return "", err
	}

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

func (t *TorrentFile) requestPeers(peerID [20]byte, port uint16) ([]Peer, error) {
	url, err := t.buildTrackerURL(peerID, port)
	if err != nil {
		return nil, err
	}

	client := &http.Client{Timeout: 15 * time.Second} // send GET request with 15 second timeout
	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	trackerResp := bencodeTrackerResp{}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	fmt.Printf("Raw response body:\n%s\n", string(body))

	err = bencode.Unmarshal(resp.Body, &trackerResp)
	if err != nil {
		return nil, err
	}

	fmt.Printf("Unmarshal check: interval=%d, peers=%d bytes\n", trackerResp.Interval, len(trackerResp.Peers))

	peers := []Peer{}
	// Parse peers, each peer is 6 bytes
	for i := 0; i < len(trackerResp.Peers); i += 6 {
		peers = append(peers, Peer{
			addr: net.IP(trackerResp.Peers[i : i+4]),
			port: binary.BigEndian.Uint16([]byte(trackerResp.Peers[i+4 : i+6])), // conver str -> byte -> uint16
		})
	}

	return peers, nil
}
