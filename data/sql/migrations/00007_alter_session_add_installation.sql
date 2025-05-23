-- +goose Up
ALTER TABLE session ADD used_installation INTEGER NOT NULL REFERENCES installation (id);

-- +goose Down
ALTER TABLE session
DROP COLUMN used_installation;
