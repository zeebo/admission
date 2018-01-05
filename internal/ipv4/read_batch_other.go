// +build !linux !amd64

package ipv4

import "net"

func ReadBatch(conn *net.UDPConn, ms []*Message) (int, error) {
	if len(ms) == 0 {
		return 0, nil
	}

	n, err := conn.Read(ms[0].buf[:])
	if err != nil {
		return 0, err
	}

	ms[0].Data = ms[0].buf[:n]
	return 1, nil
}

type iovec struct{}
