CREATE TABLE IF NOT EXISTS users (
    id          SERIAL      PRIMARY KEY,
    provider    TEXT        NOT NULL,
    provider_id TEXT        NOT NULL,
    email       TEXT        NOT NULL,
    name        TEXT        NOT NULL,
    avatar_url  TEXT        NOT NULL DEFAULT '',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_users_provider ON users(provider, provider_id);

CREATE TABLE IF NOT EXISTS user_sessions (
    token       TEXT        PRIMARY KEY,
    user_id     INTEGER     NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    expires_at  TIMESTAMPTZ NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_user_sessions_user ON user_sessions(user_id);
