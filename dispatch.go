package admission

import (
	"context"
	"net"

	"golang.org/x/net/ipv4"
)

type Dispatcher struct {
	// Handler is an interface called with each read Message.
	Handler Handler

	// PacketConn is the connection the packets are read from.
	PacketConn net.PacketConn

	// NumMessages is the number of messages to attempt to read at once. If
	// zero, 128 is used.
	NumMessages int

	// Flags are optinally flags to be passed to the batch read method.
	Flags int
}

func (d *Dispatcher) Run(ctx context.Context) (err error) {
	num_messages := d.NumMessages
	if num_messages == 0 {
		num_messages = 128
	}

	pc := ipv4.NewPacketConn(d.PacketConn)
	in := make([]*Message, num_messages)
	msgs := make([]ipv4.Message, num_messages)

	for {
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

		// now fix up valid
		for i := 0; i < n; i++ {
			in[i].Data = in[i].buf[:msgs[i].N]
			go handleMessage(ctx, d.Handler, in[i])
			in[i] = nil
		}
	}
}

func handleMessage(ctx context.Context, h Handler, m *Message) {
	h.Handle(m)
	putMessage(m)
}
