package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"strings"
	"time"

	"github.com/things-go/go-socks5"
)

var port int

func init() {
	flag.IntVar(&port, "p", 8889, "port to listen the server")
	flag.Parse()
}

func main() {
	log.Printf("starting...")

	// Create a SOCKS5 server options
	opts := []socks5.Option{
		socks5.WithDial(func(ctx context.Context, network, addr string) (net.Conn, error) {
			try := 0
			for {
				select {
				case <-ctx.Done():
					return nil, fmt.Errorf("timeout")
				default:
					if try > 3 {
						return nil, fmt.Errorf("too many retries")
					}
					conn, err := net.Dial(network, addr)
					if err != nil {
						log.Printf("conn err '%s'", err.Error())
						if strings.Contains(err.Error(), "route") {
							log.Printf("retrying connection to %s %s (%d)", network, addr, try)
							time.Sleep(100 * time.Millisecond)
							try++
							continue
						}
						return nil, err
					}
					log.Printf("CONNECT %s %s", network, addr)
					return conn, err
				}
			}
		}),
		socks5.WithLogger(socks5.NewLogger(log.New(os.Stdout, "socks5: ", log.LstdFlags))),
	}

	// Create the server
	server := socks5.NewServer(opts...)

	listenAddr := fmt.Sprintf("127.0.0.1:%d", port)
	log.Printf("SOCKS5 proxy listening on %s", listenAddr)
	ln, err := net.Listen("tcp", listenAddr)
	if err != nil {
		log.Fatal(err)
	}

	if err := server.Serve(ln); err != nil {
		log.Fatal(err)
	}
}
