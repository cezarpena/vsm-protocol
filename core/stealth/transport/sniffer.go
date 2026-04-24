package transport

import (
	"crypto/ed25519"
	"fmt"
	"net"
	"vsm-protocol/core/stealth"
	"vsm-protocol/core/stealth/session"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
)

type HandshakeCompleteFunc func(
	peerID string,
	sessionObj *stealth.NoiseSession,
	localIP, remoteIP net.IP,
	localPort, remotePort int,
	localSeq, remoteSeq uint32,
	handle *pcap.Handle,
)

type MessageReceivedFunc func(sessionID int, peerID string, data string)

// pendingHandshake tracks an in-progress Noise handshake on the server side
type pendingHandshake struct {
	noise     *stealth.NoiseSession
	localIP   net.IP
	remoteIP  net.IP
	localPort int
	remotePort int
	localSeq  uint32
	remoteSeq uint32
	stage     int // 0 = waiting for msg1, 1 = waiting for msg3
}

func ListenForKnock(device string, port int, myPrivKey ed25519.PrivateKey, onComplete HandshakeCompleteFunc, onMessage MessageReceivedFunc) {
	handle, err := pcap.OpenLive(device, 1600, true, pcap.BlockForever)
	if err != nil {
		fmt.Println("Error opening device:", err)
		return
	}
	defer handle.Close()

	filter := fmt.Sprintf("tcp and port %d", port)
	err = handle.SetBPFFilter(filter)
	if err != nil {
		fmt.Println("Error setting filter:", err)
		return
	}

	fmt.Printf(" [VSM] Listening on %s:%d...\n", device, port)

	// Track in-progress Noise handshakes by remote IP:Port
	pending := make(map[string]*pendingHandshake)

	packetSource := gopacket.NewPacketSource(handle, handle.LinkType())
	for packet := range packetSource.Packets() {
		ipLayer := packet.Layer(layers.LayerTypeIPv4)
		tcpLayer := packet.Layer(layers.LayerTypeTCP)
		if ipLayer == nil || tcpLayer == nil {
			continue
		}
		ip, _ := ipLayer.(*layers.IPv4)
		tcp, _ := tcpLayer.(*layers.TCP)

		flowKey := fmt.Sprintf("%s:%d", ip.SrcIP.String(), int(tcp.SrcPort))

		// --- 1. SYN (New knock) ---
		if tcp.SYN && !tcp.ACK {
			if stealth.VerifyKnock(myPrivKey, tcp.Seq) {
				fmt.Println(" [VSM] ACCESS GRANTED: Valid Secret Knock!")
				replyWithSynAck(handle, packet, tcp)

				noise, _ := stealth.InitializeNoise(false, myPrivKey)

				pending[flowKey] = &pendingHandshake{
					noise:      noise,
					localIP:    ip.DstIP,
					remoteIP:   ip.SrcIP,
					localPort:  int(tcp.DstPort),
					remotePort: int(tcp.SrcPort),
					localSeq:   101, // SYN-ACK seq was 100, so next is 101
					remoteSeq:  tcp.Seq + 1,
					stage:      0,
				}
			}
			continue
		}

		// --- 2. Data packets (Noise handshake or application data) ---
		if tcp.ACK && len(tcp.Payload) > 0 {

			// Check if this is part of a pending Noise handshake
			hs, isPending := pending[flowKey]
			if isPending {
				if hs.stage == 0 {
					// Noise Message 1 from initiator ("e")
					fmt.Printf(" [NOISE] Received handshake msg 1 (%d bytes)\n", len(tcp.Payload))
					hs.remoteSeq = tcp.Seq + uint32(len(tcp.Payload))
					hs.noise.HandshakeStep(tcp.Payload, false)

					// Send Noise Message 2 ("e, ee, s, es")
					msg2, _ := hs.noise.HandshakeStep(nil, true)
					fmt.Printf(" [NOISE] Sending handshake msg 2 (%d bytes)\n", len(msg2))
					hs.localSeq = InjectData(handle, hs.localIP, hs.remoteIP, hs.localPort, hs.remotePort, hs.localSeq, hs.remoteSeq, msg2)
					hs.stage = 1

				} else if hs.stage == 1 {
					// Noise Message 3 from initiator ("s, se")
					fmt.Printf(" [NOISE] Received handshake msg 3 (%d bytes)\n", len(tcp.Payload))
					hs.remoteSeq = tcp.Seq + uint32(len(tcp.Payload))
					hs.noise.HandshakeStep(tcp.Payload, false)

					fmt.Println(" [NOISE] Handshake complete. Tunnel is live.")

					// Handshake done — fire the callback with fully-keyed session
					onComplete("peer_client",
						hs.noise,
						hs.localIP, hs.remoteIP,
						hs.localPort, hs.remotePort,
						hs.localSeq, hs.remoteSeq,
						handle,
					)

					delete(pending, flowKey)
				}
				continue
			}

			// Not a pending handshake — must be application data
			s, ok := session.GetByFlow(ip.SrcIP, int(tcp.SrcPort))
			if ok {
				fmt.Printf(" [VSM] Incoming Data (%d bytes) for Session %d\n", len(tcp.Payload), s.ID)
				s.RemoteSeq = tcp.Seq + uint32(len(tcp.Payload))

				plaintext, err := s.NoiseSession.Decrypt(tcp.Payload)
				if err == nil {
					onMessage(s.ID, s.PeerID, string(plaintext))
				} else {
					fmt.Printf(" [VSM] Decryption Error: %v\n", err)
				}
			}
		}
	}
}

func replyWithSynAck(handle *pcap.Handle, original gopacket.Packet, tcp *layers.TCP) {
	ipLayer := original.Layer(layers.LayerTypeIPv4)
	if ipLayer == nil {
		return
	}
	ip, _ := ipLayer.(*layers.IPv4)

	replyIP := &layers.IPv4{
		SrcIP:    ip.DstIP,
		DstIP:    ip.SrcIP,
		Protocol: layers.IPProtocolTCP,
		Version:  4,
		TTL:      64,
	}

	replyTCP := &layers.TCP{
		SrcPort: tcp.DstPort,
		DstPort: tcp.SrcPort,
		Seq:     100,
		Ack:     tcp.Seq + 1,
		SYN:     true,
		ACK:     true,
		Window:  64240,
	}

	replyTCP.SetNetworkLayerForChecksum(replyIP)

	loopback := &layers.Loopback{Family: layers.ProtocolFamilyIPv4}

	buf := gopacket.NewSerializeBuffer()
	opts := gopacket.SerializeOptions{ComputeChecksums: true, FixLengths: true}
	gopacket.SerializeLayers(buf, opts, loopback, replyIP, replyTCP)

	fmt.Println(" [VSM] Sending SYN-ACK...")
	handle.WritePacketData(buf.Bytes())
}

