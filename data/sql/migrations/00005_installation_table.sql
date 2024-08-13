-- +goose Up
CREATE TABLE installation (
    id SERIAL PRIMARY KEY NOT NULL,
    installation_id UUID UNIQUE NOT NULL,
    notification_token TEXT,
    locale TEXT NOT NULL,
    device_manufacturer TEXT NOT NULL,
    device_os TEXT NOT NULL,
    device_os_version TEXT NOT NULL,
    app_version TEXT NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW() NOT NULL,
    updated_at TIMESTAMPTZ DEFAULT NOW() NOT NULL,
    deleted_at TIMESTAMPTZ
);

CREATE TRIGGER update_installation_updated_at_column BEFORE
UPDATE ON installation FOR EACH ROW EXECUTE PROCEDURE trigger_set_updated_at_column();

-- +goose Down
DROP TABLE installation;