package api

import (
	"errors"
	"fmt"
	db "github.com/HL/meta-bank/db/sqlc"
	"github.com/HL/meta-bank/token"
	"github.com/HL/meta-bank/util"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"net/http"
	"time"
)

type userRequest struct {
	Username string `json:"username" binding:"required,alphanum"`
	Password string `json:"password" binding:"required,min=7"`
	Email    string `json:"email" binding:"required,email"`
	FullName string `json:"fullName" binding:"required"`
	Role     string `json:"role" binding:"required,role"`
	IsActive bool   `json:"isActive"`
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

type loginRequest struct {
	Username string `json:"username" binding:"required,alphanum"`
	Password string `json:"password" binding:"required,min=7"`
}

type loginResponse struct {
	AccessToken          string           `json:"access_token"`
	AccessTokenExpiresAt *jwt.NumericDate `json:"access_token_expires_at"`
	User                 userResponse     `json:"user"`
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

	req.IsActive = true

	if err := ctx.ShouldBindJSON(&req); err != nil {
		handleUserValidationErrResponse(ctx, err)
		return
	}

	hashPassword, err := util.HashPassword(req.Password)

	if err != nil {
		ctx.JSON(http.StatusUnauthorized, handleErrorResponse(err))
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
		handleDBErrResponse(ctx, err)
		return
	}
	res := newUserResponse(user)
	ctx.JSON(http.StatusOK, res)
}

func (s *Server) loginUser(ctx *gin.Context) {

	var req loginRequest

	if err := ctx.ShouldBindJSON(&req); err != nil {
		handleUserValidationErrResponse(ctx, err)
		return
	}

	user, err := s.store.GetUser(ctx, req.Username)
	if err != nil {
		handleDBErrResponse(ctx, err)
		return
	}

	err = util.CheckPassword(req.Password, user.Password)

	if err != nil {
		ctx.JSON(http.StatusUnauthorized, handleErrorResponse(err))
		return
	}

	accessToken, accessPayload, err := s.tokenMaker.CreateToken(user.Username, user.Role, s.config.AccessTokenDuration)

	if err != nil {
		fmt.Println("code run in here")
		ctx.JSON(http.StatusInternalServerError, handleErrorResponse(err))
		return
	}

	res := loginResponse{
		AccessToken:          accessToken,
		AccessTokenExpiresAt: accessPayload.ExpiresAt,
		User:                 newUserResponse(user),
	}

	ctx.JSON(http.StatusOK, res)
}

type getUserRequest struct {
	Username string `json:"username" binding:"required,alphanum"`
}

func (s *Server) getUser(ctx *gin.Context) {
	var req getUserRequest

	if err := ctx.ShouldBindJSON(&req); err != nil {

		handleUserValidationErrResponse(ctx, err)
		return
	}

	user, err := s.store.GetUser(ctx, req.Username)

	if err != nil {
		handleDBErrResponse(ctx, err)
		return
	}

	authPayload := ctx.MustGet(s.config.AuthorizationPayloadKey).(*token.Payload)

	if authPayload.Issuer != user.Username {

		err := errors.New("request user doesn't belong to the authenticated user")
		ctx.JSON(http.StatusUnauthorized, err)
	}

	res := newUserResponse(user)

	ctx.JSON(http.StatusOK, res)
}
