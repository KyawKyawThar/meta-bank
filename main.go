package main

import (
	"context"
	"github.com/HL/meta-bank/api"
	db "github.com/HL/meta-bank/db/sqlc"
	"github.com/HL/meta-bank/util"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog/log"
)

func main() {

	config, err := util.LoadConfig(".")
	if err != nil {
		log.Fatal().Err(err).Msg("Load Config Failed")
	}

	//To connect DB string
	connPool, err := pgxpool.New(context.Background(), config.DBSource)

	if err != nil {
		log.Fatal().Err(err).Msg("Cannot connect to db")
	}
	store := db.NewStore(connPool)
	runGinServer(store, config)
}

// runGinServer server using Gin
func runGinServer(store db.Store, config util.Config) {

	//create server
	s, err := api.NewServer(store, config)

	if err != nil {
		log.Fatal().Err(err).Msg("Cannot create server")
	}

	//start server
	err = s.Start(config.HTTPServerAddress)

	if err != nil {
		log.Fatal().Err(err).Msg("Cannot Start Sever..")
	}

}
