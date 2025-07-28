-- +goose Up
CREATE TABLE system_integration (
    id SERIAL PRIMARY KEY NOT NULL,
    oauth_integration_id INTEGER NOT NULL UNIQUE REFERENCES oauth_integration(id),
    created_at TIMESTAMPTZ DEFAULT NOW() NOT NULL,
    updated_at TIMESTAMPTZ DEFAULT NOW() NOT NULL,
    deleted_at TIMESTAMPTZ
);


CREATE TRIGGER update_system_integration_updated_at_column BEFORE
UPDATE ON system_integration FOR EACH ROW EXECUTE PROCEDURE trigger_set_updated_at_column ();

CREATE VIEW active_system_integration AS
SELECT
    *
FROM
system_integration
WHERE
    deleted_at IS NULL;

-- +goose Down
DROP VIEW active_system_integration;
DROP TABLE system_integration;
