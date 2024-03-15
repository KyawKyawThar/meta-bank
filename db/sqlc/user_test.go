package db

import (
	"context"
	"fmt"
	"github.com/HL/meta-bank/util"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func CreateRandomUser(t *testing.T) User {

	hashPassword, err := util.HashPassword(util.RandomString(7))
	require.NoError(t, err)

	arg := CreateUserParams{
		Username: util.RandomOwner(),
		Password: hashPassword,
		Email:    util.RandomEmail(),
		FullName: util.RandomOwner(),
		Role:     "admin",
		IsActive: false,
	}

	user, err := store.CreateUser(context.Background(), arg)

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
	CreateRandomUser(t)
}

func TestGetUser(t *testing.T) {
	user1 := CreateRandomUser(t)

	user2, err := store.GetUser(context.Background(), user1.Username)

	fmt.Println("user2 is:", user2)

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
	oldUser := CreateRandomUser(t)
	newFullName := util.RandomOwner()
	arg := UpdateUserParams{
		Username: oldUser.Username,
		FullName: pgtype.Text{
			String: newFullName,
			Valid:  true,
		},
	}
	updateUser, err := store.UpdateUser(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, updateUser)

	require.NotEqual(t, oldUser.FullName, updateUser.FullName)

	require.Equal(t, newFullName, updateUser.FullName)
	require.Equal(t, oldUser.Username, updateUser.Username)
	require.Equal(t, oldUser.Password, updateUser.Password)
	require.Equal(t, oldUser.Email, updateUser.Email)

}

func TestDeleteUser(t *testing.T) {
	user := CreateRandomUser(t)

	err := store.DeleteUser(context.Background(), user.Username)
	require.NoError(t, err)
}
