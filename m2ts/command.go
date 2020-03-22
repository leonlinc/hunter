package m2ts

import (
	"github.com/urfave/cli/v2"
)

var (
	SendCmd *cli.Command
	RecvCmd *cli.Command
)

func init() {
	SendCmd = &cli.Command{
		Name:      "send",
		Usage:     "send TS over UDP",
		ArgsUsage: "file host port",
		Aliases:   []string{"s"},
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "i",
				Usage: "specifies the network interface",
			},
			&cli.Int64Flag{
				Name:  "b",
				Usage: "specifies the bitrate of the file",
			},
		},
		Action: func(c *cli.Context) error {
			args := c.Args().Slice()
			if len(args) < 3 {
				return cli.Exit("Not enough arguments", 1)
			}
			return send(args[0], args[1], args[2], c.String("i"), c.Int64("b"))
		},
	}
	RecvCmd = &cli.Command{
		Name:      "recv",
		Aliases:   []string{"r"},
		Usage:     "receive TS over RTP/UDP",
		ArgsUsage: "host port",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "i",
				Usage: "specifies the network interface",
			},
			&cli.StringFlag{
				Name:  "o",
				Usage: "specifies the output file",
			},
		},
		Action: func(c *cli.Context) error {
			args := c.Args().Slice()
			if len(args) < 2 {
				return cli.Exit("Not enough arguments", 1)
			}
			return recv(args[0], args[1], c.String("i"), c.String("o"))
		},
	}
}
