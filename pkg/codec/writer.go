package codec

import "encoding/binary"

// Writer writes primitive values using the protocol wire encoding.
type Writer struct {
	buf []byte
}

// NewWriter creates an empty protocol writer.
func NewWriter(capacity int) *Writer {
	return &Writer{buf: make([]byte, 0, capacity)}
}

// Bytes returns the accumulated encoded payload bytes.
func (w *Writer) Bytes() []byte {
	return w.buf
}

// WriteBool appends a boolean value as uint8 (0/1).
func (w *Writer) WriteBool(value bool) {
	if value {
		w.buf = append(w.buf, 1)
		return
	}
	w.buf = append(w.buf, 0)
}

// WriteInt32 appends a signed int32 in big-endian format.
func (w *Writer) WriteInt32(value int32) {
	w.WriteUint32(uint32(value))
}

// WriteUint16 appends a uint16 in big-endian format.
func (w *Writer) WriteUint16(value uint16) {
	var scratch [2]byte
	binary.BigEndian.PutUint16(scratch[:], value)
	w.buf = append(w.buf, scratch[:]...)
}

// WriteUint32 appends a uint32 in big-endian format.
func (w *Writer) WriteUint32(value uint32) {
	var scratch [4]byte
	binary.BigEndian.PutUint32(scratch[:], value)
	w.buf = append(w.buf, scratch[:]...)
}

// WriteString appends a UTF-8 string with uint16 length prefix.
func (w *Writer) WriteString(value string) {
	w.WriteUint16(uint16(len(value)))
	w.buf = append(w.buf, value...)
}

// WriteBytes appends raw bytes.
func (w *Writer) WriteBytes(value []byte) {
	w.buf = append(w.buf, value...)
}
