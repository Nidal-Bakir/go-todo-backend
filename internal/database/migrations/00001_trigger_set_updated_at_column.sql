-- +goose Up
-- +goose statementbegin
CREATE OR REPLACE FUNCTION trigger_set_updated_at_column() RETURNS TRIGGER AS $$ BEGIN NEW.updated_at = clock_timestamp();

RETURN NEW;

END;

$$ language 'plpgsql';

-- +goose statementend
-- +goose Down
DROP FUNCTION trigger_set_updated_at_column;