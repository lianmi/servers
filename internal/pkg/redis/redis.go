package redis

import (
	"fmt"
	"time"
	"github.com/google/wire"
	"github.com/gomodule/redigo/redis"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	// "github.com/lianmi/servers/internal/pkg/models"
)

// Options is  configuration of redis
type Options struct {
	Addr   string `yaml:"addr"` //127.0.0.1:6379
	Password   string `yaml:"password"` 
	Db   int `yaml:"db"`
	Debug bool
}

func NewRedisOptions(v *viper.Viper, logger *zap.Logger) (*Options, error) {
	var err error
	o := new(Options)
	//读取dispatcher.yaml配置文件里的redis设置
	if err = v.UnmarshalKey("redis", o); err != nil {
		return nil, errors.Wrap(err, "unmarshal db option error")
	}
	address := fmt.Sprintf("%s", o.Addr)
	logger.Info("load redis options success", zap.String("Address", address))

	return o, err
}

// Init 初始化redis连接池， 默认db=0
func New(o *Options) (*redis.Pool, error) {

	return &redis.Pool{
		MaxIdle:     2,
		IdleTimeout: 240 * time.Second,
		Dial: func() (redis.Conn, error) {
			c, err := redis.Dial("tcp", o.Addr)
			if err != nil {
				return nil, err
			}
			if o.Password != "" {
				if _, err := c.Do("AUTH", o.Password); err != nil {
					c.Close()
					return nil, err
				}
			}
			if o.Db < 0 || o.Db > 15 {
				c.Close()
				return nil, err
			}

			if _, err := c.Do("SELECT", o.Db); err != nil {
				c.Close()
				return nil, err
			}
			return c, nil
		},
		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			if time.Since(t) < time.Minute {
				return nil
			}
			_, err := c.Do("PING")
			return err
		},
	}, nil
}


var ProviderSet = wire.NewSet(New, NewRedisOptions)
