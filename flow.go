package main

import (
	"fmt"
	"encoding/binary"
	"bytes"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
)


// TcpIpFlow is used for tracking unidirectional TCP flows
type TcpIpFlow struct {
	ipFlow  gopacket.Flow
	tcpFlow gopacket.Flow
}

// NewTcpIpFlowFromLayers given IPv4 and TCP layers it returns a TcpIpFlow
func NewTcpIpFlowFromLayers(ipLayer layers.IPv4, tcpLayer layers.TCP) *TcpIpFlow {
	return &TcpIpFlow{
		ipFlow:  ipLayer.NetworkFlow(),
		tcpFlow: tcpLayer.TransportFlow(),
	}
}

// NewTcpIpFlowFromFlows given a net flow (either ipv4 or ipv6) and TCP flow returns a TcpIpFlow
func NewTcpIpFlowFromFlows(netFlow gopacket.Flow, tcpFlow gopacket.Flow) TcpIpFlow {
	// XXX todo: check that the flow types are correct
	return TcpIpFlow{
		ipFlow:  netFlow,
		tcpFlow: tcpFlow,
	}
}

// String returns the string representation of a TcpIpFlow
func (t TcpIpFlow) String() string {
	return fmt.Sprintf("%s:%s-%s:%s", t.ipFlow.Src().String(), t.tcpFlow.Src().String(), t.ipFlow.Dst().String(), t.tcpFlow.Dst().String())
}

// Reverse returns a reversed TcpIpFlow, that is to say the resulting
// TcpIpFlow flow will be made up of a reversed IP flow and a reversed
// TCP flow.
func (t *TcpIpFlow) Reverse() TcpIpFlow {
	return NewTcpIpFlowFromFlows(t.ipFlow.Reverse(), t.tcpFlow.Reverse())
}

// Equal returns true if TcpIpFlow structs t and s are equal. False otherwise.
func (t *TcpIpFlow) Equal(s *TcpIpFlow) bool {
	return t.ipFlow == s.ipFlow && t.tcpFlow == s.tcpFlow
}

// getPacketFlow returns a TcpIpFlow struct given a byte array packet
func NewTcpIpFlowFromPacket(packet []byte) (*TcpIpFlow, error) {
	var ip layers.IPv4
	var tcp layers.TCP
	decoded := []gopacket.LayerType{}
	parser := gopacket.NewDecodingLayerParser(layers.LayerTypeIPv4, &ip, &tcp)
	err := parser.DecodeLayers(packet, &decoded)
	if err != nil {
		return &TcpIpFlow{}, err
	}
	return &TcpIpFlow{
		ipFlow:  ip.NetworkFlow(),
		tcpFlow: tcp.TransportFlow(),
	}, nil
}

// Flows returns the component flow structs IPv4, TCP
func (t *TcpIpFlow) Flows() (gopacket.Flow, gopacket.Flow) {
	return t.ipFlow, t.tcpFlow
}

type HashedTcpIpv6Flow struct {
	// ipv6 16 bytes + tcp port 2 bytes == 18
	Src [18]byte
	Dst [18]byte
}

// NewHashedTcpIpv6Flow returns a comparable struct given a flow struct
func NewHashedTcpIpv6Flow(flow *TcpIpFlow) HashedTcpIpv6Flow {
	hash := HashedTcpIpv6Flow{}

	ipFlow, tcpFlow := flow.Flows()
	src := make([]byte, 18)
	copy(src, ipFlow.Src().Raw())
	copy(src[len(ipFlow.Src().Raw()):], tcpFlow.Src().Raw())
	copy(hash.Src[:], src)

	dst := make([]byte, 18)
	copy(dst, ipFlow.Dst().Raw())
	copy(dst[len(ipFlow.Dst().Raw()):], tcpFlow.Dst().Raw())
	copy(hash.Dst[:], dst)

	if bytes.Compare(hash.Src[:], hash.Dst[:]) > 0 {
		return hash
	} else {
		// reverse
		a := hash.Src
		hash.Src = hash.Dst
		hash.Dst = a
		return hash
	}
}

type HashedTcpIpv4Flow struct {
	Src uint64
	Dst uint64
}

// NewHashedTcpIpv4Flow returns a comparable struct given a flow struct
func NewHashedTcpIpv4Flow(flow *TcpIpFlow) HashedTcpIpv4Flow {
	hash := HashedTcpIpv4Flow{}

	ipFlow, tcpFlow := flow.Flows()
	src := make([]byte, 8)
	copy(src, ipFlow.Src().Raw())
	copy(src[len(ipFlow.Src().Raw()):], tcpFlow.Src().Raw())
	hash.Src = binary.BigEndian.Uint64(src)

	dst := make([]byte, 8)
	copy(dst, ipFlow.Dst().Raw())
	copy(dst[len(ipFlow.Dst().Raw()):], tcpFlow.Dst().Raw())
	hash.Dst = binary.BigEndian.Uint64(dst)

	if hash.Src > hash.Dst {
		return hash
	} else {
		// reverse
		a := hash.Src
		hash.Src = hash.Dst
		hash.Dst = a
		return hash
	}
}
