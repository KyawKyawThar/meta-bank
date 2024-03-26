package api

import (
	db "github.com/HL/meta-bank/db/sqlc"
	"github.com/HL/meta-bank/util"
	"github.com/gin-gonic/gin"
	"net/http"
	"time"
)

type userRequest struct {
	Username string `json:"username" binding:"required,alphanum"`
	Password string `json:"password" binding:"required,min=7"`
	Email    string `json:"email" binding:"required,email"`
	FullName string `json:"fullName" binding:"required"`
	Role     string `json:"role" binding:"required,role"`
	IsActive bool   `json:"isActive" binding:"required"`
}

type userResponse struct {
	Username          string    `json:"username"`
	Email             string    `json:"email"`
	FullName          string    `json:"full_name"`
	Role              string    `json:"role"`
	IsActive          bool      `json:"is_active"`
	PasswordChangedAt time.Time `json:"password_changed_at"`
	CreatedAt         time.Time `json:"created_at"`
}

func newUserResponse(user db.User) userResponse {
	return userResponse{
		Username:          user.Username,
		Email:             user.Email,
		FullName:          user.FullName,
		Role:              user.Role,
		IsActive:          user.IsActive,
		PasswordChangedAt: user.PasswordChangedAt,
		CreatedAt:         user.CreatedAt,
	}
}

func (s *Server) createUser(ctx *gin.Context) {

	var req userRequest

	if err := ctx.ShouldBindJSON(&req); err != nil {
		handleUserValidationErrResponse(ctx, err)
		return
	}

	hashPassword, err := util.HashPassword(req.Password)

	if err != nil {
		ctx.JSON(http.StatusBadRequest, handleErrorResponse(err))
		return
	}

	arg := db.CreateUserParams{
		Username: req.Username,
		Password: hashPassword,
		Email:    req.Email,
		FullName: req.FullName,
		Role:     req.Role,
		IsActive: req.IsActive,
	}

	user, err := s.store.CreateUser(ctx, arg)

	if err != nil {
		//message, statusCode := db.GetMessageFromDBError(err)
		//ctx.JSON(statusCode, handleDBErrResponse(message))
		handleDBErrResponse(ctx, err)
		return
	}
	res := newUserResponse(user)
	ctx.JSON(http.StatusOK, res)
}
