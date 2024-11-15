-- name: GetVerifyEmail :one
SELECT * FROM verify_emails WHERE username = $1 LIMIT 1;

-- name: CreateVerifyEmail :one
INSERT INTO verify_emails(
    username,email,secret_code
) VALUES(
    $1,$2,$3
) RETURNING *;

-- name: UpdateVerifyEmail :one
UPDATE verify_emails
SET is_used = TRUE
where id = @id AND
    secret_code = @secred_code AND is_used = FALSE AND expired_at > now()
RETURNING *;