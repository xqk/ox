package main

import (
	"log"
	"os"

	"ox/cmd/ox/new"
	"ox/cmd/ox/protoc"

	"github.com/urfave/cli"
)

const Version = "0.1.0"

func main() {
	app := cli.NewApp()
	app.Name = "ox"
	app.Usage = "ox tools"
	app.Version = Version
	app.Commands = []cli.Command{
		new.Cmd,
		protoc.Cmd,
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
