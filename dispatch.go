package admission

import (
	"context"
	"encoding/binary"
	"hash/crc32"
	"log"
	"net"
	"runtime"

	"golang.org/x/net/ipv4"
)

// DefaultMessages is the number of Messages passed to a ReadBatch call in the
// Dispatcher Run loop.
const DefaultMessages = 128

// Dispatcher reads Messages on the PacketConn and forwards them into the
// Handler in parallel.
type Dispatcher struct {
	// Handler is an interface called with each read Message.
	Handler Handler

	// PacketConn is the connection the packets are read from.
	PacketConn net.PacketConn

	// NumMessages is the number of messages to attempt to read at once. If
	// zero, DefaultMessages is used.
	NumMessages int

	// Flags are optinally flags to be passed to the batch read method.
	Flags int
}

func (d *Dispatcher) Run(ctx context.Context) (err error) {
	num_messages := d.NumMessages
	if num_messages == 0 {
		num_messages = DefaultMessages
	}

	pc := ipv4.NewPacketConn(d.PacketConn)
	in := make([]*Message, num_messages)
	msgs := make([]ipv4.Message, num_messages)

	for {
		// fill in any nil messages, and build up the ipv4.Message array for
		// passing to ReadBatch.
		for i := range msgs {
			if in[i] == nil {
				in[i] = getMessage()
			}
			msgs[i] = ipv4.Message{
				Buffers: in[i].buffers[:],
			}
		}

		n, err := pc.ReadBatch(msgs, d.Flags)
		if err != nil {
			return err
		}
		messagesRead.Observe(int64(n))

		// fix up the Data slices, pass them off to be handled in a goroutine
		// and clear them out of the in array for the next round of packets.
		for i := 0; i < n; i++ {
			in[i].Data = in[i].buf[:msgs[i].N]
			in[i].RemoteAddr = msgs[i].Addr
			go handleMessage(ctx, d.Handler, in[i])
			in[i] = nil
		}
	}
}

// handleMessage passes the message to the handler and returns it to the pool
// once it is done. it is the layer of safety around panics.
func handleMessage(ctx context.Context, h Handler, m *Message) {
	// first check that the message's crc is valid. silently drop any packets
	// that don't checksum.
	offset := len(m.Data) - 4
	if offset < 0 {
		return
	}
	check := crc32.Checksum(m.Data[:offset], castTable)
	got := binary.BigEndian.Uint32(m.Data[offset:])
	if check != got {
		return
	}

	// we pass addr in case m.RemoteAddr is changed during the update call
	defer func(m *Message, addr net.Addr) {
		if v := recover(); v != nil {
			// get the stack and ensure it ends with a newline
			buf := make([]byte, 64<<10)
			buf = buf[:runtime.Stack(buf, false)]
			log.Printf("admission: panic from %v: %v\n%s", addr, v, buf)
		}

		putMessage(m)
	}(m, m.RemoteAddr)

	// slice it off and pass it on
	m.Data = m.Data[:offset]
	h.Handle(m)
}
