module github.com/zatrano/packages

go 1.25.0

require (
	github.com/redis/go-redis/v9 v9.11.0
	github.com/zatrano/framework v0.0.0
	github.com/zatrano/packages/database/driver/sqlite v1.0.0
	golang.org/x/crypto v0.55.0
	golang.org/x/sys v0.47.0
	modernc.org/sqlite v1.38.2
)

require (
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	github.com/dustin/go-humanize v1.0.1 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/ncruces/go-strftime v0.1.9 // indirect
	github.com/remyoudompheng/bigfft v0.0.0-20230129092748-24d4a6f8daec // indirect
	golang.org/x/exp v0.0.0-20250620022241-b7579e27df2b // indirect
	golang.org/x/tools v0.48.0 // indirect
	modernc.org/libc v1.66.3 // indirect
	modernc.org/mathutil v1.7.1 // indirect
	modernc.org/memory v1.11.0 // indirect
)

replace github.com/zatrano/framework => ../framework

replace github.com/zatrano/packages/database/driver/sqlite => ./database/driver/sqlite

replace github.com/zatrano/packages/database/driver/mysql => ./database/driver/mysql

replace github.com/zatrano/packages/database/driver/pgsql => ./database/driver/pgsql

replace github.com/zatrano/packages/mongo => ./mongo

replace github.com/zatrano/packages/webauthn => ./webauthn
