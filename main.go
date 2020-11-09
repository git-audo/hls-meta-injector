package main

import (
	"flag"
	"fmt"
	"os"

	"./parser"
)

var (
	verbose  = flag.Bool("v", false, "show fragment file informations")
	filename = flag.String("i", "", "input fragment file")
)

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

	packetsNum := int(stat.Size()) / parser.PacketSize

	for i := 0; i < packetsNum; i++ {
		buff := make([]byte, parser.PacketSize)
		r, err := f.Read(buff)
		if err != nil || r != parser.PacketSize {
			fmt.Errorf("error reading packet, %s", err)
		}

		p := parser.NewPacket(buff)
		p.ParseHeader(buff)

		if p.Pid() == 0 {
			// pat packet
			pmtPid = ((uint16(buff[15]) & 0x1f) << 8) | uint16(buff[16])
		} else if p.Pid() == pmtPid {
			// pmt packet
			p.ParsePmt(buff)
			// pmt.NewMetaStream(102)
		} else {
			// pes packet
		}
	}
	
	/*
		if *verbose {
		fmt.Println(" Total number of packets", packetsNum)
			for k, v := range streamsPacketsCount {
				fmt.Printf(" stream %v packets %v\n", k, v)
			}
		}
	*/
}
