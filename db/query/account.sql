-- name: CreateAccount :one
INSERT INTO accounts(owner,
                     currency,
                     balance)
VALUES ($1, $2, $3) RETURNING *;

-- name: GetAccount :one
SELECT *
FROM accounts
WHERE id = $1;

-- name: UpdateAuthor :exec
UPDATE accounts
SET balance = $2
WHERE id = $1 RETURNING *;

-- name: DeleteAuthor :exec
DELETE
FROM accounts
WHERE id = $1;
