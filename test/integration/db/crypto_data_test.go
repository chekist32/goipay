package test

import (
	"context"
	"log"
	"testing"

	"github.com/chekist32/goipay/internal/db"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"
)

func createRandomXMRCryptoData(ctx context.Context, q *db.Queries) (db.XmrCryptoDatum, error) {
	return q.CreateXMRCryptoData(ctx, db.CreateXMRCryptoDataParams{PrivViewKey: uuid.NewString(), PubSpendKey: uuid.NewString()})
}

func TestCreateXMRCryptoData(t *testing.T) {
	t.Run("Should Return Valid XMR Crypto Data", func(t *testing.T) {
		runInTransaction(t, dbConnPool, func(t *testing.T, tx pgx.Tx) {
			ctx := context.Background()
			q := db.New(tx)

			_, err := createRandomXMRCryptoData(ctx, q)
			assert.NoError(t, err)
		})
	})

	t.Run("Should Return SQL Error (non unique public spend key)", func(t *testing.T) {
		runInTransaction(t, dbConnPool, func(t *testing.T, tx pgx.Tx) {
			ctx := context.Background()
			q := db.New(tx)

			xmr1, err := createRandomXMRCryptoData(ctx, q)
			assert.NoError(t, err)

			_, err = q.CreateXMRCryptoData(ctx, db.CreateXMRCryptoDataParams{PrivViewKey: uuid.NewString(), PubSpendKey: xmr1.PubSpendKey})
			var pgErr *pgconn.PgError
			assert.ErrorAs(t, err, &pgErr)
			assert.Equal(t, "23505", pgErr.Code)
		})
	})

	t.Run("Should Return SQL Error (non unique private view key)", func(t *testing.T) {
		runInTransaction(t, dbConnPool, func(t *testing.T, tx pgx.Tx) {
			ctx := context.Background()
			q := db.New(tx)

			xmr1, err := createRandomXMRCryptoData(ctx, q)
			assert.NoError(t, err)

			_, err = q.CreateXMRCryptoData(ctx, db.CreateXMRCryptoDataParams{PrivViewKey: xmr1.PrivViewKey, PubSpendKey: uuid.NewString()})
			var pgErr *pgconn.PgError
			assert.ErrorAs(t, err, &pgErr)
			assert.Equal(t, "23505", pgErr.Code)
		})
	})

}

func TestCreateCryptoData(t *testing.T) {
	t.Run("Should Return SQL Error (invalid userId)", func(t *testing.T) {
		runInTransaction(t, dbConnPool, func(t *testing.T, tx pgx.Tx) {
			ctx := context.Background()
			q := db.New(tx)

			xmr, err := createRandomXMRCryptoData(ctx, q)
			if err != nil {
				log.Fatal(err)
			}

			var userId pgtype.UUID
			if err := userId.Scan(uuid.NewString()); err != nil {
				log.Fatal(err)
			}

			_, err = q.CreateCryptoData(ctx, db.CreateCryptoDataParams{XmrID: xmr.ID, UserID: userId})
			var pgErr *pgconn.PgError
			assert.ErrorAs(t, err, &pgErr)
			assert.Equal(t, "23503", pgErr.Code)
		})
	})

	t.Run("Should Return SQL Error (invalid xmrId)", func(t *testing.T) {
		runInTransaction(t, dbConnPool, func(t *testing.T, tx pgx.Tx) {
			ctx := context.Background()
			q := db.New(tx)

			userId, err := q.CreateUser(ctx)
			if err != nil {
				log.Fatal(err)
			}

			var xmrId pgtype.UUID
			if err := xmrId.Scan(uuid.NewString()); err != nil {
				log.Fatal(err)
			}

			_, err = q.CreateCryptoData(ctx, db.CreateCryptoDataParams{XmrID: xmrId, UserID: userId})
			var pgErr *pgconn.PgError
			assert.ErrorAs(t, err, &pgErr)
			assert.Equal(t, "23503", pgErr.Code)
		})
	})

	t.Run("Should Return Valid Crypto Data", func(t *testing.T) {
		runInTransaction(t, dbConnPool, func(t *testing.T, tx pgx.Tx) {
			ctx := context.Background()
			q := db.New(tx)

			userId, err := q.CreateUser(ctx)
			if err != nil {
				log.Fatal(err)
			}
			xmr, err := createRandomXMRCryptoData(ctx, q)
			if err != nil {
				log.Fatal(err)
			}

			_, err = q.CreateCryptoData(ctx, db.CreateCryptoDataParams{XmrID: xmr.ID, UserID: userId})
			assert.NoError(t, err)
		})
	})
}

