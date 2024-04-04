package api

import (
	"github.com/HL/meta-bank/util"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"os"
	"testing"
	"time"
)

const (
	authorizationTypeBearer = "bearer"
	authorizationHeaderKey  = "authorization"
)

func newTestServer(t *testing.T) *Server {

	config := util.Config{TokenSymmetricKey: util.RandomString(32),
		AccessTokenDuration:     time.Minute,
		AuthorizationTypeBearer: authorizationTypeBearer,
		AuthorizationHeaderKey:  authorizationHeaderKey}

	server, err := NewServer(nil, config)

	require.NoError(t, err)

	return server
}

func TestMain(m *testing.M) {

	gin.SetMode(gin.TestMode)

	os.Exit(m.Run())
}
