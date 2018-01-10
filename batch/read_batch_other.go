// +build !linux !amd64

package batch

import (
	"syscall"
)

// iovec isn't used in the general case.
type iovec struct{}

// Read reads from the RawConn multiple messages in a single syscall.
func Read(sc syscall.RawConn, msgs []*Message) (int, error) {
	if len(msgs) == 0 {
		return 0, nil
	}

	var n int
	err := sc.Read(func(fd uintptr) bool {
		var err error
		n, err = syscall.Read(int(fd), msgs[0].buf[:])
		return err != syscall.EAGAIN
	})
	if err != nil {
		return 0, err
	}

	msgs[0].Data = msgs[0].buf[:n]
	return 1, nil
}
