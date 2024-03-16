package db

import (
	"context"
	"github.com/HL/meta-bank/util"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func createRandomAccount(t *testing.T) Account {

	user := createRandomUser(t)
	arg := CreateAccountParams{
		Owner:    user.Username,
		Currency: util.RandomCurrency(),
		Balance:  util.RandomAmount(),
	}

	acc, err := store.CreateAccount(context.Background(), arg)

	require.NoError(t, err)
	require.NotEmpty(t, acc)

	require.Equal(t, arg.Currency, acc.Currency)
	require.Equal(t, arg.Balance, acc.Balance)
	require.Equal(t, arg.Owner, acc.Owner)

	require.NotZero(t, arg.Balance)
	require.NotZero(t, acc.ID)
	require.NotEmpty(t, acc.CreatedAt)

	return acc
}

func TestCreateAccount(t *testing.T) {
	createRandomAccount(t)
}

func TestDeleteAccount(t *testing.T) {
	acc1 := createRandomAccount(t)
	store.DeleteAccount(context.Background(), acc1.ID)

}

func TestGetAccount(t *testing.T) {
	acc := createRandomAccount(t)

	acc2, err := store.GetAccount(context.Background(), acc.ID)
	require.NoError(t, err)
	require.NotEmpty(t, acc2)

	require.Equal(t, acc.ID, acc2.ID)
	require.Equal(t, acc.Currency, acc2.Currency)
	require.Equal(t, acc.Balance, acc2.Balance)
	require.Equal(t, acc.Owner, acc2.Owner)

	require.WithinDuration(t, acc.CreatedAt, acc2.CreatedAt, time.Second)
}

func TestUpdateAccount(t *testing.T) {
	acc1 := createRandomAccount(t)

	arg := UpdateAccountParams{
		ID:      acc1.ID,
		Balance: util.RandomAmount(),
	}
	acc2, err := store.UpdateAccount(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, acc2)

	require.NotEqual(t, acc1.Balance, acc2.Balance)
	require.Equal(t, acc1.ID, acc2.ID)
	require.Equal(t, acc1.Currency, acc2.Currency)

	require.Equal(t, acc1.Owner, acc2.Owner)
	require.WithinDuration(t, acc1.CreatedAt, acc2.CreatedAt, time.Second)
}

func TestListAccounts(t *testing.T) {
	var lastAccount Account

	for i := 0; i < 10; i++ {
		lastAccount = createRandomAccount(t)
	}

	arg := ListAccountParams{
		Owner:  lastAccount.Owner,
		Limit:  5,
		Offset: 0,
	}

	accounts, err := store.ListAccount(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, accounts)

	for _, acc := range accounts {
		require.NotEmpty(t, acc)
		require.Equal(t, lastAccount.Owner, acc.Owner)
	}
}
