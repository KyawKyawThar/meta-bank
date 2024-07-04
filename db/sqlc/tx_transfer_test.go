package db

import (
	"context"
	"fmt"
	"github.com/stretchr/testify/require"
	"testing"
)

// TransferTx
func TestTransferTx(t *testing.T) {

	//run a concurrent transfer transaction

	//before transfer balance account
	acc1 := createRandomAccount(t)
	acc2 := createRandomAccount(t)

	fmt.Println(">> before:", acc1.Balance, acc2.Balance)

	n := 2

	amount := int64(10)

	errs := make(chan error)
	results := make(chan TransferTxResult)

	for i := 0; i < n; i++ {

		txName := fmt.Sprintf("tx: %d", i+1)

		go func() {
			//for deadlock debuging
			ctx := context.WithValue(context.Background(), txKey, txName)

			result, err := testStore.TransferTx(ctx, TransferTxParams{
				ReceiveAccountID:  acc1.ID,
				TransferAccountID: acc2.ID,
				Amount:            amount,
			})

			errs <- err
			results <- result
		}()
	}

	existed := make(map[int]bool)

	//check the result that pass through from channel go routine
	for i := 0; i < n; i++ {
		err := <-errs
		require.NoError(t, err)
		result := <-results
		require.NotEmpty(t, result)

		transfer := result.Transfer
		require.NotEmpty(t, transfer)
		require.Equal(t, acc1.ID, transfer.FromAccountID)
		require.Equal(t, acc2.ID, transfer.ToAccountID)
		require.Equal(t, amount, transfer.Amount)

		require.NotZero(t, transfer.ID)
		require.NotZero(t, transfer.CreatedAt)

		_, err = testStore.GetTransfer(context.Background(), transfer.ID)
		require.NoError(t, err)

		// check entries
		fromEntry := result.FromEntry
		require.NotEmpty(t, fromEntry)

		require.Equal(t, acc1.ID, fromEntry.AccountID)
		require.Equal(t, -amount, fromEntry.Amount)

		require.NotZero(t, fromEntry.ID)
		require.NotZero(t, fromEntry.Amount)
		require.NotZero(t, fromEntry.CreatedAt)

		_, err = testStore.GetEntry(context.Background(), fromEntry.ID)
		require.NoError(t, err)

		toEntry := result.ToEntry
		require.NotEmpty(t, toEntry)
		require.Equal(t, acc2.ID, toEntry.AccountID)
		require.Equal(t, amount, toEntry.Amount)

		require.NotZero(t, toEntry.ID)
		require.NotZero(t, toEntry.Amount)
		require.NotZero(t, toEntry.CreatedAt)

		_, err = testStore.GetEntry(context.Background(), toEntry.ID)
		require.NoError(t, err)

		receiveAccount := result.ReceiveAccount
		require.NotEmpty(t, receiveAccount)
		require.Equal(t, acc1.ID, receiveAccount.ID)

		transferAccount := result.TransferAccount
		require.NotEmpty(t, transferAccount)
		require.Equal(t, acc2.ID, transferAccount.ID)

		//check and compare balance
		fmt.Println(">>> tx:", receiveAccount.Balance, transferAccount.Balance)
		diff2 := acc2.Balance - transferAccount.Balance
		diff1 := receiveAccount.Balance - acc1.Balance

		fmt.Println("diff>>: ", diff1, diff2)

		require.Equal(t, diff1, diff2)
		require.True(t, diff1%amount == 0)
		require.True(t, diff1 > 0)

		k := int(diff1 / amount)
		require.True(t, k >= 1 && k <= n)
		require.NotContains(t, existed, k)
		existed[k] = true

	}
	// check the final updated balance
	receivedAccount1, err := testStore.GetAccount(context.Background(), acc1.ID)
	require.NoError(t, err)

	transferAccount2, err := testStore.GetAccount(context.Background(), acc2.ID)
	require.NoError(t, err)

	require.Equal(t, acc1.Balance+int64(n)*amount, receivedAccount1.Balance)
	require.Equal(t, acc1.Balance-int64(n)*amount, transferAccount2.Balance)

}

func TestTransferTxDeadlock(t *testing.T) {

	acc1 := createRandomAccount(t)
	acc2 := createRandomAccount(t)

	fmt.Println(">> before:", acc1.Balance, acc2.Balance)

	n := 10

	amount := int64(10)

	errs := make(chan error)

	for i := 0; i < n; i++ {

		txName := fmt.Sprintf("tx: %d", i+1)

		fromAccountID := acc1.ID
		toAccountID := acc2.ID

		if i%2 == 0 {
			fromAccountID = acc2.ID
			toAccountID = acc1.ID
		}

		go func() {
			//for deadlock debuging
			ctx := context.WithValue(context.Background(), txKey, txName)

			_, err := testStore.TransferTx(ctx, TransferTxParams{
				ReceiveAccountID:  fromAccountID,
				TransferAccountID: toAccountID,
				Amount:            amount,
			})

			errs <- err

		}()
	}

	for i := 0; i < n; i++ {
		err := <-errs
		require.NoError(t, err)

	}
	// check the final updated balance
	receivedAccount1, err := testStore.GetAccount(context.Background(), acc1.ID)
	require.NoError(t, err)

	transferAccount2, err := testStore.GetAccount(context.Background(), acc2.ID)
	require.NoError(t, err)

	require.Equal(t, acc1.Balance, receivedAccount1.Balance)
	require.Equal(t, acc1.Balance, transferAccount2.Balance)
}
