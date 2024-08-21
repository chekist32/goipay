package test

import (
	"context"
	"log"
	"testing"
	"time"

	"github.com/chekist32/goipay/internal/db"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"
)

func TestFindCryptoCacheByCoin(t *testing.T) {
	t.Run("Should Return Valid Crypto Cache", func(t *testing.T) {
		runInTransaction(t, dbConnPool, func(t *testing.T, tx pgx.Tx) {
			ctx := context.Background()
			q := db.New(tx)

			for _, coin := range dbCoinTypes {
				_, err := q.FindCryptoCacheByCoin(ctx, coin)
				assert.NoError(t, err)
			}
		})
	})

	t.Run("Should Return SQL Error (no rows)", func(t *testing.T) {
		runInTransaction(t, dbConnPool, func(t *testing.T, tx pgx.Tx) {
			ctx := context.Background()
			q := db.New(tx)

			_, err := q.FindCryptoCacheByCoin(ctx, "test")
			var pgErr *pgconn.PgError
			assert.ErrorAs(t, err, &pgErr)
			assert.Equal(t, "22P02", pgErr.Code)
		})
	})
}

func TestUpdateCryptoCacheByCoin(t *testing.T) {
	runInTransaction(t, dbConnPool, func(t *testing.T, tx pgx.Tx) {
		ctx := context.Background()
		q := db.New(tx)

		coin := db.CoinTypeBTC

		var lastSyncedBlockHeight pgtype.Int8
		if err := lastSyncedBlockHeight.Scan(int64(123)); err != nil {
			log.Fatal(err)
		}

		updated, err := q.UpdateCryptoCacheByCoin(ctx, db.UpdateCryptoCacheByCoinParams{Coin: coin, LastSyncedBlockHeight: lastSyncedBlockHeight})
		assert.NoError(t, err)
		assert.Condition(t, func() (success bool) {
			return time.Now().UTC().Sub(updated.SyncedTimestamp.Time) < 10*time.Second
		})
	})
}
