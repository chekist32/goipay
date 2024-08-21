package util

import (
	"context"
	"encoding/json"

	"github.com/chekist32/goipay/internal/db"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

func ParseJson[T any](data []byte) (*T, error) {
	var res T
	if err := json.Unmarshal(data, &res); err != nil {
		return nil, err
	}

	return &res, nil
}

func InitDbQueriesWithTx(ctx context.Context, dbConnPool *pgxpool.Pool) (*db.Queries, pgx.Tx, error) {
	tx, err := dbConnPool.Begin(ctx)
	if err != nil {
		return nil, nil, err
	}

	return db.New(tx), tx, nil
}
