package listener

import (
	"context"
	"sync/atomic"
	"time"

	"github.com/chekist32/go-monero/daemon"
	"github.com/chekist32/goipay/internal/util"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
)

type transactionPoolSync struct {
	txs map[string]bool
}

type blockSync struct {
	lastBlockHeight atomic.Uint64
}

type DaemonRpcClientExecutor struct {
	log *zerolog.Logger

	client daemon.IDaemonRpcClient

	txPoolChns   *util.SyncMapTypeSafe[string, chan daemon.MoneroTx]
	newBlockChns *util.SyncMapTypeSafe[string, chan daemon.GetBlockResult]

	isStarted bool
	stop      chan struct{}

	blockSync           blockSync
	transactionPoolSync transactionPoolSync
}

func (d *DaemonRpcClientExecutor) syncBlock(ctx context.Context) {
	height, err := d.client.GetLastBlockHeader(true)
	if err != nil {
		d.log.Err(err).Str("method", "last_block_header").Msg(util.DefaultFailedFetchingXMRDaemonMsg)
		return
	}

	for {
		select {
		case <-ctx.Done():
			return
		default:
			if height.Result.BlockHeader.Height <= d.blockSync.lastBlockHeight.Load() {
				return
			}

			block, err := d.client.GetBlockByHeight(true, d.blockSync.lastBlockHeight.Load())
			if err != nil {
				d.log.Err(err).Str("method", "get_block").Msg(util.DefaultFailedFetchingXMRDaemonMsg)
				return
			}
			d.log.Info().Msgf("Synced blockheight: %v", block.Result.BlockHeader.Height)

			d.newBlockChns.Range(func(key string, cn chan daemon.GetBlockResult) bool {
				go func() {
					select {
					case cn <- block.Result:
						return
					case <-time.After(MIN_SYNC_TIMEOUT):
						d.newBlockChns.Delete(key)
						return
					}
				}()
				return true
			})

			d.blockSync.lastBlockHeight.Add(1)
		}
	}
}

func (d *DaemonRpcClientExecutor) syncTransactionPool() {
	txs, err := d.client.GetTransactionPool()
	if err != nil {
		d.log.Err(err).Str("method", "get_transaction_pool").Msg(util.DefaultFailedFetchingXMRDaemonMsg)
		return
	}

	fetchedTxs := txs.Transactions
	prevTxs := d.transactionPoolSync.txs
	newTxs := make(map[string]bool)

	for i := 0; i < len(fetchedTxs); i++ {
		newTxs[fetchedTxs[i].IdHash] = true

		if prevTxs[fetchedTxs[i].IdHash] {
			continue
		}

		d.txPoolChns.Range(func(key string, cn chan daemon.MoneroTx) bool {
			go func() {
				select {
				case cn <- fetchedTxs[i]:
					return
				case <-time.After(MIN_SYNC_TIMEOUT):
					d.txPoolChns.Delete(key)
					return
				}
			}()

			return true
		})
	}

	d.transactionPoolSync.txs = newTxs
}

func (d *DaemonRpcClientExecutor) sync(blockTimeout time.Duration, txPoolTimeout time.Duration) {
	t1 := time.NewTicker(blockTimeout)
	t2 := time.NewTicker(txPoolTimeout)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		for {
			s := ctx.Done()
			select {
			case <-s:
				return
			case <-t1.C:
				d.syncBlock(ctx)
			}
		}
	}()

	go func() {
		for {
			s := ctx.Done()
			select {
			case <-s:
				return
			case <-t2.C:
				d.syncTransactionPool()
			}
		}
	}()

	<-d.stop
	d.isStarted = false
}

func (d *DaemonRpcClientExecutor) Start(startBlock uint64) {
	if d.isStarted {
		return
	}
	d.isStarted = true
	d.blockSync.lastBlockHeight.Store(startBlock)

	go d.sync(MIN_SYNC_TIMEOUT, MIN_SYNC_TIMEOUT/2)
}

func (d *DaemonRpcClientExecutor) Stop() {
	d.stop <- struct{}{}
}

func (d *DaemonRpcClientExecutor) NewBlockChan() <-chan daemon.GetBlockResult {
	cn := make(chan daemon.GetBlockResult)
	d.newBlockChns.Store(uuid.NewString(), cn)
	return cn
}

func (d *DaemonRpcClientExecutor) NewTxPoolChan() <-chan daemon.MoneroTx {
	cn := make(chan daemon.MoneroTx)
	d.txPoolChns.Store(uuid.NewString(), cn)
	return cn
}

func (d *DaemonRpcClientExecutor) LastSyncedBlockHeight() uint64 {
	return d.blockSync.lastBlockHeight.Load()
}

func NewDaemonRpcClientExecutor(client daemon.IDaemonRpcClient, log *zerolog.Logger) *DaemonRpcClientExecutor {
	return &DaemonRpcClientExecutor{
		log:                 log,
		client:              client,
		transactionPoolSync: transactionPoolSync{txs: make(map[string]bool)},
		isStarted:           false,
		stop:                make(chan struct{}),
		txPoolChns:          &util.SyncMapTypeSafe[string, chan daemon.MoneroTx]{},
		newBlockChns:        &util.SyncMapTypeSafe[string, chan daemon.GetBlockResult]{},
	}
}
