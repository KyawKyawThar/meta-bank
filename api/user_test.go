package api

import (
	db "github.com/HL/meta-bank/db/sqlc"
	"github.com/HL/meta-bank/util"
	"github.com/stretchr/testify/require"
	"testing"
)

func randomUser(t *testing.T) db.User {

	hashedPassword, err := util.HashPassword(util.RandomString(8))
	require.NoError(t, err)

	return db.User{
		Username: util.RandomOwner(),
		Password: hashedPassword,
		Email:    util.RandomEmail(),
		FullName: util.RandomOwner(),
	}

}
