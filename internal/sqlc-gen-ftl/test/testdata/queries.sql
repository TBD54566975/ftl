-- name: GetUserByID :one
SELECT id, name, email FROM users WHERE id = $1;

-- name: CreateUser :exec
INSERT INTO users (name, email) VALUES ($1, $2);
