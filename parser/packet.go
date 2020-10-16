package parser

var (
	PacketSize = 188
)

type Packet struct {
	buff []byte	

	pid          uint16
	adaptationFieldControl uint8
	adaptationFieldLength uint8
	cc           uint8
}

// type Descriptor struct {
// 	descriptorTag uint8
// 	descriptorLength uint8
// 	descriptorContent []byte
// }

type Stream struct {
	streamType    uint8
	elementaryPid uint16
	esInfoLength  uint16
	descriptorList []byte
}

type Pmt struct {
	tableID                uint8
	sectionSyntaxIndicator uint8
	sectionLength          uint16

	programNumber          uint16

	pcrPid                 uint16
	programInfoLength      uint16
	
	programDescriptors     []byte
	elementaryStreams      []Stream
	crc32                  uint32	
}

func NewPacket() *Packet {
	p := new(Packet)
	p.buff = make([]byte, PacketSize)
	return p
}

func NewPmt() *Pmt {
	p := new(Pmt)
	return p
}

func (pmt *Pmt) NewMetaStream(epid uint16) {
	s := new(Stream)
	s.streamType = 21
	s.elementaryPid = epid
	s.esInfoLength = 15
	s.descriptorList = []byte("FFFF49443320FF49443320000F")

	pmt.elementaryStreams = append(pmt.elementaryStreams, *s)
	println(pmt)
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
		remainingBytes = remainingBytes - 5 - int32(stream.esInfoLength)
	}
}

func (p *Packet) Pid() uint16 {
	return p.pid
}

func (p *Packet) Afc() uint8 {
	return p.adaptationFieldControl
}
