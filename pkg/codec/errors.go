package codec

import "errors"

// ErrInvalidFrame indicates frame length or structure is invalid.
var ErrInvalidFrame = errors.New("invalid frame")

// ErrUnexpectedEOF indicates payload ended before all expected bytes were read.
var ErrUnexpectedEOF = errors.New("unexpected eof")
