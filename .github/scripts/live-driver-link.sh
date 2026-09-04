#!/usr/bin/env bash
# CI-only: link mysql/pgsql for TestLiveDriverSmoke.
# Do not commit this import into the published module — consumers' go mod tidy
# would then require nested driver modules at a fake v0.0.0 revision.
set -euo pipefail

go mod edit -replace github.com/zatrano/packages/database/driver/mysql=./database/driver/mysql
go mod edit -replace github.com/zatrano/packages/database/driver/pgsql=./database/driver/pgsql

cat > database/zz_ci_live_drivers_test.go <<'EOF'
package database_test

import (
	_ "github.com/zatrano/packages/database/driver/mysql"
	_ "github.com/zatrano/packages/database/driver/pgsql"
)
EOF
