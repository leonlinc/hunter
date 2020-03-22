package m2ts

import (
	"fmt"
	"log"
	"os"
	"path"
)

type CCAnalyzer struct {
	root    string
	errStat map[int]int
	counter map[int]int
}

func CreateCCAnalyzer(root string) *CCAnalyzer {
	return &CCAnalyzer{
		root:    root,
		errStat: make(map[int]int),
		counter: make(map[int]int),
	}
}

func (p *CCAnalyzer) Process(index int64, pkt *Packet) {
	if pkt.PID == NULL {
		return
	}
	if cc, ok := p.counter[pkt.PID]; ok {
		if cc != pkt.continuity_counter {
			p.errStat[pkt.PID] += 1
		}
	}
	p.counter[pkt.PID] = pkt.NextCC()
}

func (p *CCAnalyzer) Close() {
	logFile, err := os.Create(path.Join(p.root, "m2ts-cc.log"))
	if err != nil {
		log.Fatal(err)
	}
	defer logFile.Close()
	fmt.Fprintln(logFile, "PID", "ErrCount")
	for pid, cnt := range p.errStat {
		fmt.Fprintln(logFile, pid, cnt)
	}
}

type Metadata struct {
	PacketIndex int64
	CaptureTime int64
	MuxerTime   int64
}

type PCRAnalyzer struct {
	logFile *os.File
}

func CreatePCRAnalyzer(root string) *PCRAnalyzer {
	logFile, err := os.Create(path.Join(root, "m2ts-timing.log"))
	if err != nil {
		log.Fatal(err)
	}
	fmt.Fprintln(logFile, "Index", "PID", "PCR", "PacketIndex", "CaptureTime")
	return &PCRAnalyzer{logFile}
}

func (p *PCRAnalyzer) Process(index int64, pkt *Packet, metadata Metadata) {
	pcr := pkt.PCR()
	if pcr != -1 {
		fmt.Fprintln(p.logFile, index, pkt.PID, pcr, metadata.PacketIndex, metadata.CaptureTime)
	}
}

func (p *PCRAnalyzer) Close() {
	p.logFile.Close()
}
