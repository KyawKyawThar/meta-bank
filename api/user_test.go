package api

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/HL/meta-bank/token"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"

	mockdb "github.com/HL/meta-bank/db/mock"
	db "github.com/HL/meta-bank/db/sqlc"
	"github.com/HL/meta-bank/util"
	"github.com/HL/meta-bank/worker"
	mockwk "github.com/HL/meta-bank/worker/mock"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

// eqCreateUserParamsMatcher is a custom matcher
type eqCreateUserParamsMatcher struct {
	arg      db.CreateTxUserParams
	password string
	user     db.User
}

func (e eqCreateUserParamsMatcher) Matches(x any) bool {
	arg, ok := x.(db.CreateTxUserParams)

	if !ok {
		return false
	}
	err := util.CheckPassword(e.password, arg.Password)

	if err != nil {
		return false
	}

	fmt.Println("kkt...", e.arg, ">>>>>", arg.Password)
	e.arg.Password = arg.Password

	if !reflect.DeepEqual(e.arg.CreateUserParams, arg.CreateUserParams) {
		return false
	}

	err = arg.AfterCreate(e.user)
	return err == nil
}

func (e eqCreateUserParamsMatcher) String() string {
	return fmt.Sprintf("matches arg %v and password %v", e.arg, e.password)

}

func EqCreateUserParams(arg db.CreateTxUserParams, password string, user db.User) gomock.Matcher {

	return eqCreateUserParamsMatcher{
		arg, password, user,
	}
}

