-- +goose Up
CREATE TABLE login_option (
    id SERIAL PRIMARY KEY NOT NULL,
    login_method TEXT NOT NULL,
    access_key TEXT UNIQUE NOT NULL,
    pass VARCHAR(120),
    pass_salt VARCHAR(25),
    verified_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL,
    deleted_at TIMESTAMPTZ,
    user_id INTEGER REFERENCES users(id)
);

CREATE INDEX login_option_access_key_index ON login_option(access_key);

-- +goose Down
DROP TABLE login_option;