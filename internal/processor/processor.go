package processor

import (
	"context"
	"errors"
	"time"

	"github.com/chekist32/goipay/internal/db"
	"github.com/chekist32/goipay/internal/dto"
	"github.com/chekist32/goipay/internal/util"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog"
)

const (
	persist_cache_timeout time.Duration = 1 * time.Minute
)

var (
	unimplementedError error = errors.New("The coin is unimplemented")
)

type PaymentProcessor struct {
	dbConnPool *pgxpool.Pool

	ctx context.Context
	log *zerolog.Logger

	invoiceCn      chan db.Invoice
	newInvoicesCns *util.SyncMapTypeSafe[string, chan db.Invoice]

	xmr *xmrProcessor
}

func (p *PaymentProcessor) loadPersistedPendingInvoices() error {
	q, tx, err := util.InitDbQueriesWithTx(p.ctx, p.dbConnPool)
	if err != nil {
		p.log.Err(err).Msg(util.DefaultFailedSqlTxInitMsg)
		return err
	}

	invoices, err := q.ShiftExpiresAtForNonConfirmedInvoices(p.ctx)
	if err != nil {
		tx.Rollback(p.ctx)
		p.log.Err(err).Str("queryName", "ShiftExpiresAtForNonConfirmedInvoices").Msg(util.DefaultFailedSqlQueryMsg)
		return err
	}

	tx.Commit(p.ctx)

	for i := 0; i < len(invoices); i++ {
		switch invoices[i].Coin {
		case db.CoinTypeXMR:
			go p.xmr.handleInvoice(p.ctx, invoices[i])
		// TODO: Add impelmentation for BTC
		case db.CoinTypeBTC:
		// TODO: Add impelmentation for LTC
		case db.CoinTypeLTC:
		// TODO: Add impelmentation for ETH
		case db.CoinTypeETH:
		// TODO: Add impelmentation for TON
		case db.CoinTypeTON:
		}
	}

	return nil
}

func (p *PaymentProcessor) load() error {
	go func() {
		for {
			select {
			case tx := <-p.invoiceCn:
				p.newInvoicesCns.Range(func(key string, cn chan db.Invoice) bool {
					go func() {
						select {
						case cn <- tx:
							return
						case <-time.After(util.SEND_TIMEOUT):
							p.newInvoicesCns.Delete(key)
							return
						case <-p.ctx.Done():
							return
						}
					}()

					return true
				})

				p.log.Info().Msgf("Transaction %v changed status to %v", util.PgUUIDToString(tx.ID), tx.Status)
			case <-p.ctx.Done():
				return
			}
		}
	}()

	if err := p.loadPersistedPendingInvoices(); err != nil {
		return err
	}

	if err := p.xmr.load(p.ctx); err != nil {
		return err
	}

	return nil
}

func (p *PaymentProcessor) HandleNewInvoice(req *dto.NewInvoiceRequest) (*db.Invoice, error) {
	switch req.Coin {
	case db.CoinTypeXMR:
		return p.xmr.handleInvoicePbReq(p.ctx, req)
	// TODO: Add impelmentation for BTC
	case db.CoinTypeBTC:
		return nil, unimplementedError
	// TODO: Add impelmentation for LTC
	case db.CoinTypeLTC:
		return nil, unimplementedError
	// TODO: Add impelmentation for ETH
	case db.CoinTypeETH:
		return nil, unimplementedError
	// TODO: Add impelmentation for TON
	case db.CoinTypeTON:
		return nil, unimplementedError
	}

	return nil, errors.New("invalid coin type")
}

func (p *PaymentProcessor) NewInvoicesChan() <-chan db.Invoice {
	cn := make(chan db.Invoice)
	p.newInvoicesCns.Store(uuid.NewString(), cn)
	return cn
}

func NewPaymentProcessor(ctx context.Context, dbConnPool *pgxpool.Pool, c *dto.DaemonsConfig, log *zerolog.Logger) (*PaymentProcessor, error) {
	invoiceCn := make(chan db.Invoice)

	xmr, err := newXmrProcessor(dbConnPool, invoiceCn, c, log)
	if err != nil {
		return nil, err
	}

	pp := &PaymentProcessor{
		dbConnPool:     dbConnPool,
		invoiceCn:      invoiceCn,
		newInvoicesCns: &util.SyncMapTypeSafe[string, chan db.Invoice]{},
		xmr:            xmr,
		ctx:            ctx,
		log:            log,
	}
	if err := pp.load(); err != nil {
		return nil, err
	}

	return pp, nil
}
