package main

import (
	"log"
	"os"

	"github.com/urfave/cli/v2"
)

func main() {
	commands := []*cli.Command{
		{
			Name:      "send",
			Aliases:   []string{"s"},
			Usage:     "send UDP stream",
			ArgsUsage: "file ipaddr port",
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:  "i",
					Value: "eth0",
					Usage: "specifies network interface to use",
				},
			},
			Action: func(c *cli.Context) error {
				args := c.Args().Slice()
				if len(args) == 3 {
					send(args[0], args[1], args[2], c.String("i"))
				} else {
					cli.ShowCommandHelpAndExit(c, "s", 1)
				}
				return nil
			},
		},
		{
			Name:      "receive",
			Aliases:   []string{"r"},
			Usage:     "receive UDP stream",
			ArgsUsage: "ipaddr port",
			Action: func(c *cli.Context) error {
				args := c.Args().Slice()
				if len(args) == 2 {
				} else {
					cli.ShowCommandHelpAndExit(c, "r", 1)
				}
				return nil
			},
		},
	}

	app := &cli.App{Commands: commands}
	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}

func send(file, ipaddr, port, iface string) {
	log.Println("send args:", file, ipaddr, port, iface)
}
