
-- name: GetUser :one
SELECT * FROM users
WHERE id = $1 LIMIT 1;

-- name: ListUsers :many
SELECT * FROM users
ORDER BY name;

-- name: CreateUser :one
INSERT INTO users (
  name
) VALUES (
  $1
)
RETURNING *;

-- name: UpdateUser :exec
UPDATE users
set name = $1
WHERE id = $2
RETURNING *;

-- name: DeleteUser :exec
DELETE FROM users
WHERE id = $1;