package main

import (
	"crypto/tls"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"strings"
)

var (
	localcert   = flag.String("lcert", "", "local certificate")
	localkey    = flag.String("lkey", "", "local key")
	localprotos = flag.String("lprotos", "", "local protos")
)

func usage() {
	println("Too few arguments")
	println("Usage:")
	println("	tunnel [options] [local transport:host:port] [remote transport:remote host:port]")
	println("Example:")
	println("	tunnel plain::9999 tls:example.com:443")
	println("Supported transports:")
	println("	tls, plain")
	return
}

func main() {
	flag.Parse()
	args := flag.Args()
	if len(args) != 2 {
		usage()
		return
	}

	local, remote := args[0], args[1]

	localParts := strings.Split(local, ":")
	if len(localParts) != 3 {
		println("local address incomplete")
		usage()
		return
	}
	remoteParts := strings.Split(remote, ":")
	if len(remoteParts) != 3 {
		println("remote address incomplete")
		usage()
		return
	}

	localTransport, localHost, localPort := localParts[0], localParts[1], localParts[2]
	remoteTransport, remoteHost, remotePort := remoteParts[0], remoteParts[1], remoteParts[2]

	var serverHandler func(net.Conn) (net.Conn, error)
	switch localTransport {
	case "plain":
		serverHandler = func(c net.Conn) (net.Conn, error) {
			return c, nil
		}
	case "tls":
		certificate, err := tls.LoadX509KeyPair(*localcert, *localkey)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error while loading local certificates: %v\n", err)
			return
		}

		config := &tls.Config{
			NextProtos:   strings.Split(*localprotos, ","),
			Certificates: []tls.Certificate{certificate},
		}

		serverHandler = func(c net.Conn) (net.Conn, error) {
			return tls.Server(c, config), nil
		}
	}

	var clientHandler func() (net.Conn, error)
	switch remoteTransport {
	case "tls":
		clientHandler = func() (net.Conn, error) {
			return tls.Dial("tcp", remoteHost+":"+remotePort, &tls.Config{
				InsecureSkipVerify: true,
			})
		}
	case "plain":
		clientHandler = func() (net.Conn, error) {
			return net.Dial("tcp", remoteHost+":"+remotePort)
		}
	}

	l, err := net.Listen("tcp", localHost+":"+localPort)
	if err != nil {
		panic(err)
	}

	log.Printf("Local address: %s, destination: %s", local, remote)

	for {
		origconn, err := l.Accept()
		if err != nil {
			log.Printf("Error in accept: %v", err)
		}

		conn, err := serverHandler(origconn)
		if err != nil {
			log.Printf("%s -> %s: local transport failed: %s\n", origconn.RemoteAddr(), err)
		}

		go func() {
			pconn, err := clientHandler()
			if err != nil {
				log.Printf("%s -> %s: failed: %s\n", conn.RemoteAddr(), remote, err)
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
