package parser

//import "hash/crc32"

var (
	PacketSize = 188
)

type Packet struct {
	buff	    []byte	

	header	    *Header
	tableHeader *TableHeader
	tableSyntax *TableSyntax
	pmt	    *Pmt
	
	// TODO: add adaptation field
}

type Header struct {
	TransportErrorIndicator bool
	PayloadUnitStart	bool
	TransportPriority	bool
	Pid 			uint16
	TransportScrambling	uint8
	AdaptationFieldControl  uint8
	ContinuityCounter	uint8
	
	PointerField		uint8	
}

type TableHeader struct {
	TableId 	       uint8
	SectionSyntaxIndicator bool
	Private	      	       bool
	Reserved  	       uint8
	SectionLengthUnused    uint8
	SectionLength	       uint16

	SyntaxSection 	       *TableSyntax
}

type TableSyntax struct {
	TableIdExtension  uint16
	Reserved 	  uint8
	VersionNumber	  uint8
	Indicator	  bool
	SectionNumber	  uint8
	LastSectionNumber uint8
}

type Pmt struct {
	Reserved2		uint8
	PcrPid                  uint16
	Reserved3		uint8
	ProgramInfoLengthUnused uint8
	ProgramInfoLength       uint16
	
	ProgramDescriptors      []byte
	ElementaryStreams       []ElementaryStream
	CRC32                   uint32	
}

type ElementaryStream struct {
	StreamType     	   uint8
	Reserved3	   uint8
	ElementaryPid  	   uint16
	Reserved4 	   uint8
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

/*
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
*/
func (p *Packet) ParseHeader(buf []byte) {
	h := new(Header)
	h.Pid = ((uint16(buf[1]) & 0x1f) << 8) | uint16(buf[2])
	h.AdaptationFieldControl = (uint8(buf[3]) & 0x30)
	h.ContinuityCounter = (uint8(buf[3]) & 0x0f)

	p.header = h
}

func (p *Packet) ParsePmt(buf []byte) {
	tH := new(TableHeader)
	tH.TableId = uint8(buf[5])
	tH.SectionLength = ((uint16(buf[6]) & 0x03) << 8) | uint16(buf[7])

	tS := new(TableSyntax)

	pmt := new(Pmt)
	pmt.PcrPid = ((uint16(buf[12]) & 0x1f) << 8) | uint16(buf[13])
	pmt.ProgramInfoLength = ((uint16(buf[15]) & 0x03) << 8) | uint16(buf[16])

	remainingBytes := int32(tH.SectionLength - 13)
	for i:=0 ; remainingBytes > 0 ; i++ {
		es := new(ElementaryStream)
		es.StreamType = uint8(buf[17])
		es.ElementaryPid = ((uint16(buf[18+i*5]) & 0x3) << 8) | uint16(buf[19+i*5])
		es.EsInfoLength = ((uint16(buf[20+i*5]) & 0x3) << 8) | uint16(buf[21+i*5])
		pmt.ElementaryStreams = append(pmt.ElementaryStreams, *es)
		remainingBytes = remainingBytes - 5 - int32(es.EsInfoLength)
	}

	p.tableHeader = tH
	p.tableSyntax = tS
	p.pmt = pmt
}

func (p *Packet) Pid() uint16 {
	return p.header.Pid
}

func (p *Packet) NewES(pid uint16) {
	es := new(ElementaryStream)
	es.StreamType = 0x15
	es.Reserved3 = 0x7
	es.ElementaryPid = pid
	es.Reserved4 = 0xf
	es.EsInfoLength = 15
	es.Descriptors = []byte("260DFFFF49443320FF49443320000F")

	p.pmt.ElementaryStreams = append(p.pmt.ElementaryStreams, *es)
}
