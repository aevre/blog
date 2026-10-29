-- +goose Up
CREATE TABLE account_sessions (
    id uuid PRIMARY KEY,
    user_id bigint NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    expires_at timestamptz NOT NULL,
    revoked_at timestamptz
);
CREATE TABLE account_refresh_tokens (
    hash text PRIMARY KEY,
    session_id uuid NOT NULL REFERENCES account_sessions(id) ON DELETE CASCADE,
    used_at timestamptz
);
CREATE INDEX account_refresh_tokens_session_id ON account_refresh_tokens(session_id);

-- +goose Down
DROP TABLE account_refresh_tokens;
DROP TABLE account_sessions;
