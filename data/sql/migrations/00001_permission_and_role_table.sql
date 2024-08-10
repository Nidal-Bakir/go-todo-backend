-- +goose Up
CREATE TABLE permission (
    id SERIAL PRIMARY KEY NOT NULL,
    name TEXT UNIQUE NOT NULL CHECK (length(name) >= 1),
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL,
    deleted_at TIMESTAMPTZ
);

CREATE TABLE role (
    id SERIAL PRIMARY KEY NOT NULL,
    name TEXT UNIQUE NOT NULL CHECK (length(name) >= 1),
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL,
    deleted_at TIMESTAMPTZ
);

CREATE TABLE role_permission (
    role_id INTEGER NOT NULL REFERENCES role(id),
    permission_id INTEGER NOT NULL REFERENCES permission(id),
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL,
    PRIMARY KEY (role_id, permission_id)
);

-- +goose Down
DROP TABLE role_permission;

DROP TABLE role;

DROP TABLE permission;