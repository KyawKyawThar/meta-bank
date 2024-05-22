package api

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"time"
)

type renewAccessTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

type renewAccessTokenResponse struct {
	AccessToken          string    `json:"access_token"`
	AccessTokenExpiredAt time.Time `json:"access_token_expired_at"`
}

func (s *Server) renewAccessToken(ctx *gin.Context) {

	var req renewAccessTokenRequest

	if err := ctx.ShouldBindJSON(&req); err != nil {
		handleUserValidationErrResponse(ctx, err)
		return
	}

	refreshPayload, err := s.tokenMaker.VerifyToken(req.RefreshToken)

	if err != nil {
		ctx.JSON(http.StatusUnauthorized, handleErrorResponse(err))
		return
	}

	session, err := s.store.GetSession(ctx, refreshPayload.ID)
	if err != nil {
		handleDBErrResponse(ctx, err)
		return
	}

	if session.IsBlocked {
		err := errors.New("blocked Sessions")
		ctx.JSON(http.StatusBadRequest, handleErrorResponse(err))
		return

	}

	if session.Username != refreshPayload.Username {
		err := fmt.Errorf("incorrect session user")
		ctx.JSON(http.StatusForbidden, handleErrorResponse(err))
	}

	if session.RefreshToken != req.RefreshToken {
		err := fmt.Errorf("mismatched session token")
		ctx.JSON(http.StatusUnauthorized, handleErrorResponse(err))
	}

	if time.Now().After(session.ExpiredAt) {
		err := errors.New("expired session")
		ctx.JSON(http.StatusBadRequest, handleErrorResponse(err))
		return
	}

	accessToken, accessPayload, err := s.tokenMaker.CreateToken(refreshPayload.Username, refreshPayload.Role, s.config.AccessTokenDuration)

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, handleErrorResponse(err))
		return
	}

	res := renewAccessTokenResponse{
		AccessToken:          accessToken,
		AccessTokenExpiredAt: accessPayload.ExpiredAt,
	}

	ctx.JSON(http.StatusOK, res)
}
