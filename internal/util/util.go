package util

import (
	"context"

	"github.com/chekist32/goipay/internal/db"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

func InitDbQueriesWithTx(ctx context.Context, dbConnPool *pgxpool.Pool) (*db.Queries, pgx.Tx, error) {
	tx, err := dbConnPool.Begin(ctx)
	if err != nil {
		return nil, nil, err
	}

	return db.New(tx), tx, nil
}
