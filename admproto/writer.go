package admproto

import (
	"github.com/zeebo/incenc"
)

// Options controls options for the Writer.
type Options struct {
	// FloatEncoding is what kind of encoding to use for the floating point
	// values. The default is to use float16.
	FloatEncoding FloatEncoding
}

// Writer is a type for encoding key/value pairs to a byte buffer.
type Writer struct {
	options Options
	w       incenc.Writer
}

// NewWriterWith returns a Writer with the passed in options rather than the
// defaults.
func NewWriterWith(options Options) Writer {
	return Writer{options: options}
}

// Reset clears the state of the Writer.
func (w *Writer) Reset() {
	w.w.Reset()
}

// Begin appends header information to the buffer.
func (w *Writer) Begin(in []byte, application string, instance_id []byte, num_headers int) (
	out []byte, err error) {

	// Check that length does not exceed 255 so we can encode the length in a single byte
	if len(application) > 255 {
		return nil, Error.New("application too long")
	}
	if len(instance_id) > 255 {
		return nil, Error.New("instance_id too long")
	}
	if num_headers > 255 {
		return nil, Error.New("too many headers")
	}

	var version byte

	// signal what float encoding we're using
	switch w.options.FloatEncoding {
	case Float16Encoding:
		version |= float16Version
	case Float32Encoding:
		version |= float32Version
	case Float64Encoding:
		version |= float64Version
	default:
		return nil, Error.New("unknown float encoding: %d", w.options.FloatEncoding)
	}

	// we do not send headers yet
	if num_headers > 0 {
		version |= headersIncluded
	} else {
		version |= headersExcluded
	}

	in = append(in, version)
	in = append(in, byte(len(application)))
	in = append(in, application...)
	in = append(in, byte(len(instance_id)))
	in = append(in, instance_id...)

	if num_headers > 0 {
		in = append(in, byte(num_headers))
	}

	return in, nil
}

// AppendHeader adds the key and value to the starting bytes of the packet
func (w *Writer) AppendHeader(in, key, value []byte) (out []byte, err error) {
	if len(key) > 255 {
		return nil, Error.New("header key %s too long", key)
	}
	if len(value) > 255 {
		return nil, Error.New("header value %s too long", value)
	}
	in = append(in, byte(len(key)))
	in = append(in, key...)
	in = append(in, byte(len(value)))
	in = append(in, value...)
	return in, nil
}

// Append adds the key and value to the buffer using the last Append calls to
// reduce the amount of data it needs to write.
func (w *Writer) Append(in []byte, key string, value float64) (out []byte, err error) {
	in, err = w.w.Append(in, key)
	if err != nil {
		return nil, err
	}

	in, err = w.options.FloatEncoding.appendFloat(in, value)
	if err != nil {
		return nil, err
	}

	return in, nil
}
