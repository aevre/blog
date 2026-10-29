-- +goose Up
ALTER TABLE account_refresh_tokens ADD COLUMN expires_at timestamptz;
-- Existing opaque tokens cannot be used by the JWT refresh flow.
UPDATE account_sessions SET revoked_at = COALESCE(revoked_at, now());
UPDATE account_refresh_tokens SET expires_at = now();
ALTER TABLE account_refresh_tokens ALTER COLUMN expires_at SET NOT NULL;
CREATE INDEX account_refresh_tokens_expires_at ON account_refresh_tokens(expires_at, hash);
CREATE INDEX account_sessions_expires_at ON account_sessions(expires_at, id);

-- +goose Down
DROP INDEX account_sessions_expires_at;
DROP INDEX account_refresh_tokens_expires_at;
ALTER TABLE account_refresh_tokens DROP COLUMN expires_at;
