# Auth

Session authentication, remember-me, password reset, email verification, lockout, TOTP 2FA, and multi-device logout.

## Resolve

```go
auth.From(app)
auth.Passwords(app)
```

## Guards

Config-driven named session guards use per-guard session keys (`login_{guard}_id`):

```go
auth.Middleware(auth.From(app))           // default guard
auth.Middleware(auth.From(app), "web", "api")
auth.VerifyEmailMiddleware(auth.From(app))
```

Personal access tokens live in `packages/apitoken` (middleware), not as an auth guard driver.

## Email verification

`AUTH_MUST_VERIFY_EMAIL` / `auth.must_verify_email` (default `false`) is merged on `package:enable auth`.

- `false`: `Register` stamps `email_verified_at`; `VerifyEmailMiddleware` still requires a session but not confirmation.
- `true`: register/profile send a verification mail; `VerifyEmailMiddleware` blocks unverified users (HTML → `/auth/email/verify`, JSON 403 with `"resend"`).

The generated `User` implements `MustVerifyEmail` (it has an email). Whether confirmation is **required** is the env flag / `a.MustVerifyEmail()`, not the interface.

## HTTP surfaces

`make:auth` writes two controllers: HTML under `RegisterWeb`, JSON under `RegisterAPI` + kernel `routing.Version` → `/api/v1/auth`. There is no `packages/api`. JSON login/register return `user`; verify/resend/`GET /user`/2FA match the web flows. Signed links use `url.From(app)` (`make:auth` enables `url`).

## Lockout & 2FA

Lockout counters use the app cache when available (`SetLockoutCache`).

```go
ok, err := auth.From(app).ChallengeTwoFactor(req, code, true) // remember this device
```

Config: `auth.two_factor.issuer`, `auth.two_factor.remember_device_days`, `auth.lockout.*`, `auth.must_verify_email`.

## Scaffold

```bash
zatrano make:auth
```

`make:auth` / `make:panel` wrap account and panel routes with `VerifyEmailMiddleware`. Scaffold flashes use `auth.*` locale keys (`APP_LOCALE` + `lang/{locale}/auth.json`). Package errors such as `auth.ErrEmailTaken` return those keys from `Error()`.
