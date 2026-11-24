-- +goose Up

ALTER TABLE systems
ADD COLUMN bootstrap_token_hash VARCHAR;

-- +goose Down

ALTER TABLE systems
DROP COLUMN bootstrap_token_hash;
