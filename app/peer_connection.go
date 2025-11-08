package main

import (
	"fmt"
	"log"
	"net"
	"time"
)

const BLOCK_SIZE uint32 = 16 * 1024

type PeerConnection struct {
	net.Conn
	Status      PeerStatus
	Address     string
	PieceBuffer chan []byte
	MsgBuffer   []byte
	Torrent     TorrentInfo
}

type PeerStatus uint8

const (
	PeerIdle PeerStatus = iota
	PeerActive
	PeerDone
	PeerDisconnected
)

func NewPeerConnection(address string, info TorrentInfo) (PeerConnection, error) {
	conn, err := net.DialTimeout("tcp", address, 5*time.Second)
	if err != nil {
		return PeerConnection{}, err
	}

	return PeerConnection{
			Conn:        conn,
			Status:      PeerIdle,
			Address:     address,
			PieceBuffer: make(chan []byte, info.PieceLength), // No?
			MsgBuffer:   make([]byte, 5+BLOCK_SIZE),
			Torrent:     info},
		nil
}

func (conn *PeerConnection) DownloadPiece(index int) {
	handshakeMsg := NewHandshakeMsg(conn.Torrent)
	conn.Write(handshakeMsg.ToBytes())

	msg := conn.ReadPeerMsg()
	if _, ok := msg.(BitfieldMsg); !ok {
		fmt.Println("Expected bitfield message")
		return
	}

	conn.Write(ToBytes(InterestedMsg{}))
	msg = conn.ReadPeerMsg()
	if _, ok := msg.(UnchokeMsg); !ok {
		fmt.Println("Expected unchoke message")
		return
	}

	block, ok := conn.DownloadBlock(uint32(index), 0)
	if !ok {
		fmt.Println("Expected piece message")
		return
	}
	fmt.Printf("Received block of %d bytes: %x\n",
		len(block),
		block,
	)
}

func (conn *PeerConnection) ReadPeerMsg() PeerMessage {
	if n, err := conn.Read(conn.MsgBuffer); err != nil {
		log.Fatal("Error reading peer message:", err)
	} else if n < 5 {
		log.Fatal("Invalid peer message")
	} else {
		received := conn.MsgBuffer[:n]
		return FromBytes(received)
	}
	return nil
}

func (conn *PeerConnection) DownloadBlock(index uint32, offset uint32) (result [BLOCK_SIZE]byte, ok bool) {
	var cur uint32 = 0

	requestMsg := RequestMsg{Index: index, Begin: offset, Length: BLOCK_SIZE}
	conn.Write(ToBytes(requestMsg))

	for cur != BLOCK_SIZE {
		response := conn.ReadPeerMsg()
		pieceMsg, okMsg := response.(PieceMsg)
		if !okMsg {
			return
		}
		copy(result[cur:], pieceMsg.Block)
		cur += uint32(len(pieceMsg.Block))
	}

	ok = true
	return
}
