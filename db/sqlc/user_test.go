package db

import (
	"context"
	"github.com/HL/meta-bank/util"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func createRandomUser(t *testing.T) User {

	hashPassword, err := util.HashPassword(util.RandomString(7))
	require.NoError(t, err)

	arg := CreateUserParams{
		Username: util.RandomOwner(),
		Password: hashPassword,
		Email:    util.RandomEmail(),
		FullName: util.RandomOwner(),
		Role:     "admin",
		IsActive: true,
	}

	user, err := testStore.CreateUser(context.Background(), arg)

	require.NoError(t, err)
	require.NotEmpty(t, user)

	require.Equal(t, arg.Username, user.Username)
	require.Equal(t, arg.Password, user.Password)
	require.Equal(t, arg.Email, user.Email)
	require.Equal(t, arg.FullName, user.FullName)

	require.True(t, user.PasswordChangedAt.IsZero())

	return user

}

func TestCreateUser(t *testing.T) {
	createRandomUser(t)
}

func TestGetUser(t *testing.T) {
	user1 := createRandomUser(t)

	user2, err := testStore.GetUser(context.Background(), user1.Username)

	require.NoError(t, err)
	require.NotEmpty(t, user2)

	require.Equal(t, user1.Username, user2.Username)
	require.Equal(t, user1.Password, user2.Password)
	require.Equal(t, user1.Email, user2.Email)
	require.Equal(t, user1.FullName, user2.FullName)

	require.WithinDuration(t, user1.PasswordChangedAt, user2.PasswordChangedAt, time.Second)
	require.WithinDuration(t, user1.CreatedAt, user2.CreatedAt, time.Second)

}

func TestUpdateUserOnlyFullName(t *testing.T) {
	oldUser := createRandomUser(t)
	newFullName := util.RandomOwner()
	arg := UpdateUserParams{
		Username: oldUser.Username,
		FullName: pgtype.Text{
			String: newFullName,
			Valid:  true,
		},
	}
	updateUser, err := testStore.UpdateUser(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, updateUser)

	require.NotEqual(t, oldUser.FullName, updateUser.FullName)

	require.Equal(t, newFullName, updateUser.FullName)
	require.Equal(t, oldUser.Username, updateUser.Username)
	require.Equal(t, oldUser.Password, updateUser.Password)
	require.Equal(t, oldUser.Email, updateUser.Email)

}

func TestUpdateUserOnlyPassword(t *testing.T) {
	oldUser := createRandomUser(t)

	hashPassword, err := util.HashPassword(util.RandomString(7))

	require.NoError(t, err)

	arg := UpdateUserParams{
		Username: oldUser.Username,
		Password: pgtype.Text{
			String: hashPassword,
			Valid:  true,
		},
	}
	updateUser, err := testStore.UpdateUser(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, updateUser)

	require.NotEqual(t, oldUser.Password, updateUser.Password)

	require.Equal(t, hashPassword, updateUser.Password)
	require.Equal(t, oldUser.Username, updateUser.Username)
	require.Equal(t, oldUser.Email, updateUser.Email)

}

func TestUpdateUserOnlyEmail(t *testing.T) {
	oldUser := createRandomUser(t)

	arg := UpdateUserParams{
		Username: oldUser.Username,
		Email: pgtype.Text{
			String: util.RandomEmail(),
			Valid:  true,
		},
	}
	updateUser, err := testStore.UpdateUser(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, updateUser)

	require.NotEqual(t, oldUser.Email, updateUser.Email)

	require.Equal(t, arg.Email.String, updateUser.Email)
	require.Equal(t, oldUser.Username, updateUser.Username)
	require.Equal(t, oldUser.Password, updateUser.Password)

}

func TestDeleteUser(t *testing.T) {
	user := createRandomUser(t)

	err := testStore.DeleteUser(context.Background(), user.Username)
	require.NoError(t, err)
}
