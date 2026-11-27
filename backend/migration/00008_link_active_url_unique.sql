-- +goose Up

-- Keep URL uniqueness for active links while allowing reuse after soft deletion.
DROP INDEX links_url_key;
CREATE UNIQUE INDEX links_active_url_key ON links (url) WHERE deleted_at IS NULL;

-- +goose Down

-- Refuse rollback if multiple rows now share a URL, rather than losing history.
CREATE UNIQUE INDEX links_url_key ON links (url);
DROP INDEX links_active_url_key;
