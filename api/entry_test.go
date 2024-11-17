package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	mockdb "github.com/HL/meta-bank/db/mock"
	db "github.com/HL/meta-bank/db/sqlc"
	"github.com/HL/meta-bank/token"
	"github.com/HL/meta-bank/util"
	mockwk "github.com/HL/meta-bank/worker/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestGetEntryAPI(t *testing.T) {
	user, _ := randomUser(t)

	account := randomAccount(user.Username)

	entry := randomEntry(account.ID)
	testCases := []struct {
		name          string
		entryID       int64
		setupAuth     func(t *testing.T, request *http.Request, tokenMaker token.Maker)
		buildStubs    func(store *mockdb.MockStore)
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name:    "OK",
			entryID: entry.ID,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, tokenMaker, request, user.Username, user.Role, time.Minute, authorizationTypeBearer)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().GetEntry(gomock.Any(), entry.ID).Times(1).Return(entry, nil)
				store.EXPECT().GetAccount(gomock.Any(), account.ID).Times(1).Return(account, nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				requiredBodyMatchEntry(t, recorder.Body, entry)
			},
		},

		{
			name:    "UnauthorizedUser",
			entryID: entry.ID,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, tokenMaker, request, "unauthorizedUser", util.BANKER, time.Minute, authorizationTypeBearer)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().GetEntry(gomock.Any(), gomock.Any()).Times(1)
				store.EXPECT().GetAccount(gomock.Any(), gomock.Any()).Times(1)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recorder.Code)

			},
		},
		{
			name:    "NoAuthorization",
			entryID: entry.ID,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {

			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().GetEntry(gomock.Any(), gomock.Any()).Times(0)
				store.EXPECT().GetAccount(gomock.Any(), gomock.Any()).Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recorder.Code)

			},
		},

		{
			name:    "InvalidID",
			entryID: 0,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, tokenMaker, request, user.Username, user.Role, time.Minute, authorizationTypeBearer)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().GetEntry(gomock.Any(), gomock.Any()).Times(0)

			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)

			},
		},
		{
			name:    "NotFound",
			entryID: entry.ID,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, tokenMaker, request, user.Username, user.Role, time.Minute, authorizationTypeBearer)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().GetEntry(gomock.Any(), gomock.Eq(entry.ID)).Times(1).Return(db.Entry{}, ErrorRecordNotFound)
				store.EXPECT().GetAccount(gomock.Any(), gomock.Any()).Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusNotFound, recorder.Code)

			},
		},
		{
			name:    "accountNotFound",
			entryID: entry.ID,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, tokenMaker, request, user.Username, user.Role, time.Minute, authorizationTypeBearer)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().GetEntry(gomock.Any(), gomock.Eq(entry.ID)).Times(1).Return(entry, nil)
				store.EXPECT().GetAccount(gomock.Any(), gomock.Eq(account.ID)).Times(1).Return(db.Account{}, ErrorRecordNotFound)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusNotFound, recorder.Code)

			},
		},
		{
			name:    "InternalServerError",
			entryID: entry.ID,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, tokenMaker, request, user.Username, user.Role, time.Minute, authorizationTypeBearer)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().GetEntry(gomock.Any(), gomock.Eq(entry.ID)).Times(1).Return(db.Entry{}, ErrConnectionDone)
				store.EXPECT().GetAccount(gomock.Any(), gomock.Any()).Times(0)
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

			workerCtrl := gomock.NewController(t)
			defer workerCtrl.Finish()
			distributor := mockwk.NewMockTaskDistributor(workerCtrl)
			server := newTestServer(t, store, distributor)
			recorder := httptest.NewRecorder()

			url := fmt.Sprintf("/entries/%d", tc.entryID)

			request, err := http.NewRequest(http.MethodGet, url, nil)
			require.NoError(t, err)

			tc.setupAuth(t, request, server.tokenMaker)
			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(t, recorder)
		})
	}
}

