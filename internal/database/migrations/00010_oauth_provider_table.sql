-- +goose Up
CREATE TABLE oauth_provider (
  name VARCHAR(50) PRIMARY KEY NOT NULL, -- e.g., "google", "github"
  is_oidc_capable BOOLEAN NOT NULL DEFAULT false,
  created_at TIMESTAMP NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
  deleted_at TIMESTAMPTZ
);


CREATE TRIGGER update_oauth_provider_updated_at_column BEFORE
UPDATE ON oauth_provider FOR EACH ROW EXECUTE PROCEDURE trigger_set_updated_at_column ();

CREATE VIEW active_oauth_provider AS
SELECT
    *
FROM
    oauth_provider
WHERE
    deleted_at IS NULL;

-- +goose Down
DROP VIEW active_oauth_provider;
DROP TABLE oauth_provider;
