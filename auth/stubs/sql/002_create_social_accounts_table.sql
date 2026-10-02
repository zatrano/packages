-- make:auth --social — social_accounts (SQLite-friendly).

CREATE TABLE IF NOT EXISTS social_accounts (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL,
    provider TEXT NOT NULL,
    provider_id TEXT NOT NULL,
    provider_uid TEXT NOT NULL UNIQUE,
    name TEXT,
    email TEXT,
    avatar TEXT,
    access_token TEXT,
    created_at TEXT NOT NULL DEFAULT (datetime('now')),
    updated_at TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE INDEX IF NOT EXISTS social_accounts_user_id_idx ON social_accounts (user_id);
