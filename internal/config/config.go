package config

import (
	"flag"
)

const (
	DefaultPort = 8889
	DefaultHost = "127.0.0.1"
)

// Config holds the application configuration.
type Config struct {
	Port int
	Host string
}

// Load parses command line flags.
func Load() *Config {
	cfg := &Config{}
	flag.IntVar(&cfg.Port, "p", DefaultPort, "port to listen the server")
	flag.StringVar(&cfg.Host, "H", DefaultHost, "host to listen the server")
	flag.Parse()

	return cfg
}
