// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.26.0
// source: user.sql

package db

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
)

const createUser = `-- name: CreateUser :one
INSERT INTO users(username,
                  password,
                  email,
                  full_name, is_active,role)
VALUES ($1, $2, $3, $4, $5, $6) RETURNING username, password, email, full_name, is_active, password_changed_at, created_at, role
`

type CreateUserParams struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Email    string `json:"email"`
	FullName string `json:"full_name"`
	IsActive bool   `json:"is_active"`
	Role     string `json:"role"`
}

func (q *Queries) CreateUser(ctx context.Context, arg CreateUserParams) (User, error) {
	row := q.db.QueryRow(ctx, createUser,
		arg.Username,
		arg.Password,
		arg.Email,
		arg.FullName,
		arg.IsActive,
		arg.Role,
	)
	var i User
	err := row.Scan(
		&i.Username,
		&i.Password,
		&i.Email,
		&i.FullName,
		&i.IsActive,
		&i.PasswordChangedAt,
		&i.CreatedAt,
		&i.Role,
	)
	return i, err
}

const deleteUser = `-- name: DeleteUser :exec
DELETE
FROM users
WHERE username = $1
`

func (q *Queries) DeleteUser(ctx context.Context, username string) error {
	_, err := q.db.Exec(ctx, deleteUser, username)
	return err
}

const getUser = `-- name: GetUser :one

SELECT username, password, email, full_name, is_active, password_changed_at, created_at, role
FROM users
WHERE username = $1 LIMIT 1
`

// AND (is_active = $2 OR $2 IS NULL): This checks if is_active is equal to the second parameter ($2)
// Get all users (active and inactive)
// allUsers, err := q.GetUser(ctx, "user3", nil)
func (q *Queries) GetUser(ctx context.Context, username string) (User, error) {
	row := q.db.QueryRow(ctx, getUser, username)
	var i User
	err := row.Scan(
		&i.Username,
		&i.Password,
		&i.Email,
		&i.FullName,
		&i.IsActive,
		&i.PasswordChangedAt,
		&i.CreatedAt,
		&i.Role,
	)
	return i, err
}

const updateUser = `-- name: UpdateUser :one
Update users
SET password = coalesce($1, password),
    email    = coalesce($2, email),
    password_changed_at = coalesce($3,password_changed_at),
    full_name=coalesce($4, full_name)
WHERE username = $5 RETURNING username, password, email, full_name, is_active, password_changed_at, created_at, role
`

type UpdateUserParams struct {
	Password          pgtype.Text        `json:"password"`
	Email             pgtype.Text        `json:"email"`
	PasswordChangedAt pgtype.Timestamptz `json:"password_changed_at"`
	FullName          pgtype.Text        `json:"full_name"`
	Username          string             `json:"username"`
}

func (q *Queries) UpdateUser(ctx context.Context, arg UpdateUserParams) (User, error) {
	row := q.db.QueryRow(ctx, updateUser,
		arg.Password,
		arg.Email,
		arg.PasswordChangedAt,
		arg.FullName,
		arg.Username,
	)
	var i User
	err := row.Scan(
		&i.Username,
		&i.Password,
		&i.Email,
		&i.FullName,
		&i.IsActive,
		&i.PasswordChangedAt,
		&i.CreatedAt,
		&i.Role,
	)
	return i, err
}
