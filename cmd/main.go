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

type Descriptor struct {
	descriptorTag uint8
	descriptorLength uint8
	descriptorContent []byte
}

type Stream struct {
	streamType    uint8
	elementaryPid uint16
	esInfoLength  uint16
	descriptorList []Descriptor
}

type Pmt struct {
	tableID                uint8
	sectionSyntaxIndicator uint8
	sectionLength          uint16

	programNumber          uint16

	pcrPid                 uint16
	programInfoLength      uint16
	
	programDescriptors     []Descriptor
	elementaryStreams      []Stream
	crc32                  uint32	
}

func NewPacket() *Packet {
	p := new(Packet)
	p.buff = make([]byte, packetSize)
	return p
}

func NewPmt() *Pmt {
	p := new(Pmt)
	return p
}

func (p *Packet) ParseHeader(buf []byte) {
	p.pid = ((uint16(buf[1]) & 0x1f) << 8) | uint16(buf[2])
	p.adaptationFieldControl = (uint8(buf[3]) & 0x30)
	p.cc = (uint8(buf[3]) & 0x0f)
}

func (pmt *Pmt) ParsePmt(buf []byte) {
	pmt.tableID = uint8(buf[5])
	pmt.sectionLength = ((uint16(buf[6]) & 0x03) << 8) | uint16(buf[7])
	pmt.pcrPid = ((uint16(buf[12]) & 0x1f) << 8) | uint16(buf[13])
	pmt.programInfoLength = ((uint16(buf[15]) & 0x03) << 8) | uint16(buf[16])
	remainingBytes := int32(pmt.sectionLength - 13)
	for i:=0 ; remainingBytes > 0 ; i++ {
		stream := new(Stream)
		stream.streamType = uint8(buf[17])
		stream.elementaryPid = ((uint16(buf[18+i*5]) & 0x3) << 8) | uint16(buf[19+i*5])
		stream.esInfoLength = ((uint16(buf[20+i*5]) & 0x3) << 8) | uint16(buf[21+i*5])
		pmt.elementaryStreams = append(pmt.elementaryStreams, *stream)
		println(stream.elementaryPid)
		remainingBytes = remainingBytes - 5 - int32(stream.esInfoLength)
	}
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
		pmt := NewPmt()
		buff := make([]byte, packetSize)
		r, err := f.Read(buff)
		if err != nil || r != packetSize {
			fmt.Errorf("error reading packet, %s", err)
		}

		p.ParseHeader(buff)
		if p.adaptationFieldControl != 0x01 && p.pid == 0 {
			p.adaptationFieldLength = uint8(buff[4])
		}

		if p.pid == 0 {
			// pat packet
			// sl := ((uint16(packet[8]) & 0x3) << 8) | uint16(packet[9])
			pmtPid = ((uint16(buff[15]) & 0x1f) << 8) | uint16(buff[16])
		} else if p.pid == pmtPid {
			// pmt packet
			pmt.ParsePmt(buff)
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
