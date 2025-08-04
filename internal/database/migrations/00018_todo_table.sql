-- +goose Up
CREATE TABLE todo (
    id SERIAL PRIMARY KEY NOT NULL,
    title TEXT NOT NULL CHECK (char_length(title) <= 150),
    body TEXT NOT NULL CHECK (char_length(body) <= 10000),
    status TEXT NOT NULL CHECK (char_length(status) >= 1 AND char_length(status) <= 50),
    created_at TIMESTAMPTZ DEFAULT NOW () NOT NULL,
    updated_at TIMESTAMPTZ DEFAULT NOW () NOT NULL,
    deleted_at TIMESTAMPTZ,
    user_id INTEGER NOT NULL REFERENCES users (id)
);

CREATE TRIGGER update_todo_updated_at_column BEFORE
UPDATE ON todo FOR EACH ROW EXECUTE PROCEDURE trigger_set_updated_at_column ();

-- +goose Down
DROP TABLE todo;
