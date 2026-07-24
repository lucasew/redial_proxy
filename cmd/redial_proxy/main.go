package main

import (
	"context"
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
	defaultHost        = "127.0.0.1"
	defaultPort        = 8889
	defaultMaxRetries  = 3
	defaultRetryDelay  = 100 * time.Millisecond
	defaultDialTimeout = 10 * time.Second
)

func main() {
	var port int
	var host string
	var maxRetries int
	var retryDelay time.Duration
	var dialTimeout time.Duration
	var allowNonLoopback bool
	flag.IntVar(&port, "p", defaultPort, "port to listen the server")
	flag.StringVar(&host, "H", defaultHost, "host to listen the server")
	flag.IntVar(&maxRetries, "retries", defaultMaxRetries, "max dial/DNS retries on transient failures")
	flag.DurationVar(&retryDelay, "retry-delay", defaultRetryDelay, "delay between dial/DNS retries")
	flag.DurationVar(&dialTimeout, "dial-timeout", defaultDialTimeout, "max time for an outbound dial including retries (0 disables)")
	flag.BoolVar(&allowNonLoopback, "allow-non-loopback", false, "allow -H outside loopback (SSRF risk; default refuse)")
	flag.Parse()

	if maxRetries < 0 {
		errorreport.ReportFatal("invalid -retries", fmt.Errorf("must be >= 0, got %d", maxRetries))
	}
	if retryDelay < 0 {
		errorreport.ReportFatal("invalid -retry-delay", fmt.Errorf("must be >= 0, got %v", retryDelay))
	}
	if dialTimeout < 0 {
		errorreport.ReportFatal("invalid -dial-timeout", fmt.Errorf("must be >= 0, got %v", dialTimeout))
	}
	if err := checkListenHost(host, allowNonLoopback); err != nil {
		errorreport.ReportFatal("invalid -H", err)
	}

	slog.Info("starting...", "retries", maxRetries, "retry_delay", retryDelay, "dial_timeout", dialTimeout)

	if !isLoopbackHost(host) {
		// Only reachable with -allow-non-loopback.
		slog.Warn("proxy is bound to a non-loopback network interface, exposing it to SSRF risks")
	}

	// Pass configuration to getlistener via environment variables.
	// IPv6 literals must be bracketed: getlistener builds the address with
	// fmt.Sprintf("%s:%d", host, port), which yields an invalid "::1:port"
	// unless host is already "[::1]".
	host = hostForGetListener(host)
	if err := os.Setenv("PORT", strconv.Itoa(port)); err != nil {
		errorreport.ReportFatal("failed to set PORT env", err)
	}
	if err := os.Setenv("HOST", host); err != nil {
		errorreport.ReportFatal("failed to set HOST env", err)
	}

	// go-socks5 resolves FQDNs before Dial via NameResolver. The stock
	// DNSResolver uses net.ResolveIPAddr (no context, no retry), so flaky
	// DNS hangs or fails once without ever reaching the redialer. Share the
	// dial retry budget and bound each lookup attempt.
	resolver := &dialer.RetryResolver{
		MaxRetries: maxRetries,
		RetryDelay: retryDelay,
	}
	redialer := &dialer.Redialer{
		MaxRetries: maxRetries,
		RetryDelay: retryDelay,
	}
	// go-socks5 dials with context.Background() and no deadline, so a
	// blackholed destination can block a SOCKS handler until the OS TCP
	// timeout (often minutes). Bound the full outbound dial (including
	// redialer retries) here.
	dial := withDialTimeout(dialTimeout, redialer.DialContext)

	srv, err := socks5.New(&socks5.Config{
		Dial:     dial,
		Resolver: resolver,
		Logger:   log.New(io.Discard, "", 0),
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

// withDialTimeout returns a Dial function that cancels ctx after timeout.
// timeout <= 0 leaves the dial unbounded (previous behavior).
// The budget covers the full Redialer.DialContext call, including retries.
func withDialTimeout(timeout time.Duration, dial func(context.Context, string, string) (net.Conn, error)) func(context.Context, string, string) (net.Conn, error) {
	if dial == nil {
		return nil
	}
	if timeout <= 0 {
		return dial
	}
	return func(ctx context.Context, network, addr string) (net.Conn, error) {
		ctx, cancel := context.WithTimeout(ctx, timeout)
		defer cancel()
		return dial(ctx, network, addr)
	}
}

// checkListenHost enforces AGENTS.md: the proxy is for same-machine use and
// must not listen on non-loopback interfaces unless the operator opts in.
// Non-loopback binds turn the proxy into an open SOCKS egress (SSRF assist).
func checkListenHost(host string, allowNonLoopback bool) error {
	if isLoopbackHost(host) {
		return nil
	}
	if allowNonLoopback {
		return nil
	}
	return fmt.Errorf("%q is not a loopback address; pass -allow-non-loopback to override (SSRF risk)", host)
}

// isLoopbackHost reports whether host is a loopback address or the
// conventional "localhost" name. Used to refuse (or warn when overridden)
// when -H exposes the proxy beyond the intended local-only bind (see AGENTS.md).
// Accepts optional brackets and IPv6 zone IDs (e.g. "[::1]", "::1%lo").
func isLoopbackHost(host string) bool {
	if strings.EqualFold(host, "localhost") {
		return true
	}
	ip := net.ParseIP(ipLiteral(host))
	return ip != nil && ip.IsLoopback()
}

// hostForGetListener normalizes -H for go-getlistener, which concatenates
// host and port with fmt.Sprintf("%s:%d", …) instead of net.JoinHostPort.
// Bare IPv6 literals must be wrapped in brackets so Listen can parse them.
func hostForGetListener(host string) string {
	h := stripHostBrackets(host)
	if h == "" {
		return host
	}
	// Zone ID is not part of the address ParseIP understands; strip for the
	// IPv6 check, then put brackets around the original (zone-inclusive) form.
	base, _, _ := strings.Cut(h, "%")
	ip := net.ParseIP(base)
	if ip == nil || !strings.Contains(base, ":") {
		// Hostname or IPv4: return unbracketed base form (drop stray brackets).
		return h
	}
	return "[" + h + "]"
}

// ipLiteral returns the host form net.ParseIP understands: no brackets, no zone.
func ipLiteral(host string) string {
	h := stripHostBrackets(host)
	if i := strings.IndexByte(h, '%'); i >= 0 {
		return h[:i]
	}
	return h
}

func stripHostBrackets(host string) string {
	if len(host) >= 2 && host[0] == '[' {
		if i := strings.LastIndexByte(host, ']'); i > 0 {
			return host[1:i]
		}
	}
	return host
}
