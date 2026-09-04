package browser_test

import (
	"testing"

	"github.com/zatrano/framework/bootstrap"
	"github.com/zatrano/framework/http"
	"github.com/zatrano/packages/browser"
)

func TestBrowserVisitHome(t *testing.T) {
	t.Setenv("APP_KEY", "test-key-for-packages-browser-tests!")
	t.Setenv("APP_CONFIG_CACHE", "false")
	app := bootstrap.App()
	if err := app.Bootstrap(); err != nil {
		t.Fatal(err)
	}
	app.Router().Get("/browser-probe", func(req *http.Request) *http.Response {
		return http.HTML("probe-ok")
	})
	b, err := browser.New(app)
	if err != nil {
		t.Fatal(err)
	}
	b.Visit("/browser-probe").AssertOK().AssertSee("probe-ok").AssertPathIs("/browser-probe")
}
