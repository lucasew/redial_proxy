// Package dialer implements a custom network dialer that retries connections on specific routing errors.
// It is designed to handle transient network failures, particularly those related to routing issues,
// by automatically retrying connection attempts with a configurable backoff.
package dialer

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"strings"
	"time"
)

// Redialer is a custom dialer that transparently intercepts and retries connection attempts
// upon specific routing errors. It acts as a wrapper around the standard net.Dialer, triggering
// an automatic backoff and retry loop whenever the resulting error message contains "route".
// This component encapsulates the configuration state for the retry mechanism while remaining
// stateless across individual request executions. It is specifically designed to mitigate
// transient network issues or strict routing rules where initial attempts might fail before
// a route is successfully established.
type Redialer struct {
	// MaxRetries is the maximum number of retry attempts before giving up.
	// If set to 0, it will not retry (only the initial attempt is made).
	MaxRetries int
	// RetryDelay is the duration to wait between retry attempts.
	RetryDelay time.Duration
}

// DialContext connects to the target address on the specified network using an underlying net.Dialer.
//
// Unlike the standard dialer, this method evaluates connection failures. If an error string contains
// "route", it initiates a retry loop up to d.MaxRetries times, pausing for d.RetryDelay between attempts.
//
// To prevent hanging on cancelled requests, it strictly respects the provided context for cancellation
// both during the active connection attempt and the subsequent backoff period. If the context signals
// cancellation at any point, the operation aborts and returns immediately.
//
// Both successful connections and active retry attempts are recorded using the structured logger (slog).
func (d *Redialer) DialContext(ctx context.Context, network, addr string) (net.Conn, error) {
	var dialer net.Dialer
	try := 0
	for {
		conn, err := dialer.DialContext(ctx, network, addr)
		if err == nil {
			slog.Info("CONNECT", "network", network, "addr", addr)
			return conn, nil
		}

		if ctx.Err() != nil {
			return nil, ctx.Err()
		}

		if try >= d.MaxRetries {
			return nil, fmt.Errorf("too many retries: %w", err)
		}

		slog.Warn("conn err", "err", err.Error())
		if strings.Contains(err.Error(), "route") {
			slog.Info("retrying connection", "network", network, "addr", addr, "try", try)
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(d.RetryDelay):
				try++
				continue
			}
		}
		return nil, err
	}
}
