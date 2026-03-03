package codec

import (
	"encoding/binary"
	"math"
)

// Writer accumulates Pixel Protocol primitives into an internal buffer.
// Call Bytes() to retrieve the encoded payload, or Frame() to prepend
// the length+header frame.
type Writer struct {
	buf []byte
}

// NewWriter creates a Writer with an optional initial capacity hint.
func NewWriter(cap int) *Writer {
	if cap <= 0 {
		cap = 64
	}
	return &Writer{buf: make([]byte, 0, cap)}
}

// Reset clears the buffer for reuse.
func (w *Writer) Reset() {
	w.buf = w.buf[:0]
}

// Len returns the current byte count.
func (w *Writer) Len() int {
	return len(w.buf)
}

// Bytes returns the accumulated bytes (not a copy — do not mutate).
func (w *Writer) Bytes() []byte {
	return w.buf
}

// WriteBool writes a single byte: 1 for true, 0 for false.
func (w *Writer) WriteBool(v bool) {
	if v {
		w.buf = append(w.buf, 1)
	} else {
		w.buf = append(w.buf, 0)
	}
}

// WriteInt16 writes a big-endian signed 16-bit integer.
func (w *Writer) WriteInt16(v int16) {
	w.buf = binary.BigEndian.AppendUint16(w.buf, uint16(v))
}

// WriteUint16 writes a big-endian unsigned 16-bit integer.
func (w *Writer) WriteUint16(v uint16) {
	w.buf = binary.BigEndian.AppendUint16(w.buf, v)
}

// WriteInt32 writes a big-endian signed 32-bit integer.
func (w *Writer) WriteInt32(v int32) {
	w.buf = binary.BigEndian.AppendUint32(w.buf, uint32(v))
}

// WriteUint32 writes a big-endian unsigned 32-bit integer.
func (w *Writer) WriteUint32(v uint32) {
	w.buf = binary.BigEndian.AppendUint32(w.buf, v)
}

// WriteFloat64 writes a big-endian IEEE-754 64-bit float.
func (w *Writer) WriteFloat64(v float64) {
	w.buf = binary.BigEndian.AppendUint64(w.buf, math.Float64bits(v))
}

// WriteString writes a uint16-prefixed UTF-8 string.
func (w *Writer) WriteString(s string) {
	w.buf = binary.BigEndian.AppendUint16(w.buf, uint16(len(s)))
	w.buf = append(w.buf, s...)
}

// WriteBytes writes raw bytes with no prefix.
func (w *Writer) WriteBytes(data []byte) {
	w.buf = append(w.buf, data...)
}

// WriteListInt32 writes a uint32 count followed by that many int32 values.
func (w *Writer) WriteListInt32(vals []int32) {
	w.WriteUint32(uint32(len(vals)))
	for _, v := range vals {
		w.WriteInt32(v)
	}
}

// WriteListString writes a uint32 count followed by that many strings.
func (w *Writer) WriteListString(vals []string) {
	w.WriteUint32(uint32(len(vals)))
	for _, v := range vals {
		w.WriteString(v)
	}
}

// Frame prepends the Pixel Protocol wire frame header (uint32 length + uint16 headerID)
// to the current payload and returns the complete framed message.
// The length field equals len(headerID) + len(payload) = 2 + len(payload).
func (w *Writer) Frame(headerID uint16) []byte {
	payloadLen := len(w.buf)
	frameLen := 2 + payloadLen // header (2 bytes) + payload

	frame := make([]byte, 4+frameLen) // length prefix (4 bytes) + frame
	binary.BigEndian.PutUint32(frame[0:4], uint32(frameLen))
	binary.BigEndian.PutUint16(frame[4:6], headerID)
	copy(frame[6:], w.buf)
	return frame
}
