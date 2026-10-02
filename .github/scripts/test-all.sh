#!/usr/bin/env bash
# Test the root module and every nested module (own go.mod).
set -euo pipefail

export GOWORK=off
go test ./...

nested=(
  mongo
  webauthn
  qr
  db/postgres
  db/mysql
  db/mariadb
  db/sqlite
  db/sqlserver
  db/oracle
)
for d in "${nested[@]}"; do
  if [[ -f "$d/go.mod" ]]; then
    echo "::group::go test $d"
    (cd "$d" && go test ./...)
    echo "::endgroup::"
  fi
done
