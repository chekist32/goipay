package test

import (
	"context"
	"log"
	"math/rand"
	"testing"
	"time"

	"github.com/chekist32/goipay/internal/db"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"
)

func createRandTestInvoice(ctx context.Context, q *db.Queries, userId pgtype.UUID) (db.Invoice, error) {
	var expiresAt pgtype.Timestamptz
	if err := expiresAt.Scan(time.Now().UTC()); err != nil {
		log.Fatal(err)
	}

	return q.CreateInvoice(ctx, db.CreateInvoiceParams{
		CryptoAddress:         uuid.NewString(),
		Coin:                  dbCoinTypes[rand.Intn(len(dbCoinTypes))],
		RequiredAmount:        rand.Float64(),
		ConfirmationsRequired: int16(rand.Int()),
		ExpiresAt:             expiresAt,
		UserID:                userId,
	})
}

func TestCreateInvoice(t *testing.T) {
	t.Run("Should Create Invoice (with existing userId)", func(t *testing.T) {
		runInTransaction(t, dbConnPool, func(t *testing.T, tx pgx.Tx) {
			ctx := context.Background()
			q := db.New(tx)

			userId, err := q.CreateUser(ctx)
			if err != nil {
				log.Fatal(err)
			}

			var expiresAt pgtype.Timestamptz
			if err := expiresAt.Scan(time.Now().UTC()); err != nil {
				log.Fatal(err)
			}

			_, err = createRandTestInvoice(ctx, q, userId)

			assert.NoError(t, err)
		})
	})

	t.Run("Should Return SQL Error (no such userId)", func(t *testing.T) {
		runInTransaction(t, dbConnPool, func(t *testing.T, tx pgx.Tx) {
			ctx := context.Background()
			q := db.New(tx)

			var userId pgtype.UUID
			if err := userId.Scan(uuid.NewString()); err != nil {
				log.Fatal(err)
			}

			_, err := createRandTestInvoice(ctx, q, userId)

			var pgErr *pgconn.PgError
			assert.ErrorAs(t, err, &pgErr)
			assert.Equal(t, "23503", pgErr.Code)
		})
	})
}

func TestFindAllInvoicesByIds(t *testing.T) {
	t.Run("Should Return 2 Invoices", func(t *testing.T) {
		runInTransaction(t, dbConnPool, func(t *testing.T, tx pgx.Tx) {
			ctx := context.Background()
			q := db.New(tx)

			userId, err := q.CreateUser(ctx)
			if err != nil {
				log.Fatal(err)
			}

			var expectedInvoices [3]db.Invoice
			for i := 0; i < len(expectedInvoices); i++ {
				inv, err := createRandTestInvoice(ctx, q, userId)
				if err != nil {
					log.Fatal(err)
				}

				expectedInvoices[i] = inv
			}

			var ids [2]pgtype.UUID
			for i := 0; i < len(ids); i++ {
				ids[i] = expectedInvoices[i].ID
			}

			invoices, err := q.FindAllInvoicesByIds(ctx, ids[:])
			assert.NoError(t, err)

			for i := 0; i < len(invoices); i++ {
				assert.Equal(t, expectedInvoices[i], invoices[i])
			}
		})
	})

	t.Run("Should Return 0 Invoices", func(t *testing.T) {
		runInTransaction(t, dbConnPool, func(t *testing.T, tx pgx.Tx) {
			ctx := context.Background()
			q := db.New(tx)

			userId, err := q.CreateUser(ctx)
			if err != nil {
				log.Fatal(err)
			}

			var expectedInvoices [3]db.Invoice
			for i := 0; i < len(expectedInvoices); i++ {
				inv, err := createRandTestInvoice(ctx, q, userId)
				if err != nil {
					log.Fatal(err)
				}

				expectedInvoices[i] = inv
			}

			var ids [2]pgtype.UUID
			for i := 0; i < len(ids); i++ {
				var id pgtype.UUID
				if err := id.Scan(uuid.NewString()); err != nil {
					log.Fatal(err)
				}
				ids[i] = id
			}

			invoices, err := q.FindAllInvoicesByIds(ctx, ids[:])
			assert.NoError(t, err)

			assert.Equal(t, 0, len(invoices))
		})
	})
}

