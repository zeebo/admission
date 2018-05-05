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
	// do a huge ceremony to pipe two udp conns.
	iconn1, err := net.ListenPacket("udp", ":0")
	assertNoError(t, err)
	defer iconn1.Close()

	// type assert concrete udp stuff
	conn1 := iconn1.(*net.UDPConn)
	addr := conn1.LocalAddr().(*net.UDPAddr)

	conn2, err := net.DialUDP("udp", nil, addr)
	assertNoError(t, err)
	defer conn2.Close()

	// do a huge ceremony to read and write a packet
	sc, err := conn1.SyscallConn()
	assertNoError(t, err)

	// give it a second or ten
	chm := make(chan *Message)
	time.AfterFunc(10*time.Second, func() { close(chm) })

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
