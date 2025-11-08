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
	Status  TorrentStatus
	Path    string
	Info    MetaInfo
	Tracker TrackerResponse
	Peers   []PeerConnection

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

type MetaInfo struct {
	Announce  string
	CreatedBy string
	Info      []TorrentInfo
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

func NewMetaInfo(metainfo map[string]any) (MetaInfo, bool) {
	info := metainfo["info"].(map[string]any)

	name, nameOk := info["name"].([]byte)
	length, lengthOk := info["length"].(int)
	piece_length, pieceOk := info["piece length"].(int)
	pieces, piecesOk := info["pieces"].([]byte)

	if !(lengthOk && nameOk && pieceOk && piecesOk) {
		return MetaInfo{}, false
	}

	return MetaInfo{
		Announce:  string(metainfo["announce"].([]byte)),
		CreatedBy: string(metainfo["created by"].([]byte)),
		Info: []TorrentInfo{{
			Length:      length,
			Name:        string(name),
			PieceLength: piece_length,
			Pieces:      pieces,
		}},
	}, true
}

func DiscoverPeers(info MetaInfo) []byte {
	params := url.Values{}

	hashed := sha1.Sum(BencodeStruct(info.Info[0]))
	params.Add("info_hash", string(hashed[:]))
	params.Add("peer_id", "12345678901234567890")
	params.Add("port", "6881")
	params.Add("uploaded", "0")
	params.Add("downloaded", "0")
	params.Add("compact", "1")
	params.Add("left", fmt.Sprintf("%d", info.Info[0].Length))

	requestUrl := info.Announce + "?" + params.Encode()
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