func TestConfirmInvoiceById(t *testing.T) {
	runInTransaction(t, dbConnPool, func(t *testing.T, tx pgx.Tx) {
		ctx := context.Background()
		q := db.New(tx)

		userId, err := q.CreateUser(ctx)
		if err != nil {
			log.Fatal(err)
		}

		inv, err := createRandTestInvoice(ctx, q, userId)
		if err != nil {
			log.Fatal(err)
		}

		confirmedInv, err := q.ConfirmInvoiceById(ctx, inv.ID)
		assert.NoError(t, err)
		assert.Equal(t, db.InvoiceStatusTypeCONFIRMED, confirmedInv.Status)
		assert.True(t, confirmedInv.ConfirmedAt.Valid)
	})
}

func TestExpireInvoiceById(t *testing.T) {
	runInTransaction(t, dbConnPool, func(t *testing.T, tx pgx.Tx) {
		ctx := context.Background()
		q := db.New(tx)

		userId, err := q.CreateUser(ctx)
		if err != nil {
			log.Fatal(err)
		}

		inv, err := createRandTestInvoice(ctx, q, userId)
		if err != nil {
			log.Fatal(err)
		}

		confirmedInv, err := q.ExpireInvoiceById(ctx, inv.ID)
		assert.NoError(t, err)
		assert.Equal(t, db.InvoiceStatusTypeEXPIRED, confirmedInv.Status)
	})
}

func TestConfirmInvoiceStatusMempoolById(t *testing.T) {
	runInTransaction(t, dbConnPool, func(t *testing.T, tx pgx.Tx) {
		ctx := context.Background()
		q := db.New(tx)

		userId, err := q.CreateUser(ctx)
		if err != nil {
			log.Fatal(err)
		}

		inv, err := createRandTestInvoice(ctx, q, userId)
		if err != nil {
			log.Fatal(err)
		}

		expectedActualAmount := 1.2
		expectedTxId := "txid"

		var actualAmount pgtype.Float8
		if err := actualAmount.Scan(expectedActualAmount); err != nil {
			log.Fatal(err)
		}
		var txId pgtype.Text
		if err := txId.Scan(expectedTxId); err != nil {
			log.Fatal(err)
		}

		confirmedInv, err := q.ConfirmInvoiceStatusMempoolById(ctx, db.ConfirmInvoiceStatusMempoolByIdParams{ID: inv.ID, ActualAmount: actualAmount, TxID: txId})
		assert.NoError(t, err)
		assert.Equal(t, db.InvoiceStatusTypePENDINGMEMPOOL, confirmedInv.Status)
		assert.Equal(t, expectedActualAmount, confirmedInv.ActualAmount.Float64)
		assert.Equal(t, expectedTxId, confirmedInv.TxID.String)
	})
}

