package main

import (
	"crypto/sha1"
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
)

type Torrent struct {
	Announce  string
	CreatedBy string
	Info      []TorrentInfo
	Status    TorrentStatus
	Path      string
	Tracker   TrackerResponse
	Peers     []PeerConnection

	// Data
	File   []byte
	Pieces [][]byte
}

type TorrentStatus uint8

const (
	TorrentIdle TorrentStatus = iota
	TorrentActive
	TorrentConnecting
	TorrentFinished
)

type TorrentInfo struct {
	Length      int
	Name        string
	PieceLength int
	Pieces      []byte
}

type TrackerResponse struct {
	Interval int
	Peers    []string
}

func NewTrackerResponse(bencoded []byte) TrackerResponse {
	result := TrackerResponse{}

	peerInfo, err := Decode(bencoded)
	if err != nil {
		log.Fatal("Invalid data received from server:", err)
	}
	peerDict, ok := peerInfo.(map[string]any)
	if !ok {
		log.Fatal("Invalid data received from server")
	}
	intervalAny, ok := peerDict["interval"]
	if !ok {
		log.Fatal("List of peers not received")
	}
	interval, ok := intervalAny.(int)
	if !ok {
		log.Fatal("List of peers not received")
	}
	result.Interval = interval
	peersAny, ok := peerDict["peers"]
	if !ok {
		log.Fatal("List of peers not received")
	}
	peers, ok := peersAny.([]byte)
	if !ok {
		log.Fatal("List of peers not received")
	}

	for i := 0; i+6 <= len(peers); i += 6 {
		newPeer := ""
		for n, i := range peers[i : i+4] {
			newPeer += fmt.Sprintf("%d", i)
			if n != 3 {
				newPeer += "."
			}
		}
		newPeer += ":"
		newPeer += fmt.Sprintf("%d", binary.BigEndian.Uint16(peers[i+4:i+6]))
		result.Peers = append(result.Peers, newPeer)
	}

	return result
}

func DiscoverPeers(info TorrentInfo, path string) []byte {
	params := url.Values{}

	hashed := sha1.Sum(BencodeStruct(info))
	params.Add("info_hash", string(hashed[:]))
	params.Add("peer_id", "12345678901234567890")
	params.Add("port", "6881")
	params.Add("uploaded", "0")
	params.Add("downloaded", "0")
	params.Add("compact", "1")
	params.Add("left", fmt.Sprintf("%d", info.Length))

	requestUrl := path + "?" + params.Encode()
	resp, err := http.Get(requestUrl)
	if err != nil {
		log.Fatal(err)
	}

	body, err := io.ReadAll(resp.Body)
	if err != io.EOF {
		log.Fatal(err)
	}
	return body
}
