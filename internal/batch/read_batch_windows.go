package batch

import (
	"syscall"
)

const e_WSAEMSGSIZE = syscall.Errno(10040)

// iovec isn't used in the general case.
type iovec struct{}

// Read reads from the RawConn multiple messages in a single syscall.
func Read(sc syscall.RawConn, msgs []*Message) (int, error) {
	if len(msgs) == 0 {
		return 0, nil
	}

	//TODO: currently only reads one message at a time
	var recverr error
	var messageCount int

	err := sc.Read(func(fd uintptr) bool {
		msg := msgs[0]

		var read uint32 = 1
		var flags uint32

		var buf syscall.WSABuf
		buf.Buf = &msg.buf[0]
		buf.Len = uint32(len(msg.buf))

		recverr = syscall.WSARecv(syscall.Handle(fd), &buf, 1, &read, &flags, nil, nil)
		msg.Data = msg.buf[:read]
		if recverr != nil || read == 0 {
			return true
		}

		messageCount = 1
		return true
	})

	if recverr != nil {
		return messageCount, recverr
	}
	if err != nil {
		return messageCount, err
	}

	return messageCount, nil
}
