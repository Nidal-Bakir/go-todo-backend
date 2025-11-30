-- +goose Up
CREATE TABLE oidc_data (
    id SERIAL PRIMARY KEY NOT NULL,
    provider_name TEXT NOT NULL REFERENCES oauth_provider(name) ON DELETE CASCADE,
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

-- Partial unique indexes
CREATE UNIQUE INDEX unique_email_oidc ON oidc_data (email) WHERE email IS NOT NULL;

CREATE INDEX index_sub_oidc ON oidc_data (sub);

CREATE TRIGGER update_oidc_data_updated_at_column BEFORE
UPDATE ON oidc_data FOR EACH ROW EXECUTE PROCEDURE trigger_set_updated_at_column ();

CREATE VIEW active_oidc_data AS
SELECT
    *
FROM
    oidc_data
WHERE
    deleted_at IS NULL;

-- +goose Down
DROP VIEW active_oidc_data;
DROP TABLE oidc_data;
