package auth

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/zatrano/framework/v3/core/bootstrap/addons"
	"github.com/zatrano/framework/v3/core/contracts"
	"github.com/zatrano/framework/v3/core/ssr/starter"
	"github.com/zatrano/packages/bootutil"
)

func Commands(app contracts.App) []addons.CLICommand {
	return bootutil.CLI(
		&MakeAuthCommand{app: app},
		&MakePanelCommand{app: app},
	)
}

// MakeAuthCommand scaffolds the full application auth layer.
type MakeAuthCommand struct {
	app contracts.App
}

func (c *MakeAuthCommand) Name() string { return "make:auth" }
func (c *MakeAuthCommand) Description() string {
	return "Scaffold the auth HTTP surface (handlers/auth/{web,api}, routes/auth/{web,api}, templates)"
}

func (c *MakeAuthCommand) Handle(args []string) error {
	force := false
	templatesOnly := false
	socialFlagSet := false
	var socialProviders []string
	for _, arg := range args {
		switch {
		case arg == "--force" || arg == "-f":
			force = true
		case arg == "--templates":
			templatesOnly = true
		case arg == "--social":
			socialFlagSet = true
			socialProviders = appendUnique(socialProviders, "google")
		case strings.HasPrefix(arg, "--social="):
			socialFlagSet = true
			raw := strings.TrimSpace(strings.TrimPrefix(arg, "--social="))
			if raw == "" {
				return fmt.Errorf("--social requires a provider (example: --social=google)")
			}
			for _, p := range strings.Split(raw, ",") {
				p = strings.ToLower(strings.TrimSpace(p))
				if p == "" {
					continue
				}
				if p != "google" {
					return fmt.Errorf("unknown social provider %q (only google)", p)
				}
				socialProviders = appendUnique(socialProviders, p)
			}
			if len(socialProviders) == 0 {
				return fmt.Errorf("--social requires a provider (example: --social=google)")
			}
		}
	}
	if !socialFlagSet {
		socialProviders = nil
	}
	wantSocial := socialFlagSet
	wantSocialGo := !templatesOnly && wantSocial

	type filePair struct {
		stub string
		dest []string
	}

	pairs := []filePair{
		{"layouts/auth.html", []string{"templates", "layouts", "auth.html"}},
		{"layouts/mail.html", []string{"templates", "layouts", "mail.html"}},
		{"auth/login.html", []string{"templates", "auth", "login.html"}},
		{"auth/register.html", []string{"templates", "auth", "register.html"}},
		{"auth/forgot-password.html", []string{"templates", "auth", "forgot-password.html"}},
		{"auth/reset-password.html", []string{"templates", "auth", "reset-password.html"}},
		{"auth/confirm-password.html", []string{"templates", "auth", "confirm-password.html"}},
		{"auth/change-password.html", []string{"templates", "auth", "change-password.html"}},
		{"auth/profile.html", []string{"templates", "auth", "profile.html"}},
		{"auth/verify-email.html", []string{"templates", "auth", "verify-email.html"}},
		{"auth/two-factor-challenge.html", []string{"templates", "auth", "two-factor-challenge.html"}},
		{"auth/two-factor.html", []string{"templates", "auth", "two-factor.html"}},
		{"auth/logout-other-devices.html", []string{"templates", "auth", "logout-other-devices.html"}},
		{"mail/auth/password-reset.html", []string{"templates", "mail", "auth", "password-reset.html"}},
		{"mail/auth/verify-email.html", []string{"templates", "mail", "auth", "verify-email.html"}},
		{"mail/auth/password-changed.html", []string{"templates", "mail", "auth", "password-changed.html"}},
		{"lang/en/auth.json", []string{"lang", "en", "auth.json"}},
		{"lang/tr/auth.json", []string{"lang", "tr", "auth.json"}},
	}

	if !templatesOnly {
		pairs = append(pairs,
			filePair{"go/user_model.go.stub", []string{"app", "models", "user.go"}},
			filePair{"go/user_resource.go.stub", []string{"app", "http", "resources", "user_resource.go"}},
			filePair{"go/auth_controller.go.stub", []string{"app", "http", "handlers", "auth", "web", "auth_handler.go"}},
			filePair{"go/api_auth_controller.go.stub", []string{"app", "http", "handlers", "auth", "api", "auth_handler.go"}},
			filePair{"go/auth_service.go.stub", []string{"app", "services", "auth.go"}},
			filePair{"go/login_request.go.stub", []string{"app", "http", "requests", "auth", "login_request.go"}},
			filePair{"go/register_request.go.stub", []string{"app", "http", "requests", "auth", "register_request.go"}},
			filePair{"go/profile_update_request.go.stub", []string{"app", "http", "requests", "auth", "profile_update_request.go"}},
			filePair{"go/change_password_request.go.stub", []string{"app", "http", "requests", "auth", "change_password_request.go"}},
			filePair{"go/forgot_password_request.go.stub", []string{"app", "http", "requests", "auth", "forgot_password_request.go"}},
			filePair{"go/reset_password_request.go.stub", []string{"app", "http", "requests", "auth", "reset_password_request.go"}},
			filePair{"go/confirm_password_request.go.stub", []string{"app", "http", "requests", "auth", "confirm_password_request.go"}},
			filePair{"go/two_factor_challenge_request.go.stub", []string{"app", "http", "requests", "auth", "two_factor_challenge_request.go"}},
			filePair{"go/two_factor_confirm_request.go.stub", []string{"app", "http", "requests", "auth", "two_factor_confirm_request.go"}},
			filePair{"go/authenticate_middleware.go.stub", []string{"app", "http", "middleware", "authenticate.go"}},
			filePair{"go/routes_auth.go.stub", []string{"app", "routes", "auth", "web", "auth.go"}},
			filePair{"go/routes_auth_api.go.stub", []string{"app", "routes", "auth", "api", "auth.go"}},
			filePair{"go/auth_service_provider.go.stub", []string{"app", "providers", "auth_service_provider.go"}},
			filePair{"sql/001_create_auth_tables.sql", []string{"database", "migrations", "001_create_auth_tables.sql"}},
		)
		if wantSocialGo {
			pairs = append(pairs,
				filePair{"go/social_account_model.go.stub", []string{"app", "models", "social_account.go"}},
				filePair{"go/social_auth_service.go.stub", []string{"app", "services", "social.go"}},
				filePair{"go/social_auth_controller.go.stub", []string{"app", "http", "handlers", "auth", "web", "social_auth_handler.go"}},
				filePair{"go/api_social_auth_controller.go.stub", []string{"app", "http", "handlers", "auth", "api", "social_auth_handler.go"}},
				filePair{"sql/002_create_social_accounts_table.sql", []string{"database", "migrations", "002_create_social_accounts_table.sql"}},
			)
		}
	}

	created, skipped := 0, 0
	var routesAuthPath, routesAuthAPIPath, loginViewPath, registerViewPath, layoutAuthPath, langEnPath, langTrPath string
	for _, pair := range pairs {
		body, err := readStub(c.app, pair.stub)
		if err != nil {
			return fmt.Errorf("%s: %w", pair.stub, err)
		}
		dst := bootutil.ScaffoldDest(c.app, pair.dest)
		if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
			return err
		}
		if !force {
			if _, err := os.Stat(dst); err == nil {
				if strings.HasSuffix(dst, "auth_service_provider.go") {
					fmt.Printf("Skipped (exists): %s — add gates/policies manually or use --force\n", dst)
				} else {
					fmt.Printf("Skipped (exists): %s\n", dst)
				}
				skipped++
				continue
			}
		}
		if strings.HasSuffix(dst, ".go") {
			if conflict, err := goStubConflicts(string(body), filepath.Dir(dst), dst); err != nil {
				return err
			} else if conflict != "" {
				fmt.Printf("Skipped (declared): %s — %s\n", dst, conflict)
				skipped++
				continue
			}
		}
		if err := writeStubFile(c.app, body, dst); err != nil {
			return fmt.Errorf("%s: %w", pair.stub, err)
		}
		fmt.Printf("Created: %s\n", dst)
		created++
		switch pair.stub {
		case "go/routes_auth.go.stub":
			routesAuthPath = dst
		case "go/routes_auth_api.go.stub":
			routesAuthAPIPath = dst
		case "auth/login.html":
			loginViewPath = dst
		case "auth/register.html":
			registerViewPath = dst
		case "layouts/auth.html":
			layoutAuthPath = dst
		case "lang/en/auth.json":
			langEnPath = dst
		case "lang/tr/auth.json":
			langTrPath = dst
		}
	}

	if wantSocial {
		if !templatesOnly && routesAuthPath != "" {
			if err := injectAuthSocialWebRoutes(routesAuthPath); err != nil {
				return err
			}
		}
		if !templatesOnly && routesAuthAPIPath != "" {
			if err := injectAuthSocialAPIRoutes(routesAuthAPIPath); err != nil {
				return err
			}
		}
		if loginViewPath != "" {
			if err := injectAuthSocialLinks(loginViewPath); err != nil {
				return err
			}
		}
		if registerViewPath != "" {
			if err := injectAuthSocialLinks(registerViewPath); err != nil {
				return err
			}
		}
		if layoutAuthPath != "" {
			if err := injectAuthSocialLayout(layoutAuthPath); err != nil {
				return err
			}
		}
		if langEnPath != "" {
			if err := mergeAuthSocialLang(c.app, langEnPath, "lang/en/social.json"); err != nil {
				return err
			}
		}
		if langTrPath != "" {
			if err := mergeAuthSocialLang(c.app, langTrPath, "lang/tr/social.json"); err != nil {
				return err
			}
		}
	}
	if !templatesOnly {
		mod := bootutil.ConsumerModule(c.app)
		provider := c.app.BasePath("app", "providers", "route_service_provider.go")
		_ = bootutil.EnsureBlankImport(provider, mod+"/app/routes/auth/web")
		_ = bootutil.EnsureBlankImport(provider, mod+"/app/routes/auth/api")
	}
	if err := enableViewForAuth(c.app); err != nil {
		return err
	}
	if err := enableURLForAuth(c.app); err != nil {
		return err
	}

	fmt.Printf("\nAuth scaffold ready (%d created, %d skipped).\n", created, skipped)
	if templatesOnly {
		fmt.Println("Mode: --templates (auth HTML, layout, and mail templates only)")
	} else {
		fmt.Println("Next steps:")
		fmt.Println("  1. Enable hashing, database (packages/db), session, and auth (notification for mail). template (Canvas) and url are already enabled.")
		fmt.Println("  2. Apply SQL under database/migrations/ (001_create_auth_tables.sql[+002 social]).")
		fmt.Println("  3. Auth routes self-register from app/routes/auth/web and app/routes/auth/api.")
		fmt.Println("  4. Set AUTH_MUST_VERIFY_EMAIL=true to require email confirmation (default false).")
		if wantSocialGo {
			fmt.Println("  5. Set GOOGLE_* env vars for social login")
			fmt.Println("  6. Run your db migrate tool against database/migrations/")
		} else {
			fmt.Println("  5. Run your db migrate tool against database/migrations/")
		}
	}
	fmt.Println("Use --force to overwrite existing files. Use --templates for Canvas templates only.")
	return nil
}