func TestFindAllPendingInvoices(t *testing.T) {
	t.Run("Should Return 2 Invoices", func(t *testing.T) {
		runInTransaction(t, dbConnPool, func(t *testing.T, tx pgx.Tx) {
			ctx := context.Background()
			q := db.New(tx)

			userId, err := q.CreateUser(ctx)
			if err != nil {
				log.Fatal(err)
			}

			var expectedInvoices [3]db.Invoice
			for i := 0; i < len(expectedInvoices); i++ {
				inv, err := createRandTestInvoice(ctx, q, userId)
				if err != nil {
					log.Fatal(err)
				}

				expectedInvoices[i] = inv
			}

			var actualAmount pgtype.Float8
			if err := actualAmount.Scan(1.5); err != nil {
				log.Fatal(err)
			}
			var txId pgtype.Text
			if err := txId.Scan("txid"); err != nil {
				log.Fatal(err)
			}
			_, err = q.ConfirmInvoiceStatusMempoolById(ctx, db.ConfirmInvoiceStatusMempoolByIdParams{ID: expectedInvoices[1].ID, ActualAmount: actualAmount, TxID: txId})
			if err != nil {
				log.Fatal(err)
			}

			_, err = q.ExpireInvoiceById(ctx, expectedInvoices[len(expectedInvoices)-1].ID)
			if err != nil {
				log.Fatal(err)
			}

			invoices, err := q.FindAllPendingInvoices(ctx)
			assert.NoError(t, err)

			assert.Equal(t, 2, len(invoices))
		})
	})

	t.Run("Should Return 1 Invoices", func(t *testing.T) {
		runInTransaction(t, dbConnPool, func(t *testing.T, tx pgx.Tx) {
			ctx := context.Background()
			q := db.New(tx)

			userId, err := q.CreateUser(ctx)
			if err != nil {
				log.Fatal(err)
			}

			var expectedInvoices [3]db.Invoice
			for i := 0; i < len(expectedInvoices); i++ {
				inv, err := createRandTestInvoice(ctx, q, userId)
				if err != nil {
					log.Fatal(err)
				}

				expectedInvoices[i] = inv
			}

			var actualAmount pgtype.Float8
			if err := actualAmount.Scan(1.5); err != nil {
				log.Fatal(err)
			}
			var txId pgtype.Text
			if err := txId.Scan("txid"); err != nil {
				log.Fatal(err)
			}
			_, err = q.ConfirmInvoiceStatusMempoolById(ctx, db.ConfirmInvoiceStatusMempoolByIdParams{ID: expectedInvoices[len(expectedInvoices)-1].ID, ActualAmount: actualAmount, TxID: txId})
			if err != nil {
				log.Fatal(err)
			}

			for i := 0; i < len(expectedInvoices)-1; i++ {
				if i%2 == 0 {
					_, err := q.ExpireInvoiceById(ctx, expectedInvoices[i].ID)
					if err != nil {
						log.Fatal(err)
					}

					continue
				}
				_, err := q.ConfirmInvoiceById(ctx, expectedInvoices[i].ID)
				if err != nil {
					log.Fatal(err)
				}
			}

			invoices, err := q.FindAllPendingInvoices(ctx)
			assert.NoError(t, err)

			assert.Equal(t, 1, len(invoices))
		})
	})

	t.Run("Should Return 0 Invoices", func(t *testing.T) {
		runInTransaction(t, dbConnPool, func(t *testing.T, tx pgx.Tx) {
			ctx := context.Background()
			q := db.New(tx)

			userId, err := q.CreateUser(ctx)
			if err != nil {
				log.Fatal(err)
			}

			var expectedInvoices [3]db.Invoice
			for i := 0; i < len(expectedInvoices); i++ {
				inv, err := createRandTestInvoice(ctx, q, userId)
				if err != nil {
					log.Fatal(err)
				}

				expectedInvoices[i] = inv
			}

			for i := 0; i < len(expectedInvoices); i++ {
				if i%2 == 0 {
					_, err := q.ExpireInvoiceById(ctx, expectedInvoices[i].ID)
					if err != nil {
						log.Fatal(err)
					}

					continue
				}
				_, err := q.ConfirmInvoiceById(ctx, expectedInvoices[i].ID)
				if err != nil {
					log.Fatal(err)
				}
			}

			invoices, err := q.FindAllPendingInvoices(ctx)
			assert.NoError(t, err)

			assert.Equal(t, 0, len(invoices))
		})
	})
}

func TestShiftExpiresAtForNonConfirmedInvoices(t *testing.T) {
	runInTransaction(t, dbConnPool, func(t *testing.T, tx pgx.Tx) {
		ctx := context.Background()
		q := db.New(tx)

		userId, err := q.CreateUser(ctx)
		if err != nil {
			log.Fatal(err)
		}

		for i := 0; i < 3; i++ {
			_, err := createRandTestInvoice(ctx, q, userId)
			if err != nil {
				log.Fatal(err)
			}
		}

		invoices, err := q.ShiftExpiresAtForNonConfirmedInvoices(ctx)
		assert.NoError(t, err)

		for i := 0; i < len(invoices); i++ {
			assert.Condition(t, func() (success bool) {
				dur := invoices[i].ExpiresAt.Time.Sub(time.Now().UTC())
				return dur >= 4*time.Minute && dur <= 5*time.Minute
			})
		}
	})
}
