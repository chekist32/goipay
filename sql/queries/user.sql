-- name: CreateUser :one
INSERT INTO users DEFAULT VALUES
RETURNING *;

-- name: UserExistsById :one
SELECT EXISTS (
    SELECT 1
    FROM users
    WHERE id = $1
) AS user_exists;