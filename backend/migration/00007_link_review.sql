-- +goose Up

ALTER TABLE links
ADD COLUMN review_status VARCHAR DEFAULT 'pending',
ADD COLUMN reviewed_at TIMESTAMPTZ,
ADD COLUMN rejection_reason VARCHAR(1000),
ADD COLUMN version BIGINT NOT NULL DEFAULT 1;

UPDATE links
SET review_status = 'pending'
WHERE review_status IS NULL;

ALTER TABLE links
ALTER COLUMN review_status SET NOT NULL;

CREATE INDEX links_review_queue_idx
ON links (review_status, created_at)
WHERE deleted_at IS NULL;

CREATE INDEX links_submitter_active_idx
ON links (submitter_id, id)
WHERE deleted_at IS NULL;

ALTER TABLE links
DROP COLUMN IF EXISTS enabled;

-- +goose Down

ALTER TABLE links
ADD COLUMN enabled BOOLEAN NOT NULL DEFAULT FALSE;

UPDATE links
SET enabled = (review_status = 'approved');

DROP INDEX IF EXISTS links_submitter_active_idx;
DROP INDEX IF EXISTS links_review_queue_idx;

ALTER TABLE links
DROP COLUMN IF EXISTS review_status,
DROP COLUMN IF EXISTS reviewed_at,
DROP COLUMN IF EXISTS rejection_reason,
DROP COLUMN IF EXISTS version;
