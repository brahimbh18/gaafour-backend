package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

// MsgType identifies the kind of UDP message.
type MsgType uint8

const (
	// MsgInput is sent by a client to report movement and rotation.
	MsgInput MsgType = 0x01

	// MsgState is sent by the server to report a player's authoritative state.
	MsgState MsgType = 0x02
)

// InputPacket is the fixed-size message a client sends every frame.
//
// Wire layout (little-endian, 22 bytes):
//
//	[0]      MsgType  (0x01)
//	[1]      PlayerID uint8
//	[2..5]   DX       float32
//	[6..9]   DY       float32
//	[10..13] DZ       float32
//	[14..17] Yaw      float32
//	[18..21] Pitch    float32
type InputPacket struct {
	Type     MsgType
	PlayerID uint8
	DX       float32
	DY       float32
	DZ       float32
	Yaw      float32
	Pitch    float32
}

// InputPacketSize is the exact byte length of a serialised InputPacket.
const InputPacketSize = 22

// DecodeInput deserialises an InputPacket from raw bytes.
func DecodeInput(data []byte) (*InputPacket, error) {
	if len(data) < InputPacketSize {
		return nil, fmt.Errorf("input packet too short: got %d, want %d", len(data), InputPacketSize)
	}
	r := bytes.NewReader(data)
	p := &InputPacket{}
	if err := binary.Read(r, binary.LittleEndian, p); err != nil {
		return nil, fmt.Errorf("decode input: %w", err)
	}
	return p, nil
}

// StatePacket is the fixed-size message the server sends back each tick.
//
// Wire layout (little-endian, 22 bytes):
//
//	[0]      MsgType  (0x02)
//	[1]      PlayerID uint8
//	[2..5]   X        float32
//	[6..9]   Y        float32
//	[10..13] Z        float32
//	[14..17] Yaw      float32
//	[18..21] Pitch    float32
type StatePacket struct {
	Type     MsgType
	PlayerID uint8
	X        float32
	Y        float32
	Z        float32
	Yaw      float32
	Pitch    float32
}

// StatePacketSize is the exact byte length of a serialised StatePacket.
const StatePacketSize = 22

// EncodeState serialises a StatePacket into a fixed-size byte slice.
func EncodeState(p *StatePacket) ([]byte, error) {
	buf := new(bytes.Buffer)
	if err := binary.Write(buf, binary.LittleEndian, p); err != nil {
		return nil, fmt.Errorf("encode state: %w", err)
	}
	return buf.Bytes(), nil
}
