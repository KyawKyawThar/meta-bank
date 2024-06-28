package db

import (
	"context"
	"github.com/HL/meta-bank/util"
	"github.com/jackc/pgx/v5/pgxpool"
	"log"
	"os"
	"testing"
)

var testStore Store

func TestMain(m *testing.M) {

	config, err := util.LoadConfig("../../")

	if err != nil {
		log.Fatal("cannot load config:", err)
	}

	dbPool, err := pgxpool.New(context.Background(), config.DBSourceLocal)

	if err != nil {
		log.Fatal("Cannot connect to database", err)
		return
	}

	testStore = NewStore(dbPool)
	os.Exit(m.Run())
}
