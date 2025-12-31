package main

import (
	"context"
	"flag"
	"fmt"
	"net"
	"os"
	"strings"
	"time"

	"log"

	"github.com/armon/go-socks5"
	"github.com/lucasew/go-getlistener"
)

func init() {
	flag.IntVar(&getlistener.PORT, "p", getlistener.PORT, "port to listen the server")
	flag.Parse()
	if getlistener.PORT == 0 {
		getlistener.PORT = 8889
	}
}

const (
	maxRetries         = 3
	retrySleepDuration = 100 * time.Millisecond
)

func redial(ctx context.Context, network, addr string) (net.Conn, error) {
	try := 0
	for {
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("timeout")
		default:
			if try > maxRetries {
				return nil, fmt.Errorf("too many retries")
			}
			conn, err := net.Dial(network, addr)
			if err != nil {
				log.Printf("conn err '%s'", err.Error())
				if strings.Contains(err.Error(), "route") {
					log.Printf("retrying connection to %s %s (%d)", network, addr, try)
					time.Sleep(retrySleepDuration)
					try++
					continue
				}
				return nil, err
			}
			log.Printf("CONNECT %s %s", network, addr)
			return conn, err
		}
	}
}

func main() {
	log.Printf("starting...")
	sconfig := socks5.Config{
		Dial: redial,
	}

	proxyUsername := os.Getenv("PROXY_USERNAME")
	proxyPassword := os.Getenv("PROXY_PASSWORD")

	// If both username and password are provided, enable authentication.
	// This makes the authentication optional and avoids breaking existing setups.
	if proxyUsername != "" && proxyPassword != "" {
		log.Printf("authentication enabled")
		cred := socks5.StaticCredentials{
			proxyUsername: proxyPassword,
		}
		sconfig.AuthMethods = []socks5.Authenticator{
			socks5.UserPassAuthenticator{
				Credentials: cred,
			},
		}
	}

	srv, err := socks5.New(&sconfig)
	if err != nil {
		log.Fatal(err)
	}
	ln, err := getlistener.GetListener()
	if err != nil {
		log.Fatal(err)
	}
	err = srv.Serve(ln)
	if err != nil {
		log.Fatal(err)
	}
}
