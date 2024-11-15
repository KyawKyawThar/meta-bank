package api

import (
	"errors"
	"net/http"

	db "github.com/HL/meta-bank/db/sqlc"
	"github.com/HL/meta-bank/token"
	"github.com/gin-gonic/gin"
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
		ctx.JSON(http.StatusUnauthorized, handleCustomErrorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, acc)
}

type listAccountRequest struct {
	PageID   int32 `form:"page_id" binding:"required,min=1"`
	PageSize int32 `form:"page_size" binding:"required,min=3,max=10"`
}

func (s *Server) ListAccount(ctx *gin.Context) {

	var req listAccountRequest

	if err := ctx.ShouldBindQuery(&req); err != nil {

		if req.PageID < 1 {

			err := errors.New("PageID must be at least 1")
			ctx.JSON(http.StatusBadRequest, handleCustomErrorResponse(err))
			return
		}
		if req.PageSize < 5 || req.PageSize > 10 {

			err := errors.New("PageSize must be between 3 and 10")
			ctx.JSON(http.StatusBadRequest, handleCustomErrorResponse(err))
			return
		}
	}

	authPayload := ctx.MustGet(s.config.AuthorizationPayloadKey).(*token.Payload)

	arg := db.ListAccountParams{
		Owner:  authPayload.Username,
		Limit:  req.PageSize,
		Offset: (req.PageID - 1) * req.PageSize,
	}

	lists, err := s.store.ListAccount(ctx, arg)

	if err != nil {
		handleDBErrResponse(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, lists)
}
