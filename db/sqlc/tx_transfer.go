package db

import (
	"context"
	"fmt"
)

type TransferTxResult struct {
	Transfer        Transfer `json:"transfer"`
	ReceiveAccount  Account  `json:"receive_account"`
	TransferAccount Account  `json:"transfer_account"`
	FromEntry       Entry    `json:"from_entry"`
	ToEntry         Entry    `json:"to_entry"`
}

type TransferTxParams struct {
	TransferAccountID int64 `json:"transfer_account_id"`
	ReceiveAccountID  int64 `json:"receive_account_id"`
	Amount            int64 `json:"amount"`
}

var txKey = struct{}{}

// TransferTx performs a money transfer from one account to the other.
// It creates the transfer, add account entries, and update accounts' balance within a database transaction
func (store *SQLStore) TransferTx(ctx context.Context, arg TransferTxParams) (TransferTxResult, error) {

	var result TransferTxResult

	err := store.execTx(ctx, func(q *Queries) error {
		var err error

		txName := ctx.Value(txKey)

		fmt.Println(txName, "create transfer.")

		result.Transfer, err = q.CreateTransfer(ctx, CreateTransferParams{
			FromAccountID: arg.TransferAccountID,
			ToAccountID:   arg.ReceiveAccountID,
			Amount:        arg.Amount,
		})

		if err != nil {
			return err
		}

		fmt.Println(txName, "create entry for receiver account.")
		result.FromEntry, err = q.CreateEntry(ctx, CreateEntryParams{
			AccountID: arg.TransferAccountID,
			Amount:    -arg.Amount,
		})

		if err != nil {
			return err
		}

		fmt.Println(txName, "create entry for transfer account.")
		result.ToEntry, err = q.CreateEntry(ctx, CreateEntryParams{
			AccountID: arg.ReceiveAccountID,
			Amount:    arg.Amount,
		})

		if err != nil {
			return err
		}

		fmt.Println(txName, "get receiver account.")

		//ToAvoid deadlock
		if arg.TransferAccountID < arg.ReceiveAccountID {

			//result.TransferAccount, err = q.UpdateAccountBalance(ctx, UpdateAccountBalanceParams{
			//	Amount: -arg.Amount,
			//	ID:     arg.TransferAccountID,
			//})
			//if err != nil {
			//	return err
			//}
			//
			//result.ReceiveAccount, err = q.UpdateAccountBalance(ctx, UpdateAccountBalanceParams{
			//	Amount: arg.Amount,
			//	ID:     arg.ReceiveAccountID,
			//})
			//
			//if err != nil {
			//	return err
			//}
			result.TransferAccount, result.ReceiveAccount, err = addMoney(ctx, q, -arg.Amount, arg.Amount, arg.TransferAccountID, arg.ReceiveAccountID)
		} else {
			result.ReceiveAccount, result.TransferAccount, err = addMoney(ctx, q, arg.Amount, -arg.Amount, arg.ReceiveAccountID, arg.TransferAccountID)
			//result.ReceiveAccount, err = q.UpdateAccountBalance(ctx, UpdateAccountBalanceParams{
			//	Amount: arg.Amount,
			//	ID:     arg.ReceiveAccountID,
			//})
			//
			//if err != nil {
			//	return err
			//}
			//result.TransferAccount, err = q.UpdateAccountBalance(ctx, UpdateAccountBalanceParams{
			//	Amount: -arg.Amount,
			//	ID:     arg.TransferAccountID,
			//})
			//if err != nil {
			//	return err
			//}

		}

		//result.ReceiveAccount, err = q.UpdateAccountBalance(ctx, UpdateAccountBalanceParams{
		//	Amount: arg.Amount,
		//	ID:     arg.ReceiveAccountID,
		//})
		//
		//if err != nil {
		//	return err
		//}
		//acc1, err := q.GetAccountForUpdateTransaction(ctx, arg.ReceiveAccountID)
		//if err != nil {
		//	return err
		//}
		//
		//fmt.Println(txName, "receiver account updated")
		//result.ReceiveAccount, err = q.UpdateAccount(ctx, UpdateAccountParams{
		//	ID:      acc1.ID,
		//	Balance: acc1.Balance + arg.Amount,
		//})

		//if err != nil {
		//	return err
		//}

		fmt.Println(txName, "get transfer account")
		//acc2, err := q.GetAccountForUpdateTransaction(ctx, arg.TransferAccountID)
		//
		//if err != nil {
		//	return err
		//}
		//
		//fmt.Println(txName, "transfer account updated")
		//result.TransferAccount, err = q.UpdateAccount(ctx, UpdateAccountParams{
		//	ID:      acc2.ID,
		//	Balance: acc2.Balance - arg.Amount,
		//})
		//result.TransferAccount, err = q.UpdateAccountBalance(ctx, UpdateAccountBalanceParams{
		//	Amount: -arg.Amount,
		//	ID:     arg.TransferAccountID,
		//})
		//if err != nil {
		//	return err
		//}

		return err
	})

	return result, err
}

func addMoney(ctx context.Context, q *Queries, amount1, amount2, receiveAccountID, transferAccountID int64) (receiveAccount, transferAccount Account, err error) {

	receiveAccount, err = q.UpdateAccountBalance(ctx, UpdateAccountBalanceParams{
		Amount: amount1,
		ID:     receiveAccountID,
	})

	if err != nil {
		return
	}
	transferAccount, err = q.UpdateAccountBalance(ctx, UpdateAccountBalanceParams{
		Amount: amount2,
		ID:     transferAccountID,
	})
	if err != nil {
		return
	}
	return
}
