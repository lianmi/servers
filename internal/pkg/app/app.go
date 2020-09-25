package app

import (
	"github.com/pkg/errors"
	"os"
	"os/signal"
	"syscall"

	"github.com/google/wire"
	authNsq "github.com/lianmi/servers/internal/app/authservice/nsqBackend"
	chatKafka "github.com/lianmi/servers/internal/app/chatservice/kafkaBackend"
	orderKafka "github.com/lianmi/servers/internal/app/orderservice/kafkaBackend"
	walletKafka "github.com/lianmi/servers/internal/app/walletservice/kafkaBackend"
	"github.com/lianmi/servers/internal/pkg/transports/grpc"
	"github.com/lianmi/servers/internal/pkg/transports/http"
	"github.com/lianmi/servers/internal/pkg/transports/mqtt"
	"github.com/lianmi/servers/internal/pkg/transports/nsqclient"

	"go.uber.org/zap"
)

type Application struct {
	name              string
	logger            *zap.Logger
	httpServer        *http.Server
	grpcServer        *grpc.Server
	nsqClient         *nsqclient.NsqClient
	authNsqClient     *authNsq.NsqClient
	chatKafkaClient   *chatKafka.KafkaClient
	orderKafkaClient  *orderKafka.KafkaClient
	walletKafkaClient *walletKafka.KafkaClient
	mqttClient        *mqtt.MQTTClient
}

type Option func(app *Application) error

// func RedisOption(redisPool *redis.Pool) Option {
// 	return func(app *Application) error {
// 		app.redisPool = redisPool
// 		return nil
// 	}
// }

func HttpServerOption(svr *http.Server) Option {
	return func(app *Application) error {
		svr.Application(app.name)
		app.httpServer = svr

		return nil
	}
}

func GrpcServerOption(svr *grpc.Server) Option {
	return func(app *Application) error {
		svr.Application(app.name)
		app.grpcServer = svr
		return nil
	}
}

func NsqOption(nc *nsqclient.NsqClient) Option {
	return func(app *Application) error {
		nc.Application(app.name)
		app.nsqClient = nc
		return nil
	}
}

func AuthNsqOption(kbc *authNsq.NsqClient) Option {
	return func(app *Application) error {
		kbc.Application(app.name)
		app.authNsqClient = kbc
		return nil
	}
}

func ChatKafkaOption(kbc *chatKafka.KafkaClient) Option {
	return func(app *Application) error {
		kbc.Application(app.name)
		app.chatKafkaClient = kbc
		return nil
	}
}

func OrderKafkaOption(kbc *orderKafka.KafkaClient) Option {
	return func(app *Application) error {
		kbc.Application(app.name)
		app.orderKafkaClient = kbc
		return nil
	}
}

func WalletKafkaOption(kbc *walletKafka.KafkaClient) Option {
	return func(app *Application) error {
		kbc.Application(app.name)
		app.walletKafkaClient = kbc
		return nil
	}
}

func MQTTOption(mc *mqtt.MQTTClient) Option {
	return func(app *Application) error {
		mc.Application(app.name)
		app.mqttClient = mc
		return nil
	}
}

func New(name string, logger *zap.Logger, options ...Option) (*Application, error) {
	app := &Application{
		name:   name,
		logger: logger.With(zap.String("type", "Application")),
	}

	for _, option := range options {
		if err := option(app); err != nil {
			return nil, err
		}
	}

	return app, nil
}

func (a *Application) Start() error {

	if a.httpServer != nil {
		if err := a.httpServer.Start(); err != nil {
			return errors.Wrap(err, "http server start error")
		}
	}

	if a.grpcServer != nil {
		if err := a.grpcServer.Start(); err != nil {
			return errors.Wrap(err, "grpc server start error")
		}
	}

	if a.authNsqClient != nil {
		if err := a.authNsqClient.Start(); err != nil {
			return errors.Wrap(err, "authservice kafka backend client start error")
		}
	}

	if a.chatKafkaClient != nil {
		if err := a.chatKafkaClient.Start(); err != nil {
			return errors.Wrap(err, "chatservice kafka backend client start error")
		}
	}

	if a.orderKafkaClient != nil {
		if err := a.orderKafkaClient.Start(); err != nil {
			return errors.Wrap(err, "orderservice kafka backend client start error")
		}
	}

	if a.walletKafkaClient != nil {
		if err := a.walletKafkaClient.Start(); err != nil {
			return errors.Wrap(err, "walletservice kafka backend client start error")
		}
	}

	if a.nsqClient != nil {
		if err := a.nsqClient.Start(); err != nil {
			return errors.Wrap(err, "kafka client start error")
		}
	}

	if a.mqttClient != nil {
		if err := a.mqttClient.Start(); err != nil {
			return errors.Wrap(err, "mqtt client start error")
		}
	}

	return nil
}

func (a *Application) AwaitSignal() {
	c := make(chan os.Signal, 1)
	signal.Reset(syscall.SIGTERM, syscall.SIGINT)
	signal.Notify(c, syscall.SIGTERM, syscall.SIGINT)
	select {
	case s := <-c:
		a.logger.Info("receive a signal", zap.String("signal", s.String()))
		if a.httpServer != nil {
			if err := a.httpServer.Stop(); err != nil {
				a.logger.Warn("stop http server error", zap.Error(err))
			}
		}

		if a.grpcServer != nil {
			if err := a.grpcServer.Stop(); err != nil {
				a.logger.Warn("stop grpc server error", zap.Error(err))
			}
		}

		if a.authNsqClient != nil {
			// if err := a.authNsqClient.Stop(); err != nil {
			// 	a.logger.Warn("stop authservice  kafka client error", zap.Error(err))
			// }
		}

		if a.nsqClient != nil {
			// if err := a.nsqClient.Stop(); err != nil {
			// 	a.logger.Warn("stop kafka client error", zap.Error(err))
			// }
		}

		if a.mqttClient != nil {
			if err := a.mqttClient.Stop(); err != nil {
				a.logger.Warn("stop mqtt client error", zap.Error(err))
			}
		}

		os.Exit(0)
	}
}

var ProviderSet = wire.NewSet(New)
