package ipv4

import (
	"net"
	"syscall"
	"unsafe"
)

func ReadBatch(conn *net.UDPConn, ms []*Message) (int, error) {
	sc, err := conn.SyscallConn()
	if err != nil {
		return 0, err
	}

	hs := make([]mmsghdr, len(ms))

	for i := range ms {
		ms[i].iovec.Base = &ms[i].buf[0]
		ms[i].iovec.Len = uint64(len(ms[i].buf))
		hs[i].Hdr.Iov = &ms[i].iovec
		hs[i].Hdr.Iovlen = 1
	}

	var n int
	err = sc.Read(func(fd uintptr) (done bool) {
		var operr syscall.Errno
		n, operr = recvmmsg(fd, hs)
		return operr != syscall.EAGAIN
	})
	if err != nil {
		return n, err
	}

	for i := range ms[:n] {
		ms[i].Data = ms[i].buf[:hs[i].Len]
	}

	return n, nil
}

type mmsghdr struct {
	Hdr msghdr
	Len uint32
	_   [4]byte
}

type msghdr struct {
	Name    *byte
	Namelen uint32
	_       [4]byte
	Iov     *iovec
	Iovlen  uint64
	_       *byte
	_       uint64
	Flags   int32
	_       [4]byte
}

type iovec struct {
	Base *byte
	Len  uint64
}

func recvmmsg(fd uintptr, hs []mmsghdr) (int, syscall.Errno) {
	const sys_recvmmsg = 0x12b

	n, _, errno := syscall.Syscall6(
		sys_recvmmsg,
		fd,
		uintptr(unsafe.Pointer(&hs[0])),
		uintptr(len(hs)),
		0,
		0,
		0,
	)
	return int(n), errno
}