func TestFindCryptoDataByUserId(t *testing.T) {
	t.Run("Should Return Valid Crypto Data", func(t *testing.T) {
		runInTransaction(t, dbConnPool, func(t *testing.T, tx pgx.Tx) {
			ctx := context.Background()
			q := db.New(tx)

			userId, err := q.CreateUser(ctx)
			if err != nil {
				log.Fatal(err)
			}
			xmr, err := createRandomXMRCryptoData(ctx, q)
			if err != nil {
				log.Fatal(err)
			}
			expectedCryptoData, err := q.CreateCryptoData(ctx, db.CreateCryptoDataParams{XmrID: xmr.ID, UserID: userId})
			if err != nil {
				log.Fatal(err)
			}

			cryptoData, err := q.FindCryptoDataByUserId(ctx, userId)
			assert.NoError(t, err)
			assert.Equal(t, expectedCryptoData, cryptoData)
		})
	})

	t.Run("Should Return SQL Error (no rows)", func(t *testing.T) {
		runInTransaction(t, dbConnPool, func(t *testing.T, tx pgx.Tx) {
			ctx := context.Background()
			q := db.New(tx)

			userId, err := q.CreateUser(ctx)
			if err != nil {
				log.Fatal(err)
			}
			xmr, err := createRandomXMRCryptoData(ctx, q)
			if err != nil {
				log.Fatal(err)
			}
			_, err = q.CreateCryptoData(ctx, db.CreateCryptoDataParams{XmrID: xmr.ID, UserID: userId})
			if err != nil {
				log.Fatal(err)
			}
			var userId1 pgtype.UUID
			if err := userId1.Scan(uuid.NewString()); err != nil {
				log.Fatal(err)
			}

			_, err = q.FindCryptoDataByUserId(ctx, userId1)
			assert.ErrorIs(t, err, pgx.ErrNoRows)
		})
	})
}

func TestFindKeysAndLockXMRCryptoDataById(t *testing.T) {
	t.Run("Should Return Proper XMR Crypto Data", func(t *testing.T) {
		runInTransaction(t, dbConnPool, func(t *testing.T, tx pgx.Tx) {
			ctx := context.Background()
			q := db.New(tx)

			userId, err := q.CreateUser(ctx)
			if err != nil {
				log.Fatal(err)
			}
			xmr, err := createRandomXMRCryptoData(ctx, q)
			if err != nil {
				log.Fatal(err)
			}
			expectedCryptoData, err := q.CreateCryptoData(ctx, db.CreateCryptoDataParams{XmrID: xmr.ID, UserID: userId})
			if err != nil {
				log.Fatal(err)
			}

			keys, err := q.FindKeysAndLockXMRCryptoDataById(ctx, expectedCryptoData.XmrID)
			assert.NoError(t, err)
			assert.Equal(t, xmr.PrivViewKey, keys.PrivViewKey)
			assert.Equal(t, xmr.PubSpendKey, keys.PubSpendKey)
		})
	})

	t.Run("Should Return SQL Error (no rows)", func(t *testing.T) {
		runInTransaction(t, dbConnPool, func(t *testing.T, tx pgx.Tx) {
			ctx := context.Background()
			q := db.New(tx)

			userId, err := q.CreateUser(ctx)
			if err != nil {
				log.Fatal(err)
			}
			xmr, err := createRandomXMRCryptoData(ctx, q)
			if err != nil {
				log.Fatal(err)
			}
			_, err = q.CreateCryptoData(ctx, db.CreateCryptoDataParams{XmrID: xmr.ID, UserID: userId})
			if err != nil {
				log.Fatal(err)
			}
			var xmrId pgtype.UUID
			if err := xmrId.Scan(uuid.NewString()); err != nil {
				log.Fatal(err)
			}

			_, err = q.FindKeysAndLockXMRCryptoDataById(ctx, xmrId)
			assert.ErrorIs(t, err, pgx.ErrNoRows)
		})
	})

	t.Run("Should Return SQL Error (no rows)", func(t *testing.T) {
		runInTransaction(t, dbConnPool, func(t *testing.T, tx pgx.Tx) {
			ctx := context.Background()
			q := db.New(tx)

			userId, err := q.CreateUser(ctx)
			if err != nil {
				log.Fatal(err)
			}
			xmr, err := createRandomXMRCryptoData(ctx, q)
			if err != nil {
				log.Fatal(err)
			}
			_, err = q.CreateCryptoData(ctx, db.CreateCryptoDataParams{XmrID: xmr.ID, UserID: userId})
			if err != nil {
				log.Fatal(err)
			}
			var xmrId pgtype.UUID
			if err := xmrId.Scan(uuid.NewString()); err != nil {
				log.Fatal(err)
			}

			_, err = q.FindKeysAndLockXMRCryptoDataById(ctx, xmrId)
			assert.ErrorIs(t, err, pgx.ErrNoRows)
		})
	})

}

