package api

import (
	"fmt"
	db "github.com/HL/meta-bank/db/sqlc"
	"github.com/HL/meta-bank/token"
	"github.com/HL/meta-bank/util"
	"github.com/HL/meta-bank/worker"
	ginzerolog "github.com/dn365/gin-zerolog"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"net/http"
	"os"
)

// Server serve HTTP request for our app
type Server struct {
	store           db.Store
	router          *gin.Engine
	config          util.Config
	tokenMaker      token.Maker
	taskDistributor worker.TaskDistributor
}

// NewServer create a new http api and setup routing
func NewServer(store db.Store, config util.Config, taskDistributor worker.TaskDistributor) (*Server, error) {

	maker, err := token.NewJWTMaker(config.TokenSymmetricKey)

	if err != nil {
		return nil, fmt.Errorf("cannot create token maker %w", err)
	}

	server := &Server{
		store:           store,
		config:          config,
		tokenMaker:      maker,
		taskDistributor: taskDistributor,
	}
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		v.RegisterValidation("role", validateRole)
		v.RegisterValidation("currency", validateCurrency)
	}

	server.setUpRouter()
	return server, nil
}

// setUpRouter setup for different HTTP methods
func (s *Server) setUpRouter() {
	router := gin.Default()

	if s.config.Environment == "development" {
		router.Use(ginzerolog.Logger("GIN"))
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	}

	router.POST(util.CreateUser, s.createUser)
	router.POST(util.LoginUser, s.loginUser)

	router.POST(util.RenewToken, s.renewAccessToken)

	authRoutes := router.Group("/").Use(s.authMiddleware(s.tokenMaker))

	authRoutes.GET(util.GetUser, s.getUser)

	authRoutes.POST(util.CreateAccount, s.createAccount)
	authRoutes.GET(util.GetAccount, s.getAccount)
	s.router = router
}

// Start return the HTTP api on a specific route
func (s *Server) Start(address string) error {

	return s.router.Run(address)

}

func handleDBErrResponse(c *gin.Context, err error) {
	message, statusCode := GetMessageFromDBError(err)
	c.JSON(statusCode, gin.H{"Error": message})
}

func handleUserValidationErrResponse(c *gin.Context, err error) {
	message := GetMessageFromUserValidationError(err)
	c.JSON(http.StatusBadRequest, gin.H{"Error:": message})
}

func handleErrorResponse(err error) gin.H {
	return gin.H{"error": err.Error()}
}
