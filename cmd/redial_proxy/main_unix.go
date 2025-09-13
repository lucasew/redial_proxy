package main

import (
	"fmt"
	"net"

	"blitiri.com.ar/go/systemd"
)

func GetListener() (net.Listener, error) {
	listeners, err := systemd.Listeners()
	if err != nil {
		return nil, err
	}
	for _, v := range listeners {
		for _, listener := range v {
			return listener, nil
		}
	}
	return net.Listen("tcp", fmt.Sprintf(":%d", PORT))
}
