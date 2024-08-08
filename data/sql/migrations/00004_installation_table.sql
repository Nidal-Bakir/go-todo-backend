-- +goose Up
CREATE TABLE installation (
    id SERIAL PRIMARY KEY NOT NULL,
    installation_id TEXT UNIQUE NOT NULL,
    notification_token TEXT,
    locale TEXT,
    device_manufacturer TEXT,
    device_os TEXT,
    device_os_version TEXT,
    app_version TEXT,
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL,
    deleted_at TIMESTAMPTZ
);


-- +goose Down
DROP TABLE installation;