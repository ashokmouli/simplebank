package token

import (
	
	"testing"
	"time"

	"github.com/ashokmouli/simplebank/db/util"
	
	"github.com/stretchr/testify/require"
)

func TestPasetoOkToken(t *testing.T) {
	username := util.RandomString(6)
	issuedAt := time.Now()
	duration := time.Hour
	expiredAt := issuedAt.Add(duration)
	maker, err := NewPasetoMaker(util.RandomString(32))
	require.NoError(t, err)
	token, err := maker.CreateToken(username, duration)
	require.NoError(t, err)
	require.NotEmpty(t, token)

	payload, err := maker.VerifyToken(token)
	require.NoError(t, err)
	require.NotEmpty(t, payload)
	require.Equal(t, payload.Username, username)
	require.WithinDuration(t, payload.IssuedAt, issuedAt, time.Second)
	require.WithinDuration(t, payload.ExpiredAt, expiredAt, time.Minute)
}

func TestPasetoExpiredToken(t *testing.T) {
	maker, err := NewPasetoMaker(util.RandomString(32))
	require.NoError(t, err)
	token, err := maker.CreateToken(util.RandomOwner(), -time.Minute)
	require.NotEmpty(t, token)
	require.NoError(t, err)

	payload, err := maker.VerifyToken(token)
	require.Empty(t, payload)
	// require.True(t, errors.Is(err, paseto.ErrTokenExpired))
	require.Error(t, err)

}