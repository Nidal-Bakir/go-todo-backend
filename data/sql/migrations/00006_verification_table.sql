-- +goose Up
CREATE TABLE verification (
    id SERIAL PRIMARY KEY NOT NULL,
    code VARCHAR(10) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    deleted_at TIMESTAMPTZ,
    verifying INTEGER REFERENCES login_option(id),
    using_session_id INTEGER REFERENCES session(id)
);

CREATE INDEX verification_code_index ON verification(code);

-- +goose Down
DROP TABLE verification;