// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.25.0
// source: user.sql

package db

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgtype"
)

const createUser = `-- name: CreateUser :one
INSERT INTO users(username,
                  password,
                  email,
                  full_name, role, is_active)
VALUES ($1, $2, $3, $4, $5, $6) RETURNING username, password, email, full_name, is_active, role, password_changed_at, created_at
`

type CreateUserParams struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Email    string `json:"email"`
	FullName string `json:"full_name"`
	Role     string `json:"role"`
	IsActive bool   `json:"is_active"`
}

func (q *Queries) CreateUser(ctx context.Context, arg CreateUserParams) (User, error) {
	row := q.db.QueryRow(ctx, createUser,
		arg.Username,
		arg.Password,
		arg.Email,
		arg.FullName,
		arg.Role,
		arg.IsActive,
	)
	var i User
	err := row.Scan(
		&i.Username,
		&i.Password,
		&i.Email,
		&i.FullName,
		&i.IsActive,
		&i.Role,
		&i.PasswordChangedAt,
		&i.CreatedAt,
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

SELECT username, password, email, full_name, is_active, role, password_changed_at, created_at
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
		&i.Role,
		&i.PasswordChangedAt,
		&i.CreatedAt,
	)
	fmt.Println("error is:", err)
	return i, err
}

const listUsers = `-- name: ListUsers :many
SELECT username, password, email, full_name, is_active, role, password_changed_at, created_at
FROM users
Where role != 'admin'
  AND is_active = $1 -- Filter for active users only
ORDER BY username LIMIT $2
OFFSET $3
`

type ListUsersParams struct {
	IsActive bool  `json:"is_active"`
	Limit    int32 `json:"limit"`
	Offset   int32 `json:"offset"`
}

func (q *Queries) ListUsers(ctx context.Context, arg ListUsersParams) ([]User, error) {
	rows, err := q.db.Query(ctx, listUsers, arg.IsActive, arg.Limit, arg.Offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []User{}
	for rows.Next() {
		var i User
		if err := rows.Scan(
			&i.Username,
			&i.Password,
			&i.Email,
			&i.FullName,
			&i.IsActive,
			&i.Role,
			&i.PasswordChangedAt,
			&i.CreatedAt,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const updateUser = `-- name: UpdateUser :one
Update users
SET password  = coalesce($1, password),
    email     = coalesce($2, email),
    is_active = coalesce($3, is_active),
    full_name=coalesce($4, full_name)
WHERE username = $5 RETURNING username, password, email, full_name, is_active, role, password_changed_at, created_at
`

type UpdateUserParams struct {
	Password pgtype.Text `json:"password"`
	Email    pgtype.Text `json:"email"`
	IsActive pgtype.Bool `json:"is_active"`
	FullName pgtype.Text `json:"full_name"`
	Username string      `json:"username"`
}

func (q *Queries) UpdateUser(ctx context.Context, arg UpdateUserParams) (User, error) {
	row := q.db.QueryRow(ctx, updateUser,
		arg.Password,
		arg.Email,
		arg.IsActive,
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
		&i.Role,
		&i.PasswordChangedAt,
		&i.CreatedAt,
	)
	return i, err
}
