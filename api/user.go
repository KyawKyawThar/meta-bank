package api

import (
	"errors"
	"fmt"
	db "github.com/HL/meta-bank/db/sqlc"
	"github.com/HL/meta-bank/token"
	"github.com/HL/meta-bank/util"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"net/http"
	"time"
)

type userRequest struct {
	Username string `json:"username" binding:"required,alphanum"`
	Password string `json:"password" binding:"required,min=7"`
	Email    string `json:"email" binding:"required,email"`
	FullName string `json:"fullName" binding:"required"`
	Role     string `json:"role" `
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
	SessionID             uuid.UUID    `json:"session_id"`
	AccessToken           string       `json:"access_token"`
	RefreshToken          string       `json:"refresh_token"`
	AccessTokenExpiresAt  time.Time    `json:"access_token_expires_at"`
	RefreshTokenExpiresAt time.Time    `json:"refresh_token_expires_at"`
	User                  userResponse `json:"user"`
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

	//arg := db.CreateTxUserParams{
	//	CreateUserParams: db.CreateUserParams{
	//		Username: req.Username,
	//		Password: hashPassword,
	//		Email:    req.Email,
	//		FullName: req.FullName,
	//		Role:     req.Role,
	//		IsActive: req.IsActive,
	//	},
	//
	//	//AfterCreate func will be called after the use is created to send an
	//	// async task for email verification to Redis(DistributorSendVerifyEmail will be called).
	//	AfterCreate: func(user db.User) error {
	//		//asynq.ProcessIn(10 * time.Second) 10 sec delay mean task will only be pickup by worker after 10sec is created
	//
	//		opts := []asynq.Option{
	//			asynq.MaxRetry(10),
	//			asynq.ProcessIn(10 * time.Second), //delay is very import
	//			asynq.Queue(worker.QueueCritical),
	//		}
	//
	//		payload := &worker.PayloadSendVerifyEmail{Username: user.Username}
	//		err := s.taskDistributor.DistributorSendVerifyEmail(ctx, payload, opts...)
	//		//if err != nil {
	//		//	err := fmt.Errorf("failed to distribute task to send verify email %w", err)
	//		//	ctx.JSON(http.StatusInternalServerError, handleErrorResponse(err))
	//		//
	//		//}
	//		return err
	//	},
	//}
	//
	//txResult, err := s.store.CreateUserTx(ctx, arg)
	//
	//if err != nil {
	//	handleDBErrResponse(ctx, err)
	//	return
	//}

	//user, err := s.store.CreateUser(ctx, arg)
	//opts := []asynq.Option{
	//	asynq.MaxRetry(10),
	//	asynq.ProcessIn(10 * time.Second),
	//	asynq.Queue(worker.QueueCritical),
	//}
	//
	//payload := &worker.PayloadSendVerifyEmail{Username: user.Username}
	//err = s.taskDistributor.DistributorSendVerifyEmail(ctx, payload, opts...)
	//
	//if err != nil {
	//	err := fmt.Errorf("failed to distribute task to send verify email %w", err)
	//	ctx.JSON(http.StatusInternalServerError, handleErrorResponse(err))
	//	return
	//}
	//
	//res := newUserResponse(txResult.User)
	//ctx.JSON(http.StatusOK, res)

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
		ctx.JSON(http.StatusInternalServerError, handleErrorResponse(err))
		return
	}

	fmt.Println("content type", ctx.ContentType())
	refreshToken, refreshPayload, err := s.tokenMaker.CreateToken(user.Username, user.Role, s.config.RefreshTokenDuration)

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, handleErrorResponse(err))
		return
	}

	session, err := s.store.CreateSession(ctx, db.CreateSessionParams{
		ID:           refreshPayload.ID,
		Username:     refreshPayload.Username,
		RefreshToken: refreshToken,
		UserAgent:    ctx.Request.UserAgent(),
		ClientIp:     ctx.ClientIP(),
		IsBlocked:    false,
		ExpiredAt:    refreshPayload.ExpiredAt,
	})

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, handleErrorResponse(err))
		return
	}
	res := loginResponse{
		SessionID:             session.ID,
		AccessToken:           accessToken,
		RefreshToken:          session.RefreshToken,
		AccessTokenExpiresAt:  accessPayload.ExpiredAt,
		RefreshTokenExpiresAt: session.ExpiredAt,
		User:                  newUserResponse(user),
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

	if authPayload.Username != user.Username {

		err := errors.New("request user doesn't belong to the authenticated user")
		ctx.JSON(http.StatusUnauthorized, handleCustomErrorResponse(err))
	}

	res := newUserResponse(user)

	ctx.JSON(http.StatusOK, res)
}
