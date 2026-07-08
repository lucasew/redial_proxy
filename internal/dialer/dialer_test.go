package dialer

import (
	"context"
	"testing"
	"time"
)

const (
	testTarget         = "192.0.2.1:80"
	testContextTimeout = 100 * time.Millisecond
	testMaxDuration    = 2 * time.Second
)

func TestRedial_ContextCancellation(t *testing.T) {
	// Use an IP that is likely to cause a timeout or at least delay.
	// 192.0.2.1 is in TEST-NET-1.
	target := testTarget

	// Set a short timeout
	ctx, cancel := context.WithTimeout(context.Background(), testContextTimeout)
	defer cancel()

	d := &Redialer{
		MaxRetries: DefaultMaxRetries,
		RetryDelay: DefaultRetryDelay,
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
	if duration > testMaxDuration {
		t.Errorf("redial took too long: %v, expected ~100ms", duration)
	}

	t.Logf("Error returned: %v", err)
}
