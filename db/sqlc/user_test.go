package db

import (
	"context"
	"testing"
	"time"

	"github.com/ashokmouli/simplebank/db/util"
	"github.com/stretchr/testify/require"
)

func TestCreateUser(t *testing.T) {
	hash, err := util.HashPassword(util.RandomString(6))
	require.NoError(t, err)

	args := CreateUserParams{
		Username:       util.RandomOwner(),
		HashedPassword: hash,
		FullName:       util.RandomOwner(),
		Email:          util.RandomEmail(),
	}

	user, err := testQueries.CreateUser(context.Background(), args)
	if err != nil {
		t.Error(err)
	}
	require.NoError(t, err)
	require.NotEmpty(t, user)
	require.Equal(t, user.Username, args.Username)
	require.Equal(t, user.HashedPassword, args.HashedPassword)
	require.Equal(t, user.FullName, args.FullName)
	require.Equal(t, user.Email, args.Email)
	require.True(t, user.PasswordChangedAt.IsZero())
	require.NotZero(t, user.CreatedAt)
}

func makeUser() User {
	args := CreateUserParams{
		Username:       util.RandomOwner(),
		HashedPassword: "secret",
		FullName:       util.RandomOwner(),
		Email:          util.RandomEmail(),
	}

	user, _ := testQueries.CreateUser(context.Background(), args)

	return user

}

func TestGetUser(t *testing.T) {
	test_user := makeUser()
	user, err := testQueries.GetUser(context.Background(), test_user.Username)
	require.NoError(t, err)
	require.NotEmpty(t, user)
	require.Equal(t, test_user.Username, user.Username)
	require.Equal(t, test_user.HashedPassword, user.HashedPassword)
	require.Equal(t, test_user.FullName, user.FullName)
	require.Equal(t, test_user.Email, user.Email)
	require.WithinDuration(t, test_user.PasswordChangedAt, user.PasswordChangedAt, time.Second)
	require.WithinDuration(t, test_user.CreatedAt, user.CreatedAt, time.Second)
}
