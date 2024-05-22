package db

import "context"

type CreateTxUserParams struct {
	CreateUserParams
	AfterCreate func(User) error
}

type CreateUserTxResult struct {
	User User
}

func (s *SQLStore) CreateUserTx(ctx context.Context, arg CreateTxUserParams) (CreateUserTxResult, error) {

	var res CreateUserTxResult

	err := s.execTx(ctx, func(queries *Queries) error {

		var err error
		res.User, err = queries.CreateUser(ctx, arg.CreateUserParams)

		if err != nil {
			return err
		}
		return arg.AfterCreate(res.User)
	})
	return res, err
}
