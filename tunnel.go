package main

import (
	"crypto/tls"
	"fmt"
	"io"
	"net"
	"os"
)

func main() {
	if len(os.Args) < 3 {
		panic("Too few arguments")
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
			fmt.Printf("Conncting... %s\n", os.Args[2])
			if err != nil {
				fmt.Printf("Failed %s\n", err)
				return
			}
			defer tlsConn.Close()
			defer conn.Close()

			go io.Copy(tlsConn, conn)
			io.Copy(conn, tlsConn)
		}()
	}
}
