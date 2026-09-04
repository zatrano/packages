#!/usr/bin/env bash
# Test the root module and every nested module (own go.mod).
set -euo pipefail

go test ./...

nested=(
  mongo
  webauthn
  qr
  database/driver/sqlite
  database/driver/mysql
  database/driver/pgsql
  database/driver/mssql
  database/driver/oracle
  database/driver/mongo
)
for d in "${nested[@]}"; do
  if [[ -f "$d/go.mod" ]]; then
    echo "::group::go test $d"
    (cd "$d" && go test ./...)
    echo "::endgroup::"
  fi
done
