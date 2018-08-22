package batch

import (
	"net"
	"runtime"
	"sync"
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
	listenerRawConn, err := listenerConn.SyscallConn()
	assertNoError(t, err)

	result := make(chan []*Message, 1)

	// try to read it
	go func() {
		defer close(result)

		messages := []*Message{new(Message), new(Message), new(Message)}
		n, err := Read(listenerRawConn, messages)
		if n != 1 || err != nil {
			t.Fatalf("read error: %v %v", n, err)
		}

		result <- messages[:n]
	}()

	// write it
	time.Sleep(time.Millisecond)
	writerConn.Write([]byte("hello"))

	var messages []*Message
	select {
	case messages = <-result:
	case <-time.After(5 * time.Second):
		t.Fatal("timeout")
	}

	// check it
	if len(messages) != 1 {
		t.Fatal("invalid number of messages")
	}

	if string(messages[0].Data) != "hello" {
		t.Fatalf("got: %+v", messages[0])
	}
}

func TestReadMultiple(t *testing.T) {
	if runtime.GOOS != "linux" && runtime.GOOS != "windows" {
		t.Skip("unimplemented")
	}

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
	listenerRawConn, err := listenerConn.SyscallConn()
	assertNoError(t, err)

	var waitForWrites sync.WaitGroup
	waitForWrites.Add(1)
	result := make(chan []*Message, 1)

	// try to read it
	go func() {
		defer close(result)
		waitForWrites.Wait()

		messages := []*Message{new(Message), new(Message), new(Message)}
		n, err := Read(listenerRawConn, messages)
		if n != 2 || err != nil {
			t.Fatalf("read error: %v %v", n, err)
		}

		result <- messages[:n]
	}()

	// write it
	go func() {
		writerConn.Write([]byte("hello"))
		writerConn.Write([]byte("world"))
		waitForWrites.Done()
	}()

	var messages []*Message
	select {
	case messages = <-result:
	case <-time.After(5 * time.Second):
		t.Fatal("timeout")
	}

	// check it
	if len(messages) != 2 {
		t.Fatal("invalid number of messages")
	}

	if string(messages[0].Data) != "hello" || string(messages[1].Data) != "world" {
		t.Fatalf("got: %+v %+v", messages[0], messages[1])
	}
}

func TestClose(t *testing.T) {
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
	listenerRawConn, err := listenerConn.SyscallConn()
	assertNoError(t, err)

	done := make(chan struct{}, 1)

	// try to read it
	go func() {
		defer close(done)

		messages := []*Message{new(Message), new(Message), new(Message)}
		n, err := Read(listenerRawConn, messages)
		if err == nil {
			t.Fatalf("expected error: %v %v", n, err)
		}

		done <- struct{}{}
	}()

	err = listener.Close()
	if err != nil {
		t.Fatal(err)
	}

	<-done
}
