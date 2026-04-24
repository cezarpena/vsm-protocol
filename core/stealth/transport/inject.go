package transport

import (
	"fmt"
	"net"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
)

// InjectData sends a raw PSH/ACK TCP packet with the given payload.
// Returns the new sequence number (old seq + payload length).
func InjectData(handle *pcap.Handle, srcIP, dstIP net.IP, srcPort, dstPort int, seq, ack uint32, payload []byte) uint32 {
	ip := &layers.IPv4{
		SrcIP:    srcIP,
		DstIP:    dstIP,
		Protocol: layers.IPProtocolTCP,
		Version:  4,
		TTL:      64,
	}

	tcp := &layers.TCP{
		SrcPort: layers.TCPPort(srcPort),
		DstPort: layers.TCPPort(dstPort),
		Seq:     seq,
		Ack:     ack,
		PSH:     true,
		ACK:     true,
		Window:  64240,
	}

	tcp.SetNetworkLayerForChecksum(ip)

	buf := gopacket.NewSerializeBuffer()
	opts := gopacket.SerializeOptions{ComputeChecksums: true, FixLengths: true}
	gopacket.SerializeLayers(buf, opts,
		&layers.Loopback{Family: layers.ProtocolFamilyIPv4},
		ip, tcp, gopacket.Payload(payload))

	handle.WritePacketData(buf.Bytes())

	fmt.Printf(" [VSM] Injected %d bytes (seq=%d)\n", len(payload), seq)
	return seq + uint32(len(payload))
}
