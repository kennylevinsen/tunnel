package main

import (
	"crypto/tls"
	"io"
	"log"
	"net"
	"os"
)

func main() {
	if len(os.Args) < 4 {
		println("Too few arguments")
		println("Usage:")
		println("	tunnel [transport] [local host:port] [remote host:port]")
		println("Example:")
		println("	tunnel tls :9999 example.com:443")
		println("Supported transports:")
		println("   tls, plain")
		return
	}

	transport, local, remote := os.Args[1], os.Args[2], os.Args[3]

	var handler func() (net.Conn, error)
	switch transport {
	case "tls":
		handler = func() (net.Conn, error) {
			return tls.Dial("tcp", remote, &tls.Config{
				InsecureSkipVerify: true,
			})
		}
	case "plain":
		handler = func() (net.Conn, error) {
			return net.Dial("tcp", remote)
		}
	}

	l, err := net.Listen("tcp", local)
	if err != nil {
		panic(err)
	}

	for {
		conn, err := l.Accept()
		if err != nil {
			panic(err)
		}

		go func() {
			proxyConn, err := handler()
			if err != nil {
				log.Printf("%s -> %s: failed: %s\n", conn.RemoteAddr(), os.Args[2], err)
				return
			}
			log.Printf("%s -> %s: connected\n", conn.RemoteAddr(), proxyConn.RemoteAddr())

			defer func() {
				log.Printf("%s -> %s: disconnected\n", conn.RemoteAddr(), proxyConn.RemoteAddr())
				proxyConn.Close()
				conn.Close()
			}()

			go io.Copy(proxyConn, conn)
			io.Copy(conn, proxyConn)
		}()
	}
}
