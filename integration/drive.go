package integration

import (
	"bufio"
	"bytes"
	"io"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/zatrano/framework/v3/core/kernel"
	"github.com/zatrano/rawhttp"
)

// Exchange runs one request through Application.Handle on an in-memory connection.
func Exchange(app *kernel.Application, rawRequest string) (status int, header http.Header, body string, err error) {
	conn := &memConn{r: bytes.NewReader([]byte(rawRequest))}
	srv := &rawhttp.Server{
		ReadTimeout:  -1,
		WriteTimeout: -1,
		IdleTimeout:  -1,
		AllowUpgrade: true,
		Handler:      func(ctx *rawhttp.Ctx) { app.Handle(ctx) },
	}
	serveErr := srv.ServeConn(conn)
	if conn.w.Len() == 0 {
		return 0, nil, "", serveErr
	}
	resp, rerr := http.ReadResponse(bufio.NewReader(bytes.NewReader(conn.w.Bytes())), nil)
	if rerr != nil {
		return 0, nil, "", rerr
	}
	defer resp.Body.Close()
	raw, rerr := io.ReadAll(resp.Body)
	if rerr != nil {
		return resp.StatusCode, resp.Header, "", rerr
	}
	return resp.StatusCode, resp.Header, string(raw), nil
}

type memConn struct {
	r  *bytes.Reader
	w  bytes.Buffer
	mu sync.Mutex
}

func (c *memConn) Read(p []byte) (int, error) { return c.r.Read(p) }

func (c *memConn) Write(p []byte) (int, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.w.Write(p)
}

func (c *memConn) Close() error                     { return nil }
func (c *memConn) LocalAddr() net.Addr              { return memAddr{} }
func (c *memConn) RemoteAddr() net.Addr             { return memAddr{} }
func (c *memConn) SetDeadline(time.Time) error      { return nil }
func (c *memConn) SetReadDeadline(time.Time) error  { return nil }
func (c *memConn) SetWriteDeadline(time.Time) error { return nil }

type memAddr struct{}

func (memAddr) Network() string { return "tcp" }
func (memAddr) String() string  { return "127.0.0.1:9" }
