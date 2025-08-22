-- +goose Up
CREATE TABLE oauth_integration (
    id SERIAL PRIMARY KEY NOT NULL,
    oauth_connection_id INTEGER NOT NULL REFERENCES oauth_connection(id),
    integration_type TEXT NOT NULL,
    CONSTRAINT chk_oauth_integration_type
        CHECK (integration_type IN ('user', 'system')),
    created_at TIMESTAMPTZ DEFAULT NOW() NOT NULL,
    updated_at TIMESTAMPTZ DEFAULT NOW() NOT NULL,
    deleted_at TIMESTAMPTZ
);


CREATE TRIGGER update_oauth_integration_updated_at_column BEFORE
UPDATE ON oauth_integration FOR EACH ROW EXECUTE PROCEDURE trigger_set_updated_at_column();

CREATE VIEW active_oauth_integration AS
SELECT
    *
FROM
    oauth_integration
WHERE
    deleted_at IS NULL;

-- +goose Down
DROP VIEW active_oauth_integration;
DROP TABLE oauth_integration;
