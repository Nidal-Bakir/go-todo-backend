-- +goose Up
CREATE TABLE session (
    id SERIAL PRIMARY KEY NOT NULL,
    token TEXT UNIQUE NOT NULL CHECK (length (token) >= 50),
    created_at TIMESTAMPTZ DEFAULT NOW () NOT NULL,
    updated_at TIMESTAMPTZ DEFAULT NOW () NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    deleted_at TIMESTAMPTZ,
    originated_from INTEGER NOT NULL REFERENCES login_option (id)
);

CREATE TRIGGER update_session_updated_at_column BEFORE
UPDATE ON session FOR EACH ROW EXECUTE PROCEDURE trigger_set_updated_at_column ();

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
DROP TABLE session;
