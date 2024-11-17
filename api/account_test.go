package api

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	mockdb "github.com/HL/meta-bank/db/mock"
	db "github.com/HL/meta-bank/db/sqlc"
	"github.com/HL/meta-bank/token"
	"github.com/HL/meta-bank/util"
	mockwk "github.com/HL/meta-bank/worker/mock"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestCreateAccountAPI(t *testing.T) {

	user, _ := randomUser(t)

	account := randomAccount(user.Username)

	testCase := []struct {
		name          string
		body          gin.H
		setUpAuth     func(t *testing.T, request *http.Request, tokenMaker token.Maker)
		buildStub     func(store *mockdb.MockStore)
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			body: gin.H{
				"currency": account.Currency,
				"balance":  account.Balance,
			},
			setUpAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, tokenMaker, request, user.Username, user.Role, time.Minute, authorizationTypeBearer)
			},
			buildStub: func(store *mockdb.MockStore) {

				arg := db.CreateAccountParams{
					Owner:    user.Username,
					Currency: account.Currency,
					Balance:  account.Balance,
				}
				store.EXPECT().CreateAccount(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(account, nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				requiredBodyMatchAccount(t, recorder.Body, account)
			},
		},
		{
			name: "NoAuthorization",
			body: gin.H{
				"currency": account.Currency,
				"balance":  account.Balance,
			},
			setUpAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {},
			buildStub: func(store *mockdb.MockStore) {
				store.EXPECT().CreateAccount(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
		},
		{
			name: "InternalError",
			body: gin.H{
				"currency": account.Currency,
				"balance":  account.Balance,
			},
			setUpAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, tokenMaker, request, user.Username, user.Role, time.Minute, authorizationTypeBearer)
			},
			buildStub: func(store *mockdb.MockStore) {
				store.EXPECT().
					CreateAccount(gomock.Any(), gomock.Any()).
					Times(1).
					Return(db.Account{}, sql.ErrConnDone)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			name: "InvalidCurrency",
			body: gin.H{
				"currency": "invalid",
				"balance":  account.Balance,
			},
			setUpAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, tokenMaker, request, user.Username, user.Role, time.Minute, authorizationTypeBearer)
			},
			buildStub: func(store *mockdb.MockStore) {
				store.EXPECT().
					CreateAccount(gomock.Any(), gomock.Any()).
					Times(0)

			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
	}

	for i := range testCase {
		tc := testCase[i]

		t.Run(tc.name, func(t *testing.T) {

			ctrl := gomock.NewController(t)

			store := mockdb.NewMockStore(ctrl)

			workerCtrl := gomock.NewController(t)
			defer workerCtrl.Finish()
			distributor := mockwk.NewMockTaskDistributor(workerCtrl)
			tc.buildStub(store)

			server := newTestServer(t, store, distributor)
			recorder := httptest.NewRecorder()

			// Marshal body data to JSON
			data, err := json.Marshal(tc.body)
			require.NoError(t, err)

			url := fmt.Sprint(util.CreateAccount)
			request, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(data))
			require.NoError(t, err)

			tc.setUpAuth(t, request, server.tokenMaker)
			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(t, recorder)
		})

	}
}

func TestGetAccountAPI(t *testing.T) {

	user, _ := randomUser(t)

	account := randomAccount(user.Username)

	testCase := []struct {
		name          string
		accountID     int64
		setupAuth     func(t *testing.T, request *http.Request, tokenMaker token.Maker)
		buildStubs    func(store *mockdb.MockStore)
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name:      "OK",
			accountID: account.ID,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, tokenMaker, request, user.Username, user.Role, time.Minute, authorizationTypeBearer)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Eq(account.ID)).
					Times(1).
					Return(account, nil)

			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				requiredBodyMatchAccount(t, recorder.Body, account)

			},
		},
		{
			name:      "UnauthorizedUser",
			accountID: account.ID,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {

				addAuthorization(t, tokenMaker, request, "unauthorized_user", util.DEPOSITOR, time.Minute, authorizationTypeBearer)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().GetAccount(gomock.Any(), gomock.Eq(account.ID)).Times(1).Return(account, nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
		},
		{
			name:      "NoAuthorization",
			accountID: account.ID,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().GetAccount(gomock.Any(), gomock.Any()).Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
		},
		{
			name:      "InvalidID",
			accountID: 0,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, tokenMaker, request, user.Username, user.Role, time.Minute, authorizationTypeBearer)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().GetAccount(gomock.Any(), account.ID).Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name:      "NotFound",
			accountID: account.ID,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, tokenMaker, request, user.Username, user.Role, time.Minute, authorizationTypeBearer)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().GetAccount(gomock.Any(), gomock.Eq(account.ID)).Times(1).Return(db.Account{}, ErrorRecordNotFound)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusNotFound, recorder.Code)
			},
		},
		{
			name:      "InternalError",
			accountID: account.ID,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, tokenMaker, request, user.Username, user.Role, time.Minute, authorizationTypeBearer)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().GetAccount(gomock.Any(), gomock.Eq(account.ID)).Times(1).Return(db.Account{}, sql.ErrConnDone)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
	}

	for i := range testCase {
		tc := testCase[i]
		t.Run(tc.name, func(t *testing.T) {

			ctrl := gomock.NewController(t)
			store := mockdb.NewMockStore(ctrl)

			workerCtrl := gomock.NewController(t)
			defer workerCtrl.Finish()
			distributor := mockwk.NewMockTaskDistributor(workerCtrl)
			tc.buildStubs(store)

			server := newTestServer(t, store, distributor)
			recorder := httptest.NewRecorder()

			url := fmt.Sprintf("/account/%d", tc.accountID)
			request, err := http.NewRequest(http.MethodGet, url, nil)
			require.NoError(t, err)

			tc.setupAuth(t, request, server.tokenMaker)
			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(t, recorder)
		})
	}

}

