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
		println("	tls, plain")
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

	log.Printf("Transport: %s, binding address: %s, destination: %s", transport, local, remote)

	for {
		conn, err := l.Accept()
		if err != nil {
			panic(err)
		}

		go func() {
			pconn, err := handler()
			if err != nil {
				log.Printf("%s -> %s: failed: %s\n", conn.RemoteAddr(), os.Args[2], err)
				return
			}
			log.Printf("%s -> %s: connected\n", conn.RemoteAddr(), pconn.RemoteAddr())

			closer := make(chan bool, 1)

			go func() {
				<-closer
				conn.Close()
				pconn.Close()
				log.Printf("%s -> %s: disconnected\n", conn.RemoteAddr(), pconn.RemoteAddr())
			}()

			go func() {
				io.Copy(pconn, conn)
				closer <- true
			}()

			go func() {
				io.Copy(conn, pconn)
				closer <- true
			}()
		}()
	}
}
