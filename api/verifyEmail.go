package api

import (
	db "github.com/HL/meta-bank/db/sqlc"
	"github.com/gin-gonic/gin"
	"net/http"
)

type verifyEmailQuery struct {
	EmailId    int64  `form:"email_id" binding:"required"`
	SecretCode string `form:"secret_code" binding:"required"`
}

type verifyEmailTxResult struct {
	User        db.User
	VerifyEmail db.VerifyEmail
}

func (s *Server) verifyEmail(ctx *gin.Context) {
	var query verifyEmailQuery

	if err := ctx.ShouldBindQuery(&query); err != nil {
		handleUserValidationErrResponse(ctx, err)
	}

	arg := db.VerifyEmailTxParams{
		EmailId:    query.EmailId,
		SecretCode: query.SecretCode,
	}
	result, err := s.store.VerifyEmailTx(ctx, arg)

	if err != nil {
		handleDBErrResponse(ctx, err)
		return
	}

	res := verifyEmailTxResult{
		User:        result.User,
		VerifyEmail: result.VerifyEmail,
	}

	ctx.JSON(http.StatusOK, res.VerifyEmail)
}
