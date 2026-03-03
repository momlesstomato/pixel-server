// Package codec provides binary encoding primitives matching the Pixel Protocol
// wire format: big-endian integers, uint16-prefixed UTF-8 strings, and
// uint32-prefixed length frames.
//
// Reader and Writer are the only types used by generated packet code.
// They are allocation-free on the hot path when backed by pooled buffers.
package codec
