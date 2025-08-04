-- +goose Up
CREATE TABLE guest_login_identity (
    id SERIAL PRIMARY KEY NOT NULL,
    login_identity_id INTEGER NOT NULL UNIQUE REFERENCES login_identity(id),
    device_id TEXT NOT NULL UNIQUE CHECK (char_length(device_id) > 0 AND char_length(device_id) <= 2048),
    created_at TIMESTAMPTZ DEFAULT NOW() NOT NULL,
    updated_at TIMESTAMPTZ DEFAULT NOW() NOT NULL,
    deleted_at TIMESTAMPTZ
);
CREATE TRIGGER update_guest_login_identity_updated_at_column BEFORE
UPDATE ON guest_login_identity FOR EACH ROW EXECUTE PROCEDURE trigger_set_updated_at_column ();

CREATE VIEW active_guest_login_identity AS
SELECT
    *
FROM
    guest_login_identity
WHERE
    deleted_at IS NULL;

-- +goose Down
DROP VIEW active_guest_login_identity;
DROP TABLE guest_login_identity;

