-- +goose Up
CREATE TABLE oauth_connection (
  id SERIAL PRIMARY KEY NOT NULL,
  provider_id INTEGER NOT NULL REFERENCES oauth_provider(id),
  scopes TEXT[] NOT NULL,
  created_at TIMESTAMP NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
  deleted_at TIMESTAMPTZ,
  UNIQUE (provider_id, scopes)
);

-- +goose statementbegin
CREATE OR REPLACE FUNCTION oauth_connection_sort_and_dedupe_scopes_array_fn()
RETURNS TRIGGER AS $$
BEGIN
  NEW.scopes := (
    SELECT array_agg(DISTINCT s ORDER BY s)
    FROM unnest(NEW.scopes) s
  );
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;
-- +goose statementend

CREATE TRIGGER oauth_connection_sort_and_dedupe_scopes_trigger BEFORE
INSERT OR UPDATE ON oauth_connection FOR EACH ROW
EXECUTE FUNCTION oauth_connection_sort_and_dedupe_scopes_array_fn();

CREATE TRIGGER update_oauth_connection_updated_at_column BEFORE
UPDATE ON oauth_connection FOR EACH ROW
EXECUTE PROCEDURE trigger_set_updated_at_column();

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
DROP FUNCTION oauth_connection_sort_and_dedupe_scopes_array_fn;

