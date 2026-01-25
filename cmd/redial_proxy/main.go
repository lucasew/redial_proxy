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

	// Pass configuration to getlistener via environment variables
	os.Setenv("PORT", fmt.Sprintf("%d", port))
	os.Setenv("HOST", host)

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
