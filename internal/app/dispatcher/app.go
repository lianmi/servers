package dispatcher

import (
	"github.com/google/wire"
	"github.com/lianmi/servers/internal/app/dispatcher/nsqMq"
	"github.com/lianmi/servers/internal/pkg/app"
	"github.com/lianmi/servers/internal/pkg/transports/http"
	"github.com/lianmi/servers/internal/pkg/transports/mqtt"
	// "github.com/lianmi/servers/internal/pkg/transports/nsqclient"
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

// func NewApp(o *Options, logger *zap.Logger, kc *authNsq.NsqClient, hs *http.Server) (*app.Application, error) {
func NewApp(o *Options, logger *zap.Logger, kc *nsqMq.NsqClient, mc *mqtt.MQTTClient, hs *http.Server) (*app.Application, error) {

	a, err := app.New(o.Name, logger, app.NsqOption(kc), app.MQTTOption(mc), app.HttpServerOption(hs))

	if err != nil {
		return nil, errors.Wrap(err, "new app error")
	}

	return a, nil
}

var ProviderSet = wire.NewSet(NewOptions, NewApp)
