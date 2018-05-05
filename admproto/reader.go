package admproto

import (
	"github.com/zeebo/incenc"
)

// Reader iterates over the key/value pairs in a buffer written by a Writer.
type Reader struct {
	r        incenc.Reader
	encoding FloatEncoding
}

// NewReaderWith returns a Reader with some given scratch space as a buffer to
// avoid allocations.
func NewReaderWith(scratch []byte) Reader {
	return Reader{r: incenc.NewReaderWith(scratch)}
}

// Reset clears the state of the Reader.
func (r *Reader) Reset() {
	r.r.Reset()
	r.encoding = 0
}

// Begin returns the header information out of the packet, and the remaining
// data in the packet.
func (r *Reader) Begin(in []byte) (out, application, instance_id []byte, err error) {
	in, version, err := consume(in, 1)
	if err != nil {
		return nil, nil, nil, Error.Wrap(err)
	}

	switch version[0] {
	case float16Version:
		r.encoding = Float16Encoding
	case float32Version:
		r.encoding = Float32Encoding
	case float64Version:
		r.encoding = Float64Encoding
	default:
		return nil, nil, nil, Error.New("unknown version: %d", version[0])
	}

	in, length, err := consume(in, 1)
	if err != nil {
		return nil, nil, nil, Error.Wrap(err)
	}
	in, application, err = consume(in, int(length[0]))
	if err != nil {
		return nil, nil, nil, Error.Wrap(err)
	}

	in, length, err = consume(in, 1)
	if err != nil {
		return nil, nil, nil, Error.Wrap(err)
	}
	in, instance_id, err = consume(in, int(length[0]))

	return in, application, instance_id, err
}

// Next consumes bytes from in, returns the key and value, and returns the rest
// of the bytes as out.
func (r *Reader) Next(in []byte) (out, key []byte, value float64, err error) {
	in, key, err = r.r.Next(in)
	if err != nil {
		return nil, nil, 0, err
	}

	in, value, err = r.encoding.consumeFloat(in)
	if err != nil {
		return nil, nil, 0, err
	}

	return in, key, value, nil
}
