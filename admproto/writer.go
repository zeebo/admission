package admproto

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
