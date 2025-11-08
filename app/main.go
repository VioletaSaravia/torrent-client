package main

import (
	"log"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatal("Usage: ./app [torrent file path]")
	}

	file := DecodeFile(os.Args[1])
	torrent, ok := DecodeInto[Torrent](file)
	// torrent := Torrent{MetaInfo: info}
	if !ok {
		log.Fatal("Invalid torrent info:", file)
	}

	resp := DiscoverPeers(torrent.Info[0], torrent.Announce)
	if resp == nil {
		log.Fatal("No response received")
	}
	torrent.Tracker = NewTrackerResponse(resp)

	if len(torrent.Tracker.Peers) == 0 {
		log.Fatal("No peers received in response")
	}

	for _, i := range torrent.Tracker.Peers {
		info := torrent.Info[0]
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
		<-i.PieceBuffer
	}
}
