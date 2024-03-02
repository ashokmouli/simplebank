package api

import (
	"fmt"

	db "github.com/ashokmouli/simplebank/db/sqlc"
	"github.com/ashokmouli/simplebank/db/util"
	"github.com/ashokmouli/simplebank/token"
	"github.com/gin-gonic/gin"
)

func errorResponse(err error) gin.H {
	return gin.H{"error": err.Error()}
}

type Server struct {
	store  db.Store
	router *gin.Engine
	maker  token.Maker
	config util.Config
}

func NewServer(store db.Store, config util.Config) (*Server, error) {

	// Create a token maker
	maker, err := token.NewPasetoMaker(config.TokenSymmetricKey)
	if err != nil {
		err := fmt.Errorf("could not create a new token interface, %w", err)
		return nil, err
	}

	server := &Server{
		store:  store,
		config: config,
		maker:  maker,
	}

	// Create Routes
	server.CreateRoutes()

	return server, nil
}
func (server *Server) CreateRoutes() {
	router := gin.Default()

	router.POST("/users", server.createUser)      // Create user
	router.POST("/users/login", server.loginUser) // Login as a user

	authGroups := router.Group("/").Use(createAuthMiddleware(server.maker))

	authGroups.POST("/accounts", server.createAccount) // Create an account
	authGroups.GET("/accounts/:id", server.getAccount) // Get the account with ID equals id.
	authGroups.GET("/accounts", server.listAccount)    // List accounts

	authGroups.POST("/transfers", server.transfer)     // Perfomr account transfer
	authGroups.GET("/users/:username", server.getUser) // Get user info

	server.router = router

}

func (server *Server) StartServer(address string) error {
	return server.router.Run(address)
}
