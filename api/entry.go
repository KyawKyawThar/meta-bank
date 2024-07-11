package api

import (
	"errors"
	"fmt"
	db "github.com/HL/meta-bank/db/sqlc"
	"github.com/HL/meta-bank/token"
	"github.com/gin-gonic/gin"
	"net/http"
)

type listEntryRequest struct {
	AccountID int64 `form:"account_id" binding:"required,min=1"`
	PageID    int32 `form:"page_id" binding:"required,min=1"`
	PageSize  int32 `form:"page_size" binding:"required,min=3,max=10"`
}

func (s *Server) listEntry(ctx *gin.Context) {
	//var req listTransferRequest

	var req listEntryRequest

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

	acc, err := s.store.GetAccount(ctx, req.AccountID)
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

	arg := db.ListEntriesParams{
		AccountID: req.AccountID,
		Limit:     req.PageSize,
		Offset:    (req.PageID - 1) * req.PageSize,
	}

	entryList, err := s.store.ListEntries(ctx, arg)

	if err != nil {
		handleDBErrResponse(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, entryList)
}

type entryRequestUri struct {
	ID int64 `uri:"id" binding:"required,min=1"`
}

func (s *Server) getEntry(ctx *gin.Context) {
	var uri entryRequestUri

	if err := ctx.ShouldBindUri(&uri); err != nil {
		handleUserValidationErrResponse(ctx, err)
		return
	}

	fmt.Println("uri", uri.ID)

	entry, err := s.store.GetEntry(ctx, uri.ID)

	if err != nil {
		handleDBErrResponse(ctx, err)
		return
	}

	acc, err := s.store.GetAccount(ctx, entry.AccountID)
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

	ctx.JSON(http.StatusOK, entry)
}
