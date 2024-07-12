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
	"io"
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

		{name: "InternalServerError",
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

				store.EXPECT().TransferTx(gomock.Any(), gomock.Eq(arg)).Times(1).Return(db.TransferTxResult{}, ErrConnectionDone)

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

func TestGetTransferAPI(t *testing.T) {

	user1, _ := randomUser(t)
	user2, _ := randomUser(t)

	account1 := randomAccount(user1.Username)
	account2 := randomAccount(user2.Username)

	transfer := randomTransfer(account1.ID, account2.ID)

	testCase := []struct {
		name          string
		transferID    int64
		setupAuth     func(t *testing.T, request *http.Request, tokenMaker token.Maker)
		buildStubs    func(store *mockdb.MockStore)
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name:       "OK",
			transferID: transfer.ID,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, tokenMaker, request, user1.Username, user1.Role, time.Minute, authorizationTypeBearer)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().GetTransfer(gomock.Any(), transfer.ID).Times(1).Return(transfer, nil)
				store.EXPECT().GetAccount(gomock.Any(), account1.ID).Times(1).Return(account1, nil)

			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				requiredBodyMatchTransfer(t, recorder.Body, transfer)
			},
		},

		{
			name:       "UnauthorizedUser",
			transferID: transfer.ID,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, tokenMaker, request, user2.Username, user2.Role, time.Minute, authorizationTypeBearer)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().GetTransfer(gomock.Any(), transfer.ID).Times(1).Return(transfer, nil)
				store.EXPECT().GetAccount(gomock.Any(), account1.ID).Times(1).Return(account1, nil)

			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recorder.Code)

			},
		},
		{
			name:       "NoAuthorization",
			transferID: transfer.ID,
			setupAuth:  func(t *testing.T, request *http.Request, tokenMaker token.Maker) {},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().GetTransfer(gomock.Any(), transfer.ID).Times(0)
				store.EXPECT().GetAccount(gomock.Any(), account1.ID).Times(0)

			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recorder.Code)

			},
		},
		{
			name:       "InvalidID",
			transferID: 0,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, tokenMaker, request, user1.Username, user1.Role, time.Minute, authorizationTypeBearer)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().GetTransfer(gomock.Any(), transfer.ID).Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)

			},
		},
		{
			name:       "TransferNotFound",
			transferID: transfer.ID,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, tokenMaker, request, user1.Username, user1.Role, time.Minute, authorizationTypeBearer)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().GetTransfer(gomock.Any(), gomock.Eq(transfer.ID)).Times(1).Return(db.Transfer{}, ErrorRecordNotFound)
				store.EXPECT().GetAccount(gomock.Any(), gomock.Any()).Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusNotFound, recorder.Code)
			},
		},
		{
			name:       "AccountNotFound",
			transferID: transfer.ID,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, tokenMaker, request, user1.Username, user1.Role, time.Minute, authorizationTypeBearer)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().GetTransfer(gomock.Any(), transfer.ID).Times(1).Return(transfer, nil)
				store.EXPECT().GetAccount(gomock.Any(), gomock.Any()).Times(1).Return(db.Account{}, ErrorRecordNotFound)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusNotFound, recorder.Code)
			},
		},

		{
			name:       "InternalServerError",
			transferID: transfer.ID,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, tokenMaker, request, user1.Username, user1.Role, time.Minute, authorizationTypeBearer)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().GetTransfer(gomock.Any(), gomock.Eq(transfer.ID)).Times(1).Return(db.Transfer{}, ErrConnectionDone)
				store.EXPECT().GetAccount(gomock.Any(), gomock.Any()).Times(0)
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
			tc.buildStubs(store)

			server := newTestServer(t, store)
			recorder := httptest.NewRecorder()

			url := fmt.Sprintf("/transfer/%d", tc.transferID)

			request, err := http.NewRequest(http.MethodGet, url, nil)
			require.NoError(t, err)

			tc.setupAuth(t, request, server.tokenMaker)
			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(t, recorder)
		})
	}
}

