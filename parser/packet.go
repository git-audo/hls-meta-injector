package parser

import "hash/crc32"

var (
	PacketSize = 188
)

type Packet struct {
	buff []byte	

	h	 Header
	p	 []byte
	
	// TODO: add adaptation field
}

type Header struct {
	TransportErrorIndicator bool
	PayloadUnitStart		bool
	TransportPriority		bool
	Pid 					uint16
	TransportScrambling		uint8
	AdaptationFieldControl  uint8
	ContinuityCounter		uint8
 }

type TableHeader struct {
	TableId 			   uint8
	SectionSyntaxIndicator bool
	Private				   bool
	Reserved			   uint8
	SectionLengthUnused	   uint8
	SectionLength		   uint16

	SyntaxSection 		   TableSyntax
}

type TableSyntax struct {
	TableIdExtension  uint16
	Reserved		  uint8
	VersionNumber	  uint8
	Indicator		  bool
	SectionNumber	  uint8
	LastSectionNumber uint8
}

type Pmt struct {
	Reserved2				uint8
	PcrPid                  uint16
	Reserved3				uint8
	ProgramInfoLengthUnused uint8
	ProgramInfoLength       uint16
	
	ProgramDescriptors      []byte
	ElementaryStreams       []ElementaryStream
	CRC32                   uint32	
}

type ElementaryStream struct {
	StreamType     	   uint8
	Reserved3	       uint8
	ElementaryPid  	   uint16
	Reserved4	   	   uint8
	EsInfoLengthUnused uint8
	EsInfoLength	   uint16
	Descriptors 	   []byte
}

func NewPacket(buff []byte) *Packet {
	p := new(Packet)
//	p.buff = make([]byte, PacketSize)
	p.buff = buff
	return p
}

func NewPmt() *Pmt {
	pmt := new(Pmt)
	return pmt
}

func (pmt *Pmt) NewMetaElementaryStream(epid uint16) {
	pmt.sectionLength += 20
	pmt.programInfoLength += 17
	pd := []byte("250FFFFF49443320FF49443320001F0001")
	
	s := new(ElementaryStream)
	s.streamType = 21
	s.elementaryPid = epid
	s.esInfoLength = 15
	s.descriptorList = []byte("FFFF49443320FF49443320000F")

	pmt.programDescriptors = append(pmt.programDescriptors, pd...)
	pmt.elementaryStreams = append(pmt.elementaryStreams, *s)
	pmt.crc32 = crc32.ChecksumIEEE([]byte("FFFFFFFF"))
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
		stream := new(ElementaryStream)
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
