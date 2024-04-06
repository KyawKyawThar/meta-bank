-- name: CreateUser :one
INSERT INTO users(username,
                  password,
                  email,
                  full_name, role, is_active)
VALUES ($1, $2, $3, $4, $5, $6) RETURNING *;

-- AND (is_active = $2 OR $2 IS NULL): This checks if is_active is equal to the second parameter ($2)
-- Get all users (active and inactive)
--allUsers, err := q.GetUser(ctx, "user3", nil)

-- name: GetUser :one
SELECT *
FROM users
WHERE username = $1 LIMIT 1;

-- name: UpdateUser :one
Update users
SET password = coalesce(sqlc.narg(password), password),
    email    = coalesce(sqlc.narg(email), email),
    is_active = coalesce(sqlc.narg(is_active), is_active),
    full_name=coalesce(sqlc.narg(full_name), full_name)
WHERE username = sqlc.arg(username) RETURNING *;


-- name: DeleteUser :exec
DELETE
FROM users
WHERE username = $1;

