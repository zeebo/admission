package admmonkit

import (
	"context"
	"encoding/hex"
	"net"
	"testing"

	"github.com/spacemonkeygo/monkit/v3"
	"github.com/zeebo/assert"
)

func TestSend_NoSpinning(t *testing.T) {
	addr, err := net.ResolveUDPAddr("udp", ":0")
	assert.NoError(t, err)
	conn, err := net.ListenUDP("udp", addr)
	assert.NoError(t, err)
	defer conn.Close()

	registry := monkit.NewRegistry()
	registry.ScopeNamed("test").Event("some_event_lol")

	errc := make(chan error, 2)
	go func() {
		errc <- Send(context.Background(), Options{
			Application: "app",
			InstanceId:  []byte("inst"),
			Address:     conn.LocalAddr().String(),
			PacketSize:  10,
			Registry:    registry,
		})
	}()

	var buf [4096]byte
	n, err := conn.Read(buf[:])
	assert.NoError(t, err)
	t.Log(n, "\n"+hex.Dump(buf[:n]))

	assert.NoError(t, <-errc)
}
