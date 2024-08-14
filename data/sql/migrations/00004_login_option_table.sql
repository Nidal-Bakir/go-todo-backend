-- +goose Up
CREATE TABLE login_option (
    id SERIAL PRIMARY KEY NOT NULL,
    -- email, phone, Oauth, etc ...
    login_method TEXT NOT NULL,
    access_key TEXT UNIQUE NOT NULL CHECK (length(access_key) >= 5),
    pass VARCHAR(120),
    pass_salt VARCHAR(25),
    verified_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT NOW() NOT NULL,
    updated_at TIMESTAMPTZ DEFAULT NOW() NOT NULL,
    deleted_at TIMESTAMPTZ,
    user_id INTEGER NOT NULL REFERENCES users(id)
);

CREATE TRIGGER update_login_option_updated_at_column BEFORE
UPDATE ON login_option FOR EACH ROW EXECUTE PROCEDURE trigger_set_updated_at_column();

-- +goose Down
DROP TABLE login_option;