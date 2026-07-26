package dialer

import (
	"context"
	"errors"
	"io"
	"net"
	"os"
	"sync/atomic"
	"syscall"
	"testing"
	"time"
)

type stubConn struct{}

func (stubConn) Read([]byte) (int, error)         { return 0, io.EOF }
func (stubConn) Write([]byte) (int, error)        { return 0, io.EOF }
func (stubConn) Close() error                     { return nil }
func (stubConn) LocalAddr() net.Addr              { return nil }
func (stubConn) RemoteAddr() net.Addr             { return nil }
func (stubConn) SetDeadline(time.Time) error      { return nil }
func (stubConn) SetReadDeadline(time.Time) error  { return nil }
func (stubConn) SetWriteDeadline(time.Time) error { return nil }

func TestIsRouteError(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name string
		err  error
		want bool
	}{
		{"nil", nil, false},
		{"no route to host", errors.New("dial tcp 1.2.3.4:80: connect: no route to host"), true},
		{"network is unreachable", errors.New("dial tcp 1.2.3.4:80: connect: network is unreachable"), true},
		{"host is unreachable", errors.New("dial tcp 1.2.3.4:80: connect: host is unreachable"), true},
		{"case insensitive phrase", errors.New("No Route To Host"), true},
		// Hostnames that contain "route" must not force a retry on unrelated errors.
		{"hostname contains route", errors.New("dial tcp route.example.com:443: connect: connection refused"), false},
		{"unrelated", errors.New("connection refused"), false},
		{"ehostunreach", &os.SyscallError{Syscall: "connect", Err: syscall.EHOSTUNREACH}, true},
		{"enetunreach", &os.SyscallError{Syscall: "connect", Err: syscall.ENETUNREACH}, true},
		{"econnrefused errno", &os.SyscallError{Syscall: "connect", Err: syscall.ECONNREFUSED}, false},
		{"wrapped operror", &net.OpError{
			Op:  "dial",
			Net: "tcp",
			Err: &os.SyscallError{Syscall: "connect", Err: syscall.EHOSTUNREACH},
		}, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if got := isRouteError(tc.err); got != tc.want {
				t.Fatalf("isRouteError(%v)=%v want %v", tc.err, got, tc.want)
			}
		})
	}
}

func TestDialContext_SuccessFirstAttempt(t *testing.T) {
	t.Parallel()
	var calls atomic.Int32
	d := &Redialer{
		MaxRetries: 3,
		RetryDelay: time.Millisecond,
		dial: func(ctx context.Context, network, addr string) (net.Conn, error) {
			calls.Add(1)
			return stubConn{}, nil
		},
	}
	conn, err := d.DialContext(context.Background(), "tcp", "example.com:80")
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if conn == nil {
		t.Fatal("expected conn")
	}
	if calls.Load() != 1 {
		t.Fatalf("dial calls=%d want 1", calls.Load())
	}
}

func TestDialContext_RetriesRouteErrorThenSucceeds(t *testing.T) {
	t.Parallel()
	var calls atomic.Int32
	d := &Redialer{
		MaxRetries: 3,
		RetryDelay: time.Millisecond,
		dial: func(ctx context.Context, network, addr string) (net.Conn, error) {
			n := calls.Add(1)
			if n < 3 {
				return nil, errors.New("dial tcp: no route to host")
			}
			return stubConn{}, nil
		},
	}
	conn, err := d.DialContext(context.Background(), "tcp", "example.com:80")
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if conn == nil {
		t.Fatal("expected conn")
	}
	if calls.Load() != 3 {
		t.Fatalf("dial calls=%d want 3", calls.Load())
	}
}

func TestDialContext_NonRouteErrorDoesNotRetry(t *testing.T) {
	t.Parallel()
	var calls atomic.Int32
	wantErr := errors.New("connection refused")
	d := &Redialer{
		MaxRetries: 5,
		RetryDelay: 50 * time.Millisecond,
		dial: func(ctx context.Context, network, addr string) (net.Conn, error) {
			calls.Add(1)
			return nil, wantErr
		},
	}
	start := time.Now()
	_, err := d.DialContext(context.Background(), "tcp", "example.com:80")
	elapsed := time.Since(start)
	if !errors.Is(err, wantErr) {
		t.Fatalf("err=%v want %v", err, wantErr)
	}
	if calls.Load() != 1 {
		t.Fatalf("dial calls=%d want 1 (no retry)", calls.Load())
	}
	if elapsed > 40*time.Millisecond {
		t.Fatalf("elapsed=%v suggests unexpected backoff", elapsed)
	}
}

func TestDialContext_HostnameContainingRouteDoesNotRetry(t *testing.T) {
	t.Parallel()
	var calls atomic.Int32
	// Mimic net.Dialer error text that embeds the dial target in the message.
	wantErr := errors.New("dial tcp route.example.com:443: connect: connection refused")
	d := &Redialer{
		MaxRetries: 5,
		RetryDelay: 50 * time.Millisecond,
		dial: func(ctx context.Context, network, addr string) (net.Conn, error) {
			calls.Add(1)
			return nil, wantErr
		},
	}
	_, err := d.DialContext(context.Background(), "tcp", "route.example.com:443")
	if !errors.Is(err, wantErr) {
		t.Fatalf("err=%v want %v", err, wantErr)
	}
	if calls.Load() != 1 {
		t.Fatalf("dial calls=%d want 1 (hostname 'route' must not trigger retry)", calls.Load())
	}
}

