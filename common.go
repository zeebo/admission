package admission

import (
	"sync"
)

type Handler interface {
	Handle(m *Message)
}

type Message struct {
	// Used to keep allocations low: further consumers of the message can reuse
	// this scratch space.
	Scratch [256]byte

	// Data read from the Read call
	Data []byte

	// data buffer.
	buf [1024]byte

	// points at buf. array to avoid another allocation.
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
