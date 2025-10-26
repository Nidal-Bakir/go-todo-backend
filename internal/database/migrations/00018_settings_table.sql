-- +goose Up
CREATE TABLE settings (
    label VARCHAR(100) PRIMARY KEY NOT NULL CHECK (char_length(label) >= 1),
    value TEXT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW() NOT NULL,
    updated_at TIMESTAMPTZ DEFAULT NOW() NOT NULL
);

CREATE TRIGGER update_settings_updated_at_column BEFORE
UPDATE ON settings FOR EACH ROW EXECUTE PROCEDURE trigger_set_updated_at_column ();

-- +goose Down
DROP TABLE settings;
