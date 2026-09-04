package view

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/zatrano/framework/contracts"
	"github.com/zatrano/framework/kernel/layout"
	"github.com/zatrano/packages/assets"
	"github.com/zatrano/packages/bootutil"
	"github.com/zatrano/packages/localization"
)

func boot(app contracts.App) error {
	engine := New(layout.ViewsDir(app))
	engine.EnableCache(!app.IsDebug())
	engine.Share("appName", app.Config().GetString("app.name", "ZATRANO"))
	if tr := localization.From(app); tr != nil {
		engine.Share("locale", tr.GetLocale())
		engine.AddFunc("trans", func(localeOrKey string, args ...any) string {
			locale := ""
			key := localeOrKey
			var replace map[string]string
			if len(args) == 0 {
				return tr.Get(key)
			}
			if s, ok := args[0].(string); ok {
				locale = localeOrKey
				key = s
				if len(args) > 1 {
					replace = bootutil.CoerceStringMap(args[1])
				}
				return tr.GetFor(locale, key, replace)
			}
			replace = bootutil.CoerceStringMap(args[0])
			return tr.Get(key, replace)
		})
		engine.AddFunc("dict", func(pairs ...any) map[string]any {
			out := map[string]any{}
			for i := 0; i+1 < len(pairs); i += 2 {
				out[fmt.Sprint(pairs[i])] = pairs[i+1]
			}
			return out
		})
		engine.AddFunc("choice", func(key string, number any) string {
			n := 0
			switch v := number.(type) {
			case int:
				n = v
			case int64:
				n = int(v)
			case float64:
				n = int(v)
			case string:
				if parsed, err := strconv.Atoi(strings.TrimSpace(v)); err == nil {
					n = parsed
				}
			default:
				n, _ = strconv.Atoi(fmt.Sprint(number))
			}
			return tr.Choice(key, n)
		})
	}
	if a := assets.From(app); a != nil {
		engine.AddFunc("vite", func(path string) string {
			return a.URL(path)
		})
		engine.AddFunc("mix", func(path string) string {
			return a.URL(path)
		})
	}
	engine.SetEnvironment(app.Environment())
	app.Container().Instance("view", engine)
	return nil
}
