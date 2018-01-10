package batch

import (
	"net"
	"testing"
	"time"
)

func assertNoError(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("%+v", err)
	}
}

func TestRead(t *testing.T) {
	// do a huge ceremony to pipe two udp conns. hope this addr is usable!
	addr, err := net.ResolveUDPAddr("udp", ":16663")
	assertNoError(t, err)

	conn1, err := net.ListenUDP("udp", addr)
	assertNoError(t, err)
	defer conn1.Close()

	conn2, err := net.DialUDP("udp", nil, addr)
	assertNoError(t, err)
	defer conn2.Close()

	// do a huge ceremony to read and write a packet
	sc, err := conn1.SyscallConn()
	assertNoError(t, err)

	// give it a second
	chm := make(chan *Message)
	time.AfterFunc(time.Second, func() { close(chm) })

	// try to read it
	go func() {
		msg := new(Message)
		Read(sc, []*Message{msg})
		chm <- msg
	}()

	// write it
	conn2.Write([]byte("hello"))
	msg := <-chm

	// check it
	if msg == nil {
		t.Fatal("nil message")
	}
	if string(msg.Data) != "hello" {
		t.Fatalf("msg: %+v", msg)
	}
}
