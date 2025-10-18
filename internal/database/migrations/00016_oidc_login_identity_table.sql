-- +goose Up
CREATE TABLE oidc_login_identity (
    id SERIAL PRIMARY KEY NOT NULL,
    login_identity_id INTEGER NOT NULL UNIQUE REFERENCES login_identity(id),
    oidc_data_id INTEGER NOT NULL UNIQUE REFERENCES oidc_data(id),
    created_at TIMESTAMPTZ DEFAULT NOW() NOT NULL,
    updated_at TIMESTAMPTZ DEFAULT NOW() NOT NULL,
    deleted_at TIMESTAMPTZ
);

CREATE TRIGGER update_oidc_login_identity_updated_at_column BEFORE
UPDATE ON oidc_login_identity FOR EACH ROW EXECUTE PROCEDURE trigger_set_updated_at_column();

CREATE VIEW active_oidc_login_identity AS
SELECT
    *
FROM
    oidc_login_identity
WHERE
    deleted_at IS NULL;

-- +goose Down
DROP VIEW active_oidc_login_identity;
DROP TABLE oidc_login_identity;