func TestDialContext_MaxRetriesZeroNonRouteReturnsBareError(t *testing.T) {
	t.Parallel()
	var calls atomic.Int32
	wantErr := errors.New("connection refused")
	d := &Redialer{
		MaxRetries: 0,
		RetryDelay: 50 * time.Millisecond,
		dial: func(ctx context.Context, network, addr string) (net.Conn, error) {
			calls.Add(1)
			return nil, wantErr
		},
	}
	_, err := d.DialContext(context.Background(), "tcp", "example.com:80")
	if !errors.Is(err, wantErr) {
		t.Fatalf("err=%v want %v", err, wantErr)
	}
	// Regression: old control flow hit try >= MaxRetries before isRouteError and
	// wrapped every first failure as "too many retries", even with MaxRetries=0
	// and a non-route error.
	if err != nil && err.Error() != wantErr.Error() {
		t.Fatalf("err=%q want bare %q (no retry-exhaustion wrap)", err, wantErr)
	}
	if calls.Load() != 1 {
		t.Fatalf("dial calls=%d want 1", calls.Load())
	}
}

func TestDialContext_MaxRetriesZeroRouteErrorNoRetry(t *testing.T) {
	t.Parallel()
	var calls atomic.Int32
	routeErr := errors.New("no route to host")
	d := &Redialer{
		MaxRetries: 0,
		RetryDelay: 50 * time.Millisecond,
		dial: func(ctx context.Context, network, addr string) (net.Conn, error) {
			calls.Add(1)
			return nil, routeErr
		},
	}
	_, err := d.DialContext(context.Background(), "tcp", "example.com:80")
	if err == nil {
		t.Fatal("expected error")
	}
	if !errors.Is(err, routeErr) {
		t.Fatalf("err=%v want wrap of %v", err, routeErr)
	}
	if calls.Load() != 1 {
		t.Fatalf("dial calls=%d want 1 (MaxRetries=0 means no retry)", calls.Load())
	}
}

func TestDialContext_ExhaustsMaxRetriesOnRouteErrors(t *testing.T) {
	t.Parallel()
	var calls atomic.Int32
	routeErr := errors.New("no route to host")
	d := &Redialer{
		MaxRetries: 2,
		RetryDelay: time.Millisecond,
		dial: func(ctx context.Context, network, addr string) (net.Conn, error) {
			calls.Add(1)
			return nil, routeErr
		},
	}
	_, err := d.DialContext(context.Background(), "tcp", "example.com:80")
	if err == nil {
		t.Fatal("expected error")
	}
	if !errors.Is(err, routeErr) {
		t.Fatalf("err=%v want wrap of %v", err, routeErr)
	}
	// initial attempt + MaxRetries retries
	if calls.Load() != 3 {
		t.Fatalf("dial calls=%d want 3", calls.Load())
	}
}

func TestDialContext_ContextCanceledDuringDial(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithCancel(context.Background())
	d := &Redialer{
		MaxRetries: 3,
		RetryDelay: time.Millisecond,
		dial: func(ctx context.Context, network, addr string) (net.Conn, error) {
			cancel()
			return nil, ctx.Err()
		},
	}
	_, err := d.DialContext(ctx, "tcp", "example.com:80")
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("err=%v want context.Canceled", err)
	}
}

func TestDialContext_ContextCanceledDuringBackoff(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithCancel(context.Background())
	var calls atomic.Int32
	d := &Redialer{
		MaxRetries: 3,
		RetryDelay: 5 * time.Second,
		dial: func(ctx context.Context, network, addr string) (net.Conn, error) {
			calls.Add(1)
			return nil, errors.New("no route to host")
		},
	}
	go func() {
		// cancel shortly after first failure enters backoff
		time.Sleep(20 * time.Millisecond)
		cancel()
	}()
	start := time.Now()
	_, err := d.DialContext(ctx, "tcp", "example.com:80")
	elapsed := time.Since(start)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("err=%v want context.Canceled", err)
	}
	if calls.Load() != 1 {
		t.Fatalf("dial calls=%d want 1", calls.Load())
	}
	if elapsed > 500*time.Millisecond {
		t.Fatalf("elapsed=%v, expected quick cancel during backoff", elapsed)
	}
}

func TestRedial_ContextCancellation_Integration(t *testing.T) {
	// Live-network smoke: ensure default dial path still honors context.
	// 192.0.2.1 is TEST-NET-1; may time out or reject quickly depending on host routing.
	target := "192.0.2.1:80"

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	d := &Redialer{
		MaxRetries: 3,
		RetryDelay: 100 * time.Millisecond,
	}

	start := time.Now()
	_, err := d.DialContext(ctx, "tcp", target)
	duration := time.Since(start)

	if err == nil {
		t.Fatal("expected error, got connection")
	}

	// If redial respects context, it should return roughly around 100ms.
	// If it blocks on net.Dial (which has default system timeout ~30s+), it will take much longer
	// IF the network drops packets. If it rejects immediately, this test passes either way.
	if duration > 2*time.Second {
		t.Errorf("redial took too long: %v, expected ~100ms", duration)
	}

	t.Logf("Error returned: %v", err)
}
