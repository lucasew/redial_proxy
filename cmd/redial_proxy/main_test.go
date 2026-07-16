package main

import "testing"

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
