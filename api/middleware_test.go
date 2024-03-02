package api

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/ashokmouli/simplebank/db/util"
	"github.com/ashokmouli/simplebank/token"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func addAuthHeader(
	t *testing.T,
	req *http.Request,
	tokenMaker token.Maker,
	username string,
	duration time.Duration) {

	token, err := tokenMaker.CreateToken(username, duration)
	require.NoError(t, err)
	require.NotEmpty(t, token)
	req.Header.Set("authorization", "bearer "+token)
}
func TestAuthMiddleware(t *testing.T) {

	username := util.RandomOwner()
	duration := time.Hour
	testCases := []struct {
		name          string
		setupAuth     func(t *testing.T, req *http.Request, tokenMaker token.Maker)
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name: "ok",
			setupAuth: func(t *testing.T, req *http.Request, tokenMaker token.Maker) {
				addAuthHeader(t, req, tokenMaker, username, duration)
			},
			checkResponse: func(t *testing.T, resp *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, resp.Code)
			},
		},
		{
			name: "no authorization",
			setupAuth: func(t *testing.T, req *http.Request, tokenMaker token.Maker) {
				/* token, err := tokenMaker.CreateToken(username, duration)
				require.NoError(t, err)
				require.NotEmpty(t, token)
				req.Header.Set("authorization", "bearer " + token) */
			},
			checkResponse: func(t *testing.T, resp *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, resp.Code)
			},
		},
		{
			name: "invalid auth type",
			setupAuth: func(t *testing.T, req *http.Request, tokenMaker token.Maker) {
				token, err := tokenMaker.CreateToken(username, duration)
				require.NoError(t, err)
				require.NotEmpty(t, token)
				// Basic auth shouldn't go through.
				req.Header.Set("authorization", "basic "+token)
			},
			checkResponse: func(t *testing.T, resp *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, resp.Code)
			},
		},
		{
			name: "Expired token",
			setupAuth: func(t *testing.T, req *http.Request, tokenMaker token.Maker) {
				token, err := tokenMaker.CreateToken(username, -time.Minute)
				require.NoError(t, err)
				require.NotEmpty(t, token)
				req.Header.Set("authorization", "bearer "+token)
			},
			checkResponse: func(t *testing.T, resp *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, resp.Code)
			},
		},
	}

	for i := range testCases {
		tc := testCases[i]
		t.Run(tc.name, func(t *testing.T) {
			server := newTestServer(t, nil)

			server.router.GET(
				"/auth",
				createAuthMiddleware(server.maker),
				func(ctx *gin.Context) {
					ctx.JSON(http.StatusOK, gin.H{})
				},
			)

			req, err := http.NewRequest(http.MethodGet, "/auth", nil)
			require.NoError(t, err)
			tc.setupAuth(t, req, server.maker)
			recorder := httptest.NewRecorder()
			server.router.ServeHTTP(recorder, req)
			tc.checkResponse(t, recorder)
		})
	}
}
