-- +goose Up
CREATE TABLE permission (
    name VARCHAR(100) PRIMARY KEY NOT NULL CHECK (char_length(name) >= 1),
    created_at TIMESTAMPTZ DEFAULT NOW() NOT NULL,
    updated_at TIMESTAMPTZ DEFAULT NOW() NOT NULL,
    deleted_at TIMESTAMPTZ
);

CREATE TABLE role (
    name VARCHAR(100) PRIMARY KEY NOT NULL CHECK (char_length(name) >= 1),
    created_at TIMESTAMPTZ DEFAULT NOW() NOT NULL,
    updated_at TIMESTAMPTZ DEFAULT NOW() NOT NULL,
    deleted_at TIMESTAMPTZ
);

CREATE TABLE role_permission (
    role_name VARCHAR(100) NOT NULL REFERENCES role(name),
    permission_name VARCHAR(100) NOT NULL REFERENCES permission(name),
    created_at TIMESTAMPTZ DEFAULT NOW() NOT NULL,
    updated_at TIMESTAMPTZ DEFAULT NOW() NOT NULL,
    PRIMARY KEY (role_name, permission_name)
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
