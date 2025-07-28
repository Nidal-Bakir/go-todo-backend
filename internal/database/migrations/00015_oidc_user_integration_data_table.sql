-- +goose Up
CREATE TABLE oidc_user_integration_data (
    id SERIAL PRIMARY KEY NOT NULL,
    user_integration_id INTEGER NOT NULL UNIQUE REFERENCES user_integration(id),
    sub TEXT NOT NULL,
    email VARCHAR(255),
    iss TEXT NOT NULL,
    aud TEXT NOT NULL,
    given_name VARCHAR(250) DEFAULT '',
    family_name VARCHAR(250) DEFAULT '',
    name VARCHAR(250) DEFAULT '',
    picture TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW() NOT NULL,
    updated_at TIMESTAMPTZ DEFAULT NOW() NOT NULL,
    deleted_at TIMESTAMPTZ
);

CREATE TRIGGER update_oidc_user_integration_data_updated_at_column BEFORE
UPDATE ON oidc_user_integration_data FOR EACH ROW EXECUTE PROCEDURE trigger_set_updated_at_column ();

CREATE VIEW active_oidc_user_integration_data AS
SELECT
    *
FROM
    oidc_user_integration_data
WHERE
    deleted_at IS NULL;

-- +goose Down
DROP VIEW active_oidc_user_integration_data;
DROP TABLE oidc_user_integration_data;
