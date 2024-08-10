-- +goose Up
CREATE TABLE session (
    id SERIAL PRIMARY KEY NOT NULL,
    token TEXT UNIQUE NOT NULL CHECK (length(token) >= 50),
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    deleted_at TIMESTAMPTZ,
    originated_from INTEGER NOT NULL REFERENCES login_option(id),
    installation_id INTEGER REFERENCES installation(id)
);

CREATE INDEX session_token_index ON session(token);

-- +goose Down
DROP TABLE session;