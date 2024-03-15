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

	mockdb "github.com/ashokmouli/simplebank/db/mock"
	db "github.com/ashokmouli/simplebank/db/sqlc"
	"github.com/ashokmouli/simplebank/db/util"
	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
)

func randomUser(t *testing.T) (db.User, string) {
	password := util.RandomString(6)
	hashedPassword, err := util.HashPassword(password)
	require.NoError(t, err)

	user := db.User{
		Username:       util.RandomOwner(),
		HashedPassword: hashedPassword,
		FullName:       util.RandomString(6),
		Email:          util.RandomEmail(),
	}
	return user, password
}

// This struct is used to provide a Custom Matcher interface
type eqCreateUserParamsMatcher struct {
	expected         db.CreateUserParams
	expectedPassword string
}

func (e eqCreateUserParamsMatcher) Matches(x interface{}) bool {

	actual, ok := x.(db.CreateUserParams)
	if !ok {
		return false
	}
	err := util.CheckPassword(e.expectedPassword, actual.HashedPassword)
	if err != nil {
		return false
	}
	e.expected.HashedPassword = actual.HashedPassword
	m := gomock.Eq(e.expected)
	return m.Matches(x)
}

func (e eqCreateUserParamsMatcher) String() string {
	return fmt.Sprintf("matches arg %v and password %v", e.expected, e.expectedPassword)
}

func EqCreateUserParams(arg db.CreateUserParams, password string) gomock.Matcher {
	return eqCreateUserParamsMatcher{
		expected:         arg,
		expectedPassword: password,
	}
}

func TestCreateUser(t *testing.T) {
	user, password := randomUser(t)

	testCases := []struct {
		name        string
		body        gin.H
		buildStore  func(*testing.T, *mockdb.MockStore)
		matchResult func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name: "ok",
			body: gin.H{
				"username": user.Username,
				"password": password,
				"fullname": user.FullName,
				"email":    user.Email,
			},
			buildStore: func(t *testing.T, mock *mockdb.MockStore) {
				arg := db.CreateUserParams{
					Username: user.Username,
					FullName: user.FullName,
					Email:    user.Email,
				}

				mock.EXPECT().CreateUser(gomock.Any(), EqCreateUserParams(arg, password)).Return(user, nil).Times(1)
			},
			matchResult: func(t *testing.T, resp *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, resp.Code)
				matchReturnedUser(t, resp.Body, &user)
			},
		},
		{
			name: "InternalServerError",
			body: gin.H{
				"username": user.Username,
				"password": password,
				"fullname": user.FullName,
				"email":    user.Email,
			},
			buildStore: func(t *testing.T, mock *mockdb.MockStore) {
				arg := db.CreateUserParams{
					Username: user.Username,
					FullName: user.FullName,
					Email:    user.Email,
				}
				// Custom matcher for checking the arguments. We need to pass clearText password  to a custom matcher.
				// The custom matcher will hash this cleartext password and compare it againts what the
				// the handler has produced.
				mock.EXPECT().CreateUser(gomock.Any(), EqCreateUserParams(arg, password)).Return(db.User{}, sql.ErrConnDone).Times(1)
			},
			matchResult: func(t *testing.T, resp *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, resp.Code)
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

			data, err := json.Marshal(tc.body)
			require.NoError(t, err)

			req := httptest.NewRequest(http.MethodPost, "/users", bytes.NewReader(data))
			resp := httptest.NewRecorder()
			server.router.ServeHTTP(resp, req)
			tc.matchResult(t, resp)
		})
	}

}

func TestLoginUser(t *testing.T) {
	user, password := randomUser(t)

	testCases := []struct {
		name        string
		body        gin.H
		buildStore  func(*testing.T, *mockdb.MockStore)
		matchResult func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name: "ok",
			body: gin.H{
				"username": user.Username,
				"password": password,
			},
			buildStore: func(t *testing.T, mock *mockdb.MockStore) {
				mock.EXPECT().GetUser(gomock.Any(), gomock.Eq(user.Username)).Return(user, nil).Times(1)
				mock.EXPECT().CreateSession(gomock.Any(), gomock.Any()).Times(1)
			},
			matchResult: func(t *testing.T, resp *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, resp.Code)
				matchReturnedLogin(t, resp.Body, &user)
			},
		},
		{
			name: "Forbidden",
			body: gin.H{
				"username": user.Username,
				"password": "secret",
			},
			buildStore: func(t *testing.T, mock *mockdb.MockStore) {
				mock.EXPECT().GetUser(gomock.Any(), gomock.Eq(user.Username)).Return(user, nil).Times(1)
			},
			matchResult: func(t *testing.T, resp *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusForbidden, resp.Code)
			},
		},
		{
			name: "NotFound",
			body: gin.H{
				"username": "nouser",
				"password": password,
			},
			buildStore: func(t *testing.T, mock *mockdb.MockStore) {
				mock.EXPECT().GetUser(gomock.Any(), gomock.Eq("nouser")).Return(db.User{}, sql.ErrNoRows).Times(1)
			},
			matchResult: func(t *testing.T, resp *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusNotFound, resp.Code)
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

			data, err := json.Marshal(tc.body)
			require.NoError(t, err)

			req := httptest.NewRequest(http.MethodPost, "/users/login", bytes.NewReader(data))
			resp := httptest.NewRecorder()
			server.router.ServeHTTP(resp, req)
			tc.matchResult(t, resp)
		})
	}

}

func matchReturnedUser(t *testing.T, buffer *bytes.Buffer, user *db.User) {

	data, err := io.ReadAll(buffer)
	require.NoError(t, err)
	var gotUser db.User
	err = json.Unmarshal(data, &gotUser)
	require.NoError(t, err)
	require.Equal(t, gotUser.Username, user.Username)
	require.Equal(t, gotUser.FullName, user.FullName)
	require.Equal(t, gotUser.Email, user.Email)
	require.Empty(t, gotUser.HashedPassword)
}

func matchReturnedLogin(t *testing.T, buffer *bytes.Buffer, user *db.User) {

	data, err := io.ReadAll(buffer)
	require.NoError(t, err)
	var got createLoginResponse
	err = json.Unmarshal(data, &got)
	require.NoError(t, err)
	require.NotEmpty(t, got.AccessToken)
	require.Equal(t, got.User.Username, user.Username)
	require.Equal(t, got.User.FullName, user.FullName)
	require.Equal(t, got.User.Email, user.Email)
}
