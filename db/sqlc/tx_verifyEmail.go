package db

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5/pgtype"
)

type VerifyEmailTxParams struct {
	EmailId    int64
	SecretCode string
}

type VerifyEmailTxResult struct {
	User        User
	VerifyEmail VerifyEmail
}

func (store *SQLStore) VerifyEmailTx(ctx context.Context, arg VerifyEmailTxParams) (VerifyEmailTxResult, error) {

	var res VerifyEmailTxResult

	err := store.execTx(ctx, func(queries *Queries) error {
		var err error
		res.VerifyEmail, err = queries.UpdateVerifyEmail(ctx, UpdateVerifyEmailParams{
			ID:         arg.EmailId,
			SecredCode: arg.SecretCode,
		})

		if err != nil {
			return err
		}
		fmt.Println("code is running in here..")
		res.User, err = queries.UpdateUser(ctx, UpdateUserParams{
			IsVerifiedEmail: pgtype.Bool{
				Bool:  true,
				Valid: true,
			},
			Username: res.VerifyEmail.Username,
		})

		fmt.Println("res.User is:", res.User)
		if err != nil {
			return err
		}
		return nil
	})
	return res, err
}