func enableViewForAuth(app contracts.App) error {
	if err := bootutil.EnsureEnabledAddon(app, "template"); err != nil {
		return err
	}
	if err := bootutil.EnsureBlankImport(app.BasePath("bootstrap", "addons.go"), "github.com/zatrano/framework/v3/core/ssr"); err != nil {
		return err
	}
	return starter.Write(app)
}

func enableURLForAuth(app contracts.App) error {
	if err := bootutil.EnsureEnabledAddon(app, "url"); err != nil {
		return err
	}
	return bootutil.EnsureBlankImport(app.BasePath("bootstrap", "addons.go"), "github.com/zatrano/packages/url")
}

var (
	goTypeDecl = regexp.MustCompile(`(?m)^type\s+([A-Za-z_][A-Za-z0-9_]*)\s+`)
	goFuncDecl = regexp.MustCompile(`(?m)^func\s+([A-Za-z_][A-Za-z0-9_]*)\s*\(`)
)

// goStubConflicts returns a human reason when writing dst would redeclare symbols already in the package.
func goStubConflicts(stub, pkgDir, dst string) (string, error) {
	typeNames := uniqueMatches(goTypeDecl, stub)
	funcNames := uniqueMatches(goFuncDecl, stub)

	entries, err := os.ReadDir(pkgDir)
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", err
	}
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".go") {
			continue
		}
		path := filepath.Join(pkgDir, entry.Name())
		if filepath.Clean(path) == filepath.Clean(dst) {
			continue // same file will be overwritten under --force
		}
		body, err := os.ReadFile(path)
		if err != nil {
			return "", err
		}
		text := string(body)
		for _, name := range typeNames {
			for _, existing := range uniqueMatches(goTypeDecl, text) {
				if existing == name {
					return fmt.Sprintf("type %s already in %s", name, entry.Name()), nil
				}
			}
		}
		for _, name := range funcNames {
			for _, existing := range uniqueMatches(goFuncDecl, text) {
				if existing == name {
					return fmt.Sprintf("func %s already in %s", name, entry.Name()), nil
				}
			}
		}
		if strings.Contains(stub, "RegisterAuthWeb") && strings.Contains(text, `.As("login")`) {
			return fmt.Sprintf("login routes already in %s", entry.Name()), nil
		}
		if strings.Contains(stub, "001_create_auth_tables") &&
			(strings.Contains(text, "remember_token") || strings.Contains(text, "two_factor_secret") || (strings.Contains(text, "CREATE TABLE") && strings.Contains(text, "users"))) {
			return fmt.Sprintf("auth tables already in %s", entry.Name()), nil
		}
		if strings.Contains(stub, "002_create_social_accounts") &&
			(strings.Contains(text, "provider_uid") || strings.Contains(text, "social_accounts")) {
			return fmt.Sprintf("social accounts already in %s", entry.Name()), nil
		}
	}
	return "", nil
}

