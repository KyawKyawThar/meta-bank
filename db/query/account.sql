-- name: CreateAccount :one
INSERT INTO accounts(owner,
                     currency,
                     balance)
VALUES ($1, $2, $3) RETURNING *;

-- name: GetAccount :one
SELECT *
FROM accounts
WHERE id = $1;

-- name: ListAccount :many
SELECT *
FROM accounts
WHERE owner = $1
ORDER BY id LIMIT $2
OFFSET $3;


-- name: UpdateAccount :exec
UPDATE accounts
SET balance = $2
WHERE id = $1 RETURNING *;

-- name: DeleteAccount :exec
DELETE
FROM accounts
WHERE id = $1;
