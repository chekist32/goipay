-- name: CreateUser :one
INSERT INTO users DEFAULT VALUES
RETURNING *;

-- name: CreateUserWithId :one
INSERT INTO users(id) VALUES($1)
RETURNING *;

-- name: UserExistsById :one
SELECT EXISTS (
    SELECT 1
    FROM users
    WHERE id = $1
) AS user_exists;