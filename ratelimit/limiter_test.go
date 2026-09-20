package ratelimit_test

import (
	stdhttp "net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/zatrano/framework/v2/kernel/http"
	"github.com/zatrano/packages/ratelimit"
)

func TestNamedLimiter(t *testing.T) {
	limiter := ratelimit.New()
	limiter.For("demo", ratelimit.Limit{MaxAttempts: 2, Decay: time.Minute})
	if !limiter.Has("demo") {
		t.Fatal("expected named limit")
	}

	mw := limiter.Named("demo")
	handler := mw(func(req *http.Request) *http.Response {
		return http.JSON(map[string]any{"ok": true})
	})

	call := func() int {
		r := httptest.NewRequest(stdhttp.MethodGet, "/x", nil)
		r.RemoteAddr = "1.2.3.4:1234"
		return handler(http.NewRequest(r)).StatusCode()
	}

	first := call()
	second := call()
	if first != 200 || second != 200 {
		t.Fatal("first two should pass")
	}
	if call() != 429 {
		t.Fatal("expected 429")
	}

	missing := limiter.Named("missing")(func(req *http.Request) *http.Response {
		return http.JSON(map[string]any{"ok": true})
	})
	if missing(http.NewRequest(httptest.NewRequest(stdhttp.MethodGet, "/", nil))).StatusCode() != 500 {
		t.Fatal("undefined named policy is fail-closed")
	}
}

func TestLimiterTakeIsAtomic(t *testing.T) {
	limiter := ratelimit.New()
	const max = 25
	var allowed, blocked atomic.Int32
	done := make(chan struct{}, 80)
	for i := 0; i < 80; i++ {
		go func() {
			ok, _, _ := limiter.Take("k", max, time.Minute)
			if ok {
				allowed.Add(1)
			} else {
				blocked.Add(1)
			}
			done <- struct{}{}
		}()
	}
	for i := 0; i < 80; i++ {
		<-done
	}
	if allowed.Load() != max || blocked.Load() != 80-max {
		t.Fatalf("allowed=%d blocked=%d", allowed.Load(), blocked.Load())
	}
}
