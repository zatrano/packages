# ZATRANO packages

Official first-party **addon** packages for [ZATRANO](https://github.com/zatrano/framework): optional services and libraries that do not belong in the framework kernel.

Intelligence packages (`ai`, `rag`, `agent`) stay in the framework repository.

## Module

```
github.com/zatrano/packages
```

Import a package:

```go
import "github.com/zatrano/packages/billing"
```

Nested modules (own `go.mod`): `mongo`, `webauthn`, `qr`, and `database/driver/{mysql,pgsql,mssql,oracle,mongo}`.

## Enable a service addon

In the **application** (not the framework):

```go
import _ "github.com/zatrano/packages/billing"

app := bootstrap.App(bootstrap.WithAddons("billing"))
```

Self-registration uses `init()` → `bootstrap/addons.Register`. The framework no longer imports these packages.

## Local development

Clone next to the framework repo as `framework`:

```
replace github.com/zatrano/framework => ../framework
```

If the sibling directory is named `ZATRANO`, point the replace there instead.
