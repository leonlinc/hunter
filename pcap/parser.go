package pcap

import (
	"fmt"
	"log"
	"os"
	"path"
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
	timingLog *os.File
}

func CreateParser(root string) *Parser {
	timingLog, err := os.Create(path.Join(root, "pcap-timing.log"))
	if err != nil {
		log.Fatal(err)
	}
	fmt.Fprintln(timingLog, "Index", "CaptureTime", "Length", "PID", "PCR")
	return &Parser{
		epochTime: -1,
		timingLog: timingLog,
	}
}

func (p *Parser) Close() {
	p.timingLog.Close()
}

func (p *Parser) Parse(file string) {
	pcapFile, err := pcap.OpenOffline(file)
	if err != nil {
		log.Fatal(err)
	}
	defer pcapFile.Close()

	var idx int64
	source := gopacket.NewPacketSource(pcapFile, pcapFile.LinkType())
	for packet := range source.Packets() {
		captureTime := p.getCaptureTime(packet)
		if p.isValid(packet) {
			payload := packet.ApplicationLayer().Payload()
			for pkt := range m2ts.Packets(payload) {
				pcr := pkt.PCR()
				if pcr != -1 {
					fmt.Fprintln(p.timingLog, idx, captureTime, len(payload), pkt.PID, pcr)
				}
			}
		}
		idx += 1
	}
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