func uniqueMatches(re *regexp.Regexp, src string) []string {
	seen := map[string]struct{}{}
	var out []string
	for _, m := range re.FindAllStringSubmatch(src, -1) {
		if len(m) < 2 {
			continue
		}
		name := m[1]
		if _, ok := seen[name]; ok {
			continue
		}
		seen[name] = struct{}{}
		out = append(out, name)
	}
	return out
}

func writeStubFile(app contracts.App, body []byte, dst string) error {
	text := string(body)
	if strings.HasSuffix(dst, ".go") {
		text = bootutil.ApplyConsumerPlaceholders(app, text)
	}
	return os.WriteFile(dst, []byte(text), 0o644)
}

func containsStr(list []string, want string) bool {
	for _, v := range list {
		if v == want {
			return true
		}
	}
	return false
}

func appendUnique(list []string, v string) []string {
	if containsStr(list, v) {
		return list
	}
	return append(list, v)
}

func injectAuthSocialWebRoutes(path string) error {
	body, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	text := strings.ReplaceAll(string(body), "\r\n", "\n")
	if strings.Contains(text, "SocialAuthHandler") {
		return nil
	}
	text = strings.Replace(text, "ctrl := &authctrl.AuthHandler{App: app}",
		"ctrl := &authctrl.AuthHandler{App: app}\n	social := &authctrl.SocialAuthHandler{App: app}", 1)
	text = strings.Replace(text, "r.Post(\"/logout\", ctrl.Logout).As(\"logout\")",
		"r.Post(\"/logout\", ctrl.Logout).As(\"logout\")\n\n		r.Get(\"/google/login\", social.GoogleRedirect).As(\"login.google\")\n		r.Get(\"/google/callback\", social.GoogleCallback).As(\"login.google.callback\")", 1)
	return os.WriteFile(path, []byte(text), 0o644)
}

