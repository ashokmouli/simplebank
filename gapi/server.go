package gapi

import (
	"fmt"

	db "github.com/ashokmouli/simplebank/db/sqlc"
	"github.com/ashokmouli/simplebank/db/util"
	"github.com/ashokmouli/simplebank/pb"
	"github.com/ashokmouli/simplebank/token"
)

type Server struct {
	pb.UnimplementedSimpleBankServer
	store  db.Store
	maker  token.Maker
	config util.Config
}

// Server serves gRPC requests for our banking service
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

	return server, nil
}
