-- +goose Up
CREATE TABLE seeder_version (
  id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  version INTEGER UNIQUE NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- +goose Down
DROP TABLE seeder_version;
