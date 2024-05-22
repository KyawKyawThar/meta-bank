package db

import (
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Store define all functional for db execution and transactions
type Store interface {
	Querier
	CreateUserTx(ctx context.Context, arg CreateTxUserParams) (CreateUserTxResult, error)
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
