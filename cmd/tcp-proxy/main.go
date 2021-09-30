package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"strings"
	"sync"
)

var (
	address = flag.String("address", "127.0.0.1", "Addresses to listen on.")
)

func listen(localAddr, remoteAddr string) (err error) {
	log.Printf("Forwarding from %s -> %s\n", localAddr, remoteAddr)
	listener, err := net.Listen("tcp", localAddr)
	if err != nil {
		panic(err)
	}
	for {
		conn, err := listener.Accept()
		log.Println("New connection", conn.RemoteAddr())
		if err != nil {
			log.Println("error accepting connection", err)
			continue
		}
		go func() {
			defer conn.Close()
			conn2, err := net.Dial("tcp", remoteAddr)
			if err != nil {
				log.Println("error dialing remote addr", err)
				return
			}
			defer conn2.Close()
			closer := make(chan struct{}, 2)
			go copyWithCloser(closer, conn2, conn)
			go copyWithCloser(closer, conn, conn2)
			<-closer
			log.Println("Connection complete", conn.RemoteAddr())
		}()
	}
}

func copyWithCloser(closer chan struct{}, dst io.Writer, src io.Reader) {
	_, _ = io.Copy(dst, src)
	closer <- struct{}{} // connection is closed, send signal to stop proxy
}

func main() {
	flag.Parse()
	args := flag.Args()

	if len(args) < 2 {
		log.Fatalf("accept 2 arg(s), received %d", len(args))
	}

	remote := args[0]
	ports := args[1:]

	var wg sync.WaitGroup
	for _, port := range ports {
		ps := strings.Split(port, ":")
		if len(ps) > 2 {
			log.Fatalf("invalid port %s", port)
		}
		if len(ps) == 1 {
			ps = append(ps, ps[0])
		}
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := listen(
				fmt.Sprintf("%s:%s", *address, ps[0]),
				fmt.Sprintf("%s:%s", remote, ps[1]),
			); err != nil {
				log.Fatal(err)
			}
		}()
	}

	wg.Wait()
}
