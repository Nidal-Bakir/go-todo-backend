-- name: SettingsGetByLable :one
SELECT * FROM settings
WHERE label = @label::TEXT
LIMIT 1;


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