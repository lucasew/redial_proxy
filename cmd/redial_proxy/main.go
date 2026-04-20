// Package main implements a resilient SOCKS5 proxy server.
// It uses a custom dialer that automatically retries connections
// on routing errors, masking transient network failures from clients.
// This proxy is specifically designed to run locally (binding to loopback).
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
	// defaultPort is the port the SOCKS5 proxy will listen on if not specified.
	defaultPort = 8889
)

// main is the entry point for the redial_proxy application.
//
// It performs the following setup:
//  1. Parses command-line flags (-H for host, -p for port).
//  2. Mutates environment variables (PORT, HOST) as a side-effect
//     to configure the underlying go-getlistener package.
//  3. Initializes a custom Redialer to handle transient routing errors.
//  4. Starts a local SOCKS5 proxy server.
//
// Security Note: This proxy is strictly intended for local execution and
// should only be bound to the loopback interface (e.g., 127.0.0.1) to
// prevent unauthorized external access.
func main() {
	var port int
	var host string
	flag.IntVar(&port, "p", defaultPort, "port to listen the server")
	flag.StringVar(&host, "H", "127.0.0.1", "host to listen the server")
	flag.Parse()

	slog.Info("starting...")

	// Pass configuration to getlistener via environment variables
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
