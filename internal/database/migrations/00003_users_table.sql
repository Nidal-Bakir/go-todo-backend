-- +goose Up
CREATE TABLE users (
    id SERIAL PRIMARY KEY NOT NULL,
    username VARCHAR(50) UNIQUE NOT NULL CHECK (length(username) >= 3),
    profile_image TEXT,
    first_name VARCHAR(250) NOT NULL,
    middle_name VARCHAR(250),
    last_name VARCHAR(250),
    created_at TIMESTAMPTZ DEFAULT NOW() NOT NULL,
    updated_at TIMESTAMPTZ DEFAULT NOW() NOT NULL,
    blocked_at TIMESTAMPTZ,
    blocked_until TIMESTAMPTZ,
    deleted_at TIMESTAMPTZ,
    role_id INTEGER REFERENCES role(id)
);

CREATE TRIGGER update_users_updated_at_column BEFORE
UPDATE ON users FOR EACH ROW EXECUTE PROCEDURE trigger_set_updated_at_column();


CREATE VIEW not_deleted_users AS
SELECT
    *
FROM
    users
WHERE
    deleted_at IS NULL;


-- +goose Down
DROP VIEW not_deleted_users;
DROP TABLE users;
