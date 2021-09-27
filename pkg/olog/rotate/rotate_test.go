package rotate_test

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/xqk/ox/pkg/olog/rotate"
)

// Example of how to rotate in response to SIGHUP.
func ExampleLogger_Rotate() {
	l := &rotate.Logger{}
	log.SetOutput(l)
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGHUP)

	go func() {
		for {
			<-c
			l.Rotate()
		}
	}()
}
