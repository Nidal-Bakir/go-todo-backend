-- +goose Up
CREATE TABLE otp (
    id SERIAL PRIMARY KEY NOT NULL,
    code VARCHAR(10) NOT NULL CHECK (length(code) >= 2),
    hit_count SMALLINT NOT NULL CHECK (hit_count >= 0),
    created_at TIMESTAMPTZ DEFAULT NOW() NOT NULL,
    updated_at TIMESTAMPTZ DEFAULT NOW() NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    deleted_at TIMESTAMPTZ,
    -- the intent is for checking why we sent the otp code (creating account, forget password, etc..)
    -- you need to check it to make sure that tha operation that you are about to do is what the code was sent for.
    intent VARCHAR(50) NOT NULL CHECK (length(intent) >= 1),
    otp_for INTEGER NOT NULL REFERENCES login_option(id),
    using_session_id INTEGER NOT NULL REFERENCES session(id)
);

CREATE TRIGGER update_otp_updated_at_column BEFORE
UPDATE ON otp FOR EACH ROW EXECUTE PROCEDURE trigger_set_updated_at_column();

-- +goose Down
DROP TABLE otp;