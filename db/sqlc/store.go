package db

import (
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Store define all functional for db execution and transactions
// also called compositions and prefer ways to extends struct functionally in golang
// instead of inheritance
type Store interface {
	Querier
	CreateUserAndVerificationTx(ctx context.Context, arg CreateTxUserParams) (CreateUserAndVerificationTxResult, error)
	TransferTx(ctx context.Context, arg TransferTxParams) (TransferTxResult, error)
	VerifyEmailTx(ctx context.Context, arg VerifyEmailTxParams) (VerifyEmailTxResult, error)
}

// SQLStore provide all functionally to execute db queries and transaction
type SQLStore struct {
	*Queries
	connPoll *pgxpool.Pool
}

// NewStore create a new store
func NewStore(connPool *pgxpool.Pool) *SQLStore {

	return &SQLStore{

		Queries:  New(connPool),
		connPoll: connPool,
	}

}
