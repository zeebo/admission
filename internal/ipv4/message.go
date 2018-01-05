package ipv4

// Message is what we read from/to.
type Message struct {
	// buf contains the data that the Data slice will point at.
	buf [1024]byte

	// Scratch is to reduce allocations for consumers of Message.
	Scratch [256]byte

	// Data contained in the Message to handle.
	Data []byte

	// inlined to avoid allocations during reading
	iovec iovec
}
