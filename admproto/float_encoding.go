package admproto

import (
	"encoding/binary"
	"math"

	"github.com/zeebo/float16"
)

// FloatEncoding controls different types of encodings of floating point values.
type FloatEncoding int

const ( // an enumeration of all of the FloatEncodings.
	Float16Encoding FloatEncoding = iota
	Float32Encoding
	Float64Encoding
)

// append encodes the value and appends it to the passed in slice.
func (f FloatEncoding) appendFloat(in []byte, value float64) (out []byte, err error) {
	switch f {
	case Float16Encoding:
		value16, ok := float16.FromFloat64(value)
		if !ok {
			return nil, Error.New("value not representable in float16")
		}
		return append(in, byte(value16>>8), byte(value16)), nil

	case Float32Encoding:
		fuint32 := math.Float32bits(float32(value))
		return append(in,
			byte(fuint32>>24), byte(fuint32>>16),
			byte(fuint32>>8), byte(fuint32)), nil

	case Float64Encoding:
		fuint64 := math.Float64bits(value)
		return append(in,
			byte(fuint64>>56), byte(fuint64>>48),
			byte(fuint64>>40), byte(fuint64>>32),
			byte(fuint64>>24), byte(fuint64>>16),
			byte(fuint64>>8), byte(fuint64)), nil

	default:
		return nil, Error.New("unknown float encoding: %d", f)
	}
}

// consumeFloat consumes the float value from in and retuns the slice.
func (f FloatEncoding) consumeFloat(in []byte) (out []byte, value float64, err error) {
	switch f {
	case Float16Encoding:
		in, data, err := consume(in, 2)
		if err != nil {
			return nil, 0, err
		}
		value = float16.Float16(uint16(data[1]) | uint16(data[0])<<8).Float64()

		return in, value, nil

	case Float32Encoding:
		in, data, err := consume(in, 4)
		if err != nil {
			return nil, 0, err
		}
		value = float64(math.Float32frombits(binary.BigEndian.Uint32(data)))

		return in, value, nil

	case Float64Encoding:
		in, data, err := consume(in, 8)
		if err != nil {
			return nil, 0, err
		}
		value = math.Float64frombits(binary.BigEndian.Uint64(data))

		return in, value, nil

	default:
		return nil, 0, Error.New("unknown float encoding: %d", f)
	}
}
