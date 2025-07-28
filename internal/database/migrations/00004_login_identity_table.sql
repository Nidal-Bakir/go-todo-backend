-- +goose Up
CREATE TABLE login_identity (
    id SERIAL PRIMARY KEY NOT NULL,
    user_id INTEGER NOT NULL REFERENCES users(id),
    identity_type TEXT NOT NULL,
    CONSTRAINT chk_login_identity_type
        CHECK (identity_type IN ('email', 'phone', 'oidc', 'guest')),
    is_primary BOOLEAN DEFAULT FALSE,
    last_used_at TIMESTAMPTZ DEFAULT NOW() NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW() NOT NULL,
    updated_at TIMESTAMPTZ DEFAULT NOW() NOT NULL,
    deleted_at TIMESTAMPTZ
);

CREATE TRIGGER update_login_identity_updated_at_column BEFORE
UPDATE ON login_identity FOR EACH ROW EXECUTE PROCEDURE trigger_set_updated_at_column ();

CREATE VIEW active_login_identity AS
SELECT
    *
FROM
    login_identity
WHERE
    deleted_at IS NULL;

-- +goose Down
DROP VIEW active_login_identity;
DROP TABLE login_identity;
