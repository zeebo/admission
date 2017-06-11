package main

import (
	"context"
	"fmt"
	"net"
	"os"

	"github.com/zeebo/admission"
	"github.com/zeebo/float16"
)

func main() {
	if err := run(context.Background()); err != nil {
		fmt.Fprintln(os.Stderr, "admission:", err)
		os.Exit(1)
	}
}

func run(ctx context.Context) (err error) {
	pc, err := net.ListenPacket("udp", ":6969")
	if err != nil {
		return err
	}

	d := admission.Dispatcher{
		Handler:    printHandler{},
		PacketConn: pc,
	}

	return d.Run(ctx)
}

type printHandler struct{}

func (printHandler) Handle(m *admission.Message) {
	r := admission.NewReaderWith(m.Scratch[:])
	data := m.Data

	data, app, inst_id := r.Begin(data)
	fmt.Printf("application: %s\n", app)
	fmt.Printf("instance_id: %x\n", inst_id)

	var (
		key   []byte
		value float16.Float16
	)

	for len(data) > 0 {
		data, key, value = r.Next(data)
		fmt.Printf("\t%s: %v\n", key, value)
	}
}
