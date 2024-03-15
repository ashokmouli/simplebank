package api

import (
	"database/sql"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type RenewTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

type RenewTokenResponse struct {
	AccessToken          string    `json:"access_token"`
	AccessTokenExpiresAt time.Time `json:"access_token_expires_at"`
}

func (server *Server) renewToken(ctx *gin.Context) {

	// Unmarshal the request
	var req RenewTokenRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	// Validate the token
	payload, err := server.maker.VerifyToken(req.RefreshToken)
	if err != nil {
		ctx.JSON(http.StatusForbidden, errorResponse(err))
		return
	}

	// Fetch the session object from DB
	session, err := server.store.GetSession(ctx, payload.ID)
	if err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusNotFound, errorResponse(err))
			return
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	// Validate this token is not blocked
	if session.IsBlocked {
		err := fmt.Errorf("refresh token blocked")
		ctx.JSON(http.StatusForbidden, errorResponse(err))
		return
	}

	// Compare user names
	if payload.Username != session.Username {
		err := fmt.Errorf("mismatched user names")
		ctx.JSON(http.StatusForbidden, errorResponse(err))
	}

	// Compare the refresh token passed in with the one in DB.
	if req.RefreshToken != session.RefreshToken {
		err := fmt.Errorf("session refresh token doesn't match incoming refresh token")
		ctx.JSON(http.StatusForbidden, errorResponse(err))
		return
	}

	// Verify that the refresh token has not expired (shouldn't but just check)
	if time.Now().After(session.ExpiresAt) {
		err := fmt.Errorf("refresh token expired")
		ctx.JSON(http.StatusForbidden, errorResponse(err))
	}
	// Create access token.
	accessToken, accessPayload, err := server.maker.CreateToken(payload.Username, server.config.AccessTokenDuration)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
	}

	// Marshal back the response.
	var resp RenewTokenResponse
	resp.AccessToken = accessToken
	resp.AccessTokenExpiresAt = accessPayload.ExpiredAt

	ctx.JSON(http.StatusOK, resp)
}
