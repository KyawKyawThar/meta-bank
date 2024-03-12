-- name: CreateUser :one
INSERT INTO users(username,
                  password,
                  email,
                  full_name, role)
VALUES ($1, $2, $3, $4, $5) RETURNING *;

-- name: GetUser :one
SELECT *
FROM users
WHERE username = $1;

-- name: ListUsers :many
SELECT *
FROM users
Where role = 'admin'
  AND username = $1
ORDER BY username LIMIT $2
OFFSET $3;

-- name: UpdateUser :exec
Update users
SET password = coalesce(sqlc.narg(password), password),
    email    = coalesce(sqlc.narg(email), email),
    full_name=coalesce(sqlc.narg(full_name), fullname)
WHERE username = sqlc.arg(username) RETURNING *;

