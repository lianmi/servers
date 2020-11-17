package multichannel

import (
	"github.com/google/wire"
	"github.com/lianmi/servers/internal/pkg/models"
	"github.com/spf13/viper"
)

type NsqChannel struct {
	NsqChan chan *models.Message
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

//初始化 nsq 通道, 用于发送设备多端登录的下发
func NewChannnel() *NsqChannel {
	return &NsqChannel{
		NsqChan: make(chan *models.Message, 10),
	}
}

var ProviderSet = wire.NewSet(NewOptions, NewChannnel)
