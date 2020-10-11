package main

import (
	"os"
	"fmt"
	"flag"
)

var (
	packetSize = 188
	headerSize = 4

	verbose = flag.Bool("v", false, "show fragment file informations")
	filename = flag.String("i", "", "input fragment file")
)

type Packet struct {
	buff []byte	

	pid          uint16
	adaptationFieldControl uint8
	adaptationFieldLength uint8
	cc           uint8
}

func NewPacket() *Packet {
	p := new(Packet)
	p.buff = make([]byte, packetSize)
	return p
}

func (p *Packet) ReadPacket(buf []byte) {
	p.buff = buf
	p.pid = ((uint16(buf[1]) & 0x1f) << 8) | uint16(buf[2])
}

func (p *Packet) ParseHeader(buf []byte) {
	p.pid = ((uint16(buf[1]) & 0x1f) << 8) | uint16(buf[2])
	p.adaptationFieldControl = (uint8(buf[3]) & 0x30)
	p.cc = (uint8(buf[3]) & 0x0f)		
}
	
func main() {
	flag.Parse()
	
	f, err := os.Open(*filename)
	if err != nil {
		fmt.Errorf("error opening file, %s", err)
	}

	stat, err := f.Stat()
	if err != nil {
		return
	}

	var pmtPid uint16
	streamsPacketsCount := make(map[uint16]int)

	packetsNum := int(stat.Size())/packetSize
	
	for i:=0; i<packetsNum; i++ {
		p := NewPacket()		
		buff := make([]byte, packetSize)
		r, err := f.Read(buff)
		if err != nil || r != packetSize {
			fmt.Errorf("error reading packet, %s", err)
		}

		p.ParseHeader(buff)

		if p.adaptationFieldControl != 0x01 && p.pid == 0 {
			p.adaptationFieldLength = uint8(buff[4])
		}

		// p.pid = ((uint16(buff[1]) & 0x1f) << 8) | uint16(buff[2])
		// cc := uint8(packet[3]) & 0xf

		if p.pid == 0 {
			// pat packet
			// sl := ((uint16(packet[8]) & 0x3) << 8) | uint16(packet[9])
			pmtPid = ((uint16(buff[15]) & 0x1f) << 8) | uint16(buff[16])
		} else if p.pid == pmtPid {
			// pmt packet
			s1 := ((uint16(buff[18]) & 0x3) << 8) | uint16(buff[19])
			s2 := ((uint16(buff[23]) & 0x3) << 8) | uint16(buff[24])
			streamsPacketsCount[s1] = 0
			streamsPacketsCount[s2] = 0
		} else {
			// pes packet
			streamsPacketsCount[p.pid] += 1
		}
	}

	if *verbose {
	fmt.Println(" Total number of packets", packetsNum)
		for k, v := range streamsPacketsCount {
			fmt.Printf(" stream %v packets %v\n", k, v)
		}
	}
}
