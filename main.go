package main

import (
	"fmt"
	"net"
	"sync"

	"golang.org/x/net/ipv4"
)

func panice(err error) {
	if err != nil {
		panic(err)
	}
}

func main() {
	conn, err := net.ListenPacket("udp", "0.0.0.0:6969")
	panice(err)
	defer conn.Close()

	pc := ipv4.NewPacketConn(conn)

	in := make([]*Message, 5)
	n, err := Read(pc, in)
	panice(err)

	fmt.Printf("read %d:\n", n)
	for i, msg := range in[:n] {
		fmt.Printf("\t%d: %s\n", i+1, msg.Valid)
	}
}

func Read(pc *ipv4.PacketConn, in []*Message) (n int, err error) {
	// TODO(jeff): maybe we can remove this allocation
	msgs := make([]ipv4.Message, len(in))
	for i := range msgs {
		if in[i] == nil {
			in[i] = getMessage()
		}
		msgs[i] = in[i].AsIpv4Message()
	}

	n, err = pc.ReadBatch(msgs, 0)
	if err != nil {
		return 0, err
	}

	// fix up the valid slices
	for i := 0; i < n; i++ {
		in[i].Valid = in[i].Data[:msgs[i].N]
	}

	return n, nil
}

type Message struct {
	// Data buffer
	Data [1024]byte

	// Used to keep allocations low: further consumers of the message can reuse
	// this scratch space.
	Scratch [256]byte

	// After a read, filled with a slice of Data that is valid.
	Valid []byte

	// lazy cache around a single sice of Data
	buffers [][]byte
}

func (m *Message) AsIpv4Message() ipv4.Message {
	if len(m.buffers) == 0 {
		m.buffers = make([][]byte, 1)
		m.buffers[0] = m.Data[:]
	}

	return ipv4.Message{
		Buffers: m.buffers,
	}
}

var messagePool = sync.Pool{
	New: func() interface{} { return new(Message) },
}

func getMessage() *Message  { return messagePool.Get().(*Message) }
func putMessage(m *Message) { messagePool.Put(m) }
