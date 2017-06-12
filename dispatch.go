package admission

import (
	"context"
	"net"

	"golang.org/x/net/ipv4"
)

// DefaultMessages is the number of Messages passed to a ReadBatch call in the
// Dispatcher Run loop.
const DefaultMessages = 128

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

		// fix up the Data slices, pass them off to be handled in a goroutine
		// and clear them out of the in array for the next round of packets.
		for i := 0; i < n; i++ {
			in[i].Data = in[i].buf[:msgs[i].N]
			go handleMessage(ctx, d.Handler, in[i])
			in[i] = nil
		}
	}
}

func handleMessage(ctx context.Context, h Handler, m *Message) {
	h.Handle(m)

	// don't bother with defer sicne this is run in a goroutine anyway.
	putMessage(m)
}
