package api

import (
	"errors"
	db "github.com/HL/meta-bank/db/sqlc"
	"github.com/HL/meta-bank/token"
	"github.com/gin-gonic/gin"
	"net/http"
)

type accountRequest struct {
	Currency string `json:"currency" binding:"required,currency"`
	Balance  int64  `json:"balance" binding:"required,gt=0"`
}

func (s *Server) createAccount(ctx *gin.Context) {

	var req accountRequest

	if err := ctx.ShouldBindJSON(&req); err != nil {
		handleUserValidationErrResponse(ctx, err)
		return
	}

	authPayload := ctx.MustGet(s.config.AuthorizationPayloadKey).(*token.Payload)

	arg := db.CreateAccountParams{
		Owner:    authPayload.Issuer,
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

type getAccountRequest struct {
	ID int64 `uri:"id" binding:"required,min=1"`
}

func (s *Server) getAccount(ctx *gin.Context) {
	var req getAccountRequest

	if err := ctx.ShouldBindUri(&req); err != nil {
		handleUserValidationErrResponse(ctx, err)
		return
	}

	acc, err := s.store.GetAccount(ctx, req.ID)
	if err != nil {
		handleDBErrResponse(ctx, err)
		return
	}

	authPayload := ctx.MustGet(s.config.AuthorizationPayloadKey).(*token.Payload)

	if authPayload.Issuer != acc.Owner {

		err := errors.New("account doesn't belong to the authenticated user")
		ctx.JSON(http.StatusUnauthorized, err)
	}

	ctx.JSON(http.StatusOK, acc)
}
