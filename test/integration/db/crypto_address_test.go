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

func TestCreateCryptoAddress(t *testing.T) {
	t.Run("Should Return Valid Address", func(t *testing.T) {
		runInTransaction(t, dbConnPool, func(t *testing.T, tx pgx.Tx) {
			ctx := context.Background()
			q := db.New(tx)

			userId, err := q.CreateUser(ctx)
			if err != nil {
				log.Fatal(err)
			}

			_, err = q.CreateCryptoAddress(ctx, db.CreateCryptoAddressParams{Address: uuid.NewString(), Coin: db.CoinTypeBTC, IsOccupied: true, UserID: userId})
			assert.NoError(t, err)
		})
	})

	t.Run("Should Return SQL Error (address dublication)", func(t *testing.T) {
		runInTransaction(t, dbConnPool, func(t *testing.T, tx pgx.Tx) {
			ctx := context.Background()
			q := db.New(tx)

			userId, err := q.CreateUser(ctx)
			if err != nil {
				log.Fatal(err)
			}

			addr := uuid.NewString()

			_, err = q.CreateCryptoAddress(ctx, db.CreateCryptoAddressParams{Address: addr, Coin: db.CoinTypeBTC, IsOccupied: true, UserID: userId})
			if err != nil {
				log.Fatal(err)
			}

			_, err = q.CreateCryptoAddress(ctx, db.CreateCryptoAddressParams{Address: addr, Coin: db.CoinTypeBTC, IsOccupied: true, UserID: userId})
			var pgErr *pgconn.PgError
			assert.ErrorAs(t, err, &pgErr)
			assert.Equal(t, "23505", pgErr.Code)
		})
	})
}

func TestFindNonOccupiedCryptoAddressAndLockByUserIdAndCoin(t *testing.T) {
	t.Run("Should Return Valid Address", func(t *testing.T) {
		runInTransaction(t, dbConnPool, func(t *testing.T, tx pgx.Tx) {
			ctx := context.Background()
			q := db.New(tx)

			userId, err := q.CreateUser(ctx)
			if err != nil {
				log.Fatal(err)
			}

			expectedAddr, err := q.CreateCryptoAddress(ctx, db.CreateCryptoAddressParams{Address: uuid.NewString(), Coin: db.CoinTypeBTC, IsOccupied: false, UserID: userId})
			if err != nil {
				log.Fatal(err)
			}

			addr, err := q.FindNonOccupiedCryptoAddressAndLockByUserIdAndCoin(ctx, db.FindNonOccupiedCryptoAddressAndLockByUserIdAndCoinParams{Coin: db.CoinTypeBTC, UserID: userId})
			assert.NoError(t, err)
			assert.Equal(t, expectedAddr.Address, addr.Address)
			assert.Equal(t, expectedAddr.UserID, userId)
			assert.True(t, addr.IsOccupied)
		})
	})

	t.Run("Should Return SQL Error (no rows (coin))", func(t *testing.T) {
		runInTransaction(t, dbConnPool, func(t *testing.T, tx pgx.Tx) {
			ctx := context.Background()
			q := db.New(tx)

			userId, err := q.CreateUser(ctx)
			if err != nil {
				log.Fatal(err)
			}

			_, err = q.CreateCryptoAddress(ctx, db.CreateCryptoAddressParams{Address: uuid.NewString(), Coin: db.CoinTypeBTC, IsOccupied: false, UserID: userId})
			if err != nil {
				log.Fatal(err)
			}

			_, err = q.FindNonOccupiedCryptoAddressAndLockByUserIdAndCoin(ctx, db.FindNonOccupiedCryptoAddressAndLockByUserIdAndCoinParams{Coin: db.CoinTypeETH, UserID: userId})
			assert.ErrorIs(t, err, pgx.ErrNoRows)
		})
	})

	t.Run("Should Return SQL Error (no rows (userId))", func(t *testing.T) {
		runInTransaction(t, dbConnPool, func(t *testing.T, tx pgx.Tx) {
			ctx := context.Background()
			q := db.New(tx)

			userId, err := q.CreateUser(ctx)
			if err != nil {
				log.Fatal(err)
			}

			_, err = q.CreateCryptoAddress(ctx, db.CreateCryptoAddressParams{Address: uuid.NewString(), Coin: db.CoinTypeBTC, IsOccupied: false, UserID: userId})
			if err != nil {
				log.Fatal(err)
			}

			var userId1 pgtype.UUID
			if err := userId1.Scan(uuid.NewString()); err != nil {
				log.Fatal(err)
			}

			_, err = q.FindNonOccupiedCryptoAddressAndLockByUserIdAndCoin(ctx, db.FindNonOccupiedCryptoAddressAndLockByUserIdAndCoinParams{Coin: db.CoinTypeBTC, UserID: userId1})
			assert.ErrorIs(t, err, pgx.ErrNoRows)
		})
	})

	t.Run("Should Return SQL Error (no rows (occupied))", func(t *testing.T) {
		runInTransaction(t, dbConnPool, func(t *testing.T, tx pgx.Tx) {
			ctx := context.Background()
			q := db.New(tx)

			userId, err := q.CreateUser(ctx)
			if err != nil {
				log.Fatal(err)
			}

			_, err = q.CreateCryptoAddress(ctx, db.CreateCryptoAddressParams{Address: uuid.NewString(), Coin: db.CoinTypeBTC, IsOccupied: true, UserID: userId})
			if err != nil {
				log.Fatal(err)
			}

			_, err = q.FindNonOccupiedCryptoAddressAndLockByUserIdAndCoin(ctx, db.FindNonOccupiedCryptoAddressAndLockByUserIdAndCoinParams{Coin: db.CoinTypeBTC, UserID: userId})
			assert.ErrorIs(t, err, pgx.ErrNoRows)
		})
	})

}

