-- +goose Up
CREATE TABLE users (
    id SERIAL PRIMARY KEY NOT NULL,
    username VARCHAR(25) UNIQUE NOT NULL CHECK (length(username) >= 3),
    profile_image TEXT,
    first_name VARCHAR(120) NOT NULL,
    last_name VARCHAR(120),
    blocked_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL,
    deleted_at TIMESTAMPTZ,
    role_id INTEGER REFERENCES role(id)
);

CREATE INDEX users_username_index ON users(username);

-- +goose Down
DROP TABLE users;