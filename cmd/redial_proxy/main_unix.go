package main

import (
	"fmt"
	"net"
)

func GetListener() (net.Listener, error) {
	return net.Listen("tcp", fmt.Sprintf(":%d", PORT))
}
