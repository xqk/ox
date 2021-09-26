package server

import (
	"log"

	"ox/pkg/flag"
)

func init() {
	flag.Register(&flag.StringFlag{
		Name:    "host",
		Usage:   "--host, print host",
		Default: "127.0.0.1",
		Action: func(name string, fs *flag.FlagSet) {
			log.Printf("host flag: %v", fs.String(name))
		},
	})
}
