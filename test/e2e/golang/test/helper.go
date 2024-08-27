package e2e

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/docker/go-connections/nat"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/network"
	"github.com/testcontainers/testcontainers-go/wait"
)

func spinUpPostgresContainer() (testcontainers.Container, func(ctx context.Context), error) {
	ctx := context.Background()

	net, err := network.New(ctx)
	if err != nil {
		log.Printf("Could not create a new docker network: %s", err)
		return nil, nil, err
	}

	postgresReqEnv := map[string]string{
		"POSTGRES_DB":       "postgres",
		"POSTGRES_USER":     "postgres",
		"POSTGRES_PASSWORD": "postgres",
	}
	postgresReq := testcontainers.ContainerRequest{
		Image:        "postgres:16-alpine",
		ExposedPorts: []string{"5432/tcp"},
		Networks:     []string{net.Name},
		Env:          postgresReqEnv,
		WaitingFor: wait.ForSQL("5432/tcp", "pgx", func(host string, port nat.Port) string {
			return fmt.Sprintf("postgres://%v:%v@localhost:%s/%v", postgresReqEnv["POSTGRES_USER"], postgresReqEnv["POSTGRES_PASSWORD"], port.Port(), postgresReqEnv["POSTGRES_DB"])
		}),
	}
	postgres, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: postgresReq,
		Started:          true,
	})
	if err != nil {
		log.Printf("Could not start postgres: %s", err)
		return nil, nil, err
	}

	postgresIp, err := postgres.ContainerIP(ctx)
	if err != nil {
		log.Fatal(err)
	}

	migrationsReqEnv := map[string]string{
		"GOOSE_DRIVER": "postgres",
		"GOOSE_DBSTRING": fmt.Sprintf(
			"host=%v port=5432 user=%v password=%v dbname=%v",
			postgresIp,
			postgresReqEnv["POSTGRES_USER"],
			postgresReqEnv["POSTGRES_PASSWORD"],
			postgresReqEnv["POSTGRES_DB"],
		),
	}

	migrationsReq := testcontainers.ContainerRequest{
		Image:    "ghcr.io/kukymbr/goose-docker:3.21.1",
		Env:      migrationsReqEnv,
		Networks: []string{net.Name},
		Mounts: testcontainers.ContainerMounts{
			{
				Source: testcontainers.GenericBindMountSource{
					HostPath: fmt.Sprintf("%v/../../../../sql/migrations", os.Getenv("PWD")),
				},
				Target: testcontainers.ContainerMountTarget("/migrations"),
			},
		},
		WaitingFor: wait.ForExit(),
	}
	migrations, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: migrationsReq,
		Started:          true,
	})
	if err != nil {
		log.Printf("Could not start migrations: %s", err)
		postgres.Terminate(ctx)
		return nil, nil, err
	}

	closeHandler := func(ctx context.Context) {
		migrations.Terminate(ctx)
		postgres.Terminate(ctx)
		net.Remove(ctx)
	}

	return postgres, closeHandler, nil
}

func spinUpMoneroWalletRpcContainer() (testcontainers.Container, func(ctx context.Context), error) {
	ctx := context.Background()

	moneroRpcWalletReq := testcontainers.ContainerRequest{
		Image:        "chekist32/monero-wallet-rpc:0.18.3.4",
		ExposedPorts: []string{"38083/tcp"},
		Mounts: testcontainers.ContainerMounts{
			{
				Source: testcontainers.GenericBindMountSource{
					HostPath: fmt.Sprintf("%v/../resources/spend_wallet", os.Getenv("PWD")),
				},
				Target: testcontainers.ContainerMountTarget("/monero/wallet"),
			},
		},
		Cmd: []string{
			"--stagenet",
			"--daemon-address=stagenet.community.rino.io:38081",
			"--trusted-daemon",
			"--rpc-bind-port=38083",
			"--disable-rpc-login",
			"--wallet-dir=/monero/wallet",
		},
		WaitingFor: wait.ForLog("Starting wallet RPC server"),
	}
	moneroRpcWallet, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: moneroRpcWalletReq,
		Started:          true,
	})
	if err != nil {
		return nil, nil, err
	}

	return moneroRpcWallet, func(ctx context.Context) { moneroRpcWallet.Terminate(ctx) }, nil
}
