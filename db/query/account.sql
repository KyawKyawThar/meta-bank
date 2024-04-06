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


-- name: UpdateAccount :one
UPDATE accounts
SET balance = balance+sqlc.arg(amount)
WHERE id = $1 RETURNING *;

-- name: DeleteAccount :exec
DELETE
FROM accounts
WHERE id = $1;
