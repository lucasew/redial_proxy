package dialer

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"time"
)

// DefaultResolveTimeout is used when RetryResolver.Timeout is zero.
// Bounds a single DNS attempt; go-socks5's default resolver uses
// net.ResolveIPAddr, which ignores context and can hang until the OS DNS
// timeout on a flaky network.
const DefaultResolveTimeout = 5 * time.Second

// RetryResolver resolves hostnames with the same retry budget as Redialer.
// It is a drop-in for github.com/armon/go-socks5 NameResolver: socks5 resolves
// FQDNs before Dial, so without this, transient DNS failures never hit the
// redialer and permanent hangs never see a deadline.
type RetryResolver struct {
	// MaxRetries is the maximum number of retry attempts after the first try.
	// 0 means a single attempt only.
	MaxRetries int
	// RetryDelay is the duration to wait between retry attempts.
	RetryDelay time.Duration
	// Timeout bounds each LookupIPAddr attempt when the parent context has no
	// deadline. Zero means DefaultResolveTimeout; negative disables the
	// per-attempt bound (still respects a parent deadline if present).
	Timeout time.Duration

	// lookup, when non-nil, replaces net.DefaultResolver.LookupIPAddr.
	// Intended for tests; production code leaves this nil.
	lookup func(ctx context.Context, host string) ([]net.IPAddr, error)
}

// Resolve looks up name and returns one IP for go-socks5 (which dials a single
// address). When the lookup returns both families, IPv4 is preferred: this
// proxy targets flaky consumer networks where broken IPv6 is common and
// socks5 has no Happy Eyeballs fallback.
//
// Temporary / timeout DNS errors are retried up to MaxRetries times with
// RetryDelay between attempts. Permanent errors (e.g. NXDOMAIN) and context
// cancellation are returned immediately.
func (r *RetryResolver) Resolve(ctx context.Context, name string) (context.Context, net.IP, error) {
	try := 0
	for {
		addrs, err := r.doLookup(ctx, name)
		if err == nil {
			ip := pickIP(addrs)
			if ip == nil {
				return ctx, nil, fmt.Errorf("no addresses for %s", name)
			}
			return ctx, ip, nil
		}

		if ctx.Err() != nil {
			return ctx, nil, ctx.Err()
		}

		slog.Warn("dns err", "host", name, "err", err)
		if !isRetriableDNSError(err) {
			return ctx, nil, err
		}
		if try >= r.MaxRetries {
			return ctx, nil, fmt.Errorf("too many DNS retries: %w", err)
		}

		slog.Info("retrying DNS", "host", name, "try", try)
		timer := time.NewTimer(r.RetryDelay)
		select {
		case <-ctx.Done():
			timer.Stop()
			return ctx, nil, ctx.Err()
		case <-timer.C:
			try++
			continue
		}
	}
}

func (r *RetryResolver) doLookup(ctx context.Context, name string) ([]net.IPAddr, error) {
	attemptCtx, cancel := r.attemptContext(ctx)
	if cancel != nil {
		defer cancel()
	}
	if r.lookup != nil {
		return r.lookup(attemptCtx, name)
	}
	return net.DefaultResolver.LookupIPAddr(attemptCtx, name)
}

// attemptContext applies a per-attempt timeout when the parent has no deadline
// and Timeout is non-negative.
func (r *RetryResolver) attemptContext(ctx context.Context) (context.Context, context.CancelFunc) {
	if _, ok := ctx.Deadline(); ok {
		return ctx, nil
	}
	timeout := r.Timeout
	if timeout == 0 {
		timeout = DefaultResolveTimeout
	}
	if timeout < 0 {
		return ctx, nil
	}
	return context.WithTimeout(ctx, timeout)
}

// pickIP chooses one address from LookupIPAddr results.
// Prefer the first IPv4 address when any exist; otherwise the first result.
// go-socks5 only dials a single IP, so an IPv6-first ordering on dual-stack
// names often blackholes clients on networks with broken IPv6 routing.
func pickIP(addrs []net.IPAddr) net.IP {
	if len(addrs) == 0 {
		return nil
	}
	for _, a := range addrs {
		if a.IP == nil {
			continue
		}
		if a.IP.To4() != nil {
			return a.IP
		}
	}
	for _, a := range addrs {
		if a.IP != nil {
			return a.IP
		}
	}
	return nil
}

// isRetriableDNSError reports whether err looks like a transient DNS failure
// worth retrying. NXDOMAIN and other permanent DNS errors are not retried.
func isRetriableDNSError(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		// Parent cancellation is final. Per-attempt deadline exceeded is still
		// retriable when the parent is live (checked by the caller via ctx.Err).
		// errors.Is on a child timeout returns true for DeadlineExceeded even
		// when the parent is fine — treat bare deadline as retriable only when
		// it is not the parent. Callers already return on ctx.Err() != nil.
		return errors.Is(err, context.DeadlineExceeded)
	}
	var dnsErr *net.DNSError
	if errors.As(err, &dnsErr) {
		// Temporary covers SERVFAIL / timeout-class DNS failures on most
		// platforms; IsNotFound (NXDOMAIN) is permanent.
		if dnsErr.IsNotFound {
			return false
		}
		return dnsErr.Temporary() || dnsErr.Timeout() || dnsErr.IsTimeout
	}
	var ne net.Error
	if errors.As(err, &ne) {
		return ne.Timeout()
	}
	return false
}
