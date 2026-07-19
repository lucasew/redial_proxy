package dialer

import (
	"context"
	"errors"
	"net"
	"sync/atomic"
	"testing"
	"time"
)

func TestIsRetriableDNSError(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name string
		err  error
		want bool
	}{
		{"nil", nil, false},
		{"nxdomain", &net.DNSError{Err: "no such host", Name: "x.example", IsNotFound: true}, false},
		{"temporary", &net.DNSError{Err: "server misbehaving", Name: "x.example", IsTemporary: true}, true},
		{"timeout flag", &net.DNSError{Err: "i/o timeout", Name: "x.example", IsTimeout: true}, true},
		{"canceled", context.Canceled, false},
		{"deadline", context.DeadlineExceeded, true},
		{"unrelated", errors.New("boom"), false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if got := isRetriableDNSError(tc.err); got != tc.want {
				t.Fatalf("isRetriableDNSError(%v)=%v want %v", tc.err, got, tc.want)
			}
		})
	}
}

func TestRetryResolver_SuccessFirstAttempt(t *testing.T) {
	t.Parallel()
	var calls atomic.Int32
	want := net.ParseIP("1.2.3.4")
	r := &RetryResolver{
		MaxRetries: 3,
		RetryDelay: time.Millisecond,
		Timeout:    time.Second,
		lookup: func(ctx context.Context, host string) ([]net.IPAddr, error) {
			calls.Add(1)
			if host != "example.com" {
				t.Fatalf("host=%q", host)
			}
			return []net.IPAddr{{IP: want}}, nil
		},
	}
	_, ip, err := r.Resolve(context.Background(), "example.com")
	if err != nil {
		t.Fatalf("err=%v", err)
	}
	if !ip.Equal(want) {
		t.Fatalf("ip=%v want %v", ip, want)
	}
	if calls.Load() != 1 {
		t.Fatalf("calls=%d want 1", calls.Load())
	}
}

func TestRetryResolver_RetriesTemporaryThenSucceeds(t *testing.T) {
	t.Parallel()
	var calls atomic.Int32
	want := net.ParseIP("9.9.9.9")
	r := &RetryResolver{
		MaxRetries: 3,
		RetryDelay: time.Millisecond,
		Timeout:    time.Second,
		lookup: func(ctx context.Context, host string) ([]net.IPAddr, error) {
			n := calls.Add(1)
			if n < 3 {
				return nil, &net.DNSError{Err: "server misbehaving", Name: host, IsTemporary: true}
			}
			return []net.IPAddr{{IP: want}}, nil
		},
	}
	_, ip, err := r.Resolve(context.Background(), "flaky.example")
	if err != nil {
		t.Fatalf("err=%v", err)
	}
	if !ip.Equal(want) {
		t.Fatalf("ip=%v want %v", ip, want)
	}
	if calls.Load() != 3 {
		t.Fatalf("calls=%d want 3", calls.Load())
	}
}

func TestRetryResolver_NXDOMAINDoesNotRetry(t *testing.T) {
	t.Parallel()
	var calls atomic.Int32
	wantErr := &net.DNSError{Err: "no such host", Name: "missing.example", IsNotFound: true}
	r := &RetryResolver{
		MaxRetries: 5,
		RetryDelay: 50 * time.Millisecond,
		Timeout:    time.Second,
		lookup: func(ctx context.Context, host string) ([]net.IPAddr, error) {
			calls.Add(1)
			return nil, wantErr
		},
	}
	start := time.Now()
	_, _, err := r.Resolve(context.Background(), "missing.example")
	elapsed := time.Since(start)
	if !errors.Is(err, wantErr) {
		// DNSError may not chain via Is; compare message path
		var de *net.DNSError
		if !errors.As(err, &de) || !de.IsNotFound {
			t.Fatalf("err=%v want NXDOMAIN", err)
		}
	}
	if calls.Load() != 1 {
		t.Fatalf("calls=%d want 1", calls.Load())
	}
	if elapsed > 40*time.Millisecond {
		t.Fatalf("elapsed=%v suggests unexpected backoff", elapsed)
	}
}

func TestRetryResolver_ExhaustsMaxRetries(t *testing.T) {
	t.Parallel()
	var calls atomic.Int32
	temp := &net.DNSError{Err: "server misbehaving", Name: "x.example", IsTemporary: true}
	r := &RetryResolver{
		MaxRetries: 2,
		RetryDelay: time.Millisecond,
		Timeout:    time.Second,
		lookup: func(ctx context.Context, host string) ([]net.IPAddr, error) {
			calls.Add(1)
			return nil, temp
		},
	}
	_, _, err := r.Resolve(context.Background(), "x.example")
	if err == nil {
		t.Fatal("expected error")
	}
	if !errors.Is(err, temp) {
		var de *net.DNSError
		if !errors.As(err, &de) {
			t.Fatalf("err=%v want wrap of temporary DNS", err)
		}
	}
	// initial + MaxRetries
	if calls.Load() != 3 {
		t.Fatalf("calls=%d want 3", calls.Load())
	}
}

