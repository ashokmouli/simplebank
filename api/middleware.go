package api

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/ashokmouli/simplebank/token"
	"github.com/gin-gonic/gin"
)

const (
	authorizationHeaderKey = "authorization"
	authorizationTypeBearer = "bearer"
	authorizationPayloadKey = "auth_payload"
)

func createAuthMiddleware(tokenMaker token.Maker) gin.HandlerFunc {
	return func(ctx *gin.Context) {

		// The header should be of the form 
		// authorization: bearer <token>
		authHeader := ctx.GetHeader(authorizationHeaderKey)
		if len(authHeader) == 0 {
			err := errors.New("authorization header not found")
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, errorResponse(err))
			return
		}

		tokenString := strings.Fields(authHeader)

		// Validate that the first field is the literal "bearer"
		authType := strings.ToLower(tokenString[0])
		if authType != authorizationTypeBearer {
			err := fmt.Errorf("unsupported authorization type: %s", authType)
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, errorResponse(err))
			return
		}
		//  Pluck out the auth token and verify it.
		authToken := tokenString[1]

		payload, err := tokenMaker.VerifyToken(authToken)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, errorResponse(err))
			return
		}

		// Create a key in context with the payload
		ctx.Set(authorizationPayloadKey, payload)
		ctx.Next()
	}
}
