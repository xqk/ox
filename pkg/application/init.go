package application

import (
	"os"

	"github.com/xqk/ox/pkg"
	"github.com/xqk/ox/pkg/flag"
)

func init() {
	flag.Register(&flag.BoolFlag{
		Name:    "version",
		Usage:   "--version, print version",
		Default: false,
		Action: func(string, *flag.FlagSet) {
			pkg.PrintVersion()
			os.Exit(0)
		},
	})
}
