// Package main is the entrypoint for the redial_proxy application.
// It initializes a SOCKS5 proxy server bounded to the loopback interface,
// equipped with a custom dialer (Redialer) that intercepts connection failures
// and transparently retries requests when routing errors occur.
package main

import (
	"flag"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/armon/go-socks5"
	"github.com/lucasew/go-getlistener"
	"github.com/lucasew/redial_proxy/internal/dialer"
)

const (
	defaultPort = 8889
)

func main() {
	var port int
	var host string
	flag.IntVar(&port, "p", defaultPort, "port to listen the server")
	flag.StringVar(&host, "H", "127.0.0.1", "host to listen the server")
	flag.Parse()

	slog.Info("starting...")

	// go-getlistener library retrieves configuration strictly from environment variables.
	// We parse standard command-line flags for user convenience and bridge them into
	// the expected environment variables (PORT and HOST) before initializing the listener.
	if err := os.Setenv("PORT", fmt.Sprintf("%d", port)); err != nil {
		slog.Error("failed to set PORT env", "err", err)
		os.Exit(1)
	}
	if err := os.Setenv("HOST", host); err != nil {
		slog.Error("failed to set HOST env", "err", err)
		os.Exit(1)
	}

	d := &dialer.Redialer{
		MaxRetries: 3,
		RetryDelay: 100 * time.Millisecond,
	}

	sconfig := socks5.Config{
		Dial: d.DialContext,
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
