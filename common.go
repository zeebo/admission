// package admission is a fast way to ingest/send metrics.
package admission

import (
	"context"
	"net"
	"sync"
)

// Handler is a type that can handle messages.
type Handler interface {
	// Handle should do things with the message Data. No fields of the message
	// should be held on to after the call has returned.
	Handle(ctx context.Context, m *Message)
}

type Message struct {
	// data buffer. first to keep alignment with the rest of the fields.
	buf [1024]byte

	// Used to keep allocations low: further consumers of the message can reuse
	// this scratch space.
	Scratch [256]byte

	// Data contained in the Message to handle.
	Data []byte

	// RemoteAddr has the address that the packet was received from
	RemoteAddr net.Addr

	// points at buf. array to avoid an allocation because we need a [][]byte
	// pointed at buf eventually.
	buffers [1][]byte
}

func newMessage() (m *Message) {
	m = new(Message)
	m.buffers[0] = m.buf[:]
	return m
}

var messagePool = sync.Pool{
	New: func() interface{} { return newMessage() },
}

func getMessage() *Message  { return messagePool.Get().(*Message) }
func putMessage(m *Message) { messagePool.Put(m) }