func TestGetTransferListAPI(t *testing.T) {
	user1, _ := randomUser(t)
	user2, _ := randomUser(t)

	account1 := randomAccount(user1.Username)
	account2 := randomAccount(user2.Username)

	n := 5

	transfers := make([]db.Transfer, n)

	for i := 0; i < n; i++ {
		transfers[i] = randomTransfer(account1.ID, account2.ID)
	}

	type Query struct {
		FromAccountID int64
		ToAccountID   int64
		PageID        int
		PageSize      int
	}

	testCases := []struct {
		name          string
		query         Query
		setupAuth     func(t *testing.T, tokenMaker token.Maker, request *http.Request)
		buildStubs    func(store *mockdb.MockStore)
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			query: Query{
				FromAccountID: account1.ID,
				ToAccountID:   account2.ID,
				PageID:        1,
				PageSize:      n,
			},

			setupAuth: func(t *testing.T, tokenMaker token.Maker, request *http.Request) {
				addAuthorization(t, tokenMaker, request, user1.Username, user1.Role, time.Minute, authorizationTypeBearer)
			},
			buildStubs: func(store *mockdb.MockStore) {

				store.EXPECT().GetAccount(gomock.Any(), account1.ID).Times(1).Return(account1, nil)

				arg := db.ListTransfersParams{
					FromAccountID: account1.ID,
					Limit:         int32(n),
					Offset:        0,
				}

				store.EXPECT().ListTransfers(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(transfers, nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				requiredBodyMatchTransfers(t, recorder.Body, transfers)
			},
		},
		{
			name: "NoAuthorization",
			query: Query{
				FromAccountID: account1.ID,
				ToAccountID:   account2.ID,
				PageID:        1,
				PageSize:      n,
			},

			setupAuth: func(t *testing.T, tokenMaker token.Maker, request *http.Request) {
			},
			buildStubs: func(store *mockdb.MockStore) {

				store.EXPECT().GetAccount(gomock.Any(), gomock.Any()).Times(0)

				store.EXPECT().ListTransfers(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recorder.Code)

			},
		},

		{
			name: "NoAuthorizedUser",
			query: Query{
				FromAccountID: account1.ID,
				ToAccountID:   account2.ID,
				PageID:        1,
				PageSize:      n,
			},

			setupAuth: func(t *testing.T, tokenMaker token.Maker, request *http.Request) {
				addAuthorization(t, tokenMaker, request, "unAuthorizedUser", util.BANKER, time.Minute, authorizationTypeBearer)
			},
			buildStubs: func(store *mockdb.MockStore) {

				store.EXPECT().GetAccount(gomock.Any(), account1.ID).Times(1).Return(account1, nil)

				store.EXPECT().ListTransfers(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recorder.Code)

			},
		},
		{
			name: "InvalidPageID",
			query: Query{
				FromAccountID: account1.ID,
				ToAccountID:   account2.ID,
				PageID:        -1,
				PageSize:      n,
			},

			setupAuth: func(t *testing.T, tokenMaker token.Maker, request *http.Request) {
				addAuthorization(t, tokenMaker, request, user1.Username, user1.Role, time.Minute, authorizationTypeBearer)
			},
			buildStubs: func(store *mockdb.MockStore) {

				store.EXPECT().GetAccount(gomock.Any(), gomock.Any()).Times(0)

				store.EXPECT().ListTransfers(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)

			},
		},

		{
			name: "InvalidPageSize",
			query: Query{
				FromAccountID: account1.ID,
				ToAccountID:   account2.ID,
				PageID:        1,
				PageSize:      1000,
			},

			setupAuth: func(t *testing.T, tokenMaker token.Maker, request *http.Request) {
				addAuthorization(t, tokenMaker, request, user1.Username, user1.Role, time.Minute, authorizationTypeBearer)
			},
			buildStubs: func(store *mockdb.MockStore) {

				store.EXPECT().GetAccount(gomock.Any(), gomock.Any()).Times(0)

				store.EXPECT().ListTransfers(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)

			},
		},

		{
			name: "AccountNotFound",
			query: Query{
				FromAccountID: account1.ID,
				ToAccountID:   account2.ID,
				PageID:        1,
				PageSize:      n,
			},

			setupAuth: func(t *testing.T, tokenMaker token.Maker, request *http.Request) {
				addAuthorization(t, tokenMaker, request, user1.Username, user1.Role, time.Minute, authorizationTypeBearer)
			},
			buildStubs: func(store *mockdb.MockStore) {

				store.EXPECT().GetAccount(gomock.Any(), gomock.Eq(account1.ID)).Times(1).Return(db.Account{}, ErrorRecordNotFound)

				store.EXPECT().ListTransfers(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusNotFound, recorder.Code)

			},
		},

		{
			name: "InternalServerError",
			query: Query{
				FromAccountID: account1.ID,
				ToAccountID:   account2.ID,
				PageID:        1,
				PageSize:      n,
			},

			setupAuth: func(t *testing.T, tokenMaker token.Maker, request *http.Request) {
				addAuthorization(t, tokenMaker, request, user1.Username, user1.Role, time.Minute, authorizationTypeBearer)
			},
			buildStubs: func(store *mockdb.MockStore) {

				store.EXPECT().GetAccount(gomock.Any(), account1.ID).Times(1).Return(account1, nil)

				arg := db.ListTransfersParams{
					FromAccountID: account1.ID,
					Limit:         int32(n),
					Offset:        0,
				}

				store.EXPECT().ListTransfers(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return([]db.Transfer{}, ErrConnectionDone)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)

			},
		},
	}

	for i := range testCases {
		tc := testCases[i]
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)

			store := mockdb.NewMockStore(ctrl)
			tc.buildStubs(store)

			server := newTestServer(t, store)
			recorder := httptest.NewRecorder()

			url := fmt.Sprint(util.ListTransfer)

			request, err := http.NewRequest(http.MethodGet, url, nil)
			require.NoError(t, err)

			tc.setupAuth(t, server.tokenMaker, request)

			//add query to request parameters
			q := request.URL.Query()
			q.Add("from_account_id", fmt.Sprintf("%d", tc.query.FromAccountID))
			q.Add("to_account_id", fmt.Sprintf("%d", tc.query.ToAccountID))
			q.Add("page_id", fmt.Sprintf("%d", tc.query.PageID))
			q.Add("page_size", fmt.Sprintf("%d", tc.query.PageSize))
			request.URL.RawQuery = q.Encode()

			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(t, recorder)
		})
	}
}

func randomTransfer(fromAccountID, toAccountID int64) db.Transfer {
	return db.Transfer{
		ID:            util.RandomInteger(1, 100),
		FromAccountID: fromAccountID,
		ToAccountID:   toAccountID,
		Amount:        util.RandomAmount(),
	}
}

func requiredBodyMatchTransfer(t *testing.T, body *bytes.Buffer, transfer db.Transfer) {
	b, err := io.ReadAll(body)
	require.NoError(t, err)

	var gotTransfer db.Transfer

	err = json.Unmarshal(b, &gotTransfer)
	require.NoError(t, err)
	require.Equal(t, transfer, gotTransfer)

}

func requiredBodyMatchTransfers(t *testing.T, body *bytes.Buffer, transfer []db.Transfer) {
	b, err := io.ReadAll(body)
	require.NoError(t, err)

	var gotTransfers []db.Transfer

	err = json.Unmarshal(b, &gotTransfers)
	require.NoError(t, err)
	require.Equal(t, transfer, gotTransfers)

}
