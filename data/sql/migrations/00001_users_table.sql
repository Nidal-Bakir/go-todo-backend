-- +goose Up
CREATE TABLE users (
    id SERIAL PRIMARY KEY NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL,
    username VARCHAR(25) UNIQUE NOT NULL,
    email VARCHAR(25) NULL,
    pass_salt VARCHAR(25) NOT NULL,
    pass VARCHAR(120) NOT NULL,
    first_name VARCHAR(120) NOT NULL,
    last_name VARCHAR(120) NOT NULL,
    verified_at TIMESTAMP WITH TIME ZONE NULL
);

-- +goose Down
DROP TABLE users;