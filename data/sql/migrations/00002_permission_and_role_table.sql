-- +goose Up
CREATE TABLE permission (
    id SERIAL PRIMARY KEY NOT NULL,
    name TEXT UNIQUE NOT NULL CHECK (length(name) >= 1),
    created_at TIMESTAMPTZ DEFAULT NOW() NOT NULL,
    updated_at TIMESTAMPTZ DEFAULT NOW() NOT NULL,
    deleted_at TIMESTAMPTZ
);

CREATE TABLE role (
    id SERIAL PRIMARY KEY NOT NULL,
    name TEXT UNIQUE NOT NULL CHECK (length(name) >= 1),
    created_at TIMESTAMPTZ DEFAULT NOW() NOT NULL,
    updated_at TIMESTAMPTZ DEFAULT NOW() NOT NULL,
    deleted_at TIMESTAMPTZ
);

CREATE TABLE role_permission (
    role_id INTEGER NOT NULL REFERENCES role(id),
    permission_id INTEGER NOT NULL REFERENCES permission(id),
    created_at TIMESTAMPTZ DEFAULT NOW() NOT NULL,
    updated_at TIMESTAMPTZ DEFAULT NOW() NOT NULL,
    PRIMARY KEY (role_id, permission_id)
);

CREATE TRIGGER update_permission_updated_at_column BEFORE
UPDATE ON permission FOR EACH ROW EXECUTE PROCEDURE trigger_set_updated_at_column();

CREATE TRIGGER update_role_updated_at_column BEFORE
UPDATE ON role FOR EACH ROW EXECUTE PROCEDURE trigger_set_updated_at_column();

CREATE TRIGGER update_role_permission_updated_at_column BEFORE
UPDATE ON role_permission FOR EACH ROW EXECUTE PROCEDURE trigger_set_updated_at_column();

-- +goose Down
DROP TABLE role_permission;

DROP TABLE role;

DROP TABLE permission;