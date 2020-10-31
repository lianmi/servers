package orderservice

import (
	"github.com/google/wire"
	orderNsq "github.com/lianmi/servers/internal/app/orderservice/nsqMq"
	"github.com/lianmi/servers/internal/pkg/app"
	"github.com/lianmi/servers/internal/pkg/transports/grpc"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

type Options struct {
	Name     string
	Addr     string `yaml:"addr"` //127.0.0.1:9092
	Password string `yaml:"password"`
	Db       int    `yaml:"db"`
}

func NewOptions(v *viper.Viper, logger *zap.Logger) (*Options, error) {
	var err error
	o := new(Options)
	if err = v.UnmarshalKey("app", o); err != nil {
		return nil, errors.Wrap(err, "unmarshal app option error")
	}

	logger.Info("load application options success")

	return o, err
}

func NewApp(o *Options, logger *zap.Logger, nc *orderNsq.NsqClient, gs *grpc.Server) (*app.Application, error) {

	a, err := app.New(o.Name, logger, app.OrderNsqOption(nc), app.GrpcServerOption(gs))

	if err != nil {
		return nil, errors.Wrap(err, "new app error")
	}

	return a, nil
}

var ProviderSet = wire.NewSet(NewOptions, NewApp)
