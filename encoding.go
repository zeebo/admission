package admission

import (
	"github.com/zeebo/float16"
	"github.com/zeebo/incenc"
)

// Writer is a type for encoding key/value pairs to a byte buffer.
type Writer struct {
	w incenc.Writer
}

// Reset clears the state of the Writer.
func (w *Writer) Reset() {
	w.w.Reset()
}

// Begin appends header information to the buffer.
func (w *Writer) Begin(in []byte, application string,
	instance_id []byte) []byte {

	if len(application) > 255 {
		panic("application too long")
	}
	if len(instance_id) > 255 {
		panic("instance_id too long")
	}

	in = append(in, '\x00') // version byte
	in = append(in, byte(len(application)))
	in = append(in, application...)
	in = append(in, byte(len(instance_id)))
	in = append(in, instance_id...)

	return in
}

// Append adds the key and value to the buffer using the last Append calls to
// reduce the amount of data it needs to write.
func (w *Writer) Append(in []byte, key string, value float16.Float16) []byte {
	in = w.w.Append(in, key)
	in = append(in, byte(value>>8), byte(value))
	return in
}

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
