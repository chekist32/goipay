package e2e

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"os"
	"testing"
	"time"

	"github.com/chekist32/go-monero-rpc-client/util"
	"github.com/chekist32/go-monero/wallet"
	pb_v1 "github.com/chekist32/goipay/e2e/internal/pb/v1"
	"github.com/docker/go-connections/nat"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func spinUpGoipayContainer() (testcontainers.Container, func(ctx context.Context), error) {
	ctx := context.Background()

	postgres, cls, err := spinUpPostgresContainer()
	if err != nil {
		return nil, nil, err
	}

	nets, err := postgres.Networks(ctx)
	if err != nil {
		return nil, nil, err
	}

	postgresIp, err := postgres.ContainerIP(ctx)
	if err != nil {
		log.Fatal(err)
	}

	goipayReqEnv := map[string]string{
		"MODE": "dev",

		"SERVER_HOST": "0.0.0.0",
		"SERVER_PORT": "3000",

		"DATABASE_HOST": postgresIp,
		"DATABASE_PORT": "5432",
		"DATABASE_USER": "postgres",
		"DATABASE_PASS": "postgres",
		"DATABASE_NAME": "postgres",

		"XMR_DAEMON_URL": "http://node.monerodevs.org:38089",
	}
	goipayReq := testcontainers.ContainerRequest{
		FromDockerfile: testcontainers.FromDockerfile{
			Context: "../../../../",
		},
		ExposedPorts: []string{fmt.Sprintf("%v/tcp", goipayReqEnv["SERVER_PORT"])},
		Networks:     []string{nets[0]},
		Env:          goipayReqEnv,
		WaitingFor:   wait.ForListeningPort(nat.Port(fmt.Sprintf("%v/tcp", goipayReqEnv["SERVER_PORT"]))).WithPollInterval(5 * time.Second),
		LogConsumerCfg: &testcontainers.LogConsumerConfig{
			Consumers: []testcontainers.LogConsumer{&testcontainers.StdoutLogConsumer{}},
		},
	}

	goipay, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: goipayReq,
		Started:          true,
	})
	if err != nil {
		return nil, nil, err
	}

	closeHandler := func(ctx context.Context) {
		goipay.Terminate(ctx)
		cls(ctx)
	}

	return goipay, closeHandler, nil

}

