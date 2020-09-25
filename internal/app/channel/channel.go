package channel

import (
	"github.com/google/wire"
	"github.com/lianmi/servers/internal/pkg/models"
	"github.com/spf13/viper"
)

type NsqMqttChannel struct {
	KafkaChan chan *models.Message
	MTChan    chan *models.Message
}

type Options struct {
}

func NewOptions(v *viper.Viper) (*Options, error) {
	var (
		// err error
		o = new(Options)
	)
	return o, nil
}

//初始化 kafka 以及 mqtt通道, 用于发送到后端的业务处理服务器
func NewChannnel() *NsqMqttChannel {
	return &NsqMqttChannel{
		KafkaChan: make(chan *models.Message, 10),
		MTChan:    make(chan *models.Message, 10),
	}
}

var ProviderSet = wire.NewSet(NewOptions, NewChannnel)
