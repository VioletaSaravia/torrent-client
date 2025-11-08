package main

import (
	"encoding/binary"
)

type PeerMsgType uint8

const (
	MsgChoke PeerMsgType = iota
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

type IPeerMsg interface {
	ToBytes() []byte
	FromBytes([]byte) IPeerMsg
	Type() PeerMsgType
}

func NewIPeerMsg(received []byte) IPeerMsg {
	switch PeerMsgType(received[4]) {
	case MsgChoke:
		return ChokeMsg{}.FromBytes(received)
	case MsgUnchoke:
		return UnchokeMsg{}.FromBytes(received)
	case MsgInterested:
		return InterestedMsg{}.FromBytes(received)
	case MsgNotInterested:
		return NotInterestedMsg{}.FromBytes(received)
	case MsgHave:
		return HaveMsg{}.FromBytes(received)
	case MsgBitfield:
		return BitfieldMsg{}.FromBytes(received)
	case MsgRequest:
		return RequestMsg{}.FromBytes(received)
	case MsgPiece:
		return PieceMsg{}.FromBytes(received)
	case MsgCancel:
		return CancelMsg{}.FromBytes(received)
	case MsgPort:
		return PortMsg{}.FromBytes(received)
	}

	return nil
}

type BitfieldMsg struct {
	Bitfield []byte
}

var _ IPeerMsg = BitfieldMsg{}

func (m BitfieldMsg) ToBytes() []byte {
	result := make([]byte, 5+len(m.Bitfield))
	binary.BigEndian.PutUint32(result[:4], 1)
	result[4] = byte(m.Type())
	copy(result[5:], m.Bitfield)
	return result
}

func (m BitfieldMsg) FromBytes(b []byte) IPeerMsg {
	return BitfieldMsg{
		Bitfield: b[5:],
	}
}

func (m BitfieldMsg) Type() PeerMsgType {
	return MsgBitfield
}

type RequestMsg struct {
	Index  uint32
	Begin  uint32
	Length uint32
}

var _ IPeerMsg = RequestMsg{}

func (m RequestMsg) ToBytes() []byte {
	result := make([]byte, 5+12)
	binary.BigEndian.PutUint32(result[:4], 1)
	result[4] = byte(m.Type())
	binary.BigEndian.PutUint32(result[5:9], m.Index)
	binary.BigEndian.PutUint32(result[9:13], m.Begin)
	binary.BigEndian.PutUint32(result[13:17], m.Length)
	return result
}

func (m RequestMsg) FromBytes(b []byte) IPeerMsg {
	return RequestMsg{
		Index:  binary.BigEndian.Uint32(b[5:9]),
		Begin:  binary.BigEndian.Uint32(b[9:13]),
		Length: binary.BigEndian.Uint32(b[13:17]),
	}
}

func (m RequestMsg) Type() PeerMsgType {
	return MsgRequest
}

type PieceMsg struct {
	Index uint32
	Begin uint32
	Block []byte
}

var _ IPeerMsg = PieceMsg{}

func (m PieceMsg) ToBytes() []byte {
	result := make([]byte, 5+4+4+len(m.Block))
	binary.BigEndian.PutUint32(result[:4], 1)
	result[4] = byte(m.Type())
	binary.BigEndian.PutUint32(result[5:9], m.Index)
	binary.BigEndian.PutUint32(result[9:13], m.Begin)
	copy(result[13:], m.Block)
	return result
}

func (m PieceMsg) FromBytes(b []byte) IPeerMsg {
	return PieceMsg{
		Index: binary.BigEndian.Uint32(b[5:9]),
		Begin: binary.BigEndian.Uint32(b[9:13]),
		Block: b[13:],
	}
}

func (m PieceMsg) Type() PeerMsgType {
	return MsgPiece
}

type CancelMsg struct {
	Index  uint32
	Begin  uint32
	Length uint32
}

var _ IPeerMsg = CancelMsg{}

func (m CancelMsg) ToBytes() []byte {
	result := make([]byte, 5+12)
	binary.BigEndian.PutUint32(result[:4], 1)
	result[4] = byte(m.Type())
	binary.BigEndian.PutUint32(result[5:9], m.Index)
	binary.BigEndian.PutUint32(result[9:13], m.Begin)
	binary.BigEndian.PutUint32(result[13:17], m.Length)
	return result
}

func (m CancelMsg) FromBytes(b []byte) IPeerMsg {
	return CancelMsg{
		Index:  binary.BigEndian.Uint32(b[5:9]),
		Begin:  binary.BigEndian.Uint32(b[9:13]),
		Length: binary.BigEndian.Uint32(b[13:17]),
	}
}

func (m CancelMsg) Type() PeerMsgType {
	return MsgCancel
}

type PortMsg struct {
	ListenPort uint16
}

var _ IPeerMsg = PortMsg{}

func (m PortMsg) ToBytes() []byte {
	result := make([]byte, 5+2)
	binary.BigEndian.PutUint32(result[:4], 1)
	result[4] = byte(m.Type())
	binary.BigEndian.PutUint16(result[5:7], m.ListenPort)
	return result
}

func (m PortMsg) FromBytes(b []byte) IPeerMsg {
	return PortMsg{
		ListenPort: binary.BigEndian.Uint16(b[5:7]),
	}
}

func (m PortMsg) Type() PeerMsgType {
	return MsgPort
}

type HaveMsg struct {
	PieceIndex int32
}

var _ IPeerMsg = HaveMsg{}

func (m HaveMsg) ToBytes() []byte {
	result := make([]byte, 5+4)
	binary.BigEndian.PutUint32(result[:4], 1)
	result[4] = byte(m.Type())
	binary.BigEndian.PutUint32(result[5:], uint32(m.PieceIndex))
	return result
}

func (m HaveMsg) FromBytes(b []byte) IPeerMsg {
	return HaveMsg{
		PieceIndex: int32(binary.BigEndian.Uint32(b[5:9])),
	}
}

func (m HaveMsg) Type() PeerMsgType {
	return MsgHave
}

type NotInterestedMsg struct{}

var _ IPeerMsg = NotInterestedMsg{}

func (m NotInterestedMsg) ToBytes() []byte {
	result := make([]byte, 5)
	binary.BigEndian.PutUint32(result[:4], 1)
	result[4] = byte(m.Type())
	return result
}

func (m NotInterestedMsg) FromBytes(b []byte) IPeerMsg {
	return NotInterestedMsg{}
}

func (m NotInterestedMsg) Type() PeerMsgType {
	return MsgNotInterested
}

type InterestedMsg struct{}

var _ IPeerMsg = InterestedMsg{}

func (m InterestedMsg) ToBytes() []byte {
	result := make([]byte, 5)
	binary.BigEndian.PutUint32(result[:4], 1)
	result[4] = byte(m.Type())
	return result
}

func (m InterestedMsg) Type() PeerMsgType {
	return MsgInterested
}

func (m InterestedMsg) FromBytes(b []byte) IPeerMsg {
	return InterestedMsg{}
}

type UnchokeMsg struct{}

var _ IPeerMsg = UnchokeMsg{}

func (m UnchokeMsg) ToBytes() []byte {
	result := make([]byte, 5)
	binary.BigEndian.PutUint32(result[:4], 1)
	result[4] = byte(m.Type())
	return result
}

func (m UnchokeMsg) FromBytes(b []byte) IPeerMsg {
	return UnchokeMsg{}
}
func (m UnchokeMsg) Type() PeerMsgType {
	return MsgUnchoke
}

type ChokeMsg struct{}

var _ IPeerMsg = ChokeMsg{}

func (m ChokeMsg) ToBytes() []byte {
	result := make([]byte, 5)
	binary.BigEndian.PutUint32(result[:4], 1)
	result[4] = byte(m.Type())
	return result
}

func (m ChokeMsg) FromBytes(b []byte) IPeerMsg {
	return ChokeMsg{}
}
func (m ChokeMsg) Type() PeerMsgType {
	return MsgChoke
}

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

var _ IPeerMsg = HandshakeMsg{}

func (m HandshakeMsg) FromBytes(b []byte) IPeerMsg {
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

func (m HandshakeMsg) Type() PeerMsgType {
	return MsgHandshake
}
