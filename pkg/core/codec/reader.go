package codec

import (
	"encoding/binary"
	"fmt"
	"io"
	"math"
)

// Reader reads Pixel Protocol primitives from a byte slice.
// It tracks an internal cursor and returns errors on underflow.
type Reader struct {
	buf []byte
	pos int
}

// NewReader creates a Reader over the given byte slice.
func NewReader(data []byte) *Reader {
	return &Reader{buf: data, pos: 0}
}

// Remaining returns the number of unread bytes.
func (r *Reader) Remaining() int {
	return len(r.buf) - r.pos
}

// need checks whether n bytes are available.
func (r *Reader) need(n int) error {
	if r.pos+n > len(r.buf) {
		return fmt.Errorf("codec: read underflow: need %d bytes, have %d", n, r.Remaining())
	}
	return nil
}

// ReadBool reads a single byte and returns false for 0, true otherwise.
func (r *Reader) ReadBool() (bool, error) {
	if err := r.need(1); err != nil {
		return false, err
	}
	v := r.buf[r.pos]
	r.pos++
	return v != 0, nil
}

// ReadInt16 reads a big-endian signed 16-bit integer.
func (r *Reader) ReadInt16() (int16, error) {
	if err := r.need(2); err != nil {
		return 0, err
	}
	v := int16(binary.BigEndian.Uint16(r.buf[r.pos:]))
	r.pos += 2
	return v, nil
}

// ReadUint16 reads a big-endian unsigned 16-bit integer.
func (r *Reader) ReadUint16() (uint16, error) {
	if err := r.need(2); err != nil {
		return 0, err
	}
	v := binary.BigEndian.Uint16(r.buf[r.pos:])
	r.pos += 2
	return v, nil
}

// ReadInt32 reads a big-endian signed 32-bit integer.
func (r *Reader) ReadInt32() (int32, error) {
	if err := r.need(4); err != nil {
		return 0, err
	}
	v := int32(binary.BigEndian.Uint32(r.buf[r.pos:]))
	r.pos += 4
	return v, nil
}

// ReadUint32 reads a big-endian unsigned 32-bit integer.
func (r *Reader) ReadUint32() (uint32, error) {
	if err := r.need(4); err != nil {
		return 0, err
	}
	v := binary.BigEndian.Uint32(r.buf[r.pos:])
	r.pos += 4
	return v, nil
}

// ReadFloat64 reads a big-endian IEEE-754 64-bit float (double).
func (r *Reader) ReadFloat64() (float64, error) {
	if err := r.need(8); err != nil {
		return 0, err
	}
	bits := binary.BigEndian.Uint64(r.buf[r.pos:])
	r.pos += 8
	return math.Float64frombits(bits), nil
}

// ReadString reads a uint16-prefixed UTF-8 string.
func (r *Reader) ReadString() (string, error) {
	length, err := r.ReadUint16()
	if err != nil {
		return "", fmt.Errorf("codec: reading string length: %w", err)
	}
	if err := r.need(int(length)); err != nil {
		return "", fmt.Errorf("codec: reading string body: %w", err)
	}
	s := string(r.buf[r.pos : r.pos+int(length)])
	r.pos += int(length)
	return s, nil
}

// ReadBytes reads all remaining bytes in the buffer.
func (r *Reader) ReadBytes() ([]byte, error) {
	data := make([]byte, r.Remaining())
	copy(data, r.buf[r.pos:])
	r.pos = len(r.buf)
	return data, nil
}

// ReadBytesN reads exactly n bytes.
func (r *Reader) ReadBytesN(n int) ([]byte, error) {
	if err := r.need(n); err != nil {
		return nil, err
	}
	data := make([]byte, n)
	copy(data, r.buf[r.pos:r.pos+n])
	r.pos += n
	return data, nil
}

// ReadListInt32 reads a uint32 count followed by that many int32 values.
func (r *Reader) ReadListInt32() ([]int32, error) {
	count, err := r.ReadUint32()
	if err != nil {
		return nil, fmt.Errorf("codec: reading list count: %w", err)
	}
	if err := r.need(int(count) * 4); err != nil {
		return nil, fmt.Errorf("codec: reading list body: %w", err)
	}
	result := make([]int32, count)
	for i := uint32(0); i < count; i++ {
		result[i] = int32(binary.BigEndian.Uint32(r.buf[r.pos:]))
		r.pos += 4
	}
	return result, nil
}

// ReadListString reads a uint32 count followed by that many strings.
func (r *Reader) ReadListString() ([]string, error) {
	count, err := r.ReadUint32()
	if err != nil {
		return nil, fmt.Errorf("codec: reading list count: %w", err)
	}
	result := make([]string, count)
	for i := uint32(0); i < count; i++ {
		s, err := r.ReadString()
		if err != nil {
			return nil, fmt.Errorf("codec: reading list element %d: %w", i, err)
		}
		result[i] = s
	}
	return result, nil
}

// ReadFrom reads all bytes from an io.Reader, resets the cursor, and returns bytes read.
func (r *Reader) ReadFrom(rd io.Reader) (int64, error) {
	data, err := io.ReadAll(rd)
	if err != nil {
		return 0, err
	}
	r.buf = data
	r.pos = 0
	return int64(len(data)), nil
}
