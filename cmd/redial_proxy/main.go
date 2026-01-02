package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"net"
	"os"
	"strings"
	"time"

	"github.com/armon/go-socks5"
	"github.com/lucasew/go-getlistener"
)

const (
	defaultPort        = 8889
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
				slog.Warn("conn err", "err", err.Error())
				if strings.Contains(err.Error(), "route") {
					slog.Info("retrying connection", "network", network, "addr", addr, "try", try)
					time.Sleep(retrySleepDuration)
					try++
					continue
				}
				return nil, err
			}
			slog.Info("CONNECT", "network", network, "addr", addr)
			return conn, err
		}
	}
}

func main() {
	flag.IntVar(&getlistener.PORT, "p", getlistener.PORT, "port to listen the server")
	flag.Parse()
	if getlistener.PORT == 0 {
		getlistener.PORT = defaultPort
	}
	slog.Info("starting...")
	sconfig := socks5.Config{
		Dial: redial,
	}
	srv, err := socks5.New(&sconfig)
	if err != nil {
		slog.Error("failed to create socks5 server", "err", err)
		os.Exit(1)
	}
	ln, err := getlistener.GetListener()
	if err != nil {
		slog.Error("failed to get listener", "err", err)
		os.Exit(1)
	}
	err = srv.Serve(ln)
	if err != nil {
		slog.Error("failed to serve", "err", err)
		os.Exit(1)
	}
}
