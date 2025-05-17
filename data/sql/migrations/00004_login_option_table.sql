-- +goose Up
CREATE TABLE login_option (
    id SERIAL PRIMARY KEY NOT NULL,
    -- email, phone, guest, Oauth, etc ...
    login_method VARCHAR(25) NOT NULL CHECK (length (login_method) >= 2),
    access_key VARCHAR(200) UNIQUE NOT NULL CHECK (length (access_key) >= 5),
    hashed_pass VARCHAR(200),
    pass_salt VARCHAR(200),
    verified_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT NOW () NOT NULL,
    updated_at TIMESTAMPTZ DEFAULT NOW () NOT NULL,
    deleted_at TIMESTAMPTZ,
    user_id INTEGER NOT NULL REFERENCES users (id)
);

CREATE TRIGGER update_login_option_updated_at_column BEFORE
UPDATE ON login_option FOR EACH ROW EXECUTE PROCEDURE trigger_set_updated_at_column ();

-- +goose Down
DROP TABLE login_option;
