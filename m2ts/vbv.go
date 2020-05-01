package main

import (
	"fmt"
	"github.com/leonlinc/mpts"
)

type Record struct {
	idx int64
	pid int
	pcr int64
	dts int64
}

// func main() {
// 	// parse("UHD statmux PCR_00002_20190528122410-PCR.ts")
// 	parse("out.ts")
// }

func parse(file string) {
	var idx int64
	var records []Record

	pkts := mpts.ParseFile(file)

	for pkt := range pkts {
		var pcr int64
		var dts int64

		if pkt.Pid == 3472 {
			if currPcr, ok := pkt.PCR(); ok {
				num := int64(len(records))
				if num > 0 {
					first := records[0]
					if first.pcr != -1 {
						for _, r := range records {
							r.pcr = first.pcr + (currPcr-first.pcr)*(r.idx-first.idx)/num
							if r.dts != -1 {
								fmt.Println(r.idx, r.pid, r.pcr, r.dts*300)
							}
							// debug
							if false {
								fmt.Println(r.idx, r.pid, r.pcr, r.dts, r.pcr == first.pcr)
							}
						}
					}
				}
				records = nil

				pcr = currPcr
			} else {
				pcr = -1
			}
		}

		if pkt.Pid == 3472 && pkt.PUSI == 1 {
			var pes mpts.PesPkt
			pes.Read(pkt.Data)
			dts = pes.Dts
			if dts == 0 {
				dts = pes.Pts
			}
		} else {
			dts = -1
		}

		r := Record{idx, pkt.Pid, pcr, dts}
		records = append(records, r)

		idx += 1
	}
}
