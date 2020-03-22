package m2ts

import (
	"errors"
	"io"
	"log"
	"net"
	"os"
	"time"

	"golang.org/x/net/ipv4"
)

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
