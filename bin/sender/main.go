package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/spacemonkeygo/flagfile"
	"github.com/spacemonkeygo/spacelog/setup"
	"github.com/zeebo/admission/admmonkit"
	"gopkg.in/spacemonkeygo/monkit.v2"
	"gopkg.in/spacemonkeygo/monkit.v2/environment"
)

func main() {
	flagfile.Load()
	setup.MustSetup("sender")
	environment.Register(monkit.Default)

	if err := run(context.Background()); err != nil {
		fmt.Fprintln(os.Stderr, "admission:", err)
		os.Exit(1)
	}
}

func run(ctx context.Context) (err error) {
	opts := admmonkit.Options{
		Application: "sender",
		InstanceId:  []byte("\xca\xfe\xf0\x0d"),
		Address:     "localhost:6969",
	}

	for {
		err = admmonkit.Send(ctx, opts)
		if err != nil {
			return err
		}
		time.Sleep(5 * time.Second)
	}
}
