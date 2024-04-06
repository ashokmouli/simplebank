package main

import (
	"context"
	"database/sql"
	"log"
	"net"
	"net/http"

	"github.com/ashokmouli/simplebank/api"
	db "github.com/ashokmouli/simplebank/db/sqlc"
	"github.com/ashokmouli/simplebank/db/util"
	"github.com/ashokmouli/simplebank/gapi"
	"github.com/ashokmouli/simplebank/pb"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	_ "github.com/lib/pq"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"google.golang.org/protobuf/encoding/protojson"
)

func main() {

	config, err := util.LoadConfig("./")
	if err != nil {
		log.Fatal("could not read config file", err)
	}
	conn, err := sql.Open(config.DBDriver, config.DBSource)
	if err != nil {
		log.Fatal("cannot connect to db: ", err)
	}
	store := db.NewStore(conn)
	go runGatewayServer(store, config)
	runGrpcServer(store, config)
}

func runGatewayServer(store db.Store, config util.Config) {

	server, err := gapi.NewServer(store, config)
	if err != nil {
		log.Fatal("could not start the server: ", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// jsonOption settings below preserve the field names in proto as is. Without these options, field names are
	// camelCased.
	jsonOption := runtime.WithMarshalerOption(runtime.MIMEWildcard, &runtime.JSONPb{
		MarshalOptions: protojson.MarshalOptions{
			UseProtoNames: true,
		},
		UnmarshalOptions: protojson.UnmarshalOptions{
			DiscardUnknown: true,
		},
	})

	grpcMux := runtime.NewServeMux(jsonOption)

	err = pb.RegisterSimpleBankHandlerServer(ctx, grpcMux, server)
	if err != nil {
		log.Fatalf("cannot register handler server :%s", err)
	}

	// Create a new http mux and have it handle grpc mux.
	mux := http.NewServeMux()
	mux.Handle("/", grpcMux)

	listener, err := net.Listen("tcp", config.HTTPServerAddress)
	if err != nil {
		log.Fatalf("cannot create listener: %s", err)
	}
	log.Printf("Starting HTTP gateway server at: %s", listener.Addr().String())
	err = http.Serve(listener, mux)
	if err != nil {
		log.Fatalf("cannot start HTTP gateway server: %s", err)
	}
}

func runGrpcServer(store db.Store, config util.Config) {

	server, err := gapi.NewServer(store, config)
	if err != nil {
		log.Fatal("could not start the server: ", err)
	}

	grpcServer := grpc.NewServer()
	pb.RegisterSimpleBankServer(grpcServer, server)

	// Make services on this server visible for clients to explore.
	reflection.Register(grpcServer)

	listener, err := net.Listen("tcp", config.GRPCServerAddress)
	if err != nil {
		log.Fatal("cannot create listener")
	}
	log.Printf("Starting grpc server at: %s", listener.Addr().String())
	err = grpcServer.Serve(listener)
	if err != nil {
		log.Fatal("cannot start gRPC server")
	}
}

// runGinServer calls the Http endpoint.
func _ /* runGinServer */ (store db.Store, config util.Config) {
	server, err := api.NewServer(store, config)
	if err != nil {
		log.Fatal("could not start the server: ", err)
	}
	err = server.StartServer(config.HTTPServerAddress)
	if err != nil {
		log.Fatal("could not start the server", err)
	}
}
