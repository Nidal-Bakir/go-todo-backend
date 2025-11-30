-- +goose Up
CREATE TABLE user_integration (
    id SERIAL PRIMARY KEY NOT NULL,
    oauth_integration_id INTEGER NOT NULL UNIQUE REFERENCES oauth_integration(id) ON DELETE CASCADE,
    user_id INTEGER NOT NULL REFERENCES users(id),
    created_at TIMESTAMPTZ DEFAULT NOW() NOT NULL,
    updated_at TIMESTAMPTZ DEFAULT NOW() NOT NULL,
    deleted_at TIMESTAMPTZ
);


CREATE TRIGGER update_user_integration_updated_at_column BEFORE
UPDATE ON user_integration FOR EACH ROW EXECUTE PROCEDURE trigger_set_updated_at_column ();

CREATE VIEW active_user_integration AS
SELECT
    *
FROM
    user_integration
WHERE
    deleted_at IS NULL;

-- +goose Down
DROP VIEW active_user_integration;
DROP TABLE user_integration;
