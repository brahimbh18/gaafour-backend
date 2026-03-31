package main

import (
	"testing"
)

func TestPacketSizes(t *testing.T) {
	if InputPacketSize != StatePacketSize {
		t.Errorf("packet sizes differ: input=%d state=%d", InputPacketSize, StatePacketSize)
	}
	if InputPacketSize != 22 {
		t.Errorf("InputPacketSize = %d, want 22", InputPacketSize)
	}
}

func TestEncodeState(t *testing.T) {
	sp := &StatePacket{
		Type:     MsgState,
		PlayerID: 0,
		X:        10.0,
		Y:        5.0,
		Z:        -3.0,
		Yaw:      180.0,
		Pitch:    30.0,
	}

	b, err := EncodeState(sp)
	if err != nil {
		t.Fatalf("EncodeState: %v", err)
	}
	if len(b) != StatePacketSize {
		t.Errorf("encoded size = %d, want %d", len(b), StatePacketSize)
	}
	if MsgType(b[0]) != MsgState {
		t.Errorf("first byte = 0x%02X, want MsgState (0x%02X)", b[0], MsgState)
	}
	if b[1] != 0 {
		t.Errorf("PlayerID byte = %d, want 0", b[1])
	}
}

func TestDecodeInput(t *testing.T) {
	// Build a valid InputPacket byte slice via the state encoder (same struct layout).
	sp := &StatePacket{
		Type:     MsgState,
		PlayerID: 1,
		X:        3.0,
		Y:        0.0,
		Z:        1.0,
		Yaw:      45.0,
		Pitch:    10.0,
	}
	raw, _ := EncodeState(sp)
	raw[0] = byte(MsgInput) // fix type byte

	pkt, err := DecodeInput(raw)
	if err != nil {
		t.Fatalf("DecodeInput: %v", err)
	}
	if pkt.Type != MsgInput {
		t.Errorf("Type = %v, want MsgInput", pkt.Type)
	}
	if pkt.PlayerID != 1 {
		t.Errorf("PlayerID = %d, want 1", pkt.PlayerID)
	}
	if pkt.DX != 3.0 {
		t.Errorf("DX = %v, want 3.0", pkt.DX)
	}
	if pkt.Yaw != 45.0 {
		t.Errorf("Yaw = %v, want 45.0", pkt.Yaw)
	}
}

func TestDecodeInputTooShort(t *testing.T) {
	_, err := DecodeInput([]byte{0x01, 0x00})
	if err == nil {
		t.Error("expected error for short packet, got nil")
	}
}

func TestPlayerApplyInput(t *testing.T) {
	p := &Player{ID: 0}
	p.InputDX = 1.0
	p.InputDY = 2.0
	p.InputDZ = 3.0
	p.InputYaw = 90.0
	p.InputPitch = -10.0

	p.ApplyInput()

	if p.X != 1.0 || p.Y != 2.0 || p.Z != 3.0 {
		t.Errorf("position = (%v,%v,%v), want (1,2,3)", p.X, p.Y, p.Z)
	}
	if p.Yaw != 90.0 || p.Pitch != -10.0 {
		t.Errorf("rotation = (%v,%v), want (90,-10)", p.Yaw, p.Pitch)
	}
	// Inputs should be reset
	if p.InputDX != 0 || p.InputDY != 0 || p.InputDZ != 0 {
		t.Errorf("inputs not reset after ApplyInput")
	}

	// Second tick with no new input — position should stay the same
	p.ApplyInput()
	if p.X != 1.0 || p.Y != 2.0 || p.Z != 3.0 {
		t.Errorf("position changed on tick with no input: (%v,%v,%v)", p.X, p.Y, p.Z)
	}
}
