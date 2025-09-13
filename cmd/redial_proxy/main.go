package main

import (
	"context"
	"flag"
	"fmt"
	"net"
	"strings"
	"time"

	"log"

	"github.com/armon/go-socks5"
)

var PORT int

func init() {
	flag.IntVar(&PORT, "p", 8889, "port to listen the server")
	flag.Parse()
}

func main() {
	log.Printf("starting...")
	sconfig := socks5.Config{
		Dial: func(ctx context.Context, network, addr string) (net.Conn, error) {
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
		},
	}
	srv, err := socks5.New(&sconfig)
	if err != nil {
		log.Fatal(err)
	}
	ln, err := GetListener()
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("binding port %d...", PORT)
	err = srv.Serve(ln)
	if err != nil {
		log.Fatal(err)
	}
}