func TestSimpleRegisterUser(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	log := zerolog.New(zerolog.NewConsoleWriter()).With().Timestamp().Caller().Logger()

	goipay, goipayClean, err := spinUpGoipayContainer()
	if err != nil {
		log.Err(err).Msg("")
		assert.FailNow(t, "")
	}
	moneroWalletRpc, moneroWalletRpcClean, err := spinUpMoneroWalletRpcContainer()
	if err != nil {
		log.Err(err).Msg("")
		assert.FailNow(t, "")
	}

	goipayPort, err := goipay.MappedPort(ctx, "3000/tcp")
	if err != nil {
		log.Err(err).Msg("")
		assert.FailNow(t, "")
	}

	conn, err := grpc.NewClient(fmt.Sprintf("localhost:%v", goipayPort.Port()), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Err(err).Msg("")
		assert.FailNow(t, "")
	}

	userClient := pb_v1.NewUserServiceClient(conn)
	invoiceClient := pb_v1.NewInvoiceServiceClient(conn)

	moneroWalletRpcPort, err := moneroWalletRpc.MappedPort(ctx, "38083/tcp")
	if err != nil {
		log.Err(err).Msg("")
		assert.FailNow(t, "")
	}

	xmrWallet := wallet.New(wallet.Config{
		Address: fmt.Sprintf("http://localhost:%v", moneroWalletRpcPort.Port()),
	})
	if err := xmrWallet.OpenWallet(&wallet.RequestOpenWallet{Filename: "goipay_test", Password: os.Getenv("XMR_SPEND_WALLET_PASSWORD")}); err != nil {
		log.Err(err).Msg("")
		assert.FailNow(t, "")
	}

	t.Cleanup(func() {
		goipayClean(ctx)
		xmrWallet.CloseWallet()
		moneroWalletRpcClean(ctx)
		cancel()
	})

	// Step #1 - Register user
	log.Debug().Msg("Step #1 - Register user - Start")
	res1, err := userClient.RegisterUser(ctx, &pb_v1.RegisterUserRequest{})
	assert.NoError(t, err)
	assert.NotEmpty(t, res1.UserId)
	log.Debug().Msg("Step #1 - Register user - End")

	// Step #2 - Update XMR Crypto Keys
	log.Debug().Msg("Step #2 - Update XMR Crypto Keys - Start")
	_, err = userClient.UpdateCryptoKeys(ctx, &pb_v1.UpdateCryptoKeysRequest{
		UserId: res1.UserId,
		XmrReq: &pb_v1.XmrKeysUpdateRequest{
			PrivViewKey: "8aa763d1c8d9da4ca75cb6ca22a021b5cca376c1367be8d62bcc9cdf4b926009",
			PubSpendKey: "38e9908d33d034de0ba1281aa7afe3907b795cea14852b3d8fe276e8931cb130",
		},
	})
	assert.NoError(t, err)
	log.Debug().Msg("Step #2 - Update XMR Crypto Keys - End")

	// Step #3 - Create New Invoice + GET Invoice Status Stream

	// Step #3.1 - GET Invoice Status Stream
	log.Debug().Msg("Step #3.1 - GET Invoice Status Stream - Start")
	invoiceStatusStream, err := invoiceClient.InvoiceStatusStream(ctx, &pb_v1.InvoiceStatusStreamRequest{})
	assert.NoError(t, err)
	log.Debug().Msg("Step #3.1 - GET Invoice Status Stream - End")

	// Step #3.2 - Create New Invoice
	log.Debug().Msg("Step #3.2 - Create New Invoice - Start")
	createInvoiceReq := &pb_v1.CreateInvoiceRequest{
		UserId:  res1.UserId,
		Coin:    pb_v1.CoinType_XMR,
		Amount:  rand.Float64(),
		Timeout: uint64((10 * time.Minute).Seconds()),
	}
	res3, err := invoiceClient.CreateInvoice(ctx, createInvoiceReq)
	assert.NoError(t, err)

	expiresAt := time.Now().UTC().Add(time.Duration(createInvoiceReq.Timeout) * time.Second)

	log.Debug().Msg("Step #3.2 - Create New Invoice - End")

	// SubStep - Setup Goroutine for invoiceStatusStream
	invoiceStatusStreamCn := make(chan *pb_v1.InvoiceStatusStreamResponse)
	invoiceStatusStreamErrCn := make(chan error)

	go func() {
		for {
			select {
			case <-ctx.Done():
			default:
				res, err := invoiceStatusStream.Recv()
				if err != nil {
					invoiceStatusStreamErrCn <- err
					break
				}
				invoiceStatusStreamCn <- res
			}
		}
	}()

	// Step #5 - Send XMR on Received Address
	log.Debug().Msg("Step #5 - Send XMR on Received Address - Start")
	_, err = xmrWallet.Transfer(&wallet.RequestTransfer{
		Priority: wallet.PriorityNormal,
		RingSize: 16,
		Destinations: []*wallet.Destination{
			{Amount: util.Float64ToXMR(createInvoiceReq.Amount + 0.001), Address: res3.Address},
		},
	})
	if err != nil {
		log.Err(err).Msg("")
		assert.FailNow(t, "")
	}
	log.Debug().Msg("Step #5 - Send XMR on Received Address - End")

	// Step #6 - Wait For Status PENDING_MEMPOOL
	log.Debug().Msg("Step #6 - Wait For Status PENDING_MEMPOOL - Start")

	select {
	case r := <-invoiceStatusStreamCn:
		log.Info().Fields(map[string]interface{}{"invoice": r.Invoice}).Msg("Step #6 - Received Invoice")
		assert.Equal(t, res3.PaymentId, r.Invoice.Id)
		assert.Equal(t, pb_v1.InvoiceStatusType_PENDING_MEMPOOL, r.Invoice.Status)
	case e := <-invoiceStatusStreamErrCn:
		log.Err(e).Msg("Step #6 - An error occured")
		assert.FailNow(t, "")
	case <-time.After(expiresAt.Sub(time.Now().UTC())):
		log.Err(nil).Msg("Step #6 - Timeout has been expired")
		assert.FailNow(t, "")
	}
	log.Debug().Msg("Step #6 - Wait For Status PENDING_MEMPOOL - End")

	// Step #7 - Wait For Status CONFIRMED
	log.Debug().Msg("Step #7 - Wait For Status CONFIRMED - Start")

	select {
	case r := <-invoiceStatusStreamCn:
		log.Info().Fields(map[string]interface{}{"invoice": r.Invoice}).Msg("Step #7 - Received Invoice")
		assert.Equal(t, r.Invoice.Id, res3.PaymentId)
		assert.Equal(t, pb_v1.InvoiceStatusType_CONFIRMED, r.Invoice.Status)
	case e := <-invoiceStatusStreamErrCn:
		log.Err(e).Msg("Step #7 - An error occured")
		assert.FailNow(t, "")
	case <-time.After(expiresAt.Sub(time.Now().UTC())):
		log.Err(nil).Msg("Step #7 - Timeout has been expired")
		assert.FailNow(t, "")
	}

	log.Debug().Msg("Step #7 - Wait For Status CONFIRMED - End")
}
