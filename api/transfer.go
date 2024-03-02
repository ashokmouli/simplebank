package api

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"

	db "github.com/ashokmouli/simplebank/db/sqlc"
	"github.com/ashokmouli/simplebank/token"
	"github.com/gin-gonic/gin"
)

type transferRequest struct {
	FromAccountID int64  `json:"from_account_id" binding:"required,min=1"`
	ToAccountID   int64  `json:"to_account_id" binding:"required,min=1"`
	Amount        int64  `json:"amount" binding:"required,gt=0"`
	Currency      string `json:"currency" binding:"required,oneof=USD EUR CAD"`
}

func (server *Server) transfer(ctx *gin.Context) {
	var req transferRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}
	account, valid := validateAccount(ctx, server.store, req.FromAccountID, req.Currency)
	if !valid {
		return
	}

	_, valid = validateAccount(ctx, server.store, req.ToAccountID, req.Currency)
	if !valid {
		return
	}

	// Check that the from account is the authorized user.
	payload := (ctx.MustGet(authorizationPayloadKey)).(*token.Payload)
	if (account.Owner != payload.Username) {
		ctx.JSON(http.StatusForbidden, errorResponse(errors.New("account does not belong to logged in user")))
		return
	}

	input := db.TransferTxParams{
		FromAccountID: req.FromAccountID,
		ToAccountID:   req.ToAccountID,
		Amount:        req.Amount,
	}

	results, err := server.store.TransferTx(ctx, &input)
	if err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusNotFound, errorResponse(err))
			return
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	ctx.JSON(http.StatusOK, results)
}

// Returns true if the currency on the account object pointed to by account id matches 'currency'
func validateAccount(ctx *gin.Context, store db.Store, accountId int64, currency string) (db.Account, bool) {
	account, err := store.GetAccount(ctx, accountId)
	if err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusBadRequest, errorResponse(err))
			return account, false
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return account, false
	}
	if account.Currency != currency {
		err := fmt.Errorf("currency on account does not match the specificed currency")
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return account, false
	}
	return account, true
}
