package websocket

import (
	"bufio"
	"context"
	"io"
	"net"
	"strings"
	"testing"
	"time"

	"github.com/zatrano/framework/v3/core/kernel"
	"github.com/zatrano/framework/v3/core/kernel/http"
	"github.com/zatrano/rawhttp"
)

func TestShutdownSends1001AndRejectsNewUpgrade(t *testing.T) {
	http.ResetShutdownState()
	closing.Store(false)
	http.RegisterShutdownHook("websocket", shutdownAll)
	t.Cleanup(func() {
		http.ResetShutdownState()
		closing.Store(false)
		http.RegisterShutdownHook("websocket", shutdownAll)
	})

	app := kernel.NewApplication(t.TempDir())
	t.Cleanup(func() {
		if app == nil || app.Logger() == nil {
			return
		}
		if c, ok := app.Logger().(interface{ Close() error }); ok {
			_ = c.Close()
		}
	})
	app.Router().Get("/ws", Upgrade(func(c *Conn) error {
		_, _, err := c.ReadMessage()
		return err
	}))
	if err := app.Start(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = app.Stop(context.Background()) })

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	srv := &rawhttp.Server{
		Handler:           func(ctx *rawhttp.Ctx) { app.Handle(ctx) },
		AllowUpgrade:      true,
		KeepHijackedConns: true,
		ReadTimeout:       -1,
		WriteTimeout:      -1,
		IdleTimeout:       -1,
	}
	go func() { _ = srv.Serve(ln) }()
	t.Cleanup(func() { _ = srv.Close() })
	addr := ln.Addr().String()

	clients := make([]net.Conn, 3)
	for i := range clients {
		c, err := dialWS(t, addr)
		if err != nil {
			t.Fatal(err)
		}
		clients[i] = c
		defer c.Close()
	}
	deadline := time.Now().Add(2 * time.Second)
	for liveCount() < 3 && time.Now().Before(deadline) {
		time.Sleep(5 * time.Millisecond)
	}
	if liveCount() != 3 {
		t.Fatalf("tracked %d", liveCount())
	}

	frames := make(chan []byte, 3)
	for _, c := range clients {
		go func(c net.Conn) {
			_ = c.SetReadDeadline(time.Now().Add(3 * time.Second))
			buf := make([]byte, 8)
			n, _ := c.Read(buf)
			frames <- buf[:n]
		}(c)
	}

	started := time.Now()
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if errs := http.RunShutdownHooks(ctx); len(errs) != 0 {
		t.Fatal(errs)
	}
	if st := upgradeStatus(t, addr); st != 503 {
		t.Fatalf("upgrade after shutdown status=%d", st)
	}
	if err := srv.Shutdown(ctx); err != nil {
		t.Fatal(err)
	}
	if time.Since(started) > 2*time.Second {
		t.Fatalf("shutdown took %s", time.Since(started))
	}
	if liveCount() != 0 {
		t.Fatalf("tracked %d after shutdown", liveCount())
	}
	for i := 0; i < 3; i++ {
		buf := <-frames
		if len(buf) < 4 || buf[0]&0x0f != 0x8 || buf[2] != 0x03 || buf[3] != 0xe9 {
			t.Fatalf("client %d frame %x", i, buf)
		}
	}
}

func dialWS(t *testing.T, addr string) (net.Conn, error) {
	t.Helper()
	c, err := net.Dial("tcp", addr)
	if err != nil {
		return nil, err
	}
	if _, err := io.WriteString(c, "GET /ws HTTP/1.1\r\nHost: localhost\r\nUpgrade: websocket\r\nConnection: Upgrade\r\nSec-WebSocket-Key: dGhlIHNhbXBsZSBub25jZQ==\r\nSec-WebSocket-Version: 13\r\n\r\n"); err != nil {
		c.Close()
		return nil, err
	}
	_ = c.SetReadDeadline(time.Now().Add(2 * time.Second))
	br := bufio.NewReader(c)
	status, err := br.ReadString('\n')
	if err != nil {
		c.Close()
		return nil, err
	}
	if !strings.Contains(status, "101") {
		c.Close()
		t.Fatalf("handshake %q", status)
	}
	for {
		line, err := br.ReadString('\n')
		if err != nil || line == "\r\n" {
			break
		}
	}
	_ = c.SetReadDeadline(time.Time{})
	return c, nil
}

func upgradeStatus(t *testing.T, addr string) int {
	t.Helper()
	c, err := net.Dial("tcp", addr)
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()
	_ = c.SetDeadline(time.Now().Add(2 * time.Second))
	if _, err := io.WriteString(c, "GET /ws HTTP/1.1\r\nHost: localhost\r\nUpgrade: websocket\r\nConnection: Upgrade\r\nSec-WebSocket-Key: dGhlIHNhbXBsZSBub25jZQ==\r\nSec-WebSocket-Version: 13\r\n\r\n"); err != nil {
		t.Fatal(err)
	}
	br := bufio.NewReader(c)
	line, err := br.ReadString('\n')
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(line, "503") {
		return 503
	}
	if strings.Contains(line, "101") {
		return 101
	}
	t.Fatalf("status line %q", line)
	return 0
}
