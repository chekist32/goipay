package test

import (
	"context"
	"fmt"
	"log"
	"os"
	"testing"

	"github.com/chekist32/goipay/internal/db"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	dbConnPool  *pgxpool.Pool
	dbCoinTypes [5]db.CoinType = [5]db.CoinType{db.CoinTypeXMR, db.CoinTypeBTC, db.CoinTypeLTC, db.CoinTypeETH, db.CoinTypeTON}
)

func TestMain(m *testing.M) {
	ctx := context.Background()

	postgres, close, err := spinUpPostgresContainer()
	if err != nil {
		log.Fatal(err)
	}
	defer close(ctx)

	dbUser := "postgres"
	dbPass := dbUser
	dbName := dbUser
	dbHost := "localhost"
	dbPort, err := postgres.MappedPort(ctx, "5432/tcp")
	if err != nil {
		log.Fatal(err)
	}

	dbUrl := fmt.Sprintf("postgresql://%v:%v@%v:%v/%v", dbUser, dbPass, dbHost, dbPort.Port(), dbName)
	connPool, err := pgxpool.New(ctx, dbUrl)
	if err != nil {
		log.Fatal(err)
	}
	dbConnPool = connPool

	os.Exit(m.Run())
}
