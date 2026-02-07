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
	"github.com/lucasew/redial_proxy/internal/utils"
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
	if err := os.Setenv("PORT", fmt.Sprintf("%d", port)); err != nil {
		utils.ReportFatal(err, "failed to set PORT environment variable")
	}
	if err := os.Setenv("HOST", host); err != nil {
		utils.ReportFatal(err, "failed to set HOST environment variable")
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
		utils.ReportFatal(err, "failed to create socks5 server")
	}
	ln, err := getlistener.GetListener()
	if err != nil {
		utils.ReportFatal(err, "failed to get listener")
	}
	err = srv.Serve(ln)
	if err != nil {
		utils.ReportFatal(err, "failed to serve")
	}
}
