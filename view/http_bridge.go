package view

import (
	"fmt"
	stdhttp "net/http"

	"github.com/zatrano/framework/contracts"
	"github.com/zatrano/framework/kernel/http"
)

type httpBridge struct {
	app contracts.App
}

func installHTTPBridge(app contracts.App) {
	app.SetHTTPBridge(&httpBridge{app: app})
}

func (b *httpBridge) Middleware() []any { return nil }

func (b *httpBridge) Finalize(w stdhttp.ResponseWriter, reqAny any, respAny any) any {
	resp, _ := respAny.(*http.Response)
	return RenderView(b.app, resp)
}

// RenderView executes a view response when the engine is bound.
func RenderView(app contracts.App, resp *http.Response) *http.Response {
	if resp == nil || resp.ViewName() == "" {
		return resp
	}
	engine := From(app)
	if engine == nil {
		return resp
	}
	data := resp.ViewData()
	if data == nil {
		data = map[string]any{}
	}
	html, err := engine.Render(resp.ViewName(), data)
	if err != nil {
		if app.IsDebug() {
			return http.HTML(fmt.Sprintf("<h1>View Error</h1><pre>%v</pre>", err)).Status(500)
		}
		return http.Abort(500, "View rendering failed")
	}
	resp.SetContent([]byte(html), "text/html; charset=utf-8")
	return resp
}
