package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"log/slog"
	"net"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/armon/go-socks5"
	"github.com/lucasew/go-getlistener"
	"github.com/lucasew/redial_proxy/internal/dialer"
	"github.com/lucasew/redial_proxy/internal/errorreport"
)

const (
	defaultHost       = "127.0.0.1"
	defaultPort       = 8889
	defaultMaxRetries = 3
	defaultRetryDelay = 100 * time.Millisecond
)

func main() {
	var port int
	var host string
	var maxRetries int
	var retryDelay time.Duration
	flag.IntVar(&port, "p", defaultPort, "port to listen the server")
	flag.StringVar(&host, "H", defaultHost, "host to listen the server")
	flag.IntVar(&maxRetries, "retries", defaultMaxRetries, "max dial retries on route-like errors")
	flag.DurationVar(&retryDelay, "retry-delay", defaultRetryDelay, "delay between dial retries")
	flag.Parse()

	if maxRetries < 0 {
		errorreport.ReportFatal("invalid -retries", fmt.Errorf("must be >= 0, got %d", maxRetries))
	}
	if retryDelay < 0 {
		errorreport.ReportFatal("invalid -retry-delay", fmt.Errorf("must be >= 0, got %v", retryDelay))
	}

	slog.Info("starting...", "retries", maxRetries, "retry_delay", retryDelay)

	if !isLoopbackHost(host) {
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
			MaxRetries: maxRetries,
			RetryDelay: retryDelay,
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

// isLoopbackHost reports whether host is a loopback address or the
// conventional "localhost" name. Used to warn when -H exposes the proxy
// beyond the intended local-only bind (see AGENTS.md).
func isLoopbackHost(host string) bool {
	if strings.EqualFold(host, "localhost") {
		return true
	}
	ip := net.ParseIP(host)
	return ip != nil && ip.IsLoopback()
}
