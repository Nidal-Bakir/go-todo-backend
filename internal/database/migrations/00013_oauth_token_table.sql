-- +goose Up
CREATE TABLE oauth_token (
  id SERIAL PRIMARY KEY NOT NULL,
  oauth_integration_id INTEGER NOT NULL UNIQUE REFERENCES oauth_integration(id) ON DELETE CASCADE,
  access_token VARCHAR(2048),
  refresh_token VARCHAR(2048),
  token_type VARCHAR(50) NOT NULL DEFAULT 'Bearer',
  expires_at TIMESTAMP,
  issued_at TIMESTAMP NOT NULL DEFAULT NOW(),
  created_at TIMESTAMPTZ DEFAULT NOW() NOT NULL,
  updated_at TIMESTAMPTZ DEFAULT NOW() NOT NULL,
  deleted_at TIMESTAMPTZ
);

CREATE TRIGGER update_oauth_token_updated_at_column BEFORE
UPDATE ON oauth_token FOR EACH ROW EXECUTE PROCEDURE trigger_set_updated_at_column();

CREATE VIEW active_oauth_token AS
SELECT
    *
FROM
    oauth_token
WHERE
    deleted_at IS NULL;

-- +goose Down
DROP VIEW active_oauth_token;
DROP TABLE oauth_token;
