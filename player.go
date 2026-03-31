package main

import (
	"net"
	"sync"
)

// Player holds the authoritative state for a single connected player.
type Player struct {
	mu   sync.Mutex
	ID   uint8
	Addr *net.UDPAddr

	// Position in world space
	X, Y, Z float32

	// Rotation
	Yaw, Pitch float32

	// Pending movement input accumulated between ticks
	InputDX, InputDY, InputDZ float32
	InputYaw, InputPitch       float32
}

// ApplyInput applies the latest input values and resets the pending fields.
// Call this once per tick while holding the lock.
func (p *Player) ApplyInput() {
	p.X += p.InputDX
	p.Y += p.InputDY
	p.Z += p.InputDZ
	p.Yaw = p.InputYaw
	p.Pitch = p.InputPitch

	p.InputDX = 0
	p.InputDY = 0
	p.InputDZ = 0
	p.InputYaw = 0
	p.InputPitch = 0
}
