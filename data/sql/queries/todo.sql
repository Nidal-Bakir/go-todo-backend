-- name: TodoSoftDeleteTodoLinkedToUser :exec
UPDATE todo
SET deleted_at = NOW()
WHERE id = $1
    AND user_id = $2;


-- name: TodoCreateTodo :one
INSERT INTO
	todo (title, body, status, user_id)
VALUES
	($1, $2, $3, $4)
RETURNING
	*;


-- name: TodoGetTodoLinkedToUser :one
SELECT * FROM todo
WHERE id = $1
    AND user_id = $2
    AND deleted_at IS NULL
LIMIT 1;


-- name: TodoUpdateTodo :one
UPDATE todo
SET
	title = COALESCE(sqlc.narg ('title'), title),
	body = COALESCE(sqlc.narg ('body'), body),
	status = COALESCE(sqlc.narg ('status'), status)
WHERE
	id = $1
	AND user_id = $2
    AND deleted_at IS NULL
RETURNING
	*;

-- name: TodoGetTodosForUser :many
SELECT
    *
FROM todo
WHERE user_id = $1
   AND deleted_at IS NULL
ORDER BY created_at DESC
OFFSET $2
LIMIT $3;
