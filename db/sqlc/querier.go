// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.27.0

package db

import (
	"context"

	"github.com/google/uuid"
)

type Querier interface {
	CreateAccount(ctx context.Context, arg CreateAccountParams) (Account, error)
	CreateEntry(ctx context.Context, arg CreateEntryParams) (Entry, error)
	CreateSession(ctx context.Context, arg CreateSessionParams) (Session, error)
	CreateTransfer(ctx context.Context, arg CreateTransferParams) (Transfer, error)
	CreateUser(ctx context.Context, arg CreateUserParams) (User, error)
	CreateVerifyEmail(ctx context.Context, arg CreateVerifyEmailParams) (VerifyEmail, error)
	DeleteAccount(ctx context.Context, id int64) error
	DeleteUser(ctx context.Context, username string) error
	GetAccount(ctx context.Context, id int64) (Account, error)
	GetAccountForUpdateTransaction(ctx context.Context, id int64) (Account, error)
	GetEntry(ctx context.Context, id int64) (Entry, error)
	GetSession(ctx context.Context, id uuid.UUID) (Session, error)
	GetTransfer(ctx context.Context, id int64) (Transfer, error)
	// AND (is_active = $2 OR $2 IS NULL): This checks if is_active is equal to the second parameter ($2)
	// Get all users (active and inactive)
	//allUsers, err := q.GetUser(ctx, "user3", nil)
	GetUser(ctx context.Context, username string) (User, error)
	GetVerifyEmail(ctx context.Context, username string) (VerifyEmail, error)
	ListAccount(ctx context.Context, arg ListAccountParams) ([]Account, error)
	ListEntries(ctx context.Context, arg ListEntriesParams) ([]Entry, error)
	ListTransfers(ctx context.Context, arg ListTransfersParams) ([]Transfer, error)
	UpdateAccount(ctx context.Context, arg UpdateAccountParams) (Account, error)
	UpdateAccountBalance(ctx context.Context, arg UpdateAccountBalanceParams) (Account, error)
	UpdateUser(ctx context.Context, arg UpdateUserParams) (User, error)
	UpdateVerifyEmail(ctx context.Context, arg UpdateVerifyEmailParams) (VerifyEmail, error)
}

var _ Querier = (*Queries)(nil)
