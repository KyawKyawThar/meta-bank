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
		// Fatal starts a new message with fatal level(4). The os.Exit(1) function
		// is called by the Msg method.
		// You must call Msg on the returned event in order to send the event.

		//log.Error().Err(err)
		//Error starts a new message with error level.(level 3)
		log.Fatal().Err(err).Msg("Cannot connect to db")
	}

	store := db.NewStore(connPool)

	//redsOpts := asynq.RedisClientOpt{
	//	Addr: config.RedisAddress,
	//}
	//
	//taskDistributor := worker.NewRedisTaskDistributor(redsOpts)

	//go runTaskProcessor(redsOpts, store)
	runGinServer(store, config)
}

//func runTaskProcessor(redisOpts asynq.RedisClientOpt, store db.Store) {
//	taskProcessor := worker.NewRedisTaskProcessor(redisOpts, store)
//	log.Info().Msg("Starting task processor from main")
//
//	err := taskProcessor.Start()
//
//	if err != nil {
//		log.Fatal().Err(err).Msg("Failed to start task processor")
//	}
//}

// runGinServer server using Gin
//func runGinServer(store db.Store, config util.Config, taskDistributor worker.TaskDistributor) {
//
//	//create server
//	s, err := api.NewServer(store, config, taskDistributor)
//
//	if err != nil {
//		log.Fatal().Err(err).Msg("Cannot create server")
//	}
//
//	//start server
//	err = s.Start(config.HTTPServerAddress)
//
//	if err != nil {
//		log.Fatal().Err(err).Msg("Cannot Start Sever..")
//	}
//
//}

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
