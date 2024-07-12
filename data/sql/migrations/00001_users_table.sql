-- +goose Up
CREATE TABLE users (
    id SERIAL PRIMARY KEY NOT NULL,
    username VARCHAR(25) UNIQUE NOT NULL,
    email VARCHAR(25) NULL,
    pass_salt VARCHAR(25) NOT NULL,
    pass VARCHAR(120) NOT NULL,
    firstName VARCHAR(120) NOT NULL,
    lastName VARCHAR(120) NOT NULL
);

-- +goose Down
DROP TABLE users;