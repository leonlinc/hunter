package pcap

import (
	"os"
	"path"

	"github.com/urfave/cli/v2"
)

var (
	ParseTS *cli.Command
)

func init() {
	ParseTS = &cli.Command{
		Name:      "pcap-parse-ts",
		Usage:     "parse TS in PCAP file",
		ArgsUsage: "pcap-file",
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:  "x",
				Usage: "extract payload to pcap-file.ts",
			},
		},
		Action: func(c *cli.Context) error {
			args := c.Args().Slice()
			if len(args) < 1 {
				return cli.Exit("Not enough arguments", 1)
			}
			file := args[0]
			root := path.Base(file) + ".log"
			if err := os.MkdirAll(root, os.ModePerm); err != nil {
				return err
			}
			p := CreateParser()
			p.Parse(file, root)
			return nil
		},
	}
}
