package api

import (
	"fmt"
	"github.com/HL/meta-bank/token"
	"github.com/HL/meta-bank/util"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func addAuthorization(t *testing.T, tokenMaker token.Maker, request *http.Request, username string, role string, duration time.Duration, tokenType string) {

	token, payload, err := tokenMaker.CreateToken(username, role, duration)
	require.NoError(t, err)
	require.NotEmpty(t, payload)

	authorizationHeader := fmt.Sprintf("%s %s", tokenType, token)
	request.Header.Set(authorizationHeaderKey, authorizationHeader)

}

func TestAuthMiddleware(t *testing.T) {

	owner := util.RandomOwner()
	role := util.DEPOSITOR

	testCase := []struct {
		name          string
		setupAuth     func(t *testing.T, tokenMaker token.Maker, request *http.Request)
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			setupAuth: func(t *testing.T, tokenMaker token.Maker, request *http.Request) {
				addAuthorization(t, tokenMaker, request, owner, role, time.Minute, authorizationTypeBearer)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
			},
		},
		{
			name: "NotAuthorization",
			setupAuth: func(t *testing.T, tokenMaker token.Maker, request *http.Request) {
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
		},

		{
			name: "UnsupportedAuthorization",
			setupAuth: func(t *testing.T, tokenMaker token.Maker, request *http.Request) {
				addAuthorization(t, tokenMaker, request, owner, role, time.Minute, "unsupported")
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
		},

		{
			name: "InvalidAuthorizationHeaderFormat",
			setupAuth: func(t *testing.T, tokenMaker token.Maker, request *http.Request) {
				addAuthorization(t, tokenMaker, request, owner, role, time.Minute, "")
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
		},
		{
			name: "ExpiredToken",
			setupAuth: func(t *testing.T, tokenMaker token.Maker, request *http.Request) {
				addAuthorization(t, tokenMaker, request, owner, role, -time.Minute, authorizationTypeBearer)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
		},
	}

	for i := range testCase {
		tc := testCase[i]

		t.Run(tc.name, func(t *testing.T) {
			server := newTestServer(t)
			authPath := "/auth"
			server.router.GET(authPath, server.authMiddleware(server.tokenMaker), func(context *gin.Context) {
				context.JSON(http.StatusOK, gin.H{})

			})
			recorder := httptest.NewRecorder()

			request, err := http.NewRequest(http.MethodGet, authPath, nil)
			require.NoError(t, err)

			tc.setupAuth(t, server.tokenMaker, request)
			server.router.ServeHTTP(recorder, request)
			fmt.Println("recorder is:", recorder)
			tc.checkResponse(t, recorder)
		})
	}
}
