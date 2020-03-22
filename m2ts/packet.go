package m2ts

import (
	_ "log"
)

const (
	NULL int = 8191
)

type Packet struct {
	sync_byte                    int
	transport_error_indicator    bool
	payload_unit_start_indicator bool
	transport_priority           bool
	PID                          int
	transport_scrambling_control int
	adaptation_field_control     int
	continuity_counter           int
	adaptation_field
	data_byte []byte
}

type adaptation_field struct {
	adaptation_field_length              int
	discontinuity_indicator              bool
	random_access_indicator              bool
	elementary_stream_priority_indicator bool
	PCR_flag                             bool
	OPCR_flag                            bool
	splicing_point_flag                  bool
	transport_private_data_flag          bool
	adaptation_field_extension_flag      bool
	program_clock_reference_base         int64
	program_clock_reference_extension    int64
	private_data_byte                    []byte
}

func (pkt *Packet) PCR() int64 {
	if pkt.PCR_flag {
		return pkt.program_clock_reference_base*300 + pkt.program_clock_reference_extension
	} else {
		return -1
	}
}

func (pkt *Packet) NextCC() int {
	if pkt.adaptation_field_control == 0 || pkt.adaptation_field_control == 2 {
		return pkt.continuity_counter
	} else {
		return (pkt.continuity_counter + 1) % 16
	}
}

func packet(data []byte) *Packet {
	r := &Reader{Data: data}
	pkt := &Packet{}
	pkt.sync_byte = r.ReadBits(8)
	pkt.transport_error_indicator = r.ReadBit()
	pkt.payload_unit_start_indicator = r.ReadBit()
	pkt.transport_priority = r.ReadBit()
	pkt.PID = r.ReadBits(13)
	pkt.transport_scrambling_control = r.ReadBits(2)
	pkt.adaptation_field_control = r.ReadBits(2)
	pkt.continuity_counter = r.ReadBits(4)
	if pkt.adaptation_field_control == 2 || pkt.adaptation_field_control == 3 {
		pkt.adaptation_field_length = r.ReadBits(8)
		if pkt.adaptation_field_length > 0 {
			pkt.discontinuity_indicator = r.ReadBit()
			pkt.random_access_indicator = r.ReadBit()
			pkt.elementary_stream_priority_indicator = r.ReadBit()
			pkt.PCR_flag = r.ReadBit()
			pkt.OPCR_flag = r.ReadBit()
			pkt.splicing_point_flag = r.ReadBit()
			pkt.transport_private_data_flag = r.ReadBit()
			pkt.adaptation_field_extension_flag = r.ReadBit()
			if pkt.PCR_flag {
				pkt.program_clock_reference_base, pkt.program_clock_reference_extension = r.ReadPCR()
			}
			if pkt.OPCR_flag {
				_, _ = r.ReadPCR()
			}
			if pkt.splicing_point_flag {
				r.ReadBits(8)
			}
			if pkt.transport_private_data_flag {
				transport_private_data_length := r.ReadBits(8)
				pkt.private_data_byte = r.Data[r.Base : r.Base+transport_private_data_length]
			}
		}
		pkt.data_byte = data[4+1+pkt.adaptation_field_length:]
	} else {
		pkt.data_byte = data[4:]
	}
	return pkt
}

func Packets(data []byte) chan *Packet {
	out := make(chan *Packet)
	go func() {
		for i := 0; i+188 <= len(data); i += 188 {
			out <- packet(data[i : i+188])
		}
		close(out)
	}()
	return out
}
