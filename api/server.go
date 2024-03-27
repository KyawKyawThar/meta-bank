package api

import (
	db "github.com/HL/meta-bank/db/sqlc"
	"github.com/HL/meta-bank/util"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
	"net/http"
)

// Server serve HTTP request for our app
type Server struct {
	store  db.Store
	router *gin.Engine
	config util.Config
}

// NewServer create a new http api and setup routing
func NewServer(store db.Store, config util.Config) (*Server, error) {

	server := &Server{
		store:  store,
		config: config,
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

	router.POST(util.CreateUser, s.createUser)
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
	return gin.H{"error": err}
}