package codec

import "encoding/binary"

// Reader reads primitive values using the protocol wire encoding.
type Reader struct {
	buf []byte
	off int
}

// NewReader creates a reader over a byte slice.
func NewReader(body []byte) *Reader {
	return &Reader{buf: body}
}

// Remaining returns unread byte count.
func (r *Reader) Remaining() int {
	return len(r.buf) - r.off
}

// ReadBool reads a boolean encoded as uint8 (0/1).
func (r *Reader) ReadBool() (bool, error) {
	value, err := r.readByte()
	if err != nil {
		return false, err
	}
	return value != 0, nil
}

// ReadInt32 reads a signed int32 encoded as big-endian.
func (r *Reader) ReadInt32() (int32, error) {
	value, err := r.ReadUint32()
	return int32(value), err
}

// ReadUint16 reads a uint16 encoded as big-endian.
func (r *Reader) ReadUint16() (uint16, error) {
	if r.Remaining() < 2 {
		return 0, ErrUnexpectedEOF
	}
	value := binary.BigEndian.Uint16(r.buf[r.off : r.off+2])
	r.off += 2
	return value, nil
}

// ReadUint32 reads a uint32 encoded as big-endian.
func (r *Reader) ReadUint32() (uint32, error) {
	if r.Remaining() < 4 {
		return 0, ErrUnexpectedEOF
	}
	value := binary.BigEndian.Uint32(r.buf[r.off : r.off+4])
	r.off += 4
	return value, nil
}

// ReadString reads a UTF-8 string with uint16 length prefix.
func (r *Reader) ReadString() (string, error) {
	length, err := r.ReadUint16()
	if err != nil {
		return "", err
	}
	if r.Remaining() < int(length) {
		return "", ErrUnexpectedEOF
	}
	value := string(r.buf[r.off : r.off+int(length)])
	r.off += int(length)
	return value, nil
}

// ReadBytes reads a fixed-length raw byte segment.
func (r *Reader) ReadBytes(length int) ([]byte, error) {
	if length < 0 || r.Remaining() < length {
		return nil, ErrUnexpectedEOF
	}
	value := r.buf[r.off : r.off+length]
	r.off += length
	return value, nil
}

func (r *Reader) readByte() (byte, error) {
	if r.Remaining() < 1 {
		return 0, ErrUnexpectedEOF
	}
	value := r.buf[r.off]
	r.off++
	return value, nil
}
