package main

import (
	"os"
	"fmt"
	"flag"
)

var (
	packetSize = 188

	verbose = flag.Bool("v", false, "show fragment file informations")
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
	streamsPacketsCount := make(map[uint16]int)

	packetsNum := int(stat.Size())/packetSize
	
	for i:=0; i<packetsNum; i++ {
		packet := make([]byte, packetSize)
		r, err := f.Read(packet)
		if err != nil || r != packetSize {
			fmt.Errorf("error reading packet, %s", err)
		}

		pid := ((uint16(packet[1]) & 0x1f) << 8) | uint16(packet[2])
		// cc := uint8(packet[3]) & 0xf

		if pid == 0 {
			// pat packet
			// sl := ((uint16(packet[8]) & 0x3) << 8) | uint16(packet[9])
			pmtPid = ((uint16(packet[15]) & 0x1f) << 8) | uint16(packet[16])
		} else if pid == pmtPid {
			// pmt packet
			s1 := ((uint16(packet[18]) & 0x3) << 8) | uint16(packet[19])
			s2 := ((uint16(packet[23]) & 0x3) << 8) | uint16(packet[24])
			streamsPacketsCount[s1] = 0
			streamsPacketsCount[s2] = 0
		} else {
			// pes packet
			streamsPacketsCount[pid] += 1
		}
	}

	if *verbose {
	fmt.Println(" Total number of packets", packetsNum)
		for k, v := range streamsPacketsCount {
			fmt.Printf(" stream %v packets %v\n", k, v)
		}
	}
}
