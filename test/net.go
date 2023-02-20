package test

import (
	"bytes"
	"crypto/rand"
	"io"
	"net"
	"strconv"
	"sync"
	"time"
)

func PickPort(network string, host string) int {
	switch network {
	case "tcp":
		for retry := 0; retry < 16; retry++ {
			l, err := net.Listen("tcp", host+":0")
			if err != nil {
				continue
			}
			defer l.Close()
			_, port, err := net.SplitHostPort(l.Addr().String())
			Must(err)
			p, err := strconv.ParseInt(port, 10, 32)
			Must(err)
			return int(p)
		}
	case "udp":
		for retry := 0; retry < 16; retry++ {
			conn, err := net.ListenPacket("udp", host+":0")
			if err != nil {
				continue
			}
			defer conn.Close()
			_, port, err := net.SplitHostPort(conn.LocalAddr().String())
			Must(err)
			p, err := strconv.ParseInt(port, 10, 32)
			Must(err)
			return int(p)
		}
	default:
		return 0
	}
	return 0
}

func RunTCPEchoServer(addr string) {
	listener, err := net.Listen("tcp", addr)
	Must(err)

	go func() {
		defer listener.Close()
		for {
			conn, err := listener.Accept()
			if err != nil {
				return
			}
			go func(conn net.Conn) {
				defer conn.Close()
				for {
					buf := make([]byte, 2048)
					conn.SetDeadline(time.Now().Add(time.Second * 5))
					n, err := conn.Read(buf)
					conn.SetDeadline(time.Time{})
					if err != nil {
						return
					}
					_, err = conn.Write(buf[0:n])
					if err != nil {
						return
					}
				}
			}(conn)
		}
	}()
}

func RunUDPEchoServer(addr string) {
	conn, err := net.ListenPacket("udp", addr)
	Must(err)

	go func() {
		for {
			buf := make([]byte, 1024*8)
			n, addr, err := conn.ReadFrom(buf)
			if err != nil {
				return
			}
			conn.WriteTo(buf[0:n], addr)
		}
	}()
}

func GeneratePayload(length int) []byte {
	buf := make([]byte, length)
	io.ReadFull(rand.Reader, buf)
	return buf
}

// CheckConn checks if two netConn were connected and work properly
func CheckConn(a net.Conn, b net.Conn) bool {
	payload1 := make([]byte, 1024)
	payload2 := make([]byte, 1024)

	result1 := make([]byte, 1024)
	result2 := make([]byte, 1024)

	rand.Reader.Read(payload1)
	rand.Reader.Read(payload2)

	wg := sync.WaitGroup{}
	wg.Add(2)

	go func() {
		a.Write(payload1)
		a.Read(result2)
		wg.Done()
	}()

	go func() {
		b.Read(result1)
		b.Write(payload2)
		wg.Done()
	}()

	wg.Wait()

	return bytes.Equal(payload1, result1) && bytes.Equal(payload2, result2)
}
