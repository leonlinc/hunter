package pcap

import (
	"log"
	"time"

	"github.com/google/gopacket"
	"github.com/google/gopacket/pcap"
	"github.com/leonlinc/hunter/m2ts"
)

type Parser struct {
	srcAddr   string
	srcPort   string
	dstAddr   string
	dstPort   string
	epochTime int64
}

func CreateParser() *Parser {
	return &Parser{
		epochTime: -1,
	}
}

func (p *Parser) Parse(file string, root string) {
	pcapFile, err := pcap.OpenOffline(file)
	if err != nil {
		log.Fatal(err)
	}
	defer pcapFile.Close()

	ccAnalyzer := m2ts.CreateCCAnalyzer(root)
	pcrAnalyzer := m2ts.CreatePCRAnalyzer(root)

	var idx, packetIndex int64
	source := gopacket.NewPacketSource(pcapFile, pcapFile.LinkType())
	for packet := range source.Packets() {
		captureTime := p.getCaptureTime(packet)
		if p.isValid(packet) {
			payload := packet.ApplicationLayer().Payload()
			for pkt := range m2ts.Packets(payload) {
				ccAnalyzer.Process(idx, pkt)
				pcrAnalyzer.Process(idx, pkt, m2ts.Metadata{packetIndex, captureTime, -1})
				idx += 1
			}
		}
		packetIndex += 1
	}

	ccAnalyzer.Close()
	pcrAnalyzer.Close()
}

func (p *Parser) getCaptureTime(packet gopacket.Packet) int64 {
	captureTime := packet.Metadata().CaptureInfo.Timestamp.UnixNano() / int64(time.Microsecond)
	if p.epochTime == -1 {
		p.epochTime = captureTime
	}
	return captureTime - p.epochTime
}

func (p *Parser) isValid(packet gopacket.Packet) bool {
	if packet.Metadata().Truncated {
		return false
	}
	if packet.ApplicationLayer() == nil {
		return false
	}
	n := packet.NetworkLayer().NetworkFlow()
	t := packet.TransportLayer().TransportFlow()
	a := match(p.srcAddr, n.Src().String())
	b := match(p.srcPort, t.Src().String())
	c := match(p.dstAddr, n.Dst().String())
	d := match(p.dstPort, t.Dst().String())
	return a && b && c && d
}

func match(x, y string) bool {
	return x == "" || x == y
}
