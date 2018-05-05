package admission

import (
	"context"
	"syscall"

	"github.com/zeebo/admission/internal/batch"
	"github.com/zeebo/errs"
)

const (
	// DefaultMessages is the number of Messages passed to a ReadBatch call
	// in the Dispatcher Run loop.
	DefaultMessages = 16

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
	Conn syscall.RawConn

	// NumMessages is the number of messages to attempt to read at once. If
	// zero, DefaultMessages is used.
	NumMessages int

	// InFlight controls the number of parallel calls to the Handler. Zero is
	// DefaultInFlight.
	InFlight int

	// Hooks provide callbacks for events in the dispatcher.
	Hooks struct {
		// when messages were read with how many.
		ReadMessages func(ctx context.Context, n int)
		// when a message is dropped.
		DroppedMessage func(ctx context.Context)
	}
}

// Run reads messages and passes them to the handler in their own goroutines
// until the context is cancelled.
func (d Dispatcher) Run(ctx context.Context) (err error) {
	num_messages := d.NumMessages
	if num_messages == 0 {
		num_messages = DefaultMessages
	}
	in_flight := d.InFlight
	if in_flight == 0 {
		in_flight = DefaultInFlight
	}

	done := ctx.Done()
	msgs := make([]*Message, num_messages)
	sem := make(chan struct{}, in_flight)

	for {
		// check our context.
		//
		// TODO(jeff): it'd be nice if there was a way to cancel the read but
		// i think that requires a bunch of read deadline business that i'd
		// like to avoid.
		select {
		case <-done:
			return nil
		default:
		}

		// fill in any nil messages, and build up the Message array for
		// passing to batch.Read.
		for i := range msgs {
			if msgs[i] == nil {
				msgs[i] = getMessage()
			}
		}

		n, err := batch.Read(d.Conn, msgs)
		if err != nil {
			return errs.Wrap(err)
		}
		if d.Hooks.ReadMessages != nil {
			d.Hooks.ReadMessages(ctx, n)
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
					d.Hooks.DroppedMessage(ctx)
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
