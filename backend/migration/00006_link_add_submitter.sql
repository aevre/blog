-- +goose Up

ALTER TABLE links
ADD COLUMN submitter_id BIGINT;

-- +goose statementbegin
DO $$
DECLARE
    admin_user_id BIGINT;
BEGIN
    IF EXISTS (SELECT 1 FROM links) THEN
        SELECT id
        INTO admin_user_id
        FROM users
        WHERE role = 'admin'
        ORDER BY id ASC
        LIMIT 1;

        IF admin_user_id IS NULL THEN
            RAISE EXCEPTION 'Cannot migrate: no admin user exists.';
        END IF;

        UPDATE links
        SET submitter_id = admin_user_id;
    END IF;
END;
$$;
-- +goose statementend

ALTER TABLE links
ALTER COLUMN submitter_id SET NOT NULL;

CREATE INDEX links_submitter_id_idx
ON links (submitter_id);

-- +goose Down

DROP INDEX IF EXISTS links_submitter_id_idx;

ALTER TABLE links
DROP COLUMN IF EXISTS submitter_id;