func TestRetryResolver_PerAttemptTimeoutIsRetried(t *testing.T) {
	t.Parallel()
	var calls atomic.Int32
	want := net.ParseIP("8.8.8.8")
	r := &RetryResolver{
		MaxRetries: 2,
		RetryDelay: time.Millisecond,
		Timeout:    20 * time.Millisecond,
		lookup: func(ctx context.Context, host string) ([]net.IPAddr, error) {
			n := calls.Add(1)
			if n == 1 {
				<-ctx.Done()
				return nil, ctx.Err()
			}
			return []net.IPAddr{{IP: want}}, nil
		},
	}
	_, ip, err := r.Resolve(context.Background(), "slow.example")
	if err != nil {
		t.Fatalf("err=%v", err)
	}
	if !ip.Equal(want) {
		t.Fatalf("ip=%v want %v", ip, want)
	}
	if calls.Load() != 2 {
		t.Fatalf("calls=%d want 2", calls.Load())
	}
}

func TestRetryResolver_ParentCancelDuringBackoff(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithCancel(context.Background())
	var calls atomic.Int32
	r := &RetryResolver{
		MaxRetries: 3,
		RetryDelay: 5 * time.Second,
		Timeout:    time.Second,
		lookup: func(ctx context.Context, host string) ([]net.IPAddr, error) {
			calls.Add(1)
			return nil, &net.DNSError{Err: "server misbehaving", Name: host, IsTemporary: true}
		},
	}
	go func() {
		time.Sleep(20 * time.Millisecond)
		cancel()
	}()
	start := time.Now()
	_, _, err := r.Resolve(ctx, "x.example")
	elapsed := time.Since(start)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("err=%v want context.Canceled", err)
	}
	if calls.Load() != 1 {
		t.Fatalf("calls=%d want 1", calls.Load())
	}
	if elapsed > 500*time.Millisecond {
		t.Fatalf("elapsed=%v, expected quick cancel during backoff", elapsed)
	}
}

func TestRetryResolver_EmptyResult(t *testing.T) {
	t.Parallel()
	r := &RetryResolver{
		lookup: func(ctx context.Context, host string) ([]net.IPAddr, error) {
			return nil, nil
		},
	}
	_, _, err := r.Resolve(context.Background(), "empty.example")
	if err == nil {
		t.Fatal("expected error for empty address list")
	}
}

func TestPickIP(t *testing.T) {
	t.Parallel()
	v4a := net.ParseIP("1.2.3.4")
	v4b := net.ParseIP("5.6.7.8")
	v6a := net.ParseIP("2001:db8::1")
	v6b := net.ParseIP("2001:db8::2")

	cases := []struct {
		name  string
		addrs []net.IPAddr
		want  net.IP
	}{
		{"empty", nil, nil},
		{"ipv4 only", []net.IPAddr{{IP: v4a}}, v4a},
		{"ipv6 only", []net.IPAddr{{IP: v6a}}, v6a},
		{"ipv6 first dual-stack prefers v4", []net.IPAddr{{IP: v6a}, {IP: v4a}}, v4a},
		{"ipv4 first dual-stack keeps first v4", []net.IPAddr{{IP: v4a}, {IP: v6a}, {IP: v4b}}, v4a},
		{"multiple v6 no v4", []net.IPAddr{{IP: v6a}, {IP: v6b}}, v6a},
		{"nil IP entries skipped", []net.IPAddr{{IP: nil}, {IP: v6a}, {IP: v4a}}, v4a},
		{"all nil", []net.IPAddr{{IP: nil}}, nil},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := pickIP(tc.addrs)
			if tc.want == nil {
				if got != nil {
					t.Fatalf("pickIP=%v want nil", got)
				}
				return
			}
			if got == nil || !got.Equal(tc.want) {
				t.Fatalf("pickIP=%v want %v", got, tc.want)
			}
		})
	}
}

func TestRetryResolver_PrefersIPv4WhenDualStack(t *testing.T) {
	t.Parallel()
	v4 := net.ParseIP("93.184.216.34")
	v6 := net.ParseIP("2606:2800:220:1:248:1893:25c8:1946")
	r := &RetryResolver{
		lookup: func(ctx context.Context, host string) ([]net.IPAddr, error) {
			// AAAA-first ordering is common; socks5 would previously dial only v6.
			return []net.IPAddr{{IP: v6}, {IP: v4}}, nil
		},
	}
	_, ip, err := r.Resolve(context.Background(), "example.com")
	if err != nil {
		t.Fatalf("err=%v", err)
	}
	if !ip.Equal(v4) {
		t.Fatalf("ip=%v want IPv4 %v", ip, v4)
	}
}
