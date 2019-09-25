// package admmonkit hooks up admission with monkit.v2
package admmonkit

import (
	"context"
	"log"
	"net"

	"github.com/zeebo/admission/admproto"
	"gopkg.in/spacemonkeygo/monkit/v3"
)

// Options allows you to control where and how Send sends the data.
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

	// ProtoOps allows you to set protocol options.
	ProtoOpts admproto.Options
}

// Send will push all of the metrics in the registry to the address with the
// application and instance id in the options.
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
		w   = admproto.NewWriterWith(opts.ProtoOpts)
	)

	opts.Registry.Stats(func(name string, value float64) {
		// if we have any errors, stop.
		if err != nil {
			return
		}

		for {
			// keep track of the buffer before we send
			before := buf

			// always ensure the buffer has the prefix in it.
			if len(buf) == 0 {
				// if we can't add the application and instance id, it's fatal.
				buf, err = w.Begin(buf, opts.Application, opts.InstanceId)
				if err != nil {
					return
				}
			}

			// add the value to the buffer
			buf, err = w.Append(buf, name, value)
			if err != nil {
				// not fatal, just back up to before, but let someone know
				// it has been skipped.
				log.Println("skipped metric", name, "because", err)
				buf, err = before, nil
				return
			}

			// if we're still in the packet size, then get the next metric.
			if len(buf)+4 <= opts.PacketSize {
				return
			}

			// if we're over the packet size, send the previous value and start
			// over. be sure to account for the checksum that sendPacket adds.
			// if buf was empty at the start, we should just send it.
			// otherwise we should send the previous value.
			if len(before) == 0 {
				sendPacket(ctx, conn, buf)
			} else {
				sendPacket(ctx, conn, before)
			}

			// after sending the packet, we should reset the buffer and try to
			// add the point again.
			w.Reset()
			buf = buf[:0]
		}
	})

	// send off any remainder buf. we're guaranteed by the loop above that if
	// there is any data in buf it forms a valid packet with metrics in it.
	if err == nil && len(buf) > 0 {
		sendPacket(ctx, conn, buf)
	}

	return err
}

// sendPacket is a helper that adds a checksum to the provided buffer and sends
// it to the conn. It logs if there was an error.
func sendPacket(ctx context.Context, conn *net.UDPConn, buf []byte) {
	_, err := conn.Write(admproto.AddChecksum(buf))
	if err != nil {
		log.Println("failed to send packet:", err)
	}
}
