-- +goose Up
CREATE TABLE installation (
    id SERIAL PRIMARY KEY NOT NULL,
    installation_id UUID UNIQUE NOT NULL,
    notification_token VARCHAR(200),
    locale VARCHAR(16) NOT NULL CHECK (length(locale) >= 2),
    timezone_Offset_in_minutes INTEGER NOT NULL,
    device_manufacturer VARCHAR(50) NOT NULL,
    device_os VARCHAR(50) NOT NULL,
    device_os_version VARCHAR(50) NOT NULL,
    app_version VARCHAR(50) NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW() NOT NULL,
    updated_at TIMESTAMPTZ DEFAULT NOW() NOT NULL,
    deleted_at TIMESTAMPTZ,
    attach_to INTEGER NULL REFERENCES users(id),
    last_attach_to INTEGER NULL REFERENCES users(id),
);

CREATE TRIGGER update_installation_updated_at_column BEFORE
UPDATE ON installation FOR EACH ROW EXECUTE PROCEDURE trigger_set_updated_at_column();

-- +goose Down
DROP TABLE installation;