package api

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	mockdb "github.com/ashokmouli/simplebank/db/mock"
	db "github.com/ashokmouli/simplebank/db/sqlc"
	"github.com/ashokmouli/simplebank/db/util"
	"github.com/ashokmouli/simplebank/token"
	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
)

func TestCreateAccountAPI(t *testing.T) {

	username := util.RandomOwner()
	account := randomAccount(username)

	testCases := []struct {
		name        string
		body        gin.H
		setupAuth   func(t *testing.T, req *http.Request, tokenMaker token.Maker, username string, duration time.Duration)
		buildStore  func(t *testing.T, store *mockdb.MockStore)
		matchResult func(t *testing.T, resp *httptest.ResponseRecorder)
	}{
		{
			name: "ok",
			body: gin.H{
				"currency": account.Currency,
			},
			setupAuth: func(t *testing.T, req *http.Request, tokenMaker token.Maker, username string, duration time.Duration) {
				addAuthHeader(t, req, tokenMaker, username, duration)
			},
			buildStore: func(t *testing.T, store *mockdb.MockStore) {
				arg := db.CreateAccountParams{
					Owner:    username,
					Balance:  0,
					Currency: account.Currency,
				}
				store.EXPECT().CreateAccount(gomock.Any(), gomock.Eq(arg)).Return(account, nil).Times(1)
			},
			matchResult: func(t *testing.T, resp *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, resp.Code)
				matchReturnedAccount(t, resp.Body, &account)
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockStore := mockdb.NewMockStore(ctrl)
			tc.buildStore(t, mockStore)

			// Marshal the body data.
			data, err := json.Marshal(tc.body)
			require.NoError(t, err)

			// Create a new test server.
			server := newTestServer(t, mockStore)

			// Create a new (POST) request.
			req := httptest.NewRequest(http.MethodPost, "/accounts", bytes.NewReader(data))

			// Add the auth token as a header
			tc.setupAuth(t, req, server.maker, username, time.Hour)

			// Create a response and make the http call.
			resp := httptest.NewRecorder()
			server.router.ServeHTTP(resp, req)
			tc.matchResult(t, resp)
		})
	}

}

func TestGetAccountAPI(t *testing.T) {
	username := util.RandomOwner()
	account := randomAccount(username)

	testCases := []struct {
		name        string
		accountID   int64
		setupAuth   func(t *testing.T, req *http.Request, tokenMaker token.Maker, username string, duration time.Duration)
		buildStore  func(*testing.T, *mockdb.MockStore)
		matchResult func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name:      "ok",
			accountID: account.ID,
			setupAuth: func(t *testing.T, req *http.Request, tokenMaker token.Maker, username string, duration time.Duration) {
				addAuthHeader(t, req, tokenMaker, username, duration)
			},
			buildStore: func(t *testing.T, store *mockdb.MockStore) {
				store.EXPECT().GetAccount(gomock.Any(), gomock.Eq(account.ID)).Return(account, nil).Times(1)
			},
			matchResult: func(t *testing.T, resp *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, resp.Code)
				matchReturnedAccount(t, resp.Body, &account)
			},
		},
		{
			name:      "Not Found",
			accountID: account.ID,
			setupAuth: func(t *testing.T, req *http.Request, tokenMaker token.Maker, username string, duration time.Duration) {
				addAuthHeader(t, req, tokenMaker, username, duration)
			},
			buildStore: func(t *testing.T, store *mockdb.MockStore) {
				store.EXPECT().GetAccount(gomock.Any(), gomock.Eq(account.ID)).Return(db.Account{}, sql.ErrNoRows).Times(1)
			},
			matchResult: func(t *testing.T, resp *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusNotFound, resp.Code)
			},
		},
		{
			name:      "InternalError",
			accountID: account.ID,
			setupAuth: func(t *testing.T, req *http.Request, tokenMaker token.Maker, username string, duration time.Duration) {
				addAuthHeader(t, req, tokenMaker, username, duration)
			},
			buildStore: func(t *testing.T, store *mockdb.MockStore) {
				store.EXPECT().GetAccount(gomock.Any(), gomock.Eq(account.ID)).Return(account, sql.ErrConnDone).Times(1)
			},
			matchResult: func(t *testing.T, resp *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, resp.Code)
			},
		},
		{
			name:      "BadRequest",
			accountID: 0,
			setupAuth: func(t *testing.T, req *http.Request, tokenMaker token.Maker, username string, duration time.Duration) {
				addAuthHeader(t, req, tokenMaker, username, duration)
			},
			buildStore: func(t *testing.T, store *mockdb.MockStore) {
				store.EXPECT().GetAccount(gomock.Any(), gomock.Eq(account.ID)).Return(account, nil).Times(0)
			},
			matchResult: func(t *testing.T, resp *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, resp.Code)
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockStore := mockdb.NewMockStore(ctrl)
			tc.buildStore(t, mockStore)

			server := newTestServer(t, mockStore)
			url := fmt.Sprintf("/accounts/%d", tc.accountID)
			req := httptest.NewRequest(http.MethodGet, url, nil)
			tc.setupAuth(t, req, server.maker, username, time.Hour)
			resp := httptest.NewRecorder()
			server.router.ServeHTTP(resp, req)
			tc.matchResult(t, resp)
		})
	}
}

