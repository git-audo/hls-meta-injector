package main

import (
	"os"
	"fmt"
)

var (
	packetSize = 188
)

func main() {

	f, err := os.Open("./frag.ts")
	if err != nil {
		fmt.Errorf("error opening file, %s", err)
	}

	var pmtPid uint16
	var streamPids []uint16
	
	for i:=0; i<5; i++ {
		packet := make([]byte, packetSize)
		r, err := f.Read(packet)
		if err != nil || r != packetSize {
			fmt.Errorf("error reading packet, %s", err)
		}

		pid := ((uint16(packet[1]) & 0x1f) << 8) | uint16(packet[2])
		cc := uint8(packet[3]) & 0xf
		fmt.Println("pid", pid, "cc", cc)

		if pid == 0 {
			sl := ((uint16(packet[8]) & 0x3) << 8) | uint16(packet[9])
			pmtPid = ((uint16(packet[15]) & 0x1f) << 8) | uint16(packet[16])
			fmt.Println(" sl", sl, "pmt", pmtPid)			
		}

		if pid == pmtPid {
			streamPids = append(streamPids, ((uint16(packet[18]) & 0x3) << 8) | uint16(packet[19]))
			streamPids = append(streamPids, ((uint16(packet[23]) & 0x3) << 8) | uint16(packet[24]))
			fmt.Println("streams", streamPids)
		}
	}
}