func TestFindIndicesAndLockXMRCryptoDataById(t *testing.T) {
	t.Run("Should Return Valid XMR Indices", func(t *testing.T) {
		runInTransaction(t, dbConnPool, func(t *testing.T, tx pgx.Tx) {
			ctx := context.Background()
			q := db.New(tx)

			xmr, err := createRandomXMRCryptoData(ctx, q)
			if err != nil {
				log.Fatal(err)
			}

			indices, err := q.FindIndicesAndLockXMRCryptoDataById(ctx, xmr.ID)
			assert.NoError(t, err)
			assert.Equal(t, xmr.LastMajorIndex, indices.LastMajorIndex)
			assert.Equal(t, xmr.LastMinorIndex, indices.LastMinorIndex)
		})
	})

	t.Run("Should Return SQL Error (no rows)", func(t *testing.T) {
		runInTransaction(t, dbConnPool, func(t *testing.T, tx pgx.Tx) {
			ctx := context.Background()
			q := db.New(tx)

			_, err := createRandomXMRCryptoData(ctx, q)
			if err != nil {
				log.Fatal(err)
			}

			var xmrId pgtype.UUID
			if err := xmrId.Scan(uuid.NewString()); err != nil {
				log.Fatal(err)
			}

			_, err = q.FindIndicesAndLockXMRCryptoDataById(ctx, xmrId)
			assert.ErrorIs(t, err, pgx.ErrNoRows)
		})
	})
}

func TestUpdateIndicesXMRCryptoDataById(t *testing.T) {
	runInTransaction(t, dbConnPool, func(t *testing.T, tx pgx.Tx) {
		ctx := context.Background()
		q := db.New(tx)

		xmr, err := createRandomXMRCryptoData(ctx, q)
		if err != nil {
			log.Fatal(err)
		}

		_, err = q.UpdateIndicesXMRCryptoDataById(ctx, db.UpdateIndicesXMRCryptoDataByIdParams{ID: xmr.ID, LastMajorIndex: 1, LastMinorIndex: 1})
		assert.NoError(t, err)
	})
}

func TestUpdateKeysXMRCryptoDataById(t *testing.T) {
	runInTransaction(t, dbConnPool, func(t *testing.T, tx pgx.Tx) {
		ctx := context.Background()
		q := db.New(tx)

		xmr, err := createRandomXMRCryptoData(ctx, q)
		if err != nil {
			log.Fatal(err)
		}

		_, err = q.UpdateKeysXMRCryptoDataById(ctx, db.UpdateKeysXMRCryptoDataByIdParams{ID: xmr.ID, PrivViewKey: uuid.NewString(), PubSpendKey: uuid.NewString()})
		assert.NoError(t, err)
	})
}

func TestSetXMRCryptoDataByUserId(t *testing.T) {
	t.Run("Should Return Valid Crypto Data", func(t *testing.T) {
		runInTransaction(t, dbConnPool, func(t *testing.T, tx pgx.Tx) {
			ctx := context.Background()
			q := db.New(tx)

			userId, err := q.CreateUser(ctx)
			if err != nil {
				log.Fatal(err)
			}
			xmr, err := createRandomXMRCryptoData(ctx, q)
			if err != nil {
				log.Fatal(err)
			}
			_, err = q.CreateCryptoData(ctx, db.CreateCryptoDataParams{XmrID: xmr.ID, UserID: userId})
			if err != nil {
				log.Fatal(err)
			}
			xmr1, err := createRandomXMRCryptoData(ctx, q)
			if err != nil {
				log.Fatal(err)
			}

			_, err = q.SetXMRCryptoDataByUserId(ctx, db.SetXMRCryptoDataByUserIdParams{UserID: userId, XmrID: xmr1.ID})
			assert.NoError(t, err)
		})
	})

	t.Run("Should Return SQL Error (invalid xmr_id)", func(t *testing.T) {
		runInTransaction(t, dbConnPool, func(t *testing.T, tx pgx.Tx) {
			ctx := context.Background()
			q := db.New(tx)

			userId, err := q.CreateUser(ctx)
			if err != nil {
				log.Fatal(err)
			}
			xmr, err := createRandomXMRCryptoData(ctx, q)
			if err != nil {
				log.Fatal(err)
			}
			_, err = q.CreateCryptoData(ctx, db.CreateCryptoDataParams{XmrID: xmr.ID, UserID: userId})
			if err != nil {
				log.Fatal(err)
			}

			var xmrId pgtype.UUID
			if err := xmrId.Scan(uuid.NewString()); err != nil {
				log.Fatal(err)
			}

			_, err = q.SetXMRCryptoDataByUserId(ctx, db.SetXMRCryptoDataByUserIdParams{UserID: userId, XmrID: xmrId})
			var pgErr *pgconn.PgError
			assert.ErrorAs(t, err, &pgErr)
			assert.Equal(t, "23503", pgErr.Code)
		})
	})

}
