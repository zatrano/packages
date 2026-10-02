package session

import (
	stdhttp "net/http"
	"strings"
	"testing"

	"github.com/zatrano/framework/v3/core/kernel"
	"github.com/zatrano/framework/v3/core/kernel/http"
)

func TestFinalizeFailsLoudWithoutCanvasEngine(t *testing.T) {
	app := kernel.NewApplication(t.TempDir())
	b := &httpBridge{app: app}
	raw, err := stdhttp.NewRequest(stdhttp.MethodGet, "/", nil)
	if err != nil {
		t.Fatal(err)
	}
	req := http.RequestFromHTTP(raw)
	out, _ := b.Finalize(req, http.Template("web.welcome")).(*http.Response)
	if out == nil || out.StatusCode() != 500 {
		t.Fatalf("status=%v", out)
	}
	body := string(out.Content())
	if !strings.Contains(body, "Canvas engine not bound") && !strings.Contains(body, "Template rendering failed") {
		t.Fatalf("body=%s", body)
	}
}
