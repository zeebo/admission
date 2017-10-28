package main

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/zeebo/admission"
	"github.com/zeebo/admission/admproto"
	"gopkg.in/spacemonkeygo/monkit.v2"
	"gopkg.in/spacemonkeygo/monkit.v2/environment"
)

var mon = monkit.Package()

func main() {
	environment.Register(monkit.Default)
	if err := run(context.Background()); err != nil {
		fmt.Fprintln(os.Stderr, "admission:", err)
		os.Exit(1)
	}
}

func run(ctx context.Context) (err error) {
	go func() {
		for {
			time.Sleep(time.Second)
			cmd := exec.Command("clear")
			cmd.Stdout = os.Stdout
			cmd.Run()
			monkit.Default.Stats(func(name string, value float64) {
				if strings.Contains(name, "times") &&
					!strings.Contains(name, "count") {
					duration := time.Duration(float64(time.Second) * value)
					fmt.Println(name, duration)
				} else {
					fmt.Println(name, value)
				}
			})
		}
	}()

	pc, err := net.ListenPacket("udp", ":6969")
	if err != nil {
		return err
	}

	d := admission.Dispatcher{
		Handler:    noopHandler{},
		PacketConn: pc,
	}

	return d.Run(ctx)
}

type noopHandler struct{}

var noopHandler_Handle_Task = mon.Task()

func (noopHandler) Handle(m *admission.Message) {
	finish := noopHandler_Handle_Task(nil)

	r := admproto.NewReaderWith(m.Scratch[:])

	data, _, _ := r.Begin(m.Data)
	for len(data) > 0 {
		data, _, _ = r.Next(data)
	}

	finish(nil)
}