func injectAuthSocialAPIRoutes(path string) error {
	body, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	text := strings.ReplaceAll(string(body), "\r\n", "\n")
	if strings.Contains(text, "SocialAuthHandler") {
		return nil
	}
	text = strings.Replace(text, "ctrl := &apictrl.AuthHandler{App: app}",
		"ctrl := &apictrl.AuthHandler{App: app}\n	social := &apictrl.SocialAuthHandler{App: app}", 1)
	text = strings.Replace(text, "r.Post(\"/register\", ctrl.Register).As(\"api.v1.register\").Through(ratelimit.From(app).Named(\"login\"))",
		"r.Post(\"/register\", ctrl.Register).As(\"api.v1.register\").Through(ratelimit.From(app).Named(\"login\"))\n			r.Get(\"/google\", social.GoogleRedirect).As(\"api.v1.login.google\")\n			r.Get(\"/google/callback\", social.GoogleCallback).As(\"api.v1.login.google.callback\")\n			r.Post(\"/google/callback\", social.GoogleCallback).As(\"api.v1.login.google.callback.store\")", 1)
	return os.WriteFile(path, []byte(text), 0o644)
}

func injectAuthSocialLinks(path string) error {
	body, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	text := strings.ReplaceAll(string(body), "\r\n", "\n")
	if strings.Contains(text, "auth-social") {
		return nil
	}
	snippet := "    <p class=\"auth-divider\">@lang('auth.or_continue_with')</p>\n    <div class=\"auth-social\">\n        <a href=\"/auth/google/login\">@lang('auth.continue_google')</a>\n    </div>\n"
	text = strings.Replace(text, "    <p class=\"auth-links\">", snippet+"    <p class=\"auth-links\">", 1)
	return os.WriteFile(path, []byte(text), 0o644)
}

