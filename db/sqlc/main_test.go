package db

import (
	"context"
	"fmt"
	"github.com/HL/meta-bank/util"
	"github.com/jackc/pgx/v5/pgxpool"
	"log"
	"os"
	"testing"
)

var store Store

func TestMain(m *testing.M) {

	config, err := util.LoadConfig("../../")

	if err != nil {
		log.Fatal("cannot load config:", err)
	}

	dbPool, err := pgxpool.New(context.Background(), config.DBSource)

	if err != nil {
		log.Fatal("Cannot connect to database", err)
		return
	}
	fmt.Println("connected to database")

	store = NewStore(dbPool)
	fmt.Println("store is:", store)
	os.Exit(m.Run())
}
