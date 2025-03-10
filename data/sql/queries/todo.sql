-- name: TodoSoftDeleteTodo :exec
UPDATE todo
SET deleted_at = NOW()
WHERE id = $1;