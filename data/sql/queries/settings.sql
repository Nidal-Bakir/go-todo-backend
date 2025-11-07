-- name: SettingsGetByLable :one
SELECT * FROM settings
WHERE label = @label::TEXT
LIMIT 1;

-- name: SettingsDeleteByLable :exec
DELETE FROM settings
WHERE label = @label::TEXT;

-- name: SettingsSetSetting :exec
INSERT INTO settings (
    label,
    value
)
VALUES (
    @label::TEXT,
    sqlc.narg(value)::TEXT
)
ON CONFLICT (label)
DO UPDATE
SET value = EXCLUDED.value;

-- name: SettingsCreateLabel :exec
INSERT INTO settings (
    label
)
VALUES (
    @label::TEXT
)
ON CONFLICT (label)
DO NOTHING;
