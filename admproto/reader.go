package admproto

import (
	"github.com/zeebo/float16"
	"github.com/zeebo/incenc"
)

// Reader iterates over the key/value pairs in a buffer written by a Writer.
type Reader struct {
	r incenc.Reader
}

// NewReaderWith returns a Reader with some given scratch space as a buffer to
// avoid allocations.
func NewReaderWith(scratch []byte) Reader {
	return Reader{
		r: incenc.Reader{
			Scratch: scratch,
		},
	}
}

// Begin returns the header information out of the packet, and the remaining
// data in the packet.
func (r *Reader) Begin(in []byte) (out, application, instance_id []byte) {
	if in[0] != '\x00' {
		panic("unknown version")
	}
	in = in[1:]

	end := 1 + int(in[0])
	application = in[1:end]
	in = in[end:]

	end = 1 + int(in[0])
	instance_id = in[1:end]
	in = in[end:]

	return in, application, instance_id
}

// Next consumes bytes from in, returns the key and value, and returns the rest
// of the bytes as out.
func (r *Reader) Next(in []byte) (out, key []byte, value float16.Float16) {
	in, key = r.r.Next(in)
	value = float16.Float16(uint16(in[0])<<8 | uint16(in[1]))
	return in[2:], key, value
}
