-- +goose Up
CREATE TABLE password_login_identity (
    id SERIAL PRIMARY KEY NOT NULL,
    login_identity_id INTEGER NOT NULL UNIQUE REFERENCES login_identity(id),
    email VARCHAR(255),
    phone VARCHAR(16), -- E.164 format
    hashed_pass VARCHAR(128) NOT NULL,
    pass_salt VARCHAR(64) NOT NULL,
    verified_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT NOW() NOT NULL,
    updated_at TIMESTAMPTZ DEFAULT NOW() NOT NULL,
    deleted_at TIMESTAMPTZ,

    -- Enforce only one of phone/email
    CHECK (
      (email IS NOT NULL AND phone IS NULL)
      OR
      (phone IS NOT NULL AND email IS NULL)
    ),
    -- Enforce E.164 format
    CHECK (
      phone IS NULL OR phone ~ '^\+[1-9]\d{1,14}$'
    )
);

-- Partial unique indexes
CREATE UNIQUE INDEX unique_email_login ON password_login_identity (email) WHERE email IS NOT NULL;
CREATE UNIQUE INDEX unique_phone_login ON password_login_identity (phone) WHERE phone IS NOT NULL;

CREATE TRIGGER update_password_login_identity_updated_at_column BEFORE
UPDATE ON password_login_identity FOR EACH ROW EXECUTE PROCEDURE trigger_set_updated_at_column ();

CREATE VIEW active_password_login_identity AS
SELECT
    *
FROM
    password_login_identity
WHERE
    verified_at IS NOT NULL
    And deleted_at IS NULL;

-- +goose Down
DROP VIEW active_password_login_identity;
DROP TABLE password_login_identity;
