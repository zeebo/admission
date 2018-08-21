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
	listener, err := net.ListenPacket("udp", ":0")
	assertNoError(t, err)
	defer listener.Close()

	// type assert concrete udp stuff
	listenerConn := listener.(*net.UDPConn)
	addr := listenerConn.LocalAddr().(*net.UDPAddr)

	writerConn, err := net.DialUDP("udp", nil, addr)
	assertNoError(t, err)
	defer writerConn.Close()

	// do a huge ceremony to read and write a packet
	listnerRawConn, err := listenerConn.SyscallConn()
	assertNoError(t, err)

	// give it a second or ten
	chm := make(chan *Message)
	time.AfterFunc(10*time.Second, func() { close(chm) })

	// try to read it
	go func() {
		msg := new(Message)
		n, err := Read(listnerRawConn, []*Message{msg})
		if n != 1 || err != nil {
			t.Fatalf("read error: %v %v", n, err)
		}
		chm <- msg
	}()

	// write it
	writerConn.Write([]byte("hello"))
	msg := <-chm

	// check it
	if msg == nil {
		t.Fatal("nil message")
	}
	if string(msg.Data) != "hello" {
		t.Fatalf("msg: %+v", msg)
	}
}
