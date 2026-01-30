//go:build windows

package getlistener

import (
	"fmt"
	"net"
)

func GetListener() (net.Listener, error) {
	cfg, err := loadConfig()
	if err != nil {
		return nil, err
	}
	return net.Listen("tcp", fmt.Sprintf("%s:%d", cfg.Host, cfg.Port))
}
