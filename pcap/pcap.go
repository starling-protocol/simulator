package pcap

import (
	"encoding/binary"
	"encoding/hex"
	"os"
	"strings"
	"time"
)

type PCAPFile struct {
	buffer []byte
}

func NewPCAPFile(timeOffset time.Time) *PCAPFile {

	pcap := PCAPFile{
		buffer: []byte{},
	}

	// Reference for file format: https://wiki.wireshark.org/Development/LibpcapFileFormat
	buf := []byte{}
	buf = binary.LittleEndian.AppendUint32(buf, 0xa1b2c3d4) // Magic number
	buf = binary.LittleEndian.AppendUint16(buf, 0x02)       // Major version
	buf = binary.LittleEndian.AppendUint16(buf, 0x04)       // Minor version
	buf = binary.LittleEndian.AppendUint32(buf, 0x0)        // Not used
	buf = binary.LittleEndian.AppendUint32(buf, 0x0)        // Not used
	buf = binary.LittleEndian.AppendUint32(buf, 0xffff)     // Max length of packets
	buf = binary.LittleEndian.AppendUint32(buf, 0x01)       //0xfb)       // Data link type (in this case bluetooth LE: https://www.tcpdump.org/linktypes.html)

	pcap.buffer = append(pcap.buffer, buf...)

	return &pcap
}

func (p *PCAPFile) WriteFile(name string) error {
	if !strings.HasSuffix(name, ".pcap") {
		name += ".pcap"
	}
	f, err := os.Create(name)
	if err != nil {
		return err
	}

	defer f.Close()

	_, err = f.Write(p.buffer)
	if err != nil {
		return err
	}

	return nil
}

func (p *PCAPFile) AddPacket(packet []byte, timestamp time.Time, destination string, source string) {

	dst, err := hex.DecodeString(destination)
	if err != nil {
		panic("error decoding destination")
	}
	src, err := hex.DecodeString(source)
	if err != nil {
		panic("error decoding source")
	}

	// Construct content
	content := []byte{}
	content = binary.LittleEndian.AppendUint32(content, binary.LittleEndian.Uint32(dst[0:4])) // "Ethernet" destination  //TODO: Implement sender and receiver properly
	content = binary.LittleEndian.AppendUint16(content, binary.LittleEndian.Uint16(dst[4:6]))

	content = binary.LittleEndian.AppendUint32(content, binary.LittleEndian.Uint32(src[0:4])) // "Ethernet" source
	content = binary.LittleEndian.AppendUint16(content, binary.LittleEndian.Uint16(src[4:6]))

	content = binary.LittleEndian.AppendUint16(content, 0x0) // Non-descript protocol for data

	content = append(content, packet...) // Add actual packet

	// Construct header
	header := []byte{}
	header = binary.LittleEndian.AppendUint32(header, uint32(timestamp.Unix()))                                 // TODO: Validate that this is correct
	header = binary.LittleEndian.AppendUint32(header, uint32(timestamp.UnixMicro()-(timestamp.Unix()*1000000))) // Microseconds
	header = binary.LittleEndian.AppendUint32(header, uint32(len(content)))                                     // Size
	header = binary.LittleEndian.AppendUint32(header, uint32(len(content)))                                     // Size

	p.buffer = append(p.buffer, header...)
	p.buffer = append(p.buffer, content...)
}
