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
	"github.com/lucasew/redial_proxy/internal/errorreport"
)

const (
	defaultPort       = 8889
	defaultMaxRetries = 3
	defaultRetryDelay = 100 * time.Millisecond
)

func setupEnv(port int, host string) {
	if err := os.Setenv("PORT", fmt.Sprintf("%d", port)); err != nil {
		errorreport.ReportFatal("failed to set PORT env", err)
	}
	if err := os.Setenv("HOST", host); err != nil {
		errorreport.ReportFatal("failed to set HOST env", err)
	}
}

func runServer(d *dialer.Redialer) {
	sconfig := socks5.Config{
		Dial: d.DialContext,
	}
	srv, err := socks5.New(&sconfig)
	if err != nil {
		errorreport.ReportFatal("failed to create socks5 server", err)
	}
	ln, err := getlistener.GetListener()
	if err != nil {
		errorreport.ReportFatal("failed to get listener", err)
	}
	err = srv.Serve(ln)
	if err != nil {
		errorreport.ReportFatal("failed to serve", err)
	}
}

func main() {
	var port int
	var host string
	flag.IntVar(&port, "p", defaultPort, "port to listen the server")
	flag.StringVar(&host, "H", "127.0.0.1", "host to listen the server")
	flag.Parse()

	slog.Info("starting...")

	// Pass configuration to getlistener via environment variables
	setupEnv(port, host)

	d := &dialer.Redialer{
		MaxRetries: defaultMaxRetries,
		RetryDelay: defaultRetryDelay,
	}

	runServer(d)
}
