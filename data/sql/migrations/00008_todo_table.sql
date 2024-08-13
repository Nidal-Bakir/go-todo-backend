-- +goose Up
CREATE TABLE todo (
    id SERIAL PRIMARY KEY NOT NULL,
    title TEXT NOT NULL CHECK (length(title) >= 1),
    body TEXT,
    status TEXT NOT NULL CHECK (length(status) >= 1),
    created_at TIMESTAMPTZ DEFAULT NOW() NOT NULL,
    updated_at TIMESTAMPTZ DEFAULT NOW() NOT NULL,
    deleted_at TIMESTAMPTZ,
    user_id INTEGER NOT NULL REFERENCES users(id)
);

CREATE TRIGGER update_todo_updated_at_column BEFORE
UPDATE ON todo FOR EACH ROW EXECUTE PROCEDURE trigger_set_updated_at_column();

-- +goose Down
DROP TABLE todo;