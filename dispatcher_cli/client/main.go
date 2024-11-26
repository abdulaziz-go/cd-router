package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
	"time"
)

var (
	version  = "1.0"
	baseHost = "" // tunnel.cloud-dispatcher.uz
)

func main() {
	log.SetFlags(0)

	if len(os.Args) < 3 {
		log.Fatalf("Usage: jprq <PROTOCOL> <PORT> [-subdomain=<SUBDOMAIN>]\n"+
			"  Supported Protocols: [tcp, http]\n"+
			"  Optional Argument: -subdomain\n"+
			"  Client Version: %s\n", version)
	}

	protocol := os.Args[1]
	if protocol != "http" {
		log.Fatalf("invalid protocol %s. only http supported", protocol)
	}

	port, err := strconv.Atoi(os.Args[2])
	if err != nil || port < 0 || port > 65535 {
		log.Fatalf("Invalid port Number %d", port)
	}

	subdomain := ""
	if len(os.Args) > 3 {
		lastArg := os.Args[len(os.Args)-1]
		if strings.HasPrefix(lastArg, "-subdomain=") {
			subdomain = strings.TrimPrefix(lastArg, "-subdomain=")
		} else {
			log.Fatalf("Invalid usage: -subdomain must be the last argument")
		}
	}
	if !canReachServer(port) {
		log.Fatalf("No server is running on port: %d\n", port)
	}
	ctx := context.Background()
	ctx, cancelFunc := context.WithCancel(ctx)
	signalChan := make(chan os.Signal, 1)
	go openHttpTunnel(port, subdomain, ctx)
	<-signalChan
	cancelFunc()
}

func canReachServer(port int) bool {
	timeout := 500 * time.Millisecond
	address := fmt.Sprintf("127.0.0.1:%d", port)
	conn, err := net.DialTimeout("tcp", address, timeout)
	if err != nil {
		return false
	}
	conn.Close()
	return true
}
