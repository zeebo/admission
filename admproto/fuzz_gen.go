// +build ignore

package main

import (
	"math/rand"

	"github.com/dvyukov/go-fuzz/gen"
	"github.com/zeebo/admission/v2/admproto"
)

func main() {
	var buf []byte

	for {
		var w admproto.Writer
		buf = buf[:0]
		buf, _ = w.Begin(buf, "app", []byte("ins"), 0)
		values := rand.Intn(10)
		for i := 0; i < values; i++ {
			buf, _ = w.Append(buf, string(randomValue()), rand.Float64())
		}
		gen.Emit(buf, nil, true)
	}
}

func randomValue() []byte {
	out := make([]byte, rand.Intn(10))
	for i := range out {
		out[i] = byte(rand.Intn(10))
	}
	return out
}
