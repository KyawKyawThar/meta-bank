package api

import (
	"errors"
	"fmt"
	db "github.com/HL/meta-bank/db/sqlc"
	"github.com/HL/meta-bank/token"
	"github.com/gin-gonic/gin"
	"net/http"
)

type transferTxRequest struct {
	TransferAccountID int64  `json:"transfer_account_id" binding:"required,min=1"`
	ReceiveAccountID  int64  `json:"receive_account_id" binding:"required,min=1"`
	Amount            int64  `json:"amount" binding:"required,gt=0"`
	Currency          string `json:"currency" binding:"required,currency"`
}

func (s *Server) createTransferTx(ctx *gin.Context) {

	var req transferTxRequest

	if err := ctx.ShouldBindJSON(&req); err != nil {
		handleUserValidationErrResponse(ctx, err)
		return
	}

	authPayload := ctx.MustGet(s.config.AuthorizationPayloadKey).(*token.Payload)

	transferAccount, valid := s.validAccount(ctx, req.TransferAccountID, req.Currency, req.Amount)

	if !valid {
		return
	}

	if authPayload.Username != transferAccount.Owner {

		err := errors.New("request user doesn't belong to the authenticated user")
		ctx.JSON(http.StatusUnauthorized, handleCustomErrorResponse(err))
	}

	_, valid = s.validAccount(ctx, req.ReceiveAccountID, req.Currency, req.Amount)

	if !valid {
		return
	}

	arg := db.TransferTxParams{
		TransferAccountID: req.TransferAccountID,
		ReceiveAccountID:  req.ReceiveAccountID,
		Amount:            req.Amount,
	}

	result, err := s.store.TransferTx(ctx, arg)

	if err != nil {
		handleDBErrResponse(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, result)

}

func (s *Server) validAccount(ctx *gin.Context, accountID int64, currency string, amount int64) (db.Account, bool) {

	account, err := s.store.GetAccount(ctx, accountID)

	if err != nil {
		handleDBErrResponse(ctx, err)
		return account, false
	}

	if account.Balance <= 0 || amount > account.Balance {
		err := fmt.Errorf("your amount is %d and you don't have sufficient balance to transform this action", account.Balance)

		ctx.JSON(http.StatusBadRequest, handleCustomErrorResponse(err))

		return account, false
	}

	if account.Currency != currency {
		err := fmt.Errorf("account name %s is currency mismatch: %s vs %s", account.Owner, account.Currency, currency)

		ctx.JSON(http.StatusBadRequest, handleCustomErrorResponse(err))

		return account, false
	}

	return account, true
}

type transferRequestUri struct {
	ID int64 `uri:"id" binding:"required,min=1"`
}

func (s *Server) getTransfer(ctx *gin.Context) {
	var uri transferRequestUri

	if err := ctx.ShouldBindUri(&uri); err != nil {
		handleUserValidationErrResponse(ctx, err)
	}

	transfer, err := s.store.GetTransfer(ctx, uri.ID)

	if err != nil {
		handleDBErrResponse(ctx, err)
		return
	}

	account, err := s.store.GetAccount(ctx, transfer.FromAccountID)
	if err != nil {
		handleDBErrResponse(ctx, err)
		return
	}

	authPayload := ctx.MustGet(s.config.AuthorizationPayloadKey).(*token.Payload)

	if authPayload.Username != account.Owner {

		err := errors.New("account doesn't belong to the authenticated user")
		ctx.JSON(http.StatusUnauthorized, handleCustomErrorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, transfer)
}

type listTransferRequest struct {
	FromAccountID int64 `form:"from_account_id" binding:"required,min=1"`
	PageID        int32 `form:"page_id" binding:"required,min=1"`
	PageSize      int32 `form:"page_size" binding:"required,min=3,max=10"`
}

func (s *Server) ListTransfer(ctx *gin.Context) {
	var req listTransferRequest

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
		handleUserValidationErrResponse(ctx, err)
		return
	}

	acc, err := s.store.GetAccount(ctx, req.FromAccountID)
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

	arg := db.ListTransfersParams{
		FromAccountID: req.FromAccountID,
		Limit:         req.PageSize,
		Offset:        (req.PageID - 1) * req.PageSize,
	}

	transferLists, err := s.store.ListTransfers(ctx, arg)

	if err != nil {
		handleDBErrResponse(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, transferLists)
}
