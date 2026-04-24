package transport

import (
	"crypto/ed25519"
	"fmt"
	"net"
	"vsm-protocol/core/stealth"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
)

func StealthDial(
	device string,
	targetIP string,
	targetPort int,
	privKey ed25519.PrivateKey) (
	peerID string,
	noiseSession *stealth.NoiseSession,
	localIP, remoteIP net.IP,
	localPort, remotePort int,
	localSeq, remoteSeq uint32,
	handle *pcap.Handle,
	err error,
) {
	// 1. Generate secret signature
	knock := stealth.GenerateKnock(privKey)

	srcIP := net.ParseIP("127.0.0.1")
	dstIP := net.ParseIP(targetIP)
	srcPort := 12345

	// 2. Build the IP Layer
	ip := &layers.IPv4{
		SrcIP:    srcIP,
		DstIP:    dstIP,
		Protocol: layers.IPProtocolTCP,
		Version:  4,
		TTL:      64,
	}

	// 3. Build the TCP Layer
	tcp := &layers.TCP{
		SrcPort: layers.TCPPort(srcPort),
		DstPort: layers.TCPPort(targetPort),
		Seq:     knock,
		SYN:     true,
		Window:  64240,
	}

	tcp.SetNetworkLayerForChecksum(ip)

	buf := gopacket.NewSerializeBuffer()
	opts := gopacket.SerializeOptions{ComputeChecksums: true, FixLengths: true}

	loopback := &layers.Loopback{Family: layers.ProtocolFamilyIPv4}
	gopacket.SerializeLayers(buf, opts, loopback, ip, tcp)

	// 4. Inject the knock
	handle, err = pcap.OpenLive(device, 1600, true, pcap.BlockForever)
	if err != nil {
		return "", nil, nil, nil, 0, 0, 0, 0, nil, err
	}

	// Filter to only our conversation
	filter := fmt.Sprintf("tcp and port %d and port %d", targetPort, srcPort)
	handle.SetBPFFilter(filter)

	fmt.Printf(" [STEALTH] Injecting Knock: %d\n", knock)
	handle.WritePacketData(buf.Bytes())

	fmt.Println(" [STEALTH] Waiting for SYN-ACK...")

	// 5. Wait for SYN-ACK
	localSeq = knock + 1
	packetSource := gopacket.NewPacketSource(handle, handle.LinkType())
	for packet := range packetSource.Packets() {
		tcpLayer := packet.Layer(layers.LayerTypeTCP)
		if tcpLayer == nil {
			continue
		}
		tcpPkt, _ := tcpLayer.(*layers.TCP)

		if tcpPkt.SrcPort == layers.TCPPort(targetPort) && tcpPkt.DstPort == layers.TCPPort(srcPort) {
			if tcpPkt.SYN && tcpPkt.ACK && tcpPkt.Ack == knock+1 {
				fmt.Printf(" [STEALTH] HANDSHAKE SUCCESS! Server acknowledged ISN: %d\n", tcpPkt.Ack)
				remoteSeq = tcpPkt.Seq + 1

				// 6. TCP handshake done. Now perform Noise XX handshake.
				noise, _ := stealth.InitializeNoise(true, privKey)

				// --- Noise Message 1: Initiator → Responder ("e") ---
				msg1, _ := noise.HandshakeStep(nil, true)
				fmt.Printf(" [NOISE] Sending handshake msg 1 (%d bytes)\n", len(msg1))
				localSeq = InjectData(handle, srcIP, dstIP, srcPort, targetPort, localSeq, remoteSeq, msg1)

				// --- Wait for Noise Message 2: Responder → Initiator ("e, ee, s, es") ---
				for pkt2 := range packetSource.Packets() {
					tcp2 := pkt2.Layer(layers.LayerTypeTCP)
					if tcp2 == nil {
						continue
					}
					t2, _ := tcp2.(*layers.TCP)
					if t2.SrcPort == layers.TCPPort(targetPort) && len(t2.Payload) > 0 {
						fmt.Printf(" [NOISE] Received handshake msg 2 (%d bytes)\n", len(t2.Payload))
						remoteSeq = t2.Seq + uint32(len(t2.Payload))
						noise.HandshakeStep(t2.Payload, false)

						// --- Noise Message 3: Initiator → Responder ("s, se") ---
						msg3, _ := noise.HandshakeStep(nil, true)
						fmt.Printf(" [NOISE] Sending handshake msg 3 (%d bytes)\n", len(msg3))
						localSeq = InjectData(handle, srcIP, dstIP, srcPort, targetPort, localSeq, remoteSeq, msg3)

						fmt.Println(" [NOISE] Handshake complete. Tunnel is live.")

						return "server_peer", noise,
							srcIP, dstIP,
							srcPort, targetPort,
							localSeq, remoteSeq,
							handle, nil
					}
				}
			}
		}
	}

	return "", nil, nil, nil, 0, 0, 0, 0, nil, fmt.Errorf("stealth handshake timed out")
}

