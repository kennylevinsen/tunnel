package main

import (
	"crypto/tls"
	"fmt"
	"io"
	"log"
	"net"
	"os"
)

func main() {
	if len(os.Args) < 3 {
		fmt.Printf("Too few arguments\n")
		fmt.Printf("Usage:\n")
		fmt.Printf("	tunnel [local host:port] [remote host:port]\n")
		fmt.Printf("Example:\n")
		fmt.Printf("	tunnel :9999 example.com:443\n")
		return
	}

	l, err := net.Listen("tcp", os.Args[1])
	if err != nil {
		panic(err)
	}

	for {
		conn, err := l.Accept()
		if err != nil {
			panic(err)
		}

		go func() {
			tlsConn, err := tls.Dial("tcp", os.Args[2], &tls.Config{
				InsecureSkipVerify: true,
			})

			if err != nil {
				log.Printf("%s -> %s: failed: %s\n", conn.RemoteAddr(), os.Args[2], err)
				return
			}
			log.Printf("%s -> %s: connected\n", conn.RemoteAddr(), tlsConn.RemoteAddr())

			defer func() {
				log.Printf("%s -> %s: disconnected\n", conn.RemoteAddr(), tlsConn.RemoteAddr())
				tlsConn.Close()
				conn.Close()
			}()

			go io.Copy(tlsConn, conn)
			io.Copy(conn, tlsConn)
		}()
	}
}
