-- +goose Up
CREATE TABLE oauth_connection (
  id SERIAL PRIMARY KEY NOT NULL,
  provider_id INTEGER NOT NULL REFERENCES oauth_provider(id),
  scopes TEXT NOT NULL,
  created_at TIMESTAMP NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
  deleted_at TIMESTAMPTZ,
  UNIQUE (provider_id, scopes)
);


CREATE TRIGGER update_oauth_connection_updated_at_column BEFORE
UPDATE ON oauth_connection FOR EACH ROW EXECUTE PROCEDURE trigger_set_updated_at_column ();

CREATE VIEW active_oauth_connection AS
SELECT
    *
FROM
    oauth_connection
WHERE
    deleted_at IS NULL;

-- +goose Down
DROP VIEW active_oauth_connection;
DROP TABLE oauth_connection;
