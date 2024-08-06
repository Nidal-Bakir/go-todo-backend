-- +goose Up
CREATE TABLE session (
    id SERIAL PRIMARY KEY NOT NULL,
    token TEXT UNIQUE NOT NULL,
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    originated_from INTEGER REFERENCES user_login_option(id),
    installation_id INTEGER REFERENCES installation(id),
    
);

-- +goose Down
DROP TABLE session;