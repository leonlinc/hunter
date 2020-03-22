package main

import (
	"errors"
	"io"
	"log"
	"net"
	"os"
	"time"

	"github.com/leonlinc/hunter/pcap"
	"github.com/urfave/cli/v2"
	"golang.org/x/net/ipv4"
)

var (
	root string = "hunter.log"
)

func init() {
	err := os.MkdirAll(root, os.ModePerm)
	if err != nil {
		log.Fatal(err)
	}
}

func main() {
	send := cli.Command{
		Name:      "send",
		Aliases:   []string{"s"},
		Usage:     "send TS over UDP",
		ArgsUsage: "file host port",
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
			if len(args) == 3 {
				return send(args[0], args[1], args[2], c.String("i"), c.Int64("b"))
			} else {
				cli.ShowCommandHelpAndExit(c, "s", 1)
				return nil
			}
		},
	}
	recv := cli.Command{
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
			if len(args) == 2 {
				return recv(args[0], args[1], c.String("i"), c.String("o"))
			} else {
				cli.ShowCommandHelpAndExit(c, "r", 1)
				return nil
			}
		},
	}
	parseTsPcap := cli.Command{
		Name:      "parse-ts-pcap",
		Usage:     "parse TS in pcap",
		ArgsUsage: "file",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "x",
				Usage: "extract TS to file",
			},
		},
		Action: func(c *cli.Context) error {
			args := c.Args().Slice()
			if len(args) == 1 {
				p := pcap.CreateParser(root)
				p.Parse(args[0])
				p.Close()
				return nil
			} else {
				cli.ShowCommandHelpAndExit(c, "parse-pcap", 1)
				return nil
			}
		},
	}
	app := &cli.App{
		Commands: []*cli.Command{
			&send,
			&recv,
			&parseTsPcap,
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}

func send(file, host, port, iface string, bitrate int64) error {
	f, err := os.Open(file)
	if err != nil {
		return err
	}
	defer f.Close()

	addr, err := net.ResolveUDPAddr("udp", net.JoinHostPort(host, port))
	if err != nil {
		return err
	}

	c, err := net.ListenUDP("udp", nil)
	if err != nil {
		return err
	}
	conn := ipv4.NewPacketConn(c)
	defer conn.Close()

	if iface != "" {
		ifi, err := net.InterfaceByName(iface)
		if err != nil {
			return err
		}
		err = conn.SetMulticastInterface(ifi)
		if err != nil {
			return err
		}
	}

	if bitrate == 0 {
		return errors.New("for now bitrate must not be 0")
	}

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
				conn.WriteTo(packet, nil, addr)
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

func recv(host, port, iface, file string) error {
	addr, err := net.ResolveUDPAddr("udp", net.JoinHostPort(host, port))
	if err != nil {
		return err
	}

	c, err := net.ListenUDP("udp", addr)
	if err != nil {
		return err
	}
	conn := ipv4.NewPacketConn(c)
	defer conn.Close()

	var ifi *net.Interface
	if addr.IP.IsMulticast() {
		if iface == "" {
			return errors.New("please specifiy which network interface to join multicast")
		}
		if ifi, err = net.InterfaceByName(iface); err != nil {
			return err
		}
		if err = conn.JoinGroup(ifi, addr); err != nil {
			return err
		}
	}

	var out *os.File
	if file != "" {
		if out, err = os.Create(file); err != nil {
			return err
		}
	}

	buffer := make([]byte, 1500)
	for {
		n, _, _, err := conn.ReadFrom(buffer)
		if err != nil {
			return err
		}
		if out != nil {
			out.Write(buffer[:n])
		}
	}
}
