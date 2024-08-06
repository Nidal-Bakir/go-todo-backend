-- +goose Up
CREATE TABLE users (
    id SERIAL PRIMARY KEY NOT NULL,
    username VARCHAR(25) UNIQUE NOT NULL,
    profile_image TEXT,
    first_name VARCHAR(120) NOT NULL,
    last_name VARCHAR(120) NOT NULL,
    deleted_at TIMESTAMPTZ,
    blocked_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL
);

-- +goose Down
DROP TABLE users;