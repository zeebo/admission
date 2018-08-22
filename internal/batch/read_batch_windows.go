package batch

import (
	"runtime"
	"syscall"
	"unsafe"
)

const (
	e_WSAEMSGSIZE = syscall.Errno(10040)
)

// iovec isn't used in the general case.
type iovec struct{}

// Read reads from the RawConn multiple messages in a single syscall.
func Read(sc syscall.RawConn, msgs []*Message) (int, error) {
	if len(msgs) == 0 {
		return 0, nil
	}

	var pollerr, recverr error
	var messageCount int

	err := sc.Read(func(fd uintptr) bool {
		for _, msg := range msgs {
			// check whether there is more data
			var poll wsapollfd
			poll.fd = syscall.Handle(fd)
			poll.events = pollrdnorm

			pollerr = wsapoll(&poll, 1, 0)
			if pollerr != nil || poll.events != poll.revents {
				return true
			}

			var read uint32
			var flags uint32

			var buf syscall.WSABuf
			buf.Buf = &msg.buf[0]
			buf.Len = uint32(len(msg.buf))

			recverr = syscall.WSARecv(syscall.Handle(fd), &buf, 1, &read, &flags, nil, nil)
			msg.Data = msg.buf[:read]
			if recverr != nil || read == 0 {
				return true
			}

			messageCount++
		}
		return true
	})

	if pollerr != nil {
		return messageCount, pollerr
	}
	if recverr != nil {
		return messageCount, recverr
	}
	if err != nil {
		return messageCount, err
	}

	return messageCount, nil
}

var (
	modws2_32   = syscall.NewLazyDLL("ws2_32.dll")
	procWSAPoll = modws2_32.NewProc("WSAPoll")
)

const (
	pollrdnorm   = 0x0100
	socket_error = uintptr(^uint32(0))
)

type wsapollfd struct {
	fd      syscall.Handle
	events  uint16
	revents uint16
}

func wsapoll(polls *wsapollfd, count uint64, timeout int32) (err error) {
	r1, _, e1 := syscall.Syscall(
		procWSAPoll.Addr(), 3,
		uintptr(unsafe.Pointer(polls)), uintptr(count), uintptr(timeout))

	if r1 == socket_error {
		if e1 != 0 {
			err = syscall.Errno(e1)
		} else {
			err = syscall.EINVAL
		}
	}

	runtime.KeepAlive(polls)
	return err
}
