package api

import (
	db "github.com/HL/meta-bank/db/sqlc"
	"github.com/HL/meta-bank/util"
	"github.com/HL/meta-bank/worker"
	"github.com/gin-gonic/gin"
	"github.com/hibiken/asynq"
	"github.com/stretchr/testify/require"
	"os"
	"testing"
	"time"
)

const (
	authorizationTypeBearer = "bearer"
	authorizationHeaderKey  = "authorization"
)

func newTestServer(t *testing.T, store db.Store) *Server {

	config := util.Config{TokenSymmetricKey: util.RandomString(32),
		AccessTokenDuration:     time.Minute,
		AuthorizationTypeBearer: authorizationTypeBearer,
		AuthorizationHeaderKey:  authorizationHeaderKey}

	redsOpts := asynq.RedisClientOpt{
		Addr: config.RedisAddress,
	}
	taskDistributor := worker.NewRedisTaskDistributor(redsOpts)
	server, err := NewServer(store, config, taskDistributor)

	//server, err := NewServer(store, config)
	require.NoError(t, err)

	return server
}

func TestMain(m *testing.M) {

	gin.SetMode(gin.TestMode)

	os.Exit(m.Run())
}
