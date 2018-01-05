// package admproto describes the admission protocol
package admproto

import (
	"encoding/binary"
	"hash/crc32"

	"github.com/zeebo/errs"
)

var Error = errs.Class("admproto")

var castTable = crc32.MakeTable(crc32.Castagnoli)

func AddChecksum(buf []byte) []byte {
	var scratch [4]byte
	check := crc32.Checksum(buf, castTable)
	binary.BigEndian.PutUint32(scratch[:], check)
	return append(buf, scratch[:]...)
}

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
