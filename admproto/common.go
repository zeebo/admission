// package admproto describes the admission protocol
package admproto

import (
	"encoding/binary"
	"errors"
	"hash/crc32"

	"github.com/zeebo/errs"
)

// Error wraps all of the errors coming out of admproto.
var Error = errs.Class("admproto")

var castTable = crc32.MakeTable(crc32.Castagnoli)

// AddChecksum appends a checksum to the byte slice.
func AddChecksum(buf []byte) []byte {
	var scratch [4]byte
	check := crc32.Checksum(buf, castTable)
	binary.BigEndian.PutUint32(scratch[:], check)
	return append(buf, scratch[:]...)
}

// CheckChecksum removes an appended checksum and errors if either it cannot or
// if the checksum does not match.
func CheckChecksum(buf []byte) ([]byte, error) {
	offset := len(buf) - 4
	if offset < 0 {
		return nil, Error.New("buffer too small")
	}
	check := crc32.Checksum(buf[:offset], castTable)
	got := binary.BigEndian.Uint32(buf[offset:])
	if check != got {
		return nil, Error.New("checksum mismatch")
	}
	return buf[:offset], nil
}

// bufferTooSmall is the error these functiosn returns to ensure that they can
// be inlined. A non-leaf function currently cannot be inlined, so we must
// return the same error every time.
var bufferTooSmall = errors.New("buffer too small")

// consume attempts to read n bytes from the buffer.
func consume(in []byte, n int) (out, data []byte, err error) {
	if n < 0 || len(in) < n {
		return nil, nil, bufferTooSmall
	}
	return in[n:], in[:n], nil
}

// a list of versions and what they mean
const (
	floatMask      byte = 0b11 // mask to select the float encoding version
	float16Version byte = 0b00 // incenc encoded keys with float16 values
	float32Version byte = 0b01 // incenc encoded keys with float32 values
	float64Version byte = 0b10 // incenc encoded keys with float64 values

	headerMask      byte = 0b100 // mask to select the header's included version
	headersExcluded byte = 0b000 // headers are not included in the packet
	headersIncluded byte = 0b100 // headers are included in the packet
)
