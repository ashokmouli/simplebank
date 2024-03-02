package api

import (
	"database/sql"
	"net/http"

	db "github.com/ashokmouli/simplebank/db/sqlc"
	"github.com/ashokmouli/simplebank/token"
	"github.com/gin-gonic/gin"
	"github.com/lib/pq"
)

type createAccountRequest struct {
	Currency string `json:"currency" binding:"required,oneof=USD EUR"`
}

func (server *Server) createAccount(ctx *gin.Context) {
	var json createAccountRequest
	if err := ctx.ShouldBindJSON(&json); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}
	// Retrieve the user name from the authorized payload and make it the
	// owner of this account.
	payload := (ctx.MustGet(authorizationPayloadKey)).(*token.Payload)

	account, err := server.store.CreateAccount(ctx, db.CreateAccountParams{
		Owner:    payload.Username,
		Balance:  0,
		Currency: json.Currency,
	})

	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok {
			switch pqErr.Code.Name() {
			case "foreign_key_violation", "unique_violation":
				ctx.JSON(http.StatusForbidden, errorResponse(err))
				return
			}
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	ctx.JSON(http.StatusOK, account)
}

type getAccountRequest struct {
	ID int64 `uri:"id" binding:"required,min=1"`
}

func (server *Server) getAccount(ctx *gin.Context) {
	var req getAccountRequest
	if err := ctx.ShouldBindUri(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}
	account, err := server.store.GetAccount(ctx, req.ID)
	if err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusNotFound, errorResponse(err))
			return
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	// Validate that the account being requested is owned by the authorized user
	payload := (ctx.MustGet(authorizationPayloadKey)).(*token.Payload)
	if payload.Username != account.Owner {
		ctx.JSON(http.StatusUnauthorized, errorResponse(err))
	}
	ctx.JSON(http.StatusOK, account)
}

type listAccountRequest struct {
	PageID   int32 `form:"page_id" binding:"required,min=1"`
	PageSize int32 `form:"page_size" binding:"required,min=5,max=10"`
}

func (server *Server) listAccount(ctx *gin.Context) {
	var req listAccountRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	// Retrieve the user name from the authorized payload and make it the
	// owner of this account.
	payload := (ctx.MustGet(authorizationPayloadKey)).(*token.Payload)
	accounts, err := server.store.ListAccounts(ctx, db.ListAccountsParams{
		Owner: payload.Username,
		Limit:  req.PageSize,
		Offset: (req.PageID - 1) * req.PageSize,
	})
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	ctx.JSON(http.StatusOK, accounts)
}
