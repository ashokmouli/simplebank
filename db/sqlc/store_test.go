package db

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTransferTx(t *testing.T) {
	store := NewStore(testDB)
	account1 := makeAccount()
	account2 := makeAccount()
	amount := int64(100)
	errs := make(chan error)
	results := make(chan TransferTxResults)
	const n = 5
	fmt.Printf("Balance 1 %d, Balance 2 %d, Amount %d\n", account1.Balance, account2.Balance, amount*n)
	for i := 0; i < n; i++ {
		txName := fmt.Sprintf("tx: %d", i)
		go func() {
			ctx := context.WithValue(context.Background(), txKey, txName)
			result, err := store.TransferTx(ctx, &TransferTxParams{
				FromAccountID: account1.ID,
				ToAccountID:   account2.ID,
				Amount:        amount,
			})
			errs <- err
			results <- result
		}()
	}
	for i := 0; i < n; i++ {
		err := <-errs
		require.NoError(t, err)
		result := <-results
		require.NotEmpty(t, result)
		transfer := result.Transfer

		require.NotEmpty(t, transfer)
		require.Equal(t, transfer.FromAccount, account1.ID)
		require.Equal(t, transfer.ToAccount, account2.ID)
		require.Equal(t, transfer.Amount, amount)
		require.NotZero(t, transfer.ID)
		require.NotZero(t, transfer.CreatedAt)

		_, err = store.GetTransfer(context.Background(), transfer.ID)
		require.NoError(t, err)

		entry := result.FromEntry
		require.NotEmpty(t, entry)
		require.Equal(t, entry.AccountID, account1.ID)
		require.Equal(t, entry.Amount, -amount)
		require.NotZero(t, entry.ID)
		require.NotZero(t, entry.CreatedAt)

		entry = result.ToEntry
		require.NotEmpty(t, entry)
		require.Equal(t, entry.AccountID, account2.ID)
		require.Equal(t, entry.Amount, amount)
		require.NotZero(t, entry.ID)
		require.NotZero(t, entry.CreatedAt)

		fromAccount := result.FromAccount
		require.NotEmpty(t, fromAccount)
		require.Equal(t, fromAccount.ID, account1.ID)

		toAccount := result.ToAccount
		require.NotEmpty(t, toAccount)
		require.Equal(t, toAccount.ID, account2.ID)

		fromDiff := account1.Balance - fromAccount.Balance
		toDiff := toAccount.Balance - account2.Balance
		require.Equal(t, fromDiff, toDiff)
		require.True(t, fromDiff%amount == 0)
		k := int(fromDiff / amount)
		require.True(t, k >= 1 && k <= n)
	}

	updatedAccount1, err := store.GetAccount(context.Background(), account1.ID)
	require.NoError(t, err)
	require.Equal(t, updatedAccount1.Balance, account1.Balance-n*amount)

	updatedAccount2, err := store.GetAccount(context.Background(), account2.ID)
	require.NoError(t, err)
	require.Equal(t, updatedAccount2.Balance, account2.Balance+n*amount)
	fmt.Printf("Balance 1 %d, Balance 2 %d, Amount %d\n", updatedAccount1.Balance, updatedAccount2.Balance, amount)

}

func TestTransferDeadlock(t *testing.T) {
	store := NewStore(testDB)
	account1 := makeAccount()
	account2 := makeAccount()
	amount := int64(100)
	errs := make(chan error)
	const n = 10

	for i := 0; i < n; i++ {
		txName := fmt.Sprintf("tx: %d", i)
		fromAccountID := account1.ID
		toAccountID := account2.ID
		if i%2 == 1 {
			fromAccountID = account2.ID
			toAccountID = account1.ID
		}
		go func() {
			ctx := context.WithValue(context.Background(), txKey, txName)
			_, err := store.TransferTx(ctx, &TransferTxParams{
				FromAccountID: fromAccountID,
				ToAccountID:   toAccountID,
				Amount:        amount,
			})
			errs <- err
		}()
	}
	for i := 0; i < n; i++ {
		err := <-errs
		require.NoError(t, err)
	}

	updatedAccount1, err := store.GetAccount(context.Background(), account1.ID)
	require.NoError(t, err)
	require.Equal(t, updatedAccount1.Balance, account1.Balance)

	updatedAccount2, err := store.GetAccount(context.Background(), account2.ID)
	require.NoError(t, err)
	require.Equal(t, updatedAccount2.Balance, account2.Balance)
}
