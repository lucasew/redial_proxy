package main

import (
	"context"
	"errors"
	"net"
	"sync/atomic"
	"testing"
	"time"
)

func TestIsLoopbackHost(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name string
		host string
		want bool
	}{
		{"ipv4 loopback", "127.0.0.1", true},
		{"ipv4 loopback other", "127.0.0.2", true},
		{"ipv4 loopback high", "127.255.255.254", true},
		{"ipv6 loopback", "::1", true},
		{"ipv6 loopback bracketed", "[::1]", true},
		{"ipv6 loopback expanded", "0:0:0:0:0:0:0:1", true},
		{"ipv6 loopback expanded bracketed", "[0:0:0:0:0:0:0:1]", true},
		{"localhost", "localhost", true},
		{"localhost upper", "LOCALHOST", true},
		{"localhost mixed", "LocalHost", true},
		{"unspecified ipv4", "0.0.0.0", false},
		{"unspecified ipv6", "::", false},
		{"unspecified ipv6 bracketed", "[::]", false},
		{"private lan", "192.168.1.1", false},
		{"public", "8.8.8.8", false},
		{"hostname", "example.com", false},
		{"empty", "", false},
		{"garbage", "not-an-ip", false},
		{"bracketed non-loopback", "[2001:db8::1]", false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if got := isLoopbackHost(tc.host); got != tc.want {
				t.Fatalf("isLoopbackHost(%q)=%v want %v", tc.host, got, tc.want)
			}
		})
	}
}

func TestCheckListenHost(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name             string
		host             string
		allowNonLoopback bool
		wantErr          bool
	}{
		{"loopback ok", "127.0.0.1", false, false},
		{"localhost ok", "localhost", false, false},
		{"ipv6 loopback ok", "::1", false, false},
		{"refuse 0.0.0.0", "0.0.0.0", false, true},
		{"refuse lan", "192.168.1.1", false, true},
		{"refuse public", "8.8.8.8", false, true},
		{"refuse hostname", "example.com", false, true},
		{"override 0.0.0.0", "0.0.0.0", true, false},
		{"override lan", "192.168.0.10", true, false},
		// Override still ok for loopback (no-op).
		{"override loopback", "127.0.0.1", true, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			err := checkListenHost(tc.host, tc.allowNonLoopback)
			if tc.wantErr && err == nil {
				t.Fatalf("checkListenHost(%q, allow=%v)=nil want error", tc.host, tc.allowNonLoopback)
			}
			if !tc.wantErr && err != nil {
				t.Fatalf("checkListenHost(%q, allow=%v)=%v want nil", tc.host, tc.allowNonLoopback, err)
			}
		})
	}
}

func TestHostForGetListener(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name string
		in   string
		want string
	}{
		{"ipv4 unchanged", "127.0.0.1", "127.0.0.1"},
		{"hostname unchanged", "localhost", "localhost"},
		{"ipv6 bare gets brackets", "::1", "[::1]"},
		{"ipv6 already bracketed", "[::1]", "[::1]"},
		{"ipv6 expanded", "0:0:0:0:0:0:0:1", "[0:0:0:0:0:0:0:1]"},
		{"ipv6 public", "2001:db8::1", "[2001:db8::1]"},
		{"ipv4-mapped needs brackets", "::ffff:127.0.0.1", "[::ffff:127.0.0.1]"},
		{"empty passthrough", "", ""},
		{"garbage passthrough", "not-an-ip", "not-an-ip"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if got := hostForGetListener(tc.in); got != tc.want {
				t.Fatalf("hostForGetListener(%q)=%q want %q", tc.in, got, tc.want)
			}
		})
	}
}

func TestWithDialTimeout_DisablesWhenNonPositive(t *testing.T) {
	t.Parallel()
	var calls atomic.Int32
	base := func(ctx context.Context, network, addr string) (net.Conn, error) {
		calls.Add(1)
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		// no deadline expected when timeout is disabled
		if _, ok := ctx.Deadline(); ok {
			t.Fatal("unexpected deadline on context")
		}
		return nil, errors.New("done")
	}
	for _, timeout := range []time.Duration{0, -time.Second} {
		calls.Store(0)
		dial := withDialTimeout(timeout, base)
		_, err := dial(context.Background(), "tcp", "example.com:80")
		if err == nil || err.Error() != "done" {
			t.Fatalf("timeout=%v err=%v", timeout, err)
		}
		if calls.Load() != 1 {
			t.Fatalf("timeout=%v calls=%d", timeout, calls.Load())
		}
	}
}

func TestWithDialTimeout_EnforcesDeadline(t *testing.T) {
	t.Parallel()
	base := func(ctx context.Context, network, addr string) (net.Conn, error) {
		// block until canceled by the wrapper deadline
		<-ctx.Done()
		return nil, ctx.Err()
	}
	dial := withDialTimeout(40*time.Millisecond, base)
	start := time.Now()
	_, err := dial(context.Background(), "tcp", "example.com:80")
	elapsed := time.Since(start)
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("err=%v want context.DeadlineExceeded", err)
	}
	if elapsed < 20*time.Millisecond {
		t.Fatalf("elapsed=%v, deadline returned too early", elapsed)
	}
	if elapsed > 500*time.Millisecond {
		t.Fatalf("elapsed=%v, expected ~40ms deadline", elapsed)
	}
}

func TestWithDialTimeout_NilDial(t *testing.T) {
	t.Parallel()
	if got := withDialTimeout(time.Second, nil); got != nil {
		t.Fatalf("got non-nil dial func for nil input")
	}
}
