package getlistener

import (
	"fmt"
	"log/slog"
	"net"
	"os"
	"strconv"
)

type Config struct {
	Host string
	Port int
}

// GetAvailablePort get the number of an available port
func GetAvailablePort() (int, error) {
	listener, err := net.Listen("tcp", ":0")
	if err != nil {
		return 0, err
	}
	defer listener.Close()
	addr := listener.Addr().(*net.TCPAddr)
	return addr.Port, nil
}

func loadConfig() (*Config, error) {
	cfg := &Config{
		Host: "127.0.0.1",
		Port: 0,
	}
	envPort := os.Getenv("PORT")
	if envPort != "" {
		selectedPort, err := strconv.Atoi(envPort)
		if err != nil {
			return nil, fmt.Errorf("the environment variable PORT was provided to setup a port but has an invalid value: '%s'", envPort)
		}
		cfg.Port = selectedPort
	}
	envHost := os.Getenv("HOST")
	if envHost != "" {
		cfg.Host = envHost
		if cfg.Host != "127.0.0.1" && cfg.Host != "localhost" {
			slog.Warn(
				"SECURITY WARNING: The HOST environment variable is set to a non-local address, which may expose the service to the network. Please ensure this is intentional.",
				"host", cfg.Host,
			)
		}
	}
	return cfg, nil
}
