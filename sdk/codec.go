package sdk

import (
	"encoding/binary"
	"errors"
	"io"
)

var errInsufficientData = errors.New("insufficient data")

// Reader reads Habbo protocol primitive types from a byte slice.
type Reader struct {
	data   []byte
	offset int
}

// NewReader creates a protocol reader from raw bytes.
func NewReader(data []byte) *Reader {
	return &Reader{data: data}
}

// ReadInt32 reads a big-endian 32-bit integer.
func (r *Reader) ReadInt32() (int32, error) {
	if r.offset+4 > len(r.data) {
		return 0, errInsufficientData
	}
	v := int32(binary.BigEndian.Uint32(r.data[r.offset:]))
	r.offset += 4
	return v, nil
}

// ReadString reads a length-prefixed UTF-8 string.
func (r *Reader) ReadString() (string, error) {
	if r.offset+2 > len(r.data) {
		return "", errInsufficientData
	}
	length := int(binary.BigEndian.Uint16(r.data[r.offset:]))
	r.offset += 2
	if r.offset+length > len(r.data) {
		return "", errInsufficientData
	}
	s := string(r.data[r.offset : r.offset+length])
	r.offset += length
	return s, nil
}

// ReadBool reads a single boolean byte.
func (r *Reader) ReadBool() (bool, error) {
	if r.offset >= len(r.data) {
		return false, io.EOF
	}
	v := r.data[r.offset] != 0
	r.offset++
	return v, nil
}

// Remaining returns unread byte count.
func (r *Reader) Remaining() int {
	if r.offset >= len(r.data) {
		return 0
	}
	return len(r.data) - r.offset
}

// Writer builds Habbo protocol primitive types into a byte slice.
type Writer struct {
	buf []byte
}

// NewWriter creates a protocol writer.
func NewWriter() *Writer {
	return &Writer{}
}

// WriteInt32 appends a big-endian 32-bit integer.
func (w *Writer) WriteInt32(v int32) {
	b := make([]byte, 4)
	binary.BigEndian.PutUint32(b, uint32(v))
	w.buf = append(w.buf, b...)
}

// WriteString appends a length-prefixed UTF-8 string.
func (w *Writer) WriteString(v string) {
	b := make([]byte, 2)
	binary.BigEndian.PutUint16(b, uint16(len(v)))
	w.buf = append(w.buf, b...)
	w.buf = append(w.buf, []byte(v)...)
}

// WriteBool appends a single boolean byte.
func (w *Writer) WriteBool(v bool) {
	if v {
		w.buf = append(w.buf, 1)
	} else {
		w.buf = append(w.buf, 0)
	}
}

// Bytes returns the accumulated byte slice.
func (w *Writer) Bytes() []byte {
	return w.buf
}
