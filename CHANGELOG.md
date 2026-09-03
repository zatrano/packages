# Changelog

## 0.1.0 - 2026-09-03

### Added

- First-party addon packages split from `github.com/zatrano/framework` with git history preserved (`git filter-repo`).
- Import path `github.com/zatrano/packages/<name>`.
- `mongo` and `webauthn` self-register via `init()` (no parent-module cycle).
