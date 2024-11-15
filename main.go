package main

import (
	"context"
	"github.com/HL/meta-bank/mail"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/HL/meta-bank/api"
	db "github.com/HL/meta-bank/db/sqlc"
	"github.com/HL/meta-bank/util"
	"github.com/HL/meta-bank/worker"
	"github.com/hibiken/asynq"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog/log"
)

func main() {

	config, err := util.LoadConfig(".")

	if err != nil {
		log.Fatal().Err(err).Msg("Load Config Failed")
	}

	//To connect DB string
	connPool, err := pgxpool.New(context.Background(), config.DBSourceLocal)

	if err != nil {
		// Fatal starts a new message with fatal level(4). The os.Exit(1) function
		// is called by the Msg method.
		// You must call Msg on the returned event in order to send the event.

		//log.Error().Err(err)
		//Error starts a new message with error level.(level 3)
		log.Fatal().Err(err).Msg("Cannot connect to db")
	}

	store := db.NewStore(connPool)

	redsOpts := asynq.RedisClientOpt{
		Addr: config.RedisAddress,
	}

	taskDistributor := worker.NewRedisTaskDistributor(redsOpts)

	go runTaskProcessor(config, redsOpts, store)
	runRedisGinServer(store, config, taskDistributor)
	go listenForShutdown()
}

// runTaskProcessor function picking tasks value from redis
func runTaskProcessor(util util.Config, redisOpts asynq.RedisClientOpt, store db.Store) {

	sender := mail.NewGmailSender(util.EmailSenderName, util.EmailSenderAddress, util.EmailSenderPassword)

	taskProcessor := worker.NewRedisTaskProcessor(redisOpts, store, sender)
	log.Info().Msg("Starting task processor from main")

	err := taskProcessor.Start()

	if err != nil {
		log.Fatal().Err(err).Msg("Failed to start task processor")
	}
}

// runGinServer server using Gin
func runRedisGinServer(store db.Store, config util.Config, taskDistributor worker.TaskDistributor) {

	//create server
	s, err := api.NewServer(store, config, taskDistributor)

	if err != nil {
		log.Fatal().Err(err).Msg("Cannot create server")
	}

	//start server
	err = s.Start(config.HTTPServerAddress)

	if err != nil {
		log.Fatal().Err(err).Msg("Cannot Start Sever..")
	}

}

// runGinServer server using Gin
//func runGinServer(store db.Store, config util.Config) {
//
//	//create server
//	s, err := api.NewServer(store, config)
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

// listenForShutdown The main purpose of this function is to ensure that
// your application can shut down gracefully when it receives a termination signal.
// This is especially important for long-running applications, such as web servers or
// background workers, that need to clean up resources and finish any in-progress
// tasks before stopping
func listenForShutdown() {
	var wg *sync.WaitGroup

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	<-quit
	shutdown(wg)
	os.Exit(0)
}
func shutdown(wg *sync.WaitGroup) {

	log.Info().Msg("would run clean up tasks...")

	wg.Wait()

	log.Info().Msg("closing channels and shutting down applications...")

}
