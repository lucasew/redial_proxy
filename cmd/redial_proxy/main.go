package main

import (
	"flag"
	"io"
	"log"
	"log/slog"
	"os"
	"strconv"
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

func main() {
	var port int
	var host string
	flag.IntVar(&port, "p", defaultPort, "port to listen the server")
	flag.StringVar(&host, "H", "127.0.0.1", "host to listen the server")
	flag.Parse()

	slog.Info("starting...")

	if host != "127.0.0.1" && host != "localhost" && host != "::1" {
		slog.Warn("proxy is bound to a non-loopback network interface, exposing it to SSRF risks")
	}

	// Pass configuration to getlistener via environment variables
	if err := os.Setenv("PORT", strconv.Itoa(port)); err != nil {
		errorreport.ReportFatal("failed to set PORT env", err)
	}
	if err := os.Setenv("HOST", host); err != nil {
		errorreport.ReportFatal("failed to set HOST env", err)
	}

	srv, err := socks5.New(&socks5.Config{
		Dial: (&dialer.Redialer{
			MaxRetries: defaultMaxRetries,
			RetryDelay: defaultRetryDelay,
		}).DialContext,
		Logger: log.New(io.Discard, "", 0),
	})
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
