package main

import (
	"crypto/rand"
	"crypto/sha1"
	"encoding/binary"
)

type PeerMessageCode uint8

const (
	MsgChoke PeerMessageCode = iota
	MsgUnchoke
	MsgInterested
	MsgNotInterested
	MsgHave
	MsgBitfield
	MsgRequest
	MsgPiece
	MsgCancel
	MsgPort
	MsgHandshake = 255 // Doesn't actually have an ID.
)

type HandshakeMsg struct {
	InfoHash []byte
	PeerId   []byte
}

func NewHandshakeMsg(info TorrentInfo) HandshakeMsg {
	result := HandshakeMsg{
		InfoHash: make([]byte, 20),
		PeerId:   make([]byte, 20),
	}

	hash := sha1.Sum(BencodeStruct(info))
	copy(result.InfoHash, hash[:])

	peerId := [20]byte{}
	rand.Read(peerId[:])
	copy(result.PeerId, peerId[:])

	return result
}

func (m HandshakeMsg) FromBytes(b []byte) HandshakeMsg {
	return HandshakeMsg{
		InfoHash: b[28:48],
		PeerId:   b[48:68],
	}
}

func (m HandshakeMsg) ToBytes() []byte {
	result := make([]byte, 68)
	result[0] = byte(19)
	copy(result[1:20], "BitTorrent protocol")
	copy(result[28:48], m.InfoHash)
	copy(result[48:68], m.PeerId)
	return result
}

type ChokeMsg struct{}

type UnchokeMsg struct{}

type InterestedMsg struct{}

type NotInterestedMsg struct{}

type HaveMsg struct {
	PieceIndex int32
}

type BitfieldMsg struct {
	Bitfield []byte
}

type RequestMsg struct {
	Index  uint32
	Begin  uint32
	Length uint32
}

type PieceMsg struct {
	Index uint32
	Begin uint32
	Block []byte
}

type CancelMsg struct {
	Index  uint32
	Begin  uint32
	Length uint32
}

type PortMsg struct {
	ListenPort uint16
}

type PeerMessage interface {
	__isPeerMessage()
}

func (HandshakeMsg) __isPeerMessage()     {}
func (ChokeMsg) __isPeerMessage()         {}
func (UnchokeMsg) __isPeerMessage()       {}
func (InterestedMsg) __isPeerMessage()    {}
func (NotInterestedMsg) __isPeerMessage() {}
func (HaveMsg) __isPeerMessage()          {}
func (BitfieldMsg) __isPeerMessage()      {}
func (RequestMsg) __isPeerMessage()       {}
func (PieceMsg) __isPeerMessage()         {}
func (CancelMsg) __isPeerMessage()        {}
func (PortMsg) __isPeerMessage()          {}

func ToBytes(msg PeerMessage) []byte {
	switch m := msg.(type) {
	case ChokeMsg:
		result := make([]byte, 5)
		binary.BigEndian.PutUint32(result[:4], 1)
		result[4] = byte(MsgChoke)
		return result

	case UnchokeMsg:
		result := make([]byte, 5)
		binary.BigEndian.PutUint32(result[:4], 1)
		result[4] = byte(MsgUnchoke)
		return result

	case InterestedMsg:
		result := make([]byte, 5)
		binary.BigEndian.PutUint32(result[:4], 1)
		result[4] = byte(MsgInterested)
		return result

	case NotInterestedMsg:
		result := make([]byte, 5)
		binary.BigEndian.PutUint32(result[:4], 1)
		result[4] = byte(MsgNotInterested)
		return result

	case HaveMsg:
		result := make([]byte, 5+4)
		binary.BigEndian.PutUint32(result[:4], 1)
		result[4] = byte(MsgHave)
		binary.BigEndian.PutUint32(result[5:], uint32(m.PieceIndex))
		return result

	case BitfieldMsg:
		result := make([]byte, 5+len(m.Bitfield))
		binary.BigEndian.PutUint32(result[:4], 1)
		result[4] = byte(MsgBitfield)
		copy(result[5:], m.Bitfield)
		return result

	case RequestMsg:
		result := make([]byte, 5+12)
		binary.BigEndian.PutUint32(result[:4], 1)
		result[4] = byte(MsgRequest)
		binary.BigEndian.PutUint32(result[5:9], m.Index)
		binary.BigEndian.PutUint32(result[9:13], m.Begin)
		binary.BigEndian.PutUint32(result[13:17], m.Length)
		return result

	case PieceMsg:
		result := make([]byte, 5+4+4+len(m.Block))
		binary.BigEndian.PutUint32(result[:4], 1)
		result[4] = byte(MsgPiece)
		binary.BigEndian.PutUint32(result[5:9], m.Index)
		binary.BigEndian.PutUint32(result[9:13], m.Begin)
		copy(result[13:], m.Block)
		return result

	case CancelMsg:
		result := make([]byte, 5+12)
		binary.BigEndian.PutUint32(result[:4], 1)
		result[4] = byte(MsgCancel)
		binary.BigEndian.PutUint32(result[5:9], m.Index)
		binary.BigEndian.PutUint32(result[9:13], m.Begin)
		binary.BigEndian.PutUint32(result[13:17], m.Length)
		return result

	case PortMsg:
		result := make([]byte, 5+2)
		binary.BigEndian.PutUint32(result[:4], 1)
		result[4] = byte(MsgPort)
		binary.BigEndian.PutUint16(result[5:7], m.ListenPort)
		return result
	}

	return nil
}

func FromBytes(b []byte) PeerMessage {
	if len(b) < 5 {
		return nil
	}

	switch PeerMessageCode(b[4]) {
	case MsgChoke:
		return ChokeMsg{}

	case MsgUnchoke:
		return UnchokeMsg{}

	case MsgInterested:
		return InterestedMsg{}

	case MsgNotInterested:
		return NotInterestedMsg{}

	case MsgHave:
		return HaveMsg{
			PieceIndex: int32(binary.BigEndian.Uint32(b[5:9])),
		}

	case MsgBitfield:
		return BitfieldMsg{
			Bitfield: b[5:],
		}

	case MsgRequest:
		return RequestMsg{
			Index:  binary.BigEndian.Uint32(b[5:9]),
			Begin:  binary.BigEndian.Uint32(b[9:13]),
			Length: binary.BigEndian.Uint32(b[13:17]),
		}

	case MsgPiece:
		return PieceMsg{
			Index: binary.BigEndian.Uint32(b[5:9]),
			Begin: binary.BigEndian.Uint32(b[9:13]),
			Block: b[13:],
		}

	case MsgCancel:
		return CancelMsg{
			Index:  binary.BigEndian.Uint32(b[5:9]),
			Begin:  binary.BigEndian.Uint32(b[9:13]),
			Length: binary.BigEndian.Uint32(b[13:17]),
		}

	case MsgPort:
		return PortMsg{
			ListenPort: binary.BigEndian.Uint16(b[5:7]),
		}
	}

	return nil
}
