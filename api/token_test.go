package api

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	mockdb "github.com/ashokmouli/simplebank/db/mock"
	db "github.com/ashokmouli/simplebank/db/sqlc"
	"github.com/ashokmouli/simplebank/db/util"
	"github.com/ashokmouli/simplebank/token"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
)
func TestRenewTokenAPI(t *testing.T) {

	username := util.RandomOwner()
	

	testCases := []struct {
		name        string
		createToken func(t *testing.T, tokenMaker token.Maker) (string, *token.Payload)
		buildStore  func(t *testing.T, store *mockdb.MockStore, maker token.Maker, token string, payload *token.Payload)
		matchResult func(t *testing.T, resp *httptest.ResponseRecorder, maker token.Maker, payload *token.Payload)
	}{
		{
			name: "ok",
			
			createToken: func(t *testing.T, tokenMaker token.Maker) (string, *token.Payload) {
				token, payload, err := tokenMaker.CreateToken(username, time.Hour)
				require.NoError(t, err)
				require.NotEmpty(t, payload)
				return token, payload
			},
		
			buildStore: func(t *testing.T, store *mockdb.MockStore, maker token.Maker, token string, payload *token.Payload) {

				session := db.Session {
					ID: payload.ID,
					Username: payload.Username,
					IsBlocked: false,
					RefreshToken: token,
					ExpiresAt: payload.ExpiredAt,
					CreatedAt: time.Now(),
				}
				store.EXPECT().GetSession(gomock.Any(), gomock.Eq(payload.ID)).Return(session, nil).Times(1)
			},
			matchResult: func(t *testing.T, resp *httptest.ResponseRecorder, maker token.Maker, payload *token.Payload) {
				require.Equal(t, http.StatusOK, resp.Code)
				// Verify the user name in the access token matches the payload.
				data, err := io.ReadAll(resp.Body)
				require.NoError(t, err)
				var got RenewTokenResponse
				err = json.Unmarshal(data, &got)
				require.NoError(t, err)
				gotPayload, err := maker.VerifyToken(got.AccessToken)
				require.NoError(t, err)
				require.Equal(t, gotPayload.Username, payload.Username)
			},
		},
		{
			name: "Forbidden",
			
			createToken: func(t *testing.T, tokenMaker token.Maker) (string, *token.Payload) {
				token, payload, err := tokenMaker.CreateToken(username, time.Hour)
				require.NoError(t, err)
				require.NotEmpty(t, payload)
				return token, payload
			},
		
			buildStore: func(t *testing.T, store *mockdb.MockStore, maker token.Maker, token1 string, payload *token.Payload) {

				// Create a different session object and return it to the handler to test error conditions.
				username := util.RandomOwner()
				token2, payload2, err := maker.CreateToken(username, time.Hour)
				require.NoError(t, err)
				require.NotEmpty(t, payload2)

				session := db.Session {
					ID: payload.ID,
					Username: payload2.Username,
					IsBlocked: false,
					RefreshToken: token2,
					ExpiresAt: payload2.ExpiredAt,
					CreatedAt: time.Now(),
				}
				store.EXPECT().GetSession(gomock.Any(), gomock.Eq(payload.ID)).Return(session, nil).Times(1)
			},
			matchResult: func(t *testing.T, resp *httptest.ResponseRecorder, maker token.Maker, payload *token.Payload) {
				require.Equal(t, http.StatusForbidden, resp.Code)
			},
		},
		{
			name: "Forbidden -- Blocked",
			
			createToken: func(t *testing.T, tokenMaker token.Maker) (string, *token.Payload) {
				token, payload, err := tokenMaker.CreateToken(username, time.Hour)
				require.NoError(t, err)
				require.NotEmpty(t, payload)
				return token, payload
			},
		
			buildStore: func(t *testing.T, store *mockdb.MockStore, maker token.Maker, token1 string, payload *token.Payload) {

				// Create a different session object and return it to the handler to test error conditions.
				username := util.RandomOwner()
				token2, payload2, err := maker.CreateToken(username, time.Hour)
				require.NoError(t, err)
				require.NotEmpty(t, payload2)

				session := db.Session {
					ID: payload.ID,
					Username: payload2.Username,
					IsBlocked: true,
					RefreshToken: token2,
					ExpiresAt: payload2.ExpiredAt,
					CreatedAt: time.Now(),
				}
				store.EXPECT().GetSession(gomock.Any(), gomock.Eq(payload.ID)).Return(session, nil).Times(1)
			},
			matchResult: func(t *testing.T, resp *httptest.ResponseRecorder, maker token.Maker, payload *token.Payload) {
				require.Equal(t, http.StatusForbidden, resp.Code)
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			// Create a mockstore.
			mockStore := mockdb.NewMockStore(ctrl)

			// Create a new test server.
			server := newTestServer(t, mockStore)

			// Create a refresh token
			token, payload := tc.createToken(t, server.maker)

			// Marshall the token
			
			data, err := json.Marshal(RenewTokenRequest {
				RefreshToken: token,
			})
			require.NoError(t, err)

			// Build out the mock store.
			tc.buildStore(t, mockStore, server.maker, token, payload)


			// Create a new (POST) request.
			req := httptest.NewRequest(http.MethodPost, "/tokens/renew_token", bytes.NewReader(data))

			// Create a response and make the http call.
			resp := httptest.NewRecorder()
			server.router.ServeHTTP(resp, req)
			tc.matchResult(t, resp, server.maker, payload)
		})
	}

}