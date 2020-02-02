package main

import (
	"errors"
	"io"
	"log"
	"net"
	"os"
	"time"

	"github.com/urfave/cli/v2"
	"golang.org/x/net/ipv4"
)

func main() {
	send := cli.Command{
		Name:      "send",
		Aliases:   []string{"s"},
		Usage:     "send UDP stream",
		ArgsUsage: "file host port",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "iface-addr",
				Aliases: []string{"i"},
				Value:   "",
				Usage:   "specifies the network interface IP address",
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
	}
	recv := cli.Command{
		Name:      "recv",
		Aliases:   []string{"r"},
		Usage:     "receive UDP stream",
		ArgsUsage: "host port",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "iface",
				Aliases: []string{"i"},
				Value:   "",
				Usage:   "specifies the network interface name",
			},
		},
		Action: func(c *cli.Context) error {
			args := c.Args().Slice()
			if len(args) == 2 {
				return recv(args[0], args[1], c.String("i"))
			} else {
				cli.ShowCommandHelpAndExit(c, "r", 1)
				return nil
			}
		},
	}
	app := &cli.App{
		Commands: []*cli.Command{
			&send,
			&recv,
		},
	}
	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}

func send(file, host, port, iface_addr string, bitrate int64) error {
	if bitrate == 0 {
		return errors.New("Sorry but please specify the bitrate for now")
	}

	f, err := os.Open(file)
	if err != nil {
		return err
	}
	defer f.Close()
	laddr, err := net.ResolveUDPAddr("udp", net.JoinHostPort(iface_addr, ""))
	if err != nil {
		return err
	}
	raddr, err := net.ResolveUDPAddr("udp", net.JoinHostPort(host, port))
	if err != nil {
		return err
	}
	conn, err := net.DialUDP("udp", laddr, raddr)
	if err != nil {
		return err
	}
	defer conn.Close()
	log.Println("local addr:", conn.LocalAddr())

	beginning := time.Now()
	packet := make([]byte, 1316)
	ticks := time.Tick(time.Millisecond)
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
					return err
				}
			}
		}
	}
	return nil
}

func recv(host, port, iface string) error {
	ifi, err := net.InterfaceByName(iface)
	if err != nil {
		return err
	}
	addr, err := net.ResolveUDPAddr("udp", net.JoinHostPort(host, port))
	if err != nil {
		return err
	}
	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		return err
	}
	defer conn.Close()
	pktConn := ipv4.NewPacketConn(conn)
	if addr.IP.IsMulticast() {
		if err := pktConn.JoinGroup(ifi, addr); err != nil {
			return err
		}
	}
	buffer := make([]byte, 1500)
	for {
		n, cm, _, err := pktConn.ReadFrom(buffer)
		if err != nil {
			return err
		}
		log.Println(n, cm)
	}
}
