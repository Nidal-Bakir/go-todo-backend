-- +goose Up
CREATE TABLE user_login_option (
    id SERIAL PRIMARY KEY NOT NULL,
    login_method TEXT NOT NULL,
    access_key TEXT UNIQUE NOT NULL,
    pass VARCHAR(120),
    pass_salt VARCHAR(25),
    verified_at TIMESTAMPTZ,
    deleted_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL,
    user_id INTEGER REFERENCES users(id)
);

-- +goose Down
DROP TABLE user_login_option;