func injectAuthSocialLayout(path string) error {
	body, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	text := strings.ReplaceAll(string(body), "\r\n", "\n")
	if strings.Contains(text, ".auth-social") {
		return nil
	}
	css := `        .auth-divider {
            margin: 0.35rem 0 0;
            text-align: center;
            color: var(--muted);
            font-size: 0.85rem;
        }
        .auth-social {
            display: grid;
            grid-template-columns: 1fr;
            gap: 0.65rem;
        }
        .auth-social a {
            display: block;
            text-align: center;
            padding: 0.75rem 0.5rem;
            border-radius: 8px;
            border: 1px solid var(--line);
            color: var(--text);
            text-decoration: none;
            font-family: Syne, sans-serif;
            font-size: 0.85rem;
            font-weight: 700;
        }
        .auth-social a:hover {
            border-color: color-mix(in srgb, var(--brand) 35%, var(--line));
            color: var(--brand);
        }
`
	text = strings.Replace(text, "        .auth-links {", css+"        .auth-links {", 1)
	return os.WriteFile(path, []byte(text), 0o644)
}

func mergeAuthSocialLang(app contracts.App, dst, stub string) error {
	extraBody, err := readStub(app, stub)
	if err != nil {
		return err
	}
	var extra map[string]string
	if err := json.Unmarshal(extraBody, &extra); err != nil {
		return err
	}
	baseBody, err := os.ReadFile(dst)
	if err != nil {
		return err
	}
	var base map[string]string
	if err := json.Unmarshal(baseBody, &base); err != nil {
		return err
	}
	if base == nil {
		base = map[string]string{}
	}
	for k, v := range extra {
		base[k] = v
	}
	out, err := json.MarshalIndent(base, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(dst, append(out, '\n'), 0o644)
}
