-- +goose Up
CREATE TABLE todo (
    id SERIAL PRIMARY KEY NOT NULL,
    title TEXT NOT NULL,
    desc TEXT,
    status TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL,
    deleted_at TIMESTAMPTZ,
    user_id INTEGER NOT NULL REFERENCES users(id)
);

-- +goose Down
DROP TABLE todo;