func TestCreateUserAPI(t *testing.T) {

	user, password := randomUser(t)
	mockSecretCode := util.RandomString(32)
	testCase := []struct {
		name          string
		body          gin.H
		buildStub     func(store *mockdb.MockStore, taskDistributor *mockwk.MockTaskDistributor)
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			body: gin.H{
				"username": user.Username,
				"password": password,
				"email":    user.Email,
				"fullName": user.FullName,
			},

			buildStub: func(store *mockdb.MockStore, taskDistributor *mockwk.MockTaskDistributor) {
				arg := db.CreateTxUserParams{
					CreateUserParams: db.CreateUserParams{
						Username: user.Username,
						Email:    user.Email,
						FullName: user.FullName,
						IsActive: true,
					},
					AfterCreate: func(createUser db.User) error {

						if createUser.Username != user.Username {
							return errors.New("AfterCreate failed")
						}
						return nil
					},
				}
				store.EXPECT().CreateUserAndVerificationTx(gomock.Any(), EqCreateUserParams(arg, password, user)).Times(1).
					Return(db.CreateUserAndVerificationTxResult{
						User: user,
						VerifyEmail: db.VerifyEmail{

							Username:   user.Username,
							Email:      user.Email,
							SecretCode: mockSecretCode,
						},
					}, nil)

				taskPayload := &worker.PayloadSendVerifyEmail{Username: user.Username}
				taskDistributor.EXPECT().DistributorSendVerifyEmail(gomock.Any(), taskPayload, gomock.Any()).Times(1).Return(nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				requireBodyMatchUser(t, recorder.Body, user)
			},
		},

		{
			name: "DuplicateUsername",
			body: gin.H{
				"username": user.Username,
				"password": password,
				"email":    user.Email,
				"fullName": user.FullName,
			},

			buildStub: func(store *mockdb.MockStore, taskDistributor *mockwk.MockTaskDistributor) {

				store.EXPECT().CreateUserAndVerificationTx(gomock.Any(), gomock.Any()).Times(1).
					Return(db.CreateUserAndVerificationTxResult{
						User:        db.User{},
						VerifyEmail: db.VerifyEmail{},
					}, ErrorUniqueViolation)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusForbidden, recorder.Code)
			},
		},
		{
			name: "InvalidUsername",
			body: gin.H{
				"username": "invalid-User#1",
				"password": password,
				"email":    user.Email,
				"fullName": user.FullName,
			},

			buildStub: func(store *mockdb.MockStore, taskDistributor *mockwk.MockTaskDistributor) {

				store.EXPECT().CreateUserAndVerificationTx(gomock.Any(), gomock.Any()).Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "InvalidEmail",
			body: gin.H{
				"username": user.Username,
				"password": password,
				"email":    "invalid-email",
				"fullName": user.FullName,
			},
			buildStub: func(store *mockdb.MockStore, taskDistributor *mockwk.MockTaskDistributor) {
				store.EXPECT().CreateUser(gomock.Any(), gomock.Any()).Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "TooShortPassword",
			body: gin.H{
				"username": user.Username,
				"password": "123",
				"email":    user.Email,
				"fullName": user.FullName,
			},
			buildStub: func(store *mockdb.MockStore, taskDistributor *mockwk.MockTaskDistributor) {
				store.EXPECT().CreateUser(gomock.Any(), gomock.Any()).Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
	}

	for i := range testCase {
		tc := testCase[i]

		t.Run(tc.name, func(t *testing.T) {
			storeCtrl := gomock.NewController(t)
			defer storeCtrl.Finish()
			store := mockdb.NewMockStore(storeCtrl)

			workerCtrl := gomock.NewController(t)
			defer workerCtrl.Finish()
			distributor := mockwk.NewMockTaskDistributor(workerCtrl)

			tc.buildStub(store, distributor)

			server := newTestServer(t, store, distributor)
			recorder := httptest.NewRecorder()

			// Marshal body data to JSON
			data, err := json.Marshal(tc.body)
			require.NoError(t, err)

			request, err := http.NewRequest(http.MethodPost, util.CreateUser, bytes.NewReader(data))
			require.NoError(t, err)

			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(t, recorder)
		})
	}

}

func TestLoginUserAPI(t *testing.T) {

	user, password := randomUser(t)

	testCase := []struct {
		name          string
		body          gin.H
		buildStubs    func(store *mockdb.MockStore, taskDistributor *mockwk.MockTaskDistributor)
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name: "Ok",
			body: gin.H{
				"username": user.Username,
				"password": password,
			},
			buildStubs: func(store *mockdb.MockStore, taskDistributor *mockwk.MockTaskDistributor) {
				store.EXPECT().
					GetUser(gomock.Any(), gomock.Eq(user.Username)).
					Times(1).Return(user, nil)
				store.EXPECT().
					CreateSession(gomock.Any(), gomock.Any()).
					Times(1)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
			},
		},
		{
			name: "InternalError",
			body: gin.H{
				"username": user.Username,
				"password": password,
			},
			buildStubs: func(store *mockdb.MockStore, taskDistributor *mockwk.MockTaskDistributor) {
				store.EXPECT().
					GetUser(gomock.Any(), gomock.Any()).
					Times(1).Return(db.User{}, sql.ErrConnDone)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			name: "InvalidUsername",
			body: gin.H{
				"username": "invalid-#1",
				"password": password,
			},
			buildStubs: func(store *mockdb.MockStore, taskDistributor *mockwk.MockTaskDistributor) {
				store.EXPECT().
					GetUser(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "UserNotFound",
			body: gin.H{
				"username": "NotFound",
				"password": password,
			},
			buildStubs: func(store *mockdb.MockStore, taskDistributor *mockwk.MockTaskDistributor) {
				store.EXPECT().
					GetUser(gomock.Any(), gomock.Eq("NotFound")).
					Times(1).Return(db.User{}, ErrorRecordNotFound)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusNotFound, recorder.Code)
			},
		},
		{
			name: "IncorrectPassword",
			body: gin.H{
				"username": user.Username,
				"password": password,
			},
			buildStubs: func(store *mockdb.MockStore, taskDistributor *mockwk.MockTaskDistributor) {
				store.EXPECT().GetUser(gomock.Any(), gomock.Any()).Times(1).Return(db.User{}, nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
		},
	}

	for i := range testCase {
		tc := testCase[i]

		t.Run(tc.name, func(t *testing.T) {
			storeCtrl := gomock.NewController(t)
			defer storeCtrl.Finish()
			store := mockdb.NewMockStore(storeCtrl)

			workerCtrl := gomock.NewController(t)
			defer workerCtrl.Finish()
			distributor := mockwk.NewMockTaskDistributor(workerCtrl)

			tc.buildStubs(store, distributor)

			server := newTestServer(t, store, distributor)
			recorder := httptest.NewRecorder()

			// Marshal body data to JSON
			data, err := json.Marshal(tc.body)
			require.NoError(t, err)

			request, err := http.NewRequest(http.MethodPost, util.LoginUser, bytes.NewReader(data))
			require.NoError(t, err)

			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(t, recorder)
		})
	}
}

func TestGetUserAPI(t *testing.T) {

	user, _ := randomUser(t)

	testCase := []struct {
		name          string
		body          gin.H
		setupAuth     func(t *testing.T, request *http.Request, tokenMaker token.Maker)
		buildStub     func(store *mockdb.MockStore, taskDistributor *mockwk.MockTaskDistributor)
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{

		{
			name: "Ok",
			body: gin.H{
				"username": user.Username,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, tokenMaker, request, user.Username, user.Role, time.Minute, authorizationTypeBearer)
			},
			buildStub: func(store *mockdb.MockStore, taskDistributor *mockwk.MockTaskDistributor) {
				store.EXPECT().
					GetUser(gomock.Any(), gomock.Eq(user.Username)).
					Times(1).Return(user, nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				requireBodyMatchUser(t, recorder.Body, user)
			},
		},

		{
			name: "InternalError",
			body: gin.H{
				"username": user.Username,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, tokenMaker, request, user.Username, user.Role, time.Minute, authorizationTypeBearer)
			},
			buildStub: func(store *mockdb.MockStore, taskDistributor *mockwk.MockTaskDistributor) {
				store.EXPECT().
					GetUser(gomock.Any(), gomock.Any()).
					Times(1).Return(db.User{}, sql.ErrConnDone)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)

			},
		},
		{
			name: "UnauthorizedUser",
			body: gin.H{
				"username": user.Username,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, tokenMaker, request, "unauthorized_user", user.Role, time.Minute, authorizationTypeBearer)
			},
			buildStub: func(store *mockdb.MockStore, taskDistributor *mockwk.MockTaskDistributor) {
				store.EXPECT().
					GetUser(gomock.Any(), gomock.Eq(user.Username)).
					Times(1).Return(user, nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recorder.Code)

			},
		},
		{
			name: "NoAuthorization",
			body: gin.H{
				"username": user.Username,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {

			},
			buildStub: func(store *mockdb.MockStore, taskDistributor *mockwk.MockTaskDistributor) {
				store.EXPECT().
					GetUser(gomock.Any(), gomock.Eq(user.Username)).
					Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recorder.Code)

			},
		},
		{
			name: "InvalidUsername",
			body: gin.H{
				"username": "invalid-#1",
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, tokenMaker, request, user.Username, user.Role, time.Minute, authorizationTypeBearer)
			},
			buildStub: func(store *mockdb.MockStore, taskDistributor *mockwk.MockTaskDistributor) {
				store.EXPECT().
					GetUser(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "UserNotFound",
			body: gin.H{
				"username": "NotFound",
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, tokenMaker, request, user.Username, user.Role, time.Minute, authorizationTypeBearer)
			},
			buildStub: func(store *mockdb.MockStore, taskDistributor *mockwk.MockTaskDistributor) {
				store.EXPECT().
					GetUser(gomock.Any(), gomock.Eq("NotFound")).
					Times(1).Return(db.User{}, ErrorRecordNotFound)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusNotFound, recorder.Code)
			},
		},
	}

	for i := range testCase {
		tc := testCase[i]

		t.Run(tc.name, func(t *testing.T) {
			storeCtrl := gomock.NewController(t)
			defer storeCtrl.Finish()
			store := mockdb.NewMockStore(storeCtrl)

			workerCtrl := gomock.NewController(t)
			defer workerCtrl.Finish()
			distributor := mockwk.NewMockTaskDistributor(workerCtrl)

			tc.buildStub(store, distributor)

			server := newTestServer(t, store, distributor)
			recorder := httptest.NewRecorder()

			// Marshal body data to JSON
			data, err := json.Marshal(tc.body)
			require.NoError(t, err)

			request, err := http.NewRequest(http.MethodGet, util.GetUser, bytes.NewReader(data))
			require.NoError(t, err)

			tc.setupAuth(t, request, server.tokenMaker)

			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(t, recorder)
		})
	}
}

func randomUser(t *testing.T) (user db.User, password string) {
	password = util.RandomString(8)

	hashedPassword, err := util.HashPassword(password)
	require.NoError(t, err)

	user = db.User{
		Username: util.RandomOwner(),
		Password: hashedPassword,
		Email:    util.RandomEmail(),
		FullName: util.RandomOwner(),
	}
	return

}

func requireBodyMatchUser(t *testing.T, body *bytes.Buffer, user db.User) {
	b, err := io.ReadAll(body)
	require.NoError(t, err)

	var gotUser db.User

	err = json.Unmarshal(b, &gotUser)
	fmt.Println("got user: ", gotUser)
	require.NoError(t, err)
	require.Equal(t, user.Username, gotUser.Username)
	require.Equal(t, user.FullName, gotUser.FullName)
	require.Equal(t, user.Email, gotUser.Email)
	require.Empty(t, gotUser.Password)

}
