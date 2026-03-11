package codec

import (
	"encoding/binary"
	"fmt"
)

// Reader provides typed reads over a packet body payload.
type Reader struct {
	// payload stores packet body bytes.
	payload []byte
	// offset stores the current read cursor.
	offset int
}

// NewReader creates a packet body reader.
func NewReader(payload []byte) *Reader {
	return &Reader{payload: payload}
}

// ReadInt32 reads one big-endian int32 from payload.
func (reader *Reader) ReadInt32() (int32, error) {
	if reader.offset+4 > len(reader.payload) {
		return 0, fmt.Errorf("insufficient bytes for int32")
	}
	value := int32(binary.BigEndian.Uint32(reader.payload[reader.offset : reader.offset+4]))
	reader.offset += 4
	return value, nil
}

// ReadUint16 reads one big-endian uint16 from payload.
func (reader *Reader) ReadUint16() (uint16, error) {
	if reader.offset+2 > len(reader.payload) {
		return 0, fmt.Errorf("insufficient bytes for uint16")
	}
	value := binary.BigEndian.Uint16(reader.payload[reader.offset : reader.offset+2])
	reader.offset += 2
	return value, nil
}

// ReadBool reads one byte and maps zero=false, non-zero=true.
func (reader *Reader) ReadBool() (bool, error) {
	if reader.offset+1 > len(reader.payload) {
		return false, fmt.Errorf("insufficient bytes for bool")
	}
	value := reader.payload[reader.offset] != 0
	reader.offset++
	return value, nil
}

// ReadString reads uint16 length-prefixed UTF-8 bytes.
func (reader *Reader) ReadString() (string, error) {
	length, err := reader.ReadUint16()
	if err != nil {
		return "", err
	}
	if reader.offset+int(length) > len(reader.payload) {
		return "", fmt.Errorf("insufficient bytes for string")
	}
	value := string(reader.payload[reader.offset : reader.offset+int(length)])
	reader.offset += int(length)
	return value, nil
}

// Remaining returns unread payload bytes.
func (reader *Reader) Remaining() int {
	return len(reader.payload) - reader.offset
}

// Writer provides typed writes for packet body payload composition.
type Writer struct {
	// payload stores composed packet body bytes.
	payload []byte
}

// NewWriter creates a packet body writer.
func NewWriter() *Writer {
	return &Writer{payload: make([]byte, 0, 32)}
}

// WriteInt32 appends one big-endian int32 to payload.
func (writer *Writer) WriteInt32(value int32) {
	buffer := make([]byte, 4)
	binary.BigEndian.PutUint32(buffer, uint32(value))
	writer.payload = append(writer.payload, buffer...)
}

// WriteUint16 appends one big-endian uint16 to payload.
func (writer *Writer) WriteUint16(value uint16) {
	buffer := make([]byte, 2)
	binary.BigEndian.PutUint16(buffer, value)
	writer.payload = append(writer.payload, buffer...)
}

// WriteBool appends one boolean byte to payload.
func (writer *Writer) WriteBool(value bool) {
	if value {
		writer.payload = append(writer.payload, 1)
		return
	}
	writer.payload = append(writer.payload, 0)
}

// WriteString appends uint16 length-prefixed UTF-8 bytes.
func (writer *Writer) WriteString(value string) error {
	if len(value) > 65535 {
		return fmt.Errorf("string length exceeds uint16 max")
	}
	writer.WriteUint16(uint16(len(value)))
	writer.payload = append(writer.payload, []byte(value)...)
	return nil
}

// Bytes returns composed payload bytes.
func (writer *Writer) Bytes() []byte {
	return writer.payload
}
