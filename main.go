package main

import (
	"io"
	"log"
	"net"
	"os"
	"time"

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
					Value: "",
					Usage: "specifies the interface address to use",
				},
				&cli.Int64Flag{
					Name:  "b",
					Value: 0,
					Usage: "specifies the bitrate of the file",
				},
			},
			Action: func(c *cli.Context) error {
				args := c.Args().Slice()
				if len(args) == 3 {
					send(args[0], args[1], args[2], c.String("i"), c.Int64("b"))
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

func send(file, ipaddr, port, iface string, bitrate int64) {
	f, err := os.Open(file)
	if err != nil {
		log.Fatalln(err)
	}
	raddr, err := net.ResolveUDPAddr("udp", net.JoinHostPort(ipaddr, port))
	if err != nil {
		log.Fatalln(err)
	}
	laddr, err := net.ResolveUDPAddr("udp", net.JoinHostPort(iface, ""))
	if err != nil {
		log.Fatalln(err)
	}
	conn, err := net.DialUDP("udp", laddr, raddr)
	if err != nil {
		log.Fatalln(err)
	}
	defer f.Close()
	defer conn.Close()

	if bitrate == 0 {
		log.Fatal("Sorry, please specify the bitrate...")
	}

	log.Println("local addr:", conn.LocalAddr())

	beginning := time.Now()
	packet := make([]byte, 1316)
	ticks := time.Tick(100 * time.Microsecond)
	var elapsed time.Duration
	var sent, target int64
	for _ = range ticks {
		elapsed = time.Since(beginning)
		target = int64(float64(bitrate) * elapsed.Seconds())
		for sent < target {
			n, err := f.Read(packet)
			if n != 0 {
				sent += int64(n * 8)
				conn.Write(packet[:n])
			}
			if err != nil {
				if err == io.EOF {
					log.Println("source loop")
					f.Seek(0, 0)
				} else {
					log.Fatalln(err)
				}
			}
		}
	}
}
