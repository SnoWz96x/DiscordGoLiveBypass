package main

import (
	"io"
	"net"
	"strings"
	"testing"
	"time"
)

func TestHTTPSProxyStartsWithTLSAndPropagatesHandshakeFailure(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer listener.Close()
	hello := make(chan byte, 1)
	go func() {
		c, err := listener.Accept()
		if err != nil {
			hello <- 0
			return
		}
		defer c.Close()
		c.SetDeadline(time.Now().Add(time.Second))
		b := make([]byte, 1)
		io.ReadFull(c, b)
		hello <- b[0]
	}()
	addr := listener.Addr().(*net.TCPAddr)
	_, err = OpenTunnel(Endpoint{"https", "127.0.0.1", addr.Port}, "discord.com", 443, time.Second)
	if err == nil || !strings.Contains(err.Error(), "TLS com o proxy") {
		t.Fatalf("wanted TLS error, got %v", err)
	}
	if b := <-hello; b != 22 {
		t.Fatalf("first byte %d: expected TLS handshake, not plaintext CONNECT", b)
	}
}
