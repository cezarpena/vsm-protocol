package main

/*
#include <stdlib.h>

typedef void (*knock_callback)(int, char*);
typedef void (*message_callback)(int, char*, char*);

static void call_knock_callback(knock_callback cb, int port, char* peer_id) {
	cb(port, peer_id);
}
static void call_message_callback(message_callback cb, int session_id, char* peer_id, char* msg) {
	cb(session_id, peer_id, msg);
}
*/
import "C"

import (
	"encoding/json"
	"fmt"
	"net"
	"runtime"
	"unsafe"
	"vsm-protocol/core/stealth"
	"vsm-protocol/core/stealth/session"
	"vsm-protocol/core/stealth/transport"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
)

// Event types for the callback dispatcher
type knockEvent struct {
	sessionID int
	peerID    string
}

type messageEvent struct {
	sessionID int
	peerID    string
	data      string
}

// Global channels and callback storage
var (
	knockChan   = make(chan knockEvent, 64)
	messageChan = make(chan messageEvent, 64)
)

// startCallbackDispatcher runs on a single, locked OS thread.
// It waits for events from the sniffer and calls the C/Python
// callbacks safely from a stable thread context.
func startCallbackDispatcher(knockCb C.knock_callback, msgCb C.message_callback) {
	go func() {
		// Pin this goroutine to one OS thread forever.
		// Python will always see callbacks from this same thread.
		runtime.LockOSThread()

		for {
			select {
			case evt := <-knockChan:
				peerID_C := C.CString(evt.peerID)
				C.call_knock_callback(knockCb, C.int(evt.sessionID), peerID_C)
				C.free(unsafe.Pointer(peerID_C))

			case evt := <-messageChan:
				pID := C.CString(evt.peerID)
				msg := C.CString(evt.data)
				C.call_message_callback(msgCb, C.int(evt.sessionID), pID, msg)
				C.free(unsafe.Pointer(pID))
				C.free(unsafe.Pointer(msg))
			}
		}
	}()
}

//export GenerateVSMIdentity
func GenerateVSMIdentity() *C.char {

	id, err := stealth.GenerateNewIdentity()
	if err != nil {
		return nil
	}

	vsmJson, err := json.Marshal(id)
	if err != nil {
		return nil
	}

	return C.CString(string(vsmJson))
}

//export FreeString
func FreeString(str *C.char) {
	C.free(unsafe.Pointer(str))
}

//export StartVSMServer
func StartVSMServer(port int, idStr *C.char, knockCb C.knock_callback, msgCb C.message_callback) {

	var identity stealth.VsmIdentity
	err := json.Unmarshal([]byte(C.GoString(idStr)), &identity)
	if err != nil {
		fmt.Printf(" [CGO] Error loading identity: %v\n", err)
		return
	}

	privKey, err := stealth.DecodeKey(identity.PrivKey)
	if err != nil {
		fmt.Printf(" [CGO] Error decoding private key: %v\n", err)
		return
	}

	fmt.Printf(" [CGO] Starting VSM Server on port %d for identity %s\n", port, identity.ID)

	// Start the thread-safe callback dispatcher
	startCallbackDispatcher(knockCb, msgCb)

	go func() {
		transport.ListenForKnock("lo0", port, privKey,
			func(pID string, nS *stealth.NoiseSession, lIP, rIP net.IP, lP, rP int, lS, rS uint32, h *pcap.Handle) {
				sessionID := session.Register(pID, nS, lIP, rIP, lP, rP, lS, rS, h)
				// Send event to the dispatcher instead of calling C directly
				knockChan <- knockEvent{sessionID: sessionID, peerID: pID}
			},
			func(sessionID int, peerID string, data string) {
				// Send event to the dispatcher instead of calling C directly
				messageChan <- messageEvent{sessionID: sessionID, peerID: peerID, data: data}
			},
		)
	}()

}

//export SendMessage
func SendMessage(sessionID int, message *C.char) bool {
	s, ok := session.GetByID(sessionID)
	if !ok {
		return false
	}

	plaintext := []byte(C.GoString(message))
	ciphertext, err := s.NoiseSession.Encrypt(plaintext)
	if err != nil {
		return false
	}

	defer func() {
		s.LocalSeq += uint32(len(ciphertext))
	}()

	ip := &layers.IPv4{
		SrcIP: s.LocalIP, DstIP: s.RemoteIP,
		Protocol: layers.IPProtocolTCP,
		Version:  4,
		TTL:      64,
	}

	tcp := &layers.TCP{
		SrcPort: layers.TCPPort(s.LocalPort),
		DstPort: layers.TCPPort(s.RemotePort),
		Seq:     s.LocalSeq,
		Ack:     s.RemoteSeq,
		PSH:     true,
		ACK:     true,
		Window:  64240,
	}

	tcp.SetNetworkLayerForChecksum(ip)

	buf := gopacket.NewSerializeBuffer()
	opts := gopacket.SerializeOptions{ComputeChecksums: true, FixLengths: true}

	gopacket.SerializeLayers(buf, opts,
		&layers.Loopback{Family: layers.ProtocolFamilyIPv4},
		ip, tcp, gopacket.Payload(ciphertext))

	fmt.Printf(" [CGO] Injecting Encrypted Message (%d bytes)...\n", len(ciphertext))
	s.DeviceHandle.WritePacketData(buf.Bytes())
	return true
}

//export VSMDial
func VSMDial(device, targetIP *C.char, port int, idStr *C.char) int {
	var identity stealth.VsmIdentity
	json.Unmarshal([]byte(C.GoString(idStr)), &identity)
	privKey, _ := stealth.DecodeKey(identity.PrivKey)
	fmt.Printf(" [CGO] Dialing %s:%d...\n", C.GoString(targetIP), port)

	pID, nS, lIP, rIP, lP, rP, lS, rS, handle, err := transport.StealthDial(
		C.GoString(device),
		C.GoString(targetIP),
		port,
		privKey,
	)

	if err != nil {
		return -1
	}
	return session.Register(pID, nS, lIP, rIP, lP, rP, lS, rS, handle)
}

func main() {}
