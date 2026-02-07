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

// Redialer is a custom dialer that retries connections upon specific routing errors.
// It wraps net.Dialer and automatically retries when the error message contains "route".
// This is particularly useful in environments with transient network issues or strict routing rules
// where initial attempts might fail before a route is established.
type Redialer struct {
	// MaxRetries is the maximum number of retry attempts before giving up.
	// If set to 0, it will not retry (only the initial attempt is made).
	MaxRetries int
	// RetryDelay is the duration to wait between retry attempts.
	RetryDelay time.Duration
}

// DialContext connects to the address on the named network using net.Dialer.
//
// If the connection fails with an error string containing "route", it will retry
// up to d.MaxRetries times, waiting d.RetryDelay between attempts.
//
// It respects the provided context for cancellation both during the connection attempt
// and the backoff period. If the context is canceled, the operation returns immediately.
//
// Successful connections and retry attempts are logged using slog.
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
