-- +goose Up
CREATE TABLE verification (
    id SERIAL PRIMARY KEY NOT NULL,
    code VARCHAR(10) NOT NULL CHECK (length(code) >= 2),
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    deleted_at TIMESTAMPTZ,
    -- the intent is for checking why we sent the verification code (creating account, forget password, etc..)
    -- you need to check it to make sure that tha operation that you are about to do is what the code was sent for.
    intent TEXT NOT NULL CHECK (length(intent) >= 1),
    verifying INTEGER NOT NULL REFERENCES login_option(id),
    using_session_id INTEGER REFERENCES session(id)
);

-- +goose Down
DROP TABLE verification;