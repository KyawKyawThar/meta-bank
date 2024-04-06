package api

import (
	"errors"
	db "github.com/HL/meta-bank/db/sqlc"
	"github.com/HL/meta-bank/token"
	"github.com/gin-gonic/gin"
	"net/http"
)

type createAccountRequest struct {
	Currency string `json:"currency" binding:"required,currency"`
	Balance  int64  `json:"balance" binding:"required,gt=0"`
}

func (s *Server) createAccount(ctx *gin.Context) {

	var req createAccountRequest

	if err := ctx.ShouldBindJSON(&req); err != nil {
		handleUserValidationErrResponse(ctx, err)
		return
	}

	authPayload := ctx.MustGet(s.config.AuthorizationPayloadKey).(*token.Payload)

	arg := db.CreateAccountParams{
		Owner:    authPayload.Username,
		Currency: req.Currency,
		Balance:  req.Balance,
	}

	acc, err := s.store.CreateAccount(ctx, arg)

	if err != nil {
		handleDBErrResponse(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, acc)
}

type accountRequestUri struct {
	ID int64 `uri:"id" binding:"required,min=1"`
}

func (s *Server) getAccount(ctx *gin.Context) {
	var uri accountRequestUri

	if err := ctx.ShouldBindUri(&uri); err != nil {
		handleUserValidationErrResponse(ctx, err)
		return
	}

	acc, err := s.store.GetAccount(ctx, uri.ID)
	if err != nil {
		handleDBErrResponse(ctx, err)
		return
	}

	authPayload := ctx.MustGet(s.config.AuthorizationPayloadKey).(*token.Payload)

	if authPayload.Username != acc.Owner {

		err := errors.New("account doesn't belong to the authenticated user")
		ctx.JSON(http.StatusUnauthorized, err)
	}

	ctx.JSON(http.StatusOK, acc)
}

type updateAccountRequest struct {
	//ID      int64 `uri:"id" binding:"required,min=1"`
	Balance int64 `json:"balance" binding:"required,gte=0"`
}

func (s *Server) updateAccount(ctx *gin.Context) {

	var req updateAccountRequest
	var uri accountRequestUri

	if err := ctx.ShouldBindUri(&uri); err != nil {
		handleUserValidationErrResponse(ctx, err)
		return
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {

		handleUserValidationErrResponse(ctx, err)
		return
	}

	arg := db.UpdateAccountParams{
		ID:     uri.ID,
		Amount: req.Balance,
	}

	acc, err := s.store.UpdateAccount(ctx, arg)

	if err != nil {
		handleDBErrResponse(ctx, err)
		return
	}

	authPayload := ctx.MustGet(s.config.AuthorizationPayloadKey).(*token.Payload)

	if authPayload.Username != acc.Owner {

		err := errors.New("account doesn't belong to the authenticated user")
		ctx.JSON(http.StatusUnauthorized, err)
	}

	ctx.JSON(http.StatusOK, acc)
}
