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

	"github.com/lucasew/go-getlistener"
	"github.com/things-go/go-socks5"
)

func init() {
	flag.IntVar(&getlistener.PORT, "p", getlistener.PORT, "port to listen the server")
	flag.Parse()
	if getlistener.PORT == 0 {
		getlistener.PORT = 8889
	}
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

	// Add username/password authentication if env vars are set.
	// To enable authentication, set the PROXY_USERNAME and PROXY_PASSWORD environment variables.
	username := os.Getenv("PROXY_USERNAME")
	password := os.Getenv("PROXY_PASSWORD")
	if username != "" && password != "" {
		creds := make(socks5.StaticCredentials)
		creds[username] = password
		cator := socks5.UserPassAuthenticator{Credentials: creds}
		opts = append(opts, socks5.WithAuthMethods([]socks5.Authenticator{cator}))
	}

	// Create the server
	server := socks5.NewServer(opts...)

	ln, err := getlistener.GetListener()
	if err != nil {
		log.Fatal(err)
	}

	if err := server.Serve(ln); err != nil {
		log.Fatal(err)
	}
}
