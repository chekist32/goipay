package app

import (
	"context"
	"fmt"
	"net"
	"os"

	"github.com/chekist32/goipay/internal/dto"
	handler_v1 "github.com/chekist32/goipay/internal/handler/v1"
	pb_v1 "github.com/chekist32/goipay/internal/pb/v1"
	"github.com/chekist32/goipay/internal/processor"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"gopkg.in/yaml.v3"
)

type AppMode string

const (
	DEV_APP_MODE  AppMode = "dev"
	PROD_APP_MODE AppMode = "prod"
)

type AppConfigDaemon struct {
	Url  string `yaml:"url"`
	User string `yaml:"user"`
	Pass string `yaml:"pass"`
}

type AppConfig struct {
	Mode AppMode `yaml:"mode"`

	Server struct {
		Host string `yaml:"host"`
		Port string `yaml:"port"`
	} `yaml:"server"`

	Database struct {
		Host string `yaml:"host"`
		Port string `yaml:"port"`
		User string `yaml:"user"`
		Pass string `yaml:"pass"`
		Name string `yaml:"name"`
	} `yaml:"database"`

	Coin struct {
		Xmr struct {
			Daemon AppConfigDaemon `yaml:"daemon"`
		} `yaml:"xmr"`
	} `yaml:"coin"`
}

func NewAppConfig(path string) (*AppConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var conf AppConfig
	if err := yaml.Unmarshal(data, &conf); err != nil {
		return nil, err
	}

	conf.Mode = AppMode(os.ExpandEnv(string(conf.Mode)))

	conf.Server.Host = os.ExpandEnv(conf.Server.Host)
	conf.Server.Port = os.ExpandEnv(conf.Server.Port)

	conf.Database.Host = os.ExpandEnv(conf.Database.Host)
	conf.Database.Port = os.ExpandEnv(conf.Database.Port)
	conf.Database.User = os.ExpandEnv(conf.Database.User)
	conf.Database.Pass = os.ExpandEnv(conf.Database.Pass)
	conf.Database.Name = os.ExpandEnv(conf.Database.Name)

	conf.Coin.Xmr.Daemon.Url = os.ExpandEnv(conf.Coin.Xmr.Daemon.Url)
	conf.Coin.Xmr.Daemon.User = os.ExpandEnv(conf.Coin.Xmr.Daemon.User)
	conf.Coin.Xmr.Daemon.Pass = os.ExpandEnv(conf.Coin.Xmr.Daemon.Pass)

	return &conf, nil
}

type App struct {
	ctxCancel context.CancelFunc

	config *AppConfig
	log    *zerolog.Logger

	dbConnPool       *pgxpool.Pool
	paymentProcessor *processor.PaymentProcessor
}

func (a *App) Start(ctx context.Context) error {
	if err := a.dbConnPool.Ping(ctx); err != nil {
		a.log.Err(err).Msg("failed to connect to database")
		return err
	}
	defer a.dbConnPool.Close()

	lis, err := net.Listen("tcp", a.config.Server.Host+":"+a.config.Server.Port)
	if err != nil {
		a.log.Fatal().Msgf("failed to listen on port %v: %v", a.config.Server.Port, err)
	}
	g := grpc.NewServer(
		grpc.UnaryInterceptor(NewRequestLoggingInterceptor(a.log).Intercepte),
	)
	pb_v1.RegisterUserServiceServer(g, handler_v1.NewUserGrpc(a.dbConnPool, a.log))
	pb_v1.RegisterInvoiceServiceServer(g, handler_v1.NewInvoiceGrpc(a.dbConnPool, a.paymentProcessor, a.log))

	if a.config.Mode == DEV_APP_MODE {
		reflection.Register(g)
	}

	ch := make(chan error, 1)
	go func() {
		if err := g.Serve(lis); err != nil {
			a.log.Err(err).Msg("failed to start server")
			ch <- err
		}
		close(ch)
	}()

	a.log.Info().Msgf("Starting server %v\n", lis.Addr())

	select {
	case err = <-ch:
		return err
	case <-ctx.Done():
		a.ctxCancel()
		g.GracefulStop()
		return nil
	}
}

func appConfigToDaemonsConfig(c *AppConfig) *dto.DaemonsConfig {
	acdTodc := func(c *AppConfigDaemon) *dto.DaemonConfig {
		return &dto.DaemonConfig{
			Url:  c.Url,
			User: c.User,
			Pass: c.Pass,
		}
	}

	return &dto.DaemonsConfig{
		Xmr: *acdTodc(&c.Coin.Xmr.Daemon),
	}
}

func getLogger() *zerolog.Logger {
	logger := zerolog.New(zerolog.NewConsoleWriter()).With().Timestamp().Caller().Logger()
	return &logger
}

func NewApp(pathToConfig string) *App {
	ctx, cancel := context.WithCancel(context.Background())
	log := getLogger()

	conf, err := NewAppConfig(pathToConfig)
	if err != nil {
		log.Fatal().Err(err)
	}

	dbUrl := fmt.Sprintf("postgresql://%v:%v@%v:%v/%v", conf.Database.User, conf.Database.Pass, conf.Database.Host, conf.Database.Port, conf.Database.Name)
	connPool, err := pgxpool.New(ctx, dbUrl)
	if err != nil {
		log.Fatal().Err(err)
	}

	pp, err := processor.NewPaymentProcessor(ctx, connPool, appConfigToDaemonsConfig(conf), log)
	if err != nil {
		log.Fatal().Err(err)
	}

	return &App{
		log:              log,
		ctxCancel:        cancel,
		config:           conf,
		dbConnPool:       connPool,
		paymentProcessor: pp,
	}
}