func TestGetEntryListAPI(t *testing.T) {

	user, _ := randomUser(t)
	account := randomAccount(user.Username)

	n := 5

	entries := make([]db.Entry, n)

	for i := 0; i < n; i++ {
		entries[i] = randomEntry(account.ID)
	}

	type Query struct {
		AccountID int64
		PageID    int
		PageSize  int
	}

	testCases := []struct {
		name          string
		query         Query
		setupAuth     func(t *testing.T, tokenMaker token.Maker, request *http.Request)
		buildStubs    func(store *mockdb.MockStore)
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{name: "OK",
			query: Query{
				AccountID: account.ID,
				PageID:    1,
				PageSize:  n,
			},
			setupAuth: func(t *testing.T, tokenMaker token.Maker, request *http.Request) {
				addAuthorization(t, tokenMaker, request, user.Username, user.Role, time.Minute, authorizationTypeBearer)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().GetAccount(gomock.Any(), account.ID).Times(1).Return(account, nil)

				arg := db.ListEntriesParams{
					AccountID: account.ID,
					Limit:     int32(n),
					Offset:    0,
				}

				store.EXPECT().ListEntries(gomock.Any(), gomock.Eq(arg)).
					Times(1).Return(entries, nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				requiredBodyMatchEntries(t, recorder.Body, entries)
			},
		},
		{name: "NoAuthorization",
			query: Query{
				AccountID: account.ID,
				PageID:    1,
				PageSize:  n,
			},
			setupAuth: func(t *testing.T, tokenMaker token.Maker, request *http.Request) {
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().GetAccount(gomock.Any(), gomock.Any()).Times(0)

				store.EXPECT().ListEntries(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recorder.Code)

			},
		},

		{name: "UnAuthorizedUser",
			query: Query{
				AccountID: account.ID,
				PageID:    1,
				PageSize:  n,
			},
			setupAuth: func(t *testing.T, tokenMaker token.Maker, request *http.Request) {
				addAuthorization(t, tokenMaker, request, "UnauthorizedUser", util.BANKER, time.Minute, authorizationTypeBearer)
			},
			buildStubs: func(store *mockdb.MockStore) {

				store.EXPECT().GetAccount(gomock.Any(), gomock.Eq(account.ID)).
					Times(1).Return(account, nil)

				store.EXPECT().ListEntries(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recorder.Code)

			},
		},

		{name: "InvalidPageID",
			query: Query{
				AccountID: account.ID,
				PageID:    -1,
				PageSize:  n,
			},
			setupAuth: func(t *testing.T, tokenMaker token.Maker, request *http.Request) {
				addAuthorization(t, tokenMaker, request, user.Username, user.Role, time.Minute, authorizationTypeBearer)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().ListEntries(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)

			},
		},

		{name: "InvalidPageSize",
			query: Query{
				AccountID: account.ID,
				PageID:    1,
				PageSize:  1000,
			},
			setupAuth: func(t *testing.T, tokenMaker token.Maker, request *http.Request) {
				addAuthorization(t, tokenMaker, request, user.Username, user.Role, time.Minute, authorizationTypeBearer)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().ListEntries(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)

			},
		},

		{name: "AccountNotFound",
			query: Query{
				AccountID: account.ID,
				PageID:    1,
				PageSize:  n,
			},
			setupAuth: func(t *testing.T, tokenMaker token.Maker, request *http.Request) {
				addAuthorization(t, tokenMaker, request, user.Username, user.Role, time.Minute, authorizationTypeBearer)
			},
			buildStubs: func(store *mockdb.MockStore) {

				store.EXPECT().GetAccount(gomock.Any(), gomock.Eq(account.ID)).
					Times(1).Return(db.Account{}, ErrorRecordNotFound)

				store.EXPECT().ListEntries(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusNotFound, recorder.Code)

			},
		},

		{name: "InternalError",
			query: Query{
				AccountID: account.ID,
				PageID:    1,
				PageSize:  n,
			},
			setupAuth: func(t *testing.T, tokenMaker token.Maker, request *http.Request) {
				addAuthorization(t, tokenMaker, request, user.Username, user.Role, time.Minute, authorizationTypeBearer)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().GetAccount(gomock.Any(), account.ID).Times(1).Return(account, nil)

				store.EXPECT().ListEntries(gomock.Any(), gomock.Any()).
					Times(1).Return([]db.Entry{}, ErrConnectionDone)
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

			workerCtrl := gomock.NewController(t)
			defer workerCtrl.Finish()
			distributor := mockwk.NewMockTaskDistributor(workerCtrl)

			server := newTestServer(t, store, distributor)
			recorder := httptest.NewRecorder()

			url := fmt.Sprint(util.ListEntry)

			request, err := http.NewRequest(http.MethodGet, url, nil)
			require.NoError(t, err)

			tc.setupAuth(t, server.tokenMaker, request)

			//add query to request parameters
			q := request.URL.Query()
			q.Add("account_id", fmt.Sprintf("%d", tc.query.AccountID))
			q.Add("page_id", fmt.Sprintf("%d", tc.query.PageID))
			q.Add("page_size", fmt.Sprintf("%d", tc.query.PageSize))
			request.URL.RawQuery = q.Encode()

			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(t, recorder)
		})
	}
}

func randomEntry(accountID int64) db.Entry {
	return db.Entry{
		ID:        util.RandomInteger(1, 100),
		AccountID: accountID,
		Amount:    util.RandomAmount(),
	}
}

func requiredBodyMatchEntry(t *testing.T, body *bytes.Buffer, entry db.Entry) {
	b, err := io.ReadAll(body)
	require.NoError(t, err)

	var gotEntry db.Entry

	err = json.Unmarshal(b, &gotEntry)
	require.NoError(t, err)
	require.Equal(t, entry, gotEntry)

}

func requiredBodyMatchEntries(t *testing.T, body *bytes.Buffer, entry []db.Entry) {
	b, err := io.ReadAll(body)
	require.NoError(t, err)

	var gotEntries []db.Entry

	err = json.Unmarshal(b, &gotEntries)
	require.NoError(t, err)
	require.Equal(t, entry, gotEntries)

}
