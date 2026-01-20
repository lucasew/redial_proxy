package main

import (
	"testing"
)

func TestGetSocksConfig(t *testing.T) {
	t.Run("NoAuth", func(t *testing.T) {
		t.Setenv("PROXY_USERNAME", "")
		t.Setenv("PROXY_PASSWORD", "")
		conf := getSocksConfig()
		if conf.Credentials != nil {
			t.Error("Expected no credentials, but got some")
		}
	})

	t.Run("WithAuth", func(t *testing.T) {
		t.Setenv("PROXY_USERNAME", "user")
		t.Setenv("PROXY_PASSWORD", "pass")
		conf := getSocksConfig()
		if conf.Credentials == nil {
			t.Fatal("Expected credentials, got nil")
		}

		if !conf.Credentials.Valid("user", "pass") {
			t.Error("Credentials should be valid for user/pass")
		}
		if conf.Credentials.Valid("user", "wrong") {
			t.Error("Credentials should be invalid for wrong pass")
		}
	})
}
