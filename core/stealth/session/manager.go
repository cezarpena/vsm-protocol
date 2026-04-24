package session

import (
	"fmt"
	"math/rand"
	"net"
	"sync"
	"vsm-protocol/core/stealth"

	"github.com/google/gopacket/pcap"
)

// VSMSession holds the state for an active encrypted connection
type VSMSession struct {
	ID           int
	PeerID       string
	NoiseSession *stealth.NoiseSession
	LocalSeq     uint32
	RemoteSeq    uint32

	// Networking context
	DeviceHandle *pcap.Handle
	LocalIP      net.IP
	RemoteIP     net.IP
	LocalPort    int
	RemotePort   int
}

// Global Registry for all active sessions
var (
	registry = sync.Map{}
)

// Register adds a new session to the global registry
func Register(peerID string, noise *stealth.NoiseSession, lIP, rIP net.IP, lPort, rPort int, lSeq, rSeq uint32, h *pcap.Handle) int {
	id := rand.Intn(1000000)
	s := &VSMSession{
		ID:           id,
		PeerID:       peerID,
		NoiseSession: noise,
		LocalSeq:     lSeq,
		RemoteSeq:    rSeq,
		DeviceHandle: h,
		LocalIP:      lIP,
		RemoteIP:     rIP,
		LocalPort:    lPort,
		RemotePort:   rPort,
	}

	// 1. Store by ID (for Outbound/API calls)
	registry.Store(id, s)

	// 2. Store by Flow (for Inbound packet routing)
	// We use the RemoteIP:RemotePort as the key because that's our source for incoming packets
	flowKey := fmt.Sprintf("%s:%d", rIP.String(), rPort)
	registry.Store(flowKey, s)

	return id
}

// GetByID retrieves a session by its integer ID
func GetByID(id int) (*VSMSession, bool) {
	val, ok := registry.Load(id)
	if !ok {
		return nil, false
	}
	s, ok := val.(*VSMSession)
	return s, ok
}

// GetByFlow retrieves a session by its source address (IP:Port)
func GetByFlow(remoteIP net.IP, remotePort int) (*VSMSession, bool) {
	flowKey := fmt.Sprintf("%s:%d", remoteIP.String(), remotePort)
	val, ok := registry.Load(flowKey)
	if !ok {
		return nil, false
	}
	s, ok := val.(*VSMSession)
	return s, ok
}
