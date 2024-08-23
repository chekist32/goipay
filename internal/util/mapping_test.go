package util

import (
	"fmt"
	"log"
	"math"
	"math/rand"
	"testing"
	"time"

	"github.com/chekist32/goipay/internal/db"
	"github.com/chekist32/goipay/internal/dto"
	pb_v1 "github.com/chekist32/goipay/internal/pb/v1"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var (
	pbCoins           []pb_v1.CoinType          = []pb_v1.CoinType{pb_v1.CoinType_XMR, pb_v1.CoinType_BTC, pb_v1.CoinType_LTC, pb_v1.CoinType_ETH, pb_v1.CoinType_TON}
	dbCoins           []db.CoinType             = []db.CoinType{db.CoinTypeXMR, db.CoinTypeBTC, db.CoinTypeLTC, db.CoinTypeETH, db.CoinTypeTON}
	dbInvoiceStatuses []db.InvoiceStatusType    = []db.InvoiceStatusType{db.InvoiceStatusTypePENDING, db.InvoiceStatusTypePENDINGMEMPOOL, db.InvoiceStatusTypeEXPIRED, db.InvoiceStatusTypeCONFIRMED}
	pbInvoiceStatuses []pb_v1.InvoiceStatusType = []pb_v1.InvoiceStatusType{pb_v1.InvoiceStatusType_PENDING, pb_v1.InvoiceStatusType_PENDING_MEMPOOL, pb_v1.InvoiceStatusType_EXPIRED, pb_v1.InvoiceStatusType_CONFIRMED}
)

func TestStringToPgUUID(t *testing.T) {
	t.Parallel()

	t.Run("Should Return Valid PgUUID", func(t *testing.T) {
		t.Parallel()

		expectedUUID := uuid.New()

		pgUUID, err := StringToPgUUID(expectedUUID.String())
		assert.NoError(t, err)
		assert.True(t, pgUUID.Valid)
		assert.Equal(t, [16]byte(expectedUUID), pgUUID.Bytes)
	})

	t.Run("Should Return Error", func(t *testing.T) {
		t.Parallel()

		expectedUUIDStr := uuid.NewString()
		expectedUUIDStr = expectedUUIDStr[1:]

		_, err := StringToPgUUID(expectedUUIDStr)
		assert.Error(t, err)
	})
}

func TestPgUUIDToString(t *testing.T) {
	t.Parallel()

	for i := 0; i < 5; i++ {
		t.Run(fmt.Sprintf("TestPgUUIDToString #%v", i), func(t *testing.T) {
			expectedUUIDStr := uuid.NewString()
			pgUUID, err := StringToPgUUID(expectedUUIDStr)
			if err != nil {
				log.Fatal(err)
			}

			assert.Equal(t, expectedUUIDStr, PgUUIDToString(*pgUUID))
		})
	}
}

func TestPbCoinToDbCoin(t *testing.T) {
	t.Parallel()

	t.Run("Should Return Valid DbCoin For PbCoin", func(t *testing.T) {
		for i := 0; i < len(pbCoins); i++ {
			t.Run(fmt.Sprintf("Should Return Valid DbCoin For PbCoin(%v)", pbCoins[i]), func(t *testing.T) {
				expectedDbCoin := dbCoins[i]

				dbCoin, err := PbCoinToDbCoin(pbCoins[i])
				assert.NoError(t, err)
				assert.Equal(t, expectedDbCoin, dbCoin)
			})
		}
	})

	t.Run("Should Return Error", func(t *testing.T) {
		_, err := PbCoinToDbCoin(math.MaxInt32)
		assert.Error(t, err)
		assert.ErrorIs(t, err, invalidProtoBufCoinTypeErr)
	})

}

func TestDbCoinToPbCoin(t *testing.T) {
	t.Parallel()

	t.Run("Should Return Valid PbCoin For DbCoin", func(t *testing.T) {
		for i := 0; i < len(pbCoins); i++ {
			t.Run(fmt.Sprintf("Should Return Valid PbCoin For DbCoin(%v)", dbCoins[i]), func(t *testing.T) {
				expectedPbCoin := pbCoins[i]

				pbCoin, err := DbCoinToPbCoin(dbCoins[i])
				assert.NoError(t, err)
				assert.Equal(t, expectedPbCoin, pbCoin)
			})
		}
	})

	t.Run("Should Return Error", func(t *testing.T) {
		_, err := DbCoinToPbCoin(db.CoinType(uuid.NewString()))
		assert.Error(t, err)
		assert.ErrorIs(t, err, invalidDbCoinTypeErr)
	})

}

func TestDbInvoiceStatusToPbInvoiceStatus(t *testing.T) {
	t.Run("Should Return Valid PbInvoiceStatus For DbInvoiceStatus", func(t *testing.T) {
		for i := 0; i < len(dbInvoiceStatuses); i++ {
			t.Run(fmt.Sprintf("Should Return Valid PbInvoiceStatus For DbInvoiceStatus(%v)", dbInvoiceStatuses[i]), func(t *testing.T) {
				expectedPbInvoiceStatus := pbInvoiceStatuses[i]

				pbInvoiceStatus, err := DbInvoiceStatusToPbInvoiceStatus(dbInvoiceStatuses[i])
				assert.NoError(t, err)
				assert.Equal(t, expectedPbInvoiceStatus, pbInvoiceStatus)
			})
		}
	})

	t.Run("Should Return Error", func(t *testing.T) {
		_, err := DbInvoiceStatusToPbInvoiceStatus(db.InvoiceStatusType(uuid.NewString()))
		assert.Error(t, err)
		assert.ErrorIs(t, err, invalidDbStatusTypeErr)
	})
}

func TestDbInvoiceToPbInvoice(t *testing.T) {
	idStr := uuid.NewString()
	actualAmountFloat64 := rand.Float64()
	createdAtTime := time.Now().UTC()
	expiresAtTime := createdAtTime.Add(time.Duration(rand.Intn(math.MaxInt)+1) * time.Minute)
	txIdStr := uuid.NewString()
	userIdStr := uuid.NewString()

	var id pgtype.UUID
	if err := id.Scan(idStr); err != nil {
		log.Fatal(err)
	}
	var actualAmount pgtype.Float8
	if err := actualAmount.Scan(actualAmountFloat64); err != nil {
		log.Fatal(err)
	}
	var createdAt pgtype.Timestamptz
	if err := createdAt.Scan(createdAtTime); err != nil {
		log.Fatal(err)
	}
	var expiresAt pgtype.Timestamptz
	if err := expiresAt.Scan(expiresAtTime); err != nil {
		log.Fatal(err)
	}
	var txId pgtype.Text
	if err := txId.Scan(txIdStr); err != nil {
		log.Fatal(err)
	}
	var userId pgtype.UUID
	if err := userId.Scan(userIdStr); err != nil {
		log.Fatal(err)
	}

	dbInv := db.Invoice{
		ID:                    id,
		CryptoAddress:         uuid.NewString(),
		Coin:                  db.CoinTypeBTC,
		RequiredAmount:        rand.Float64(),
		ActualAmount:          actualAmount,
		ConfirmationsRequired: int16(rand.Intn(math.MaxInt16)),
		CreatedAt:             createdAt,
		ConfirmedAt:           pgtype.Timestamptz{},
		Status:                db.InvoiceStatusTypePENDING,
		ExpiresAt:             expiresAt,
		TxID:                  txId,
		UserID:                userId,
	}

	expectedPbInvoice := pb_v1.Invoice{
		Id:                    idStr,
		CryptoAddress:         dbInv.CryptoAddress,
		Coin:                  pb_v1.CoinType_BTC,
		RequiredAmount:        dbInv.RequiredAmount,
		ActualAmount:          actualAmountFloat64,
		ConfirmationsRequired: uint32(dbInv.ConfirmationsRequired),
		CreatedAt:             timestamppb.New(createdAtTime),
		ConfirmedAt:           timestamppb.New(dbInv.ConfirmedAt.Time),
		Status:                pb_v1.InvoiceStatusType_PENDING,
		ExpiresAt:             timestamppb.New(expiresAtTime),
		TxId:                  txIdStr,
		UserId:                userIdStr,
	}

	assert.Equal(t, expectedPbInvoice, *DbInvoiceToPbInvoice(&dbInv))
}

func TestPbNewInvoiceToProcessorNewInvoice(t *testing.T) {
	userId := uuid.NewString()
	amount := rand.Float64()
	timeout := rand.Uint64()
	confirmations := rand.Uint32()

	newInv := pb_v1.CreateInvoiceRequest{
		UserId:        userId,
		Coin:          pb_v1.CoinType_BTC,
		Amount:        amount,
		Timeout:       timeout,
		Confirmations: confirmations,
	}

	expectedProcessorNewInvoice := dto.NewInvoiceRequest{
		UserId:        userId,
		Coin:          db.CoinTypeBTC,
		Amount:        amount,
		Timeout:       timeout,
		Confirmations: confirmations,
	}

	assert.Equal(t, expectedProcessorNewInvoice, *PbNewInvoiceToProcessorNewInvoice(&newInv))
}
