package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	mockdb "github.com/HL/meta-bank/db/mock"
	db "github.com/HL/meta-bank/db/sqlc"
	"github.com/HL/meta-bank/token"
	"github.com/HL/meta-bank/util"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestCreateTransferAPI(t *testing.T) {
	amount := int64(10)

	user1, _ := randomUser(t)
	user2, _ := randomUser(t)
	user3, _ := randomUser(t)

	account1 := randomAccount(user1.Username)
	account2 := randomAccount(user2.Username)
	account3 := randomAccount(user3.Username)

	account1.Currency = util.USD
	account2.Currency = util.USD
	account3.Currency = util.SGD

	testCase := []struct {
		name          string
		body          gin.H
		setUpAuth     func(t *testing.T, tokenMaker token.Maker, request *http.Request)
		buildStubs    func(store *mockdb.MockStore)
		checkResponse func(recorder *httptest.ResponseRecorder)
	}{
		{name: "Ok",
			body: gin.H{
				"transfer_account_id": account1.ID,
				"receive_account_id":  account2.ID,
				"amount":              amount,
				"currency":            util.USD,
			},
			setUpAuth: func(t *testing.T, tokenMaker token.Maker, request *http.Request) {
				addAuthorization(t, tokenMaker, request, user1.Username, user1.Role, time.Minute, authorizationTypeBearer)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().GetAccount(gomock.Any(), account1.ID).Times(1).Return(account1, nil)
				store.EXPECT().GetAccount(gomock.Any(), account2.ID).Times(1).Return(account2, nil)

				arg := db.TransferTxParams{
					TransferAccountID: account1.ID,
					ReceiveAccountID:  account2.ID,
					Amount:            amount,
				}

				store.EXPECT().TransferTx(gomock.Any(), gomock.Eq(arg)).Times(1)

			},

			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
			},
		},
		{name: "NoAuthorization",
			body: gin.H{
				"transfer_account_id": account1.ID,
				"receive_account_id":  account2.ID,
				"amount":              amount,
				"currency":            util.USD,
			},
			setUpAuth: func(t *testing.T, tokenMaker token.Maker, request *http.Request) {

			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().GetAccount(gomock.Any(), gomock.Any()).Times(0)
				store.EXPECT().TransferTx(gomock.Any(), gomock.Any()).Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
		},
		{
			name: "transferAccountNotFound",
			body: gin.H{
				"transfer_account_id": account1.ID,
				"receive_account_id":  account2.ID,
				"amount":              amount,
				"currency":            util.USD,
			},
			setUpAuth: func(t *testing.T, tokenMaker token.Maker, request *http.Request) {
				addAuthorization(t, tokenMaker, request, user1.Username, user1.Role, time.Minute, authorizationTypeBearer)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().GetAccount(gomock.Any(), gomock.Eq(account1.ID)).Times(1).Return(db.Account{}, ErrorRecordNotFound)
				store.EXPECT().GetAccount(gomock.Any(), gomock.Eq(account2.ID)).Times(0)
				store.EXPECT().TransferTx(gomock.Any(), gomock.Any()).Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusNotFound, recorder.Code)
			},
		},

		{
			name: "receiverAccountNotFound",
			body: gin.H{
				"transfer_account_id": account1.ID,
				"receive_account_id":  account2.ID,
				"amount":              amount,
				"currency":            util.USD,
			},
			setUpAuth: func(t *testing.T, tokenMaker token.Maker, request *http.Request) {
				addAuthorization(t, tokenMaker, request, user1.Username, user1.Role, time.Minute, authorizationTypeBearer)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().GetAccount(gomock.Any(), gomock.Eq(account1.ID)).Times(1).Return(account1, nil)
				store.EXPECT().GetAccount(gomock.Any(), gomock.Eq(account2.ID)).Times(1).Return(db.Account{}, ErrorRecordNotFound)
				store.EXPECT().TransferTx(gomock.Any(), gomock.Any()).Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusNotFound, recorder.Code)
			},
		},

		{name: "UnauthorizedUser",
			body: gin.H{
				"transfer_account_id": account1.ID,
				"receive_account_id":  account2.ID,
				"amount":              amount,
				"currency":            util.USD,
			},
			setUpAuth: func(t *testing.T, tokenMaker token.Maker, request *http.Request) {
				addAuthorization(t, tokenMaker, request, user2.Username, user2.Role, time.Minute, authorizationTypeBearer)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().GetAccount(gomock.Any(), gomock.Eq(account1.ID)).Times(1).Return(account1, nil)

				store.EXPECT().GetAccount(gomock.Any(), gomock.Eq(account2.ID)).Times(1).Return(db.Account{}, nil)
				store.EXPECT().TransferTx(gomock.Any(), gomock.Any()).Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recorder.Code)

			},
		},
		{
			name: "doNotHaveAmount",
			body: gin.H{
				"transfer_account_id": account1.ID,
				"receive_account_id":  account2.ID,
				"amount":              0,
				"currency":            util.USD,
			},
			setUpAuth: func(t *testing.T, tokenMaker token.Maker, request *http.Request) {
				addAuthorization(t, tokenMaker, request, user1.Username, user1.Role, time.Minute, authorizationTypeBearer)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().GetAccount(gomock.Any(), gomock.Eq(account3.ID)).Times(0)
				store.EXPECT().TransferTx(gomock.Any(), gomock.Any()).Times(0)
			},

			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)

			},
		},
		{
			name: "transferAccountCurrencyMismatch",
			body: gin.H{
				"transfer_account_id": account3.ID,
				"receive_account_id":  account1.ID,
				"amount":              amount,
				"currency":            util.USD,
			},
			setUpAuth: func(t *testing.T, tokenMaker token.Maker, request *http.Request) {
				addAuthorization(t, tokenMaker, request, user3.Username, user3.Role, time.Minute, authorizationTypeBearer)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().GetAccount(gomock.Any(), gomock.Eq(account3.ID)).Times(1).Return(account3, nil)
				store.EXPECT().GetAccount(gomock.Any(), gomock.Eq(account1.ID)).Times(0)
				store.EXPECT().TransferTx(gomock.Any(), gomock.Any()).Times(0)
			},

			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)

			},
		},
		{
			name: "receiverAccountCurrencyMismatch",
			body: gin.H{
				"transfer_account_id": account1.ID,
				"receive_account_id":  account3.ID,
				"amount":              amount,
				"currency":            util.USD,
			},
			setUpAuth: func(t *testing.T, tokenMaker token.Maker, request *http.Request) {
				addAuthorization(t, tokenMaker, request, user1.Username, user1.Role, time.Minute, authorizationTypeBearer)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().GetAccount(gomock.Any(), gomock.Eq(account1.ID)).Times(1).Return(account1, nil)
				store.EXPECT().GetAccount(gomock.Any(), gomock.Eq(account3.ID)).Times(1).Return(account3, nil)
				store.EXPECT().TransferTx(gomock.Any(), gomock.Any()).Times(0)
			},

			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "InvalidCurrency",
			body: gin.H{
				"transfer_account_id": account1.ID,
				"receive_account_id":  account2.ID,
				"amount":              amount,
				"currency":            "xyz",
			},
			setUpAuth: func(t *testing.T, tokenMaker token.Maker, request *http.Request) {
				addAuthorization(t, tokenMaker, request, user1.Username, user1.Role, time.Minute, authorizationTypeBearer)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().GetAccount(gomock.Any(), gomock.Any()).Times(0)
				store.EXPECT().TransferTx(gomock.Any(), gomock.Any()).Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "NegativeAmount",
			body: gin.H{
				"transfer_account_id": account1.ID,
				"receive_account_id":  account2.ID,
				"amount":              -amount,
				"currency":            util.USD,
			},
			setUpAuth: func(t *testing.T, tokenMaker token.Maker, request *http.Request) {
				addAuthorization(t, tokenMaker, request, user1.Username, user1.Role, time.Minute, authorizationTypeBearer)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().GetAccount(gomock.Any(), gomock.Any()).Times(0)
				store.EXPECT().TransferTx(gomock.Any(), gomock.Any()).Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "GetAccountError",
			body: gin.H{
				"transfer_account_id": account1.ID,
				"receive_account_id":  account2.ID,
				"amount":              amount,
				"currency":            util.USD,
			},
			setUpAuth: func(t *testing.T, tokenMaker token.Maker, request *http.Request) {
				addAuthorization(t, tokenMaker, request, user1.Username, user1.Role, time.Minute, authorizationTypeBearer)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().GetAccount(gomock.Any(), gomock.Any()).Times(1).Return(db.Account{}, ErrConnectionDone)
				store.EXPECT().TransferTx(gomock.Any(), gomock.Any()).Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
	}

	for i := range testCase {
		tc := testCase[i]
		t.Run(tc.name, func(t *testing.T) {

			ctrl := gomock.NewController(t)
			store := mockdb.NewMockStore(ctrl)

			tc.buildStubs(store)

			server := newTestServer(t, store)
			recorder := httptest.NewRecorder()

			// Marshal body data to JSON
			data, err := json.Marshal(tc.body)
			require.NoError(t, err)

			url := fmt.Sprintf(util.CreateTransfer)
			request, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(data))
			require.NoError(t, err)

			tc.setUpAuth(t, server.tokenMaker, request)
			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(recorder)
		})
	}
}
