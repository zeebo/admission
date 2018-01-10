package batch

import (
	"sync"
	"syscall"
	"unsafe"
)

// iovec for linux amd64
type iovec struct {
	Base *byte
	Len  uint64
}

// Read reads from the RawConn multiple messages in a single syscall.
func Read(sc syscall.RawConn, msgs []*Message) (int, error) {
	// get a mmsghdr slice out from the pool. if it's not big enough, allocate
	// a new one with enough space. we use a pointer to a slice to avoid an
	// allocation when placing into the pool. this does a double allocation in
	// the case of a miss, but no allocations in the case of a hit.
	hdrs_p, _ := hdrPool.Get().(*[]mmsghdr)
	if hdrs_p == nil {
		hdrs := make([]mmsghdr, len(msgs))
		hdrs_p = &hdrs
	} else if cap(*hdrs_p) < len(msgs) {
		*hdrs_p = make([]mmsghdr, len(msgs))
	}
	hdrs := *hdrs_p
	hdrs = hdrs[:len(msgs)]

	// initialize the mmsghdrs to point at the buffers in the passed in
	// Messages. we always set the iovec field on the Message in case someone
	// just passes in a fresh one.
	for i := range msgs {
		msgs[i].iovec.Base = &msgs[i].buf[0]
		msgs[i].iovec.Len = uint64(len(msgs[i].buf))

		hdrs[i] = mmsghdr{
			Hdr: msghdr{
				Iov:    &msgs[i].iovec,
				Iovlen: 1,
			},
			Len: 0,
		}
	}

	// we reduce allocations for the sc.Read method by use of the op struct.
	// we have to use a closure because no value passed in to the Read call can
	// be used to squirrel away the data. since go can't prove that the closure
	// we pass to sc.Read doesn't escape, it would be an allocation to build
	// the closure for every Read call. instead, we keep a pool of op structs
	// that store the closure and the closed upon values. this causes 2
	// allocations (the op struct, and the closure) when the pool misses, but
	// we don't need to do an allocation for every Read when the pool hits.
	type op struct {
		m    func(uintptr) bool
		hdrs []mmsghdr
		n    int
	}

	// get and initialize a *op from the pool
	o, _ := opPool.Get().(*op)
	if o == nil {
		o = new(op)
		o.m = func(fd uintptr) bool {
			var operr syscall.Errno
			o.n, operr = recvmmsg(fd, o.hdrs)
			return operr != syscall.EAGAIN
		}
	}
	o.hdrs = hdrs
	o.n = 0

	// issue the Read call and look at the results
	err := sc.Read(o.m)
	n := o.n

	// replace the op. we clear hdrs here to avoid keeping them alive inside
	// of the pool if possible.
	o.hdrs = nil
	opPool.Put(o)

	if err != nil {
		return n, err
	}

	// read the results into the msgs
	for i := range msgs[:n] {
		msgs[i].Data = msgs[i].buf[:hdrs[i].Len]
	}

	// we no longer need the mmsghdrs. return them for another call
	hdrPool.Put(hdrs_p)

	return n, nil
}

// we use these pools to reduce allocations in Read
var (
	hdrPool sync.Pool
	opPool  sync.Pool
)

type mmsghdr struct {
	Hdr msghdr
	Len uint32
	_   [4]byte // padding
}

type msghdr struct {
	_      *byte   // Name
	_      uint32  // Name len
	_      [4]byte // padding
	Iov    *iovec
	Iovlen uint64
	_      *byte   // Control
	_      uint64  // Control len
	_      int32   // Flags
	_      [4]byte // padding
}

// recvmmsg runs the recvmmsg syscall on the fd with the provided msg headers.
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
