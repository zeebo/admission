package admission

import (
	"context"
	"net"

	"github.com/zeebo/admission/internal/ipv4"
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

	// Conn is the connection the packets are read from.
	Conn *net.UDPConn

	// NumMessages is the number of messages to attempt to read at once. If
	// zero, DefaultMessages is used.
	NumMessages int

	// InFlight controls the number of parallel calls to the Handler. Zero is
	// DefaultInFlight.
	InFlight int

	// Hooks provide callbacks for events in the dispatcher.
	Hooks struct {
		ReadMessages   func(n int) // when messages were read with how many.
		DroppedMessage func()      // when a message is dropped.
	}
}

func (d Dispatcher) Run(ctx context.Context) (err error) {
	num_messages := d.NumMessages
	if num_messages == 0 {
		num_messages = DefaultMessages
	}
	in_flight := d.InFlight
	if in_flight == 0 {
		in_flight = DefaultInFlight
	}

	msgs := make([]*Message, num_messages)
	sem := make(chan struct{}, in_flight)

	for {
		// fill in any nil messages, and build up the ipv4.Message array for
		// passing to ReadBatch.
		for i := range msgs {
			if msgs[i] == nil {
				msgs[i] = getMessage()
			}
		}

		n, err := ipv4.ReadBatch(d.Conn, msgs)
		if err != nil {
			return err
		}
		if d.Hooks.ReadMessages != nil {
			d.Hooks.ReadMessages(n)
		}

		// fix up the Data slices, pass them off to be handled in a goroutine
		// and clear them out of the in array for the next round of packets.
		for i := 0; i < n; i++ {
			select {
			case sem <- struct{}{}:
				go handleMessage(ctx, sem, d.Handler, msgs[i])
				msgs[i] = nil
			default:
				if d.Hooks.DroppedMessage != nil {
					d.Hooks.DroppedMessage()
				}
			}
		}
	}
}

// handleMessage passes the message to the handler and returns it to the pool
// once it is done.
func handleMessage(ctx context.Context, sem chan struct{}, h Handler,
	m *Message) {

	h.Handle(ctx, m)
	putMessage(m)
	<-sem
}
