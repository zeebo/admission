package admmonkit

import (
	"context"
	"net"

	"github.com/spacemonkeygo/spacelog"
	"github.com/zeebo/admission"
	"github.com/zeebo/float16"
	"gopkg.in/spacemonkeygo/monkit.v2"
)

var logger = spacelog.GetLogger()

type Options struct {
	// Application to send with
	Application string

	// Instance Id to send with
	InstanceId []byte

	// Address to send packets to
	Address string

	// PacketSize controls maximum packet size. If zero, 1024 is used.
	PacketSize int

	// Registry to pull stats from. If nil, monkit.Default is used.
	Registry *monkit.Registry
}

func Send(ctx context.Context, opts Options) (err error) {
	addr, err := net.ResolveUDPAddr("udp", opts.Address)
	if err != nil {
		return err
	}

	conn, err := net.DialUDP("udp", nil, addr)
	if err != nil {
		return err
	}
	defer conn.Close()

	if opts.Registry == nil {
		opts.Registry = monkit.Default
	}
	if opts.PacketSize == 0 {
		opts.PacketSize = 1024
	}

	var (
		buf []byte
		w   admission.Writer
	)

	metrics := 0
	buf = w.Begin(buf, opts.Application, opts.InstanceId)

	opts.Registry.Stats(func(name string, value float64) {
		// keep track of the buffer before we send
		before := buf

		// add the value to the buffer
		value16, ok := float16.FromFloat64(value)
		if !ok {
			logger.Infof("skipping %q because value unrepresentable: %v",
				name, value)
			return
		}

		buf = w.Append(buf, name, value16)

		// if we're over the packet size, send the previous value and start
		// over.
		if len(buf) > opts.PacketSize {
			logger.Infof("sending packet size %d bytes containing %d metrics",
				len(before), metrics)
			sendPacket(ctx, conn, before)

			w.Reset()
			metrics = 0
			buf = w.Begin(buf[:0], opts.Application, opts.InstanceId)
			buf = w.Append(buf, name, value16)
		}

		metrics++
	})

	logger.Infof("sending packet size %d bytes containing %d metrics",
		len(buf), metrics)
	sendPacket(ctx, conn, buf)

	return nil
}

func sendPacket(ctx context.Context, conn *net.UDPConn, buf []byte) {
	_, err := conn.Write(buf)
	logger.Errore(err)
}
