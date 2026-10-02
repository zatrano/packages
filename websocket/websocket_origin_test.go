package websocket

import (
	stdhttp "net/http"
	"testing"

	"github.com/zatrano/framework/v3/core/kernel/http"
)

func TestSameOriginRejectsCrossSite(t *testing.T) {
	raw := &stdhttp.Request{Host: "app.example.com", Header: stdhttp.Header{}}
	raw.Header.Set("Origin", "https://evil.example")
	req := http.RequestFromHTTP(raw)
	if SameOrigin(req) {
		t.Fatal("cross-site Origin must be rejected")
	}
}

func TestSameOriginAllowsMatchingHost(t *testing.T) {
	raw := &stdhttp.Request{Host: "app.example.com", Header: stdhttp.Header{}}
	raw.Header.Set("Origin", "https://app.example.com")
	req := http.RequestFromHTTP(raw)
	if !SameOrigin(req) {
		t.Fatal("same-origin must be allowed")
	}
}

func TestSameOriginAllowsMissingOrigin(t *testing.T) {
	raw := &stdhttp.Request{Host: "app.example.com", Header: stdhttp.Header{}}
	req := http.RequestFromHTTP(raw)
	if !SameOrigin(req) {
		t.Fatal("missing Origin (non-browser) should be allowed")
	}
}
