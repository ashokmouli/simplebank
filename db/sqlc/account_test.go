package db

import (
	"context"
	"testing"

	"github.com/ashokmouli/simplebank/db/util"
	"github.com/stretchr/testify/require"
)

func TestCreateAccount(t *testing.T) {
	args := CreateAccountParams{
		Owner:    util.RandomOwner(),
		Balance:  util.RandomMoney(),
		Currency: util.RandomCurrency(),
	}

	account, err := testQueries.CreateAccount(context.Background(), args)
	if err != nil {
		t.Error(err)
	}
	require.NoError(t, err)
	require.NotEmpty(t, account)
	require.Equal(t, account.Owner, args.Owner)
	require.Equal(t, account.Balance, args.Balance)
	require.Equal(t, account.Currency, args.Currency)
	require.NotZero(t, account.ID)
	require.NotZero(t, account.CreatedAt)
}

func makeAccount() Account {
	args := CreateAccountParams{
		Owner:    util.RandomOwner(),
		Balance:  util.RandomMoney(),
		Currency: util.RandomCurrency(),
	}

	account, _ := testQueries.CreateAccount(context.Background(), args)

	return account

}

func TestGetAccount(t *testing.T) {
	test_account := makeAccount()
	account, err := testQueries.GetAccount(context.Background(), test_account.ID)
	require.NoError(t, err)
	require.NotEmpty(t, account)
	require.Equal(t, test_account.Owner, account.Owner)
	require.Equal(t, test_account.Balance, account.Balance)
	require.Equal(t, test_account.Currency, account.Currency)

}

func TestListAccounts(t *testing.T) {

	// Make 10 accounts.
	var test_accounts [10]Account
	for i := 0; i < 10; i++ {
		test_accounts[i] = makeAccount()
	}
	args := ListAccountsParams{
		Limit:  3,
		Offset: 0,
	}
	// Call List Account
	accounts, err := testQueries.ListAccounts(context.Background(), args)
	require.NoError(t, err)
	require.Equal(t, len(accounts), 3)

	/*
		 Not clear how to check the values returned in the test_accounts, since it will return the first n accounts, unless
		 you keep track of how many rows are there and start with that count as offset.
		for i := 0; i < 3; i++ {
			require.Equal(t, test_accounts[i].Owner, accounts[i].Owner)
			require.Equal(t, test_accounts[i].Balance, accounts[i].Balance)
			require.Equal(t, test_accounts[i].Currency, accounts[i].Currency)
		}
	*/
}

func TestUpdateAccounts(t *testing.T) {
	test_account := makeAccount()
	arg := UpdateAccountParams{
		ID:      test_account.ID,
		Balance: test_account.Balance + 100,
	}
	account, err := testQueries.UpdateAccount(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, account)
	require.Equal(t, test_account.Owner, account.Owner)
	require.Equal(t, test_account.Balance+100, account.Balance)
	require.Equal(t, test_account.Currency, account.Currency)

}

func TestDeleteAccount(t *testing.T) {
	test_account := makeAccount()
	err := testQueries.DeleteAccount(context.Background(), test_account.ID)
	require.NoError(t, err)

	account, err := testQueries.GetAccount(context.Background(), test_account.ID)
	require.NotEmpty(t, err)
	require.Empty(t, account)
}
