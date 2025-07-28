-- +goose Up
ALTER TABLE session ADD used_installation INTEGER NOT NULL REFERENCES installation(id);

CREATE VIEW active_session AS
SELECT
    *
FROM
    session
WHERE
    expires_at > NOW ()
    AND deleted_at IS NULL;

-- +goose Down
DROP VIEW active_session;
ALTER TABLE session
DROP COLUMN used_installation;
