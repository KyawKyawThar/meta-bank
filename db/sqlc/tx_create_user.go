package db

import (
	"context"
	"github.com/HL/meta-bank/util"
)

type CreateTxUserParams struct {
	CreateUserParams
	AfterCreate func(User) error // this output error will be used to decide whether to commit or rollback transaction
}

type CreateUserAndVerificationTxResult struct {
	User        User
	VerifyEmail VerifyEmail
}

func (store *SQLStore) CreateUserAndVerificationTx(ctx context.Context, arg CreateTxUserParams) (CreateUserAndVerificationTxResult, error) {

	var res CreateUserAndVerificationTxResult

	err := store.execTx(ctx, func(queries *Queries) error {

		var err error
		res.User, err = queries.CreateUser(ctx, arg.CreateUserParams)

		if err != nil {
			return err // Rollback if email verification creation fails
		}

		params := CreateVerifyEmailParams{
			Username:   res.User.Username,
			Email:      res.User.Email,
			SecretCode: util.RandomString(32),
		}

		res.VerifyEmail, err = queries.CreateVerifyEmail(ctx, params)

		if err != nil {
			return err // Rollback if email verification creation fails
		}

		if err := arg.AfterCreate(res.User); err != nil {
			return err // Rollback if AfterCreate fails
		}
		return nil // Commit transaction if all steps succeed

	})
	return res, err
}
