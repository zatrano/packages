package timing_test

import (
	stdhttp "net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/zatrano/framework/v3/core/kernel/http"
	"github.com/zatrano/packages/toolkit/timing"
)

func TestTimingMarks(t *testing.T) {
	r := httptest.NewRequest(stdhttp.MethodGet, "/", nil)
	req := http.RequestFromHTTP(r)
	timing.Add(req, "db", 12*time.Millisecond, "query")
	timing.Add(req, "template", 5*time.Millisecond)
	header := timing.Header(req, 20*time.Millisecond)
	if !strings.Contains(header, "app;dur=") || !strings.Contains(header, "db;") || !strings.Contains(header, "total;") {
		t.Fatal(header)
	}
}
