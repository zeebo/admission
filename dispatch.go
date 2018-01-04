package admission

import (
	"context"
	"log"
	"net"
	"runtime"

	"golang.org/x/net/ipv4"
)

const (
	// DefaultMessages is the number of Messages passed to a ReadBatch call
	// in the Dispatcher Run loop.
	DefaultMessages = 128

	// DefaultInFlight is the number of concurrent calls to the Handler
	// allowed in the Run loop. If it would go over, the message is dropped.
	DefaultInFlight = 256
)

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

	// InFlight controls the number of parallel calls to the Handler. Zero is
	// DefaultInFlight.
	InFlight int
}

func (d *Dispatcher) Run(ctx context.Context) (err error) {
	num_messages := d.NumMessages
	if num_messages == 0 {
		num_messages = DefaultMessages
	}
	in_flight := d.InFlight
	if in_flight == 0 {
		in_flight = DefaultInFlight
	}

	pc := ipv4.NewPacketConn(d.PacketConn)
	in := make([]*Message, num_messages)
	msgs := make([]ipv4.Message, num_messages)
	sem := make(chan struct{}, in_flight)

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
			in[i].RemoteAddr = msgs[i].Addr

			select {
			case sem <- struct{}{}:
				go handleMessage(ctx, sem, d.Handler, in[i])
				in[i] = nil
			default:
			}
		}
	}
}

// handleMessage passes the message to the handler and returns it to the pool
// once it is done. it is the layer of safety around panics.
func handleMessage(ctx context.Context, sem chan struct{}, h Handler,
	m *Message) {

	// we pass addr in case m.RemoteAddr is changed during the update call
	defer func(m *Message, addr net.Addr) {
		if v := recover(); v != nil {
			buf := make([]byte, 64<<10)
			buf = buf[:runtime.Stack(buf, false)]
			log.Printf("admission: panic from %v: %v\n%s", addr, v, buf)
		}

		putMessage(m)
		<-sem
	}(m, m.RemoteAddr)

	h.Handle(ctx, m)
}
