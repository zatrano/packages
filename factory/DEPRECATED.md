# DEPRECATED — attribute helpers only (V3)

ORM persistence (`Create` / `CreateMany`) is **removed**. Use `Make` for test attributes;
persist via `packages/db` + sqlc / `database/seeders`.

- `package:enable factory` is rejected (catalog Stability=deprecated).
- Doctor `APP-FAC-001` errors on app imports.
- Prefer not to use this package in new V3 apps.