func TestUpdateIsOccupiedByCryptoAddress(t *testing.T) {
	runInTransaction(t, dbConnPool, func(t *testing.T, tx pgx.Tx) {
		ctx := context.Background()
		q := db.New(tx)

		userId, err := q.CreateUser(ctx)
		if err != nil {
			log.Fatal(err)
		}

		createdAddr, err := q.CreateCryptoAddress(ctx, db.CreateCryptoAddressParams{Address: uuid.NewString(), Coin: db.CoinTypeBTC, IsOccupied: false, UserID: userId})
		if err != nil {
			log.Fatal(err)
		}

		addr, err := q.UpdateIsOccupiedByCryptoAddress(ctx, db.UpdateIsOccupiedByCryptoAddressParams{Address: createdAddr.Address, IsOccupied: true})
		assert.NoError(t, err)
		assert.True(t, addr.IsOccupied)
	})
}

func TestDeleteAllCryptoAddressByUserIdAndCoin(t *testing.T) {
	gen := func(ctx context.Context, q *db.Queries) ([]pgtype.UUID, []db.CryptoAddress) {
		userId1, err := q.CreateUser(ctx)
		if err != nil {
			log.Fatal(err)
		}
		userId2, err := q.CreateUser(ctx)
		if err != nil {
			log.Fatal(err)
		}

		addr1, err := q.CreateCryptoAddress(ctx, db.CreateCryptoAddressParams{Address: uuid.NewString(), Coin: db.CoinTypeBTC, IsOccupied: false, UserID: userId1})
		if err != nil {
			log.Fatal(err)
		}
		addr2, err := q.CreateCryptoAddress(ctx, db.CreateCryptoAddressParams{Address: uuid.NewString(), Coin: db.CoinTypeXMR, IsOccupied: false, UserID: userId1})
		if err != nil {
			log.Fatal(err)
		}
		addr3, err := q.CreateCryptoAddress(ctx, db.CreateCryptoAddressParams{Address: uuid.NewString(), Coin: db.CoinTypeBTC, IsOccupied: false, UserID: userId2})
		if err != nil {
			log.Fatal(err)
		}

		return []pgtype.UUID{userId1, userId2}, []db.CryptoAddress{addr1, addr2, addr3}
	}

	t.Run("Should Delete 1 Address", func(t *testing.T) {
		runInTransaction(t, dbConnPool, func(t *testing.T, tx pgx.Tx) {
			ctx := context.Background()
			q := db.New(tx)

			userIds, addrs := gen(ctx, q)

			deleted, err := q.DeleteAllCryptoAddressByUserIdAndCoin(ctx, db.DeleteAllCryptoAddressByUserIdAndCoinParams{UserID: userIds[0], Coin: db.CoinTypeBTC})
			assert.NoError(t, err)
			assert.Equal(t, 1, len(deleted))
			assert.Equal(t, addrs[0], deleted[0])
		})
	})

	t.Run("Should Delete 1 Address", func(t *testing.T) {
		runInTransaction(t, dbConnPool, func(t *testing.T, tx pgx.Tx) {
			ctx := context.Background()
			q := db.New(tx)

			userIds, addrs := gen(ctx, q)

			deleted, err := q.DeleteAllCryptoAddressByUserIdAndCoin(ctx, db.DeleteAllCryptoAddressByUserIdAndCoinParams{UserID: userIds[1], Coin: db.CoinTypeBTC})
			assert.NoError(t, err)
			assert.Equal(t, 1, len(deleted))
			assert.Equal(t, addrs[2], deleted[0])
		})
	})

	t.Run("Should Delete 0 Addresses", func(t *testing.T) {
		runInTransaction(t, dbConnPool, func(t *testing.T, tx pgx.Tx) {
			ctx := context.Background()
			q := db.New(tx)

			userIds, _ := gen(ctx, q)

			deleted, err := q.DeleteAllCryptoAddressByUserIdAndCoin(ctx, db.DeleteAllCryptoAddressByUserIdAndCoinParams{UserID: userIds[0], Coin: db.CoinTypeETH})
			assert.NoError(t, err)
			assert.Equal(t, 0, len(deleted))
		})
	})

}
