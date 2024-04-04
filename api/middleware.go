package api

import (
	"errors"
	"fmt"
	"github.com/HL/meta-bank/token"
	"github.com/gin-gonic/gin"
	"net/http"
	"strings"
)

func (s *Server) authMiddleware(tokenMaker token.Maker) gin.HandlerFunc {

	return func(ctx *gin.Context) {

		authorizationHeader := ctx.GetHeader(s.config.AuthorizationHeaderKey)

		if len(authorizationHeader) == 0 {
			err := errors.New("authorization header is not provide")

			ctx.AbortWithStatusJSON(http.StatusUnauthorized, handleErrorResponse(err))
			return
		}

		fields := strings.Fields(authorizationHeader)

		if len(fields) < 2 {
			err := errors.New("invalid authorization header format")

			ctx.AbortWithStatusJSON(http.StatusUnauthorized, handleErrorResponse(err))
			return
		}

		authorizationType, authorization := strings.ToLower(fields[0]), fields[1]

		if authorizationType != s.config.AuthorizationTypeBearer {
			err := fmt.Errorf("%s authorization type", authorizationType)
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, handleErrorResponse(err))
			return
		}

		payload, err := tokenMaker.VerifyToken(authorization)

		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, handleErrorResponse(err))
			return
		}

		ctx.Set(s.config.AuthorizationPayloadKey, payload)
		ctx.Next()
	}
}
