package db

import (
	"context"
	"github.com/HL/meta-bank/util"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func CreateRandomTransfer(t *testing.T, account1, account2 Account) Transfer {
	arg := CreateTransferParams{
		FromAccountID: account1.ID,
		ToAccountID:   account2.ID,
		Amount:        util.RandomAmount(),
	}

	transfer, err := testStore.CreateTransfer(context.Background(), arg)

	require.NoError(t, err)
	require.NotEmpty(t, transfer)

	require.Equal(t, arg.FromAccountID, transfer.FromAccountID)
	require.Equal(t, arg.ToAccountID, transfer.ToAccountID)
	require.Equal(t, arg.Amount, transfer.Amount)

	require.NotZero(t, transfer.Amount)
	require.NotZero(t, transfer.CreatedAt)

	return transfer
}

func TestCreateTransfer(t *testing.T) {

	acc1 := createRandomAccount(t)
	acc2 := createRandomAccount(t)

	CreateRandomTransfer(t, acc1, acc2)

}

func TestGetTransfer(t *testing.T) {
	acc1 := createRandomAccount(t)
	acc2 := createRandomAccount(t)

	transfer := CreateRandomTransfer(t, acc1, acc2)

	t1, err := testStore.GetTransfer(context.Background(), transfer.ID)

	require.NoError(t, err)
	require.NotEmpty(t, t1)

	require.Equal(t, transfer.ID, t1.ID)
	require.Equal(t, transfer.Amount, t1.Amount)
	require.Equal(t, transfer.FromAccountID, t1.FromAccountID)
	require.Equal(t, transfer.ToAccountID, t1.ToAccountID)

	require.WithinDuration(t, transfer.CreatedAt, t1.CreatedAt, time.Second)
}

func TestListTransfer(t *testing.T) {
	acc1 := createRandomAccount(t)
	acc2 := createRandomAccount(t)

	for i := 0; i < 5; i++ {
		CreateRandomTransfer(t, acc1, acc2)
		CreateRandomTransfer(t, acc2, acc1)
	}

	arg := ListTransfersParams{
		FromAccountID: acc1.ID,
		ToAccountID:   acc1.ID,
		Limit:         5,
		Offset:        5,
	}
	transfers, err := testStore.ListTransfers(context.Background(), arg)

	require.NoError(t, err)
	require.NotEmpty(t, transfers)
	require.Len(t, transfers, 5)

	for _, transfer := range transfers {
		t1, err := testStore.GetTransfer(context.Background(), transfer.ID)
		require.NoError(t, err)
		require.NotEmpty(t, t1)

		require.True(t, acc1.ID == transfer.FromAccountID || acc1.ID == transfer.ToAccountID)
	}
}
