package main

import (
	"log"
	"os"

	"github.com/leonlinc/hunter/m2ts"
	"github.com/leonlinc/hunter/pcap"
	"github.com/urfave/cli/v2"
)

func main() {
	commands := []*cli.Command{
		m2ts.SendCmd,
		m2ts.RecvCmd,
		pcap.ParseTS,
	}
	app := &cli.App{Commands: commands}
	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
