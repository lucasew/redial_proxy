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
		{"ipv6 loopback expanded", "0:0:0:0:0:0:0:1", true},
		{"localhost", "localhost", true},
		{"localhost upper", "LOCALHOST", true},
		{"localhost mixed", "LocalHost", true},
		{"unspecified ipv4", "0.0.0.0", false},
		{"unspecified ipv6", "::", false},
		{"private lan", "192.168.1.1", false},
		{"public", "8.8.8.8", false},
		{"hostname", "example.com", false},
		{"empty", "", false},
		{"garbage", "not-an-ip", false},
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
