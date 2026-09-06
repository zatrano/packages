package session

import (
	"fmt"
	stdhttp "net/http"
	"strings"
	"time"

	"github.com/zatrano/framework/v2/contracts"
	"github.com/zatrano/framework/v2/kernel/dirs"
	"github.com/zatrano/framework/v2/kernel/env"
	"github.com/zatrano/framework/v2/kernel/http"
	"github.com/zatrano/framework/v2/kernel/routing"
	"github.com/zatrano/packages/flash"
	"github.com/zatrano/packages/localization"
	"github.com/zatrano/packages/validation"
	"github.com/zatrano/packages/view"
)

type httpBridge struct {
	app contracts.App
}

func installHTTPBridge(app contracts.App) {
	app.SetHTTPBridge(&httpBridge{app: app})
}

func (b *httpBridge) Middleware() []any {
	return []any{
		b.sessionMiddleware(),
		b.localeMiddleware(),
	}
}

func (b *httpBridge) Finalize(w stdhttp.ResponseWriter, reqAny any, respAny any) any {
	req, _ := reqAny.(*http.Request)
	resp, _ := respAny.(*http.Response)
	return b.finalize(w, req, resp)
}

func (b *httpBridge) finalize(w stdhttp.ResponseWriter, req *http.Request, resp *http.Response) *http.Response {
	app := b.app
	if resp == nil {
		resp = http.Abort(204)
	}

	engine := view.From(app)
	if resp.ViewName() != "" && engine != nil {
		data := resp.ViewData()
		if data == nil {
			data = map[string]any{}
		}
		if sess := req.Session(); sess != nil {
			if token, ok := sess.Get("_csrf_token").(string); ok {
				data["_token"] = token
			}
			data["flash"] = flash.All(req)
			data["old"] = flash.OldInput(req)
			data["errors"] = validation.ErrorsFromSession(req)
			data["errorBags"] = validation.ErrorBagsFromSession(req)
		} else {
			data["old"] = map[string]string{}
			data["errors"] = validation.NewMessageBag(nil)
			data["errorBags"] = map[string]any{}
		}
		type authView interface {
			Check(req *http.Request) bool
			User(req *http.Request) any
		}
		var authMgr authView
		if raw, err := app.Make("auth"); err == nil {
			authMgr, _ = raw.(authView)
		}
		authenticated := authMgr != nil && authMgr.Check(req)
		data["auth"] = authenticated
		data["guest"] = !authenticated
		if tr := localization.From(app); tr != nil {
			locale := tr.GetLocale()
			if req != nil {
				if v, ok := req.Get("locale").(string); ok && strings.TrimSpace(v) != "" {
					locale = strings.TrimSpace(v)
				}
			}
			data["locale"] = locale
			langPath := dirs.LocalizationDir(app)
			data["langPublished"] = localization.Published(langPath)
			data["locales"] = localization.Options(langPath, locale)
		}
		var user any
		if authenticated {
			if u := authMgr.User(req); u != nil {
				user = u
				data["user"] = user
			}
		}
		if raw, err := app.Make("gate"); err == nil {
			type gater interface {
				Allows(user any, ability string, args ...any) bool
			}
			if gate, ok := raw.(gater); ok {
				data["__can"] = func(ability string, args ...any) bool {
					return gate.Allows(user, ability, args...)
				}
			}
		}
		html, err := engine.Render(resp.ViewName(), data)
		if err != nil {
			if app.IsDebug() {
				resp = http.HTML(fmt.Sprintf("<h1>View Error</h1><pre>%v</pre>", err)).Status(500)
			} else {
				resp = http.Abort(500, "View rendering failed")
			}
		} else {
			resp.SetContent([]byte(html), "text/html; charset=utf-8")
		}
	}

	if bag, ok := req.Session().(*Bag); ok && bag != nil {
		if sess := From(app); sess != nil {
			hadCookie := strings.TrimSpace(req.Cookie(sess.CookieName())) != ""
			if bag.Changed() || hadCookie {
				_ = sess.Save(bag)
				stdhttp.SetCookie(w, &stdhttp.Cookie{
					Name:     sess.CookieName(),
					Value:    bag.ID(),
					Path:     "/",
					HttpOnly: true,
					Secure:   req.Secure() || env.GetBool("SESSION_SECURE", false),
					SameSite: stdhttp.SameSiteLaxMode,
					MaxAge:   int(time.Hour.Seconds() * 2),
				})
			}
		}
	}
	return resp
}

func (b *httpBridge) sessionMiddleware() routing.MiddlewareFunc {
	return func(next routing.HandlerFunc) routing.HandlerFunc {
		return func(req *http.Request) *http.Response {
			sess := From(b.app)
			if sess == nil {
				return next(req)
			}
			id := req.Cookie(sess.CookieName())
			bag, err := sess.Start(id)
			if err == nil {
				req.SetSession(bag)
			}
			return next(req)
		}
	}
}

func (b *httpBridge) localeMiddleware() routing.MiddlewareFunc {
	return func(next routing.HandlerFunc) routing.HandlerFunc {
		return func(req *http.Request) *http.Response {
			tr := localization.From(b.app)
			if tr != nil {
				langPath := dirs.LocalizationDir(b.app)
				locale := ""
				if sess := req.Session(); sess != nil {
					if raw, ok := sess.Get("locale").(string); ok {
						locale = strings.TrimSpace(strings.ToLower(raw))
					}
				}
				if locale == "" {
					// Prefer APP_LOCALE (translator default) over Accept-Language.
					configured := strings.TrimSpace(strings.ToLower(tr.GetLocale()))
					if configured != "" && localization.HasLocale(langPath, configured) {
						locale = configured
					} else {
						locale = negotiateLocale(req.Header("Accept-Language"), localization.Available(langPath))
					}
				}
				if locale != "" && localization.HasLocale(langPath, locale) {
					tr.SetLocale(locale)
					_ = tr.Load(locale)
					req.Set("locale", locale)
				}
				if engine := view.From(b.app); engine != nil {
					engine.Share("locale", locale)
				}
			}
			return next(req)
		}
	}
}

func negotiateLocale(header string, available []string) string {
	header = strings.TrimSpace(header)
	if header == "" || len(available) == 0 {
		return ""
	}
	allowed := map[string]string{}
	for _, code := range available {
		code = strings.ToLower(strings.TrimSpace(code))
		allowed[code] = code
		if i := strings.IndexByte(code, '-'); i > 0 {
			allowed[code[:i]] = code
		}
	}
	for _, part := range strings.Split(header, ",") {
		tag := strings.TrimSpace(strings.Split(part, ";")[0])
		tag = strings.ToLower(tag)
		if tag == "" || tag == "*" {
			continue
		}
		if code, ok := allowed[tag]; ok {
			return code
		}
		if i := strings.IndexByte(tag, '-'); i > 0 {
			if code, ok := allowed[tag[:i]]; ok {
				return code
			}
		}
	}
	return ""
}
