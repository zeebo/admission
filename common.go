// package admission is a fast way to ingest/send metrics.
package admission

import (
	"context"
	"sync"

	"github.com/zeebo/admission/v3/internal/batch"
)

// Handler is a type that can handle messages.
type Handler interface {
	// Handle should do things with the message Data. No fields of the message
	// should be held on to after the call has returned.
	Handle(ctx context.Context, m *Message)
}

// Message is what is handled by a handler.
type Message = batch.Message

var messagePool = sync.Pool{
	New: func() interface{} { return new(Message) },
}

func getMessage() *Message  { return messagePool.Get().(*Message) }
func putMessage(m *Message) { messagePool.Put(m) }
