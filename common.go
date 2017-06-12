// package admission is a fast way to ingest/send metrics.
package admission

import (
	"sync"
)

// Handler is a type that can handle messages.
type Handler interface {
	// Handle should do things with the message Data. No fields of the message
	// should be held on to after the call has returned.
	Handle(m *Message)
}

type Message struct {
	// Used to keep allocations low: further consumers of the message can reuse
	// this scratch space.
	Scratch [256]byte

	// Data contained in the Message to handle.
	Data []byte

	// data buffer.
	buf [1024]byte

	// points at buf. array to avoid an allocation.
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
