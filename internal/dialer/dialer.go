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
type Redialer struct {
	MaxRetries int
	RetryDelay time.Duration
}

// DialContext connects to the address on the named network.
// It retries the connection if a "route" error occurs.
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
