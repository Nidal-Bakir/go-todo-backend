-- +goose Up
CREATE TABLE session (
    id SERIAL PRIMARY KEY NOT NULL,
    token VARCHAR(2048) UNIQUE NOT NULL CHECK (char_length(token) >= 50),
    -- the used ip address to create this session
    ip_address INET NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW () NOT NULL,
    updated_at TIMESTAMPTZ DEFAULT NOW () NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    deleted_at TIMESTAMPTZ,
    originated_from INTEGER NOT NULL REFERENCES login_identity(id) ON DELETE CASCADE
);

CREATE TRIGGER update_session_updated_at_column BEFORE
UPDATE ON session FOR EACH ROW EXECUTE PROCEDURE trigger_set_updated_at_column ();

-- +goose Down
DROP TABLE session;