func TestGetAccountsListAPI(t *testing.T) {
	user, _ := randomUser(t)
	n := 5

	accounts := make([]db.Account, n)

	for i := 0; i < n; i++ {
		accounts[i] = randomAccount("nicholas")
	}

	type Query struct {
		PageID   int
		PageSize int
	}

	testCases := []struct {
		name          string
		query         Query
		setupAuth     func(t *testing.T, tokenMaker token.Maker, request *http.Request)
		buildStubs    func(store *mockdb.MockStore)
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{

		{
			name: "Ok",
			query: Query{
				PageID:   1,
				PageSize: n,
			},
			setupAuth: func(t *testing.T, tokenMaker token.Maker, request *http.Request) {
				addAuthorization(t, tokenMaker, request, user.Username, user.Role, time.Minute, authorizationTypeBearer)
			},
			buildStubs: func(store *mockdb.MockStore) {
				arg := db.ListAccountParams{
					Owner:  user.Username,
					Limit:  int32(n),
					Offset: 0,
				}
				store.EXPECT().ListAccount(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(accounts, nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				requireBodyMathAccounts(t, recorder.Body, accounts)
			},
		},

		{
			name: "NoAuthorization",
			query: Query{
				PageID:   1,
				PageSize: n,
			},
			setupAuth: func(t *testing.T, tokenMaker token.Maker, request *http.Request) {

			},
			buildStubs: func(store *mockdb.MockStore) {

				store.EXPECT().ListAccount(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recorder.Code)

			},
		},
		{
			name: "InternalError",
			query: Query{
				PageID:   1,
				PageSize: n,
			},
			setupAuth: func(t *testing.T, tokenMaker token.Maker, request *http.Request) {
				addAuthorization(t, tokenMaker, request, user.Username, user.Role, time.Minute, authorizationTypeBearer)
			},
			buildStubs: func(store *mockdb.MockStore) {
				arg := db.ListAccountParams{
					Owner:  user.Username,
					Limit:  int32(n),
					Offset: 0,
				}
				store.EXPECT().ListAccount(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return([]db.Account{}, sql.ErrConnDone)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)

			},
		},
		{
			name: "InvalidPageID",
			query: Query{
				PageID:   -1,
				PageSize: n,
			},
			setupAuth: func(t *testing.T, tokenMaker token.Maker, request *http.Request) {
				addAuthorization(t, tokenMaker, request, user.Username, user.Role, time.Minute, authorizationTypeBearer)
			},
			buildStubs: func(store *mockdb.MockStore) {

				store.EXPECT().ListAccount(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)

			},
		},
		{
			name: "InvalidPageSize",
			query: Query{
				PageID:   1,
				PageSize: 1000,
			},
			setupAuth: func(t *testing.T, tokenMaker token.Maker, request *http.Request) {
				addAuthorization(t, tokenMaker, request, user.Username, user.Role, time.Minute, authorizationTypeBearer)
			},
			buildStubs: func(store *mockdb.MockStore) {

				store.EXPECT().ListAccount(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)

			},
		},
	}

	for i := range testCases {
		tc := testCases[i]

		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			store := mockdb.NewMockStore(ctrl)

			tc.buildStubs(store)

			workerCtrl := gomock.NewController(t)
			defer workerCtrl.Finish()
			distributor := mockwk.NewMockTaskDistributor(workerCtrl)

			server := newTestServer(t, store, distributor)
			recorder := httptest.NewRecorder()

			url := fmt.Sprintf(util.ListAccount)
			request, err := http.NewRequest(http.MethodGet, url, nil)
			require.NoError(t, err)

			tc.setupAuth(t, server.tokenMaker, request)

			//add query to request parameters
			q := request.URL.Query()
			q.Add("page_id", fmt.Sprintf("%d", tc.query.PageID))
			q.Add("page_size", fmt.Sprintf("%d", tc.query.PageSize))
			request.URL.RawQuery = q.Encode()

			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(t, recorder)
		})
	}

}

func randomAccount(owner string) db.Account {

	return db.Account{
		ID:       util.RandomInteger(1, 1000),
		Owner:    owner,
		Currency: util.RandomCurrency(),
		Balance:  util.RandomAmount(),
	}
}

func requiredBodyMatchAccount(t *testing.T, body *bytes.Buffer, account db.Account) {
	b, err := io.ReadAll(body)
	require.NoError(t, err)

	var gotAccount db.Account

	err = json.Unmarshal(b, &gotAccount)
	require.NoError(t, err)
	require.Equal(t, account, gotAccount)

}

func requireBodyMathAccounts(t *testing.T, body *bytes.Buffer, accounts []db.Account) {
	b, err := io.ReadAll(body)
	require.NoError(t, err)

	var gotAccounts []db.Account

	err = json.Unmarshal(b, &gotAccounts)
	require.NoError(t, err)
	require.Equal(t, accounts, gotAccounts)

}
