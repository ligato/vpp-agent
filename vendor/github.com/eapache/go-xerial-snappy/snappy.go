package snappy

import (
	"bytes"
	"encoding/binary"
	"errors"

	master "github.com/golang/snappy"
)

const (
	sizeOffset = 16
	sizeBytes  = 4
)

var (
	xerialHeader = []byte{130, 83, 78, 65, 80, 80, 89, 0}
	// ErrMalformed is returned by the decoder when the xerial framing
	// is malformed
	ErrMalformed = errors.New("malformed xerial framing")
)

// Encode encodes data as snappy with no framing header.
func Encode(src []byte) []byte {
	return master.Encode(nil, src)
}

// Decode decodes snappy data whether it is traditional unframed
// or includes the xerial framing format.
func Decode(src []byte) ([]byte, error) {
	return DecodeInto(nil, src)
}

// DecodeInto decodes snappy data whether it is traditional unframed
// or includes the xerial framing format into the specified `dst`.
// It is assumed that the entirety of `dst` including all capacity is available
// for use by this function. If `dst` is nil *or* insufficiently large to hold
// the decoded `src`, new space will be allocated.
func DecodeInto(dst, src []byte) ([]byte, error) {
	var max = len(src)
	if max < len(xerialHeader) {
		return nil, ErrMalformed
	}

	if !bytes.Equal(src[:8], xerialHeader) {
		return master.Decode(dst[:cap(dst)], src)
	}

	if max < sizeOffset+sizeBytes {
		return nil, ErrMalformed
	}

	if dst == nil {
		dst = make([]byte, 0, len(src))
	}

	dst = dst[:0]
	var (
		pos   = sizeOffset
		chunk []byte
		err       error
	)

	for pos+sizeBytes <= max {
		size := int(binary.BigEndian.Uint32(src[pos : pos+sizeBytes]))
		pos += sizeBytes

		nextPos := pos + size
		// On architectures where int is 32-bytes wide size + pos could
		// overflow so we need to check the low bound as well as the
		// high
		if nextPos < pos || nextPos > max {
			return nil, ErrMalformed
		}

		chunk, err = master.Decode(chunk[:cap(chunk)], src[pos:nextPos])
		if err != nil {
			return nil, err
		}
		pos = nextPos
		dst = append(dst, chunk...)
	}
	return dst, nil
}
