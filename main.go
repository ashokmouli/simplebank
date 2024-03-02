package main

import (
	"database/sql"
	"log"

	"github.com/ashokmouli/simplebank/api"
	db "github.com/ashokmouli/simplebank/db/sqlc"
	"github.com/ashokmouli/simplebank/db/util"
	_ "github.com/lib/pq"
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
	server, err := api.NewServer(store, config)
	if err != nil {
		log.Fatal("could not start the server: ", err)
	}
	err = server.StartServer(config.ServerAddress)
	if err != nil {
		log.Fatal("could not start the server", err)
	}
}
