package main

import (
	"log"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatal("Usage: ./app [torrent file path]")
	}

	torrent := Torrent{
		Path: os.Args[1],
	}

	file := DecodeFile(torrent.Path)
	var ok bool
	torrent.Info, ok = NewMetaInfo(file)
	if !ok {
		log.Fatal("Invalid torrent info:", file)
	}

	resp := DiscoverPeers(torrent.Info)
	if resp == nil {
		log.Fatal("No response received")
	}
	torrent.Tracker = NewTrackerResponse(resp)

	if len(torrent.Tracker.Peers) == 0 {
		log.Fatal("No peers received in response")
	}

	for _, i := range torrent.Tracker.Peers {
		info := torrent.Info.Info[0]
		conn, err := NewPeerConnection(i, info)
		if err != nil {
			return
		}
		torrent.Peers = append(torrent.Peers, conn)
		defer conn.Close()
	}

	for n, i := range torrent.Peers {
		go i.DownloadPiece(n)
	}

	for _, i := range torrent.Peers {
		piece := <-i.PieceBuffer
	}
}
