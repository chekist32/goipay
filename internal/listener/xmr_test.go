package listener

import (
	"context"
	"errors"
	"log"
	"math/rand"
	"sync/atomic"
	"testing"
	"time"

	"github.com/chekist32/go-monero/daemon"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockDaemonRpcClient struct {
	mock.Mock
}

func (m *MockDaemonRpcClient) GetLastBlockHeader(includeHex bool) (*daemon.JsonRpcGenericResponse[daemon.GetBlockHeaderResult], error) {
	args := m.Called(includeHex)
	return args.Get(0).(*daemon.JsonRpcGenericResponse[daemon.GetBlockHeaderResult]), args.Error(1)
}

func (m *MockDaemonRpcClient) GetBlockByHeight(includeHex bool, height uint64) (*daemon.JsonRpcGenericResponse[daemon.GetBlockResult], error) {
	args := m.Called(includeHex, height)
	return args.Get(0).(*daemon.JsonRpcGenericResponse[daemon.GetBlockResult]), args.Error(1)
}

func (m *MockDaemonRpcClient) GetTransactionPool() (*daemon.GetTransactionPoolResponse, error) {
	args := m.Called()
	return args.Get(0).(*daemon.GetTransactionPoolResponse), args.Error(1)
}

func (m *MockDaemonRpcClient) GetBlockByHash(fillPowHash bool, hash string) (*daemon.JsonRpcGenericResponse[daemon.GetBlockResult], error) {
	args := m.Called()
	return args.Get(0).(*daemon.JsonRpcGenericResponse[daemon.GetBlockResult]), args.Error(1)
}
func (m *MockDaemonRpcClient) GetBlockCount() (*daemon.JsonRpcGenericResponse[daemon.GetBlockCountResult], error) {
	args := m.Called()
	return args.Get(0).(*daemon.JsonRpcGenericResponse[daemon.GetBlockCountResult]), args.Error(1)
}
func (m *MockDaemonRpcClient) GetBlockHeaderByHash(fillPowHash bool, hash string) (*daemon.JsonRpcGenericResponse[daemon.GetBlockHeaderResult], error) {
	args := m.Called()
	return args.Get(0).(*daemon.JsonRpcGenericResponse[daemon.GetBlockHeaderResult]), args.Error(1)
}
func (m *MockDaemonRpcClient) GetBlockHeaderByHeight(fillPowHash bool, height uint64) (*daemon.JsonRpcGenericResponse[daemon.GetBlockHeaderResult], error) {
	args := m.Called()
	return args.Get(0).(*daemon.JsonRpcGenericResponse[daemon.GetBlockHeaderResult]), args.Error(1)
}
func (m *MockDaemonRpcClient) GetBlockHeadersRange(fillPowHash bool, startHeight uint64, endHeight uint64) (*daemon.JsonRpcGenericResponse[daemon.GetBlockHeadersRangeResult], error) {
	args := m.Called()
	return args.Get(0).(*daemon.JsonRpcGenericResponse[daemon.GetBlockHeadersRangeResult]), args.Error(1)
}
func (m *MockDaemonRpcClient) GetBlockTemplate(wallet string, reverseSize uint64) (*daemon.JsonRpcGenericResponse[daemon.GetBlockTemplateResult], error) {
	args := m.Called()
	return args.Get(0).(*daemon.JsonRpcGenericResponse[daemon.GetBlockTemplateResult]), args.Error(1)
}
func (m *MockDaemonRpcClient) GetCurrentHeight() (*daemon.GetHeightResponse, error) {
	args := m.Called()
	return args.Get(0).(*daemon.GetHeightResponse), args.Error(1)
}
func (m *MockDaemonRpcClient) GetFeeEstimate() (*daemon.JsonRpcGenericResponse[daemon.GetFeeEstimateResult], error) {
	args := m.Called()
	return args.Get(0).(*daemon.JsonRpcGenericResponse[daemon.GetFeeEstimateResult]), args.Error(1)
}
func (m *MockDaemonRpcClient) GetInfo() (*daemon.JsonRpcGenericResponse[daemon.GetInfoResult], error) {
	args := m.Called()
	return args.Get(0).(*daemon.JsonRpcGenericResponse[daemon.GetInfoResult]), args.Error(1)
}
func (m *MockDaemonRpcClient) GetTransactions(txHashes []string, decodeAsJson bool, prune bool, split bool) (*daemon.GetTransactionsResponse, error) {
	args := m.Called()
	return args.Get(0).(*daemon.GetTransactionsResponse), args.Error(1)
}
func (m *MockDaemonRpcClient) GetVersion() (*daemon.JsonRpcGenericResponse[daemon.GetVersionResult], error) {
	args := m.Called()
	return args.Get(0).(*daemon.JsonRpcGenericResponse[daemon.GetVersionResult]), args.Error(1)
}
func (m *MockDaemonRpcClient) OnGetBlockHash(height uint64) (*daemon.JsonRpcGenericResponse[daemon.OnGetBlockHashResult], error) {
	args := m.Called()
	return args.Get(0).(*daemon.JsonRpcGenericResponse[daemon.OnGetBlockHashResult]), args.Error(1)
}
func (m *MockDaemonRpcClient) SetRpcConnection(connection *daemon.RpcConnection) {}
func (m *MockDaemonRpcClient) SubmitBlock(blobData []string) (*daemon.JsonRpcGenericResponse[daemon.SubmitBlockResult], error) {
	args := m.Called()
	return args.Get(0).(*daemon.JsonRpcGenericResponse[daemon.SubmitBlockResult]), args.Error(1)
}

func TestBlockChan(t *testing.T) {
	t.Parallel()

	t.Run("Check NewBlockChan Func", func(t *testing.T) {
		d := new(MockDaemonRpcClient)
		xmr := NewDaemonRpcClientExecutor(d, zerolog.DefaultContextLogger)

		expectedBlockResult := daemon.GetBlockResult{
			BlockDetails: daemon.BlockDetails{
				Timestamp: rand.Uint32(),
			},
		}

		blockCnAmount := atomic.Int32{}
		blockCnAmount.Store(0)

		blockCn := xmr.NewBlockChan()
		xmr.newBlockChns.Range(func(key string, value chan daemon.GetBlockResult) bool {
			go func() {
				blockCnAmount.Add(1)
				value <- expectedBlockResult
			}()

			return true
		})

		actualBlockResult := daemon.GetBlockResult{}
		select {
		case actualBlockResult = <-blockCn:
			break
		case <-time.After(MIN_SYNC_TIMEOUT):
			log.Fatal(errors.New("Timeout has been expired"))
		}

		assert.Equal(t, int32(1), blockCnAmount.Load())
		assert.Equal(t, expectedBlockResult, actualBlockResult)
	})
}

func TestTxPoolChan(t *testing.T) {
	t.Parallel()

	t.Run("Check NewTxPoolChan Func", func(t *testing.T) {
		d := new(MockDaemonRpcClient)
		xmr := NewDaemonRpcClientExecutor(d, zerolog.DefaultContextLogger)

		expectedMoneroTx := daemon.MoneroTx{
			IdHash: uuid.NewString(),
		}

		txPoolCnAmount := atomic.Int32{}
		txPoolCnAmount.Store(0)

		txPoolCn := xmr.NewTxPoolChan()
		xmr.txPoolChns.Range(func(key string, value chan daemon.MoneroTx) bool {
			go func() {
				txPoolCnAmount.Add(1)
				value <- expectedMoneroTx
			}()

			return true
		})

		actualMoneroTx := daemon.MoneroTx{}
		select {
		case actualMoneroTx = <-txPoolCn:
			break
		case <-time.After(MIN_SYNC_TIMEOUT):
			log.Fatal(errors.New("Timeout has been expired"))
		}

		assert.Equal(t, int32(1), txPoolCnAmount.Load())
		assert.Equal(t, expectedMoneroTx, actualMoneroTx)
	})
}

func TestSyncBlock(t *testing.T) {
	lastBlockHeight := rand.Uint64()
	d := new(MockDaemonRpcClient)
	d.On("GetLastBlockHeader", true).Return(
		&daemon.JsonRpcGenericResponse[daemon.GetBlockHeaderResult]{
			Result: daemon.GetBlockHeaderResult{
				BlockHeader: daemon.BlockHeader{
					Height: lastBlockHeight,
				},
			},
		},
		error(nil),
	)

	expectedBlockResult := daemon.GetBlockResult{
		BlockDetails: daemon.BlockDetails{
			Timestamp: rand.Uint32(),
		},
	}
	d.On("GetBlockByHeight", true, lastBlockHeight-1).Return(
		&daemon.JsonRpcGenericResponse[daemon.GetBlockResult]{
			Result: expectedBlockResult,
		},
		error(nil),
	)

	xmr := NewDaemonRpcClientExecutor(d, &zerolog.Logger{})
	xmr.blockSync.lastBlockHeight.Store(lastBlockHeight - 1)
	blockCn := xmr.NewBlockChan()

	ctx := context.Background()
	xmr.syncBlock(ctx)

	actualBlockResult := daemon.GetBlockResult{}
	select {
	case actualBlockResult = <-blockCn:
		break
	case <-time.After(MIN_SYNC_TIMEOUT):
		log.Fatal(errors.New("Timeout has been expired"))
	}

	assert.Equal(t, lastBlockHeight, xmr.blockSync.lastBlockHeight.Load())
	assert.Equal(t, expectedBlockResult, actualBlockResult)
}

func TestSyncTransactionPool(t *testing.T) {
	// 1
	expectedTxs1Map := map[string]daemon.MoneroTx{
		"tx1": {IdHash: "tx1"},
		"tx2": {IdHash: "tx2"},
		"tx3": {IdHash: "tx3"},
		"tx4": {IdHash: "tx4"},
		"tx5": {IdHash: "tx5"},
	}
	expectedTxs1Slice := make([]daemon.MoneroTx, 0)
	for _, tx := range expectedTxs1Map {
		expectedTxs1Slice = append(expectedTxs1Slice, tx)
	}

	d := new(MockDaemonRpcClient)
	d.On("GetTransactionPool").Once().Return(
		&daemon.GetTransactionPoolResponse{
			Transactions: expectedTxs1Slice,
		},
		error(nil),
	)

	xmr := NewDaemonRpcClientExecutor(d, &zerolog.Logger{})
	txPoolCn := xmr.NewTxPoolChan()

	xmr.syncTransactionPool()

	txs1 := make(map[string]daemon.MoneroTx, 0)
	for i := 0; i < len(expectedTxs1Map); i++ {
		select {
		case tx := <-txPoolCn:
			txs1[tx.IdHash] = tx
		case <-time.After(MIN_SYNC_TIMEOUT):
			log.Fatal(errors.New("Timeout has been expired"))
		}
	}

	assert.Equal(t, expectedTxs1Map, txs1)
	assert.Condition(t, func() (success bool) {
		for id := range expectedTxs1Map {
			if !xmr.transactionPoolSync.txs[id] {
				return false
			}
		}

		return true
	})

	// 2
	expectedTxs2Map := map[string]daemon.MoneroTx{
		"tx1": {IdHash: "tx1"},
		"tx3": {IdHash: "tx3"},
		"tx5": {IdHash: "tx5"},
		"tx7": {IdHash: "tx7"},
		"tx6": {IdHash: "tx6"},
	}
	expectedTxs2Slice := make([]daemon.MoneroTx, 0)
	for _, tx := range expectedTxs2Map {
		expectedTxs2Slice = append(expectedTxs2Slice, tx)
	}
	d.On("GetTransactionPool").Return(
		&daemon.GetTransactionPoolResponse{
			Transactions: expectedTxs2Slice,
		},
		error(nil),
	)

	xmr.syncTransactionPool()

	txs2 := make(map[string]daemon.MoneroTx, 0)
	for i := 0; i < 2; i++ {
		select {
		case tx := <-txPoolCn:
			txs2[tx.IdHash] = tx
		case <-time.After(MIN_SYNC_TIMEOUT):
			log.Fatal(errors.New("Timeout has been expired"))
		}
	}

	assert.Equal(t, map[string]daemon.MoneroTx{"tx7": {IdHash: "tx7"}, "tx6": {IdHash: "tx6"}}, txs2)
	assert.Condition(t, func() (success bool) {
		for id := range expectedTxs2Map {
			if !xmr.transactionPoolSync.txs[id] {
				return false
			}
		}

		return true
	})
}