type Query struct {
	pageID  int
	pageSize int
}

func TestListAccountAPI(t *testing.T) {

	username := util.RandomOwner()

	n := 5
	accounts := make([]db.Account, n)
	for i := 0; i < n; i++ {
		accounts[i] = randomAccount(username)
	}

	testCases := []struct {
		name string
		query Query
		setupAuth   func(t *testing.T, req *http.Request, tokenMaker token.Maker, username string, duration time.Duration) 
		buildStore  func(*testing.T, *mockdb.MockStore)
		matchResult func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name:      "ok",
			query: Query {
				pageID: 1,
				pageSize: n,
			},
			setupAuth: func(t *testing.T, req *http.Request, tokenMaker token.Maker, username string, duration time.Duration) {
				addAuthHeader(t, req, tokenMaker, username, duration)
			},
			buildStore: func(t *testing.T, store *mockdb.MockStore) {
				arg := db.ListAccountsParams {
					Owner: username,
					Limit: int32(n),
					Offset: 0,
				}
				store.EXPECT().ListAccounts(gomock.Any(), gomock.Eq(arg)).Return(accounts, nil).Times(1)
			},
			matchResult: func(t *testing.T, resp *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, resp.Code)
				matchReturnedAccounts(t, resp.Body, accounts)
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockStore := mockdb.NewMockStore(ctrl)
			tc.buildStore(t, mockStore)

			server := newTestServer(t, mockStore)

			url := "/accounts"
			req := httptest.NewRequest(http.MethodGet, url, nil)

			q := req.URL.Query()
			q.Add("page_id", fmt.Sprintf("%d", tc.query.pageID))
			q.Add("page_size", fmt.Sprintf("%d", tc.query.pageSize))
			req.URL.RawQuery = q.Encode()

			tc.setupAuth(t, req, server.maker, username, time.Hour)
			resp := httptest.NewRecorder()
			server.router.ServeHTTP(resp, req)
			tc.matchResult(t, resp)
		})
	}
}

func randomAccount(user string) db.Account {
	return db.Account{
		ID:       util.RandomInt(1, 1000),
		Owner:    user,
		Balance:  util.RandomMoney(),
		Currency: util.RandomCurrency(),
	}
}

func matchReturnedAccount(t *testing.T, buffer *bytes.Buffer, account *db.Account) {

	data, err := io.ReadAll(buffer)
	require.NoError(t, err)
	var gotAccount db.Account
	err = json.Unmarshal(data, &gotAccount)
	require.NoError(t, err)
	require.Equal(t, gotAccount.ID, account.ID)
	require.Equal(t, gotAccount.Owner, account.Owner)
	require.Equal(t, gotAccount.Balance, account.Balance)
	require.Equal(t, gotAccount.Currency, account.Currency)
}

func matchReturnedAccounts(t *testing.T, buffer *bytes.Buffer, accounts []db.Account) {

	data, err := io.ReadAll(buffer)
	require.NoError(t, err)

	var gotAccounts []db.Account
	err = json.Unmarshal(data, &gotAccounts)
	require.NoError(t, err)
	require.Equal(t, len(gotAccounts), len(accounts))
	for i := range(gotAccounts) {
		require.Equal(t, gotAccounts[i].ID, accounts[i].ID)
		require.Equal(t, gotAccounts[i].Owner, accounts[i].Owner)
		require.Equal(t, gotAccounts[i].Balance, accounts[i].Balance)
		require.Equal(t, gotAccounts[i].Currency, accounts[i].Currency)
	}
}
