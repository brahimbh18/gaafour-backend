package main

import (
	"log"
	"net"
	"sync"
	"time"
)

const (
	listenAddr = ":9999"
	tickRate   = 60 // Hz
	maxPlayers = 2
)

// Server holds all server state.
type Server struct {
	conn    *net.UDPConn
	players [maxPlayers]*Player
	mu      sync.RWMutex // guards players slice
}

func main() {
	addr, err := net.ResolveUDPAddr("udp", listenAddr)
	if err != nil {
		log.Fatalf("resolve addr: %v", err)
	}

	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		log.Fatalf("listen UDP: %v", err)
	}
	defer conn.Close()

	log.Printf("gaafour-backend listening on UDP %s at %d Hz", listenAddr, tickRate)

	s := &Server{conn: conn}

	go s.readLoop()
	s.tickLoop()
}

// readLoop receives UDP datagrams and dispatches them to the appropriate player.
func (s *Server) readLoop() {
	buf := make([]byte, 1024)
	for {
		n, remoteAddr, err := s.conn.ReadFromUDP(buf)
		if err != nil {
			log.Printf("read error: %v", err)
			continue
		}

		data := buf[:n]
		if len(data) == 0 {
			continue
		}

		switch MsgType(data[0]) {
		case MsgInput:
			s.handleInput(data, remoteAddr)
		default:
			log.Printf("unknown message type 0x%02X from %s", data[0], remoteAddr)
		}
	}
}

// handleInput processes an InputPacket from a client.
func (s *Server) handleInput(data []byte, addr *net.UDPAddr) {
	pkt, err := DecodeInput(data)
	if err != nil {
		log.Printf("bad input packet from %s: %v", addr, err)
		return
	}

	if pkt.PlayerID >= maxPlayers {
		log.Printf("invalid player ID %d from %s", pkt.PlayerID, addr)
		return
	}

	s.mu.Lock()
	p := s.players[pkt.PlayerID]
	if p == nil {
		// First packet from this player — register them.
		p = &Player{ID: pkt.PlayerID, Addr: addr}
		s.players[pkt.PlayerID] = p
		log.Printf("player %d connected from %s", pkt.PlayerID, addr)
	}
	s.mu.Unlock()

	p.mu.Lock()
	// Accumulate input; the tick loop will consume it.
	p.InputDX = pkt.DX
	p.InputDY = pkt.DY
	p.InputDZ = pkt.DZ
	p.InputYaw = pkt.Yaw
	p.InputPitch = pkt.Pitch
	p.mu.Unlock()
}

// tickLoop runs the fixed-rate game loop.
func (s *Server) tickLoop() {
	ticker := time.NewTicker(time.Second / tickRate)
	defer ticker.Stop()

	for range ticker.C {
		s.tick()
	}
}

// tick advances the game state by one tick and broadcasts authoritative state.
func (s *Server) tick() {
	s.mu.RLock()
	players := s.players
	s.mu.RUnlock()

	// Update each connected player's position from their latest input.
	for _, p := range players {
		if p == nil {
			continue
		}
		p.mu.Lock()
		p.ApplyInput()
		p.mu.Unlock()
	}

	// Broadcast each player's state to all other connected players.
	for i, src := range players {
		if src == nil {
			continue
		}

		src.mu.Lock()
		pkt := &StatePacket{
			Type:     MsgState,
			PlayerID: src.ID,
			X:        src.X,
			Y:        src.Y,
			Z:        src.Z,
			Yaw:      src.Yaw,
			Pitch:    src.Pitch,
		}
		src.mu.Unlock()

		payload, err := EncodeState(pkt)
		if err != nil {
			log.Printf("encode state for player %d: %v", src.ID, err)
			continue
		}

		for j, dst := range players {
			if j == i || dst == nil {
				continue
			}
			dst.mu.Lock()
			dstAddr := dst.Addr
			dst.mu.Unlock()

			if _, err := s.conn.WriteToUDP(payload, dstAddr); err != nil {
				log.Printf("send state player %d → player %d: %v", src.ID, dst.ID, err)
			}
		}
	}
}
