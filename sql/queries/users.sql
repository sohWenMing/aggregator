-- name: CreateUser :one

INSERT INTO users (id, created_at, updated_at, name)
VALUES($1, $2, $3, $4)
RETURNING *;

-- name: RetrieveUser :one
SELECT * from users
WHERE name = $1
LIMIT 1;

-- name: ResetUsers :exec
DELETE from users;