# Changelog

All notable changes to `github.com/zatrano/packages` are documented in this file.

## Unreleased

### Added

- First-party addon packages split from the ZATRANO framework module (Stage B).
- Consumers blank-import a package (for example `github.com/zatrano/packages/billing`) and enable it with `bootstrap.WithAddons(...)`.

### Notes

- Go module path is `github.com/zatrano/packages` (no `/v2` suffix). Do **not** tag `v2.0.0-alpha`: the Go toolchain would require `github.com/zatrano/packages/v2`. Use `v0.x` / `v1.x` tags (for example `v1.0.0-alpha`) until a real v1/v2 module decision.
- Local development: `replace github.com/zatrano/framework => ../framework`.
