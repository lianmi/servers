package nsqclient

import (
	// "encoding/hex"
	"encoding/json"
	"strings"

	"fmt"
	"os"
	"os/signal"
	"syscall"

	"time"

	"github.com/gomodule/redigo/redis"
	"github.com/google/wire"
	// "github.com/pkg/errors"
	"github.com/spf13/viper"
	"go.uber.org/zap"

	"github.com/lianmi/servers/internal/app/channel"
	"github.com/lianmi/servers/internal/pkg/models"

	"github.com/nsqio/go-nsq"
)

type NsqOptions struct {
	Broker       string //127.0.0.1:4161
	ProducerAddr string //127.0.0.1:4150
	Topics       string //以逗号隔开: Auth.Frontend Users.Frontend ... etc.
	ChnanelName  string //channel名称
}

type nsqHandler struct {
	nsqConsumer      *nsq.Consumer
	messagesReceived int
	nsqMqttChannel   *channel.NsqMqttChannel

	logger *zap.Logger
}

type nsqProducer struct {
	*nsq.Producer
}

type NsqClient struct {
	o              *NsqOptions
	app            string
	topics         []string
	nsqMqttChannel *channel.NsqMqttChannel

	Producer *nsqProducer

	logger    *zap.Logger
	redisPool *redis.Pool
}

func NewNsqOptions(v *viper.Viper) (*NsqOptions, error) {
	var (
		err error
		o   = new(NsqOptions)
	)

	if err = v.UnmarshalKey("nsq", o); err != nil {
		return nil, err
	}

	return o, err
}

//初始化消费者
func initConsumer(topic, channelName, addr string, nqChannel *channel.NsqMqttChannel, logger *zap.Logger) error {
	cfg := nsq.NewConfig()

	//设置轮询时间间隔，最小10ms， 最大 5m， 默认60s
	cfg.LookupdPollInterval = 3 * time.Second

	c, err := nsq.NewConsumer(topic, channelName, cfg)
	if err != nil {
		return err
	}
	c.SetLoggerLevel(nsq.LogLevelWarning) // 设置警告级别

	handler := &nsqHandler{
		nsqConsumer:    c,
		nsqMqttChannel: nqChannel,
		logger:         logger,
	}
	c.AddHandler(handler)

	err = c.ConnectToNSQLookupd(addr)
	if err != nil {
		return err
	}
	return nil
}

//处理后端服务发来的消息,JSON格式
func (nh *nsqHandler) HandleMessage(msg *nsq.Message) error {
	nh.messagesReceived++
	nh.logger.Debug(fmt.Sprintf("receive ID: %s, addr: %s", msg.ID, msg.NSQDAddress))

	var backendMessage models.Message

	//反序化
	if err := json.Unmarshal(msg.Body, &backendMessage); err == nil {

		businessType := backendMessage.GetBusinessType()
		businessSubType := backendMessage.GetBusinessSubType()
		taskId := backendMessage.GetTaskID()
		businessTypeName := backendMessage.GetBusinessTypeName()
		code := backendMessage.GetCode()

		//根据目标target,  组装数据包， 向mqtt的channel写入
		nh.logger.Info("Receive message Ffom backend service",
			zap.Uint32("taskId:", taskId),
			zap.String("BusinessTypeName:", businessTypeName), //业务
			zap.Uint32("businessType:", businessType),         // 业务类型
			zap.Uint32("businessSubType:", businessSubType),   // 业务子类型
			zap.Int32("code:", code),                          // 状态码
			zap.String("Source:", backendMessage.GetSource()), //发送者
			zap.String("Target:", backendMessage.GetTarget()), //接收者
		)

		//向MTChan通道写入数据, 实现向客户端发送数据
		nh.nsqMqttChannel.MTChan <- &backendMessage
	}
	return nil
}

//初始化生产者
func initProducer(addr string) (*nsqProducer, error) {
	// fmt.Println("init producer address:", addr)
	producer, err := nsq.NewProducer(addr, nsq.NewConfig())
	if err != nil {
		return nil, err
	}
	return &nsqProducer{producer}, nil
}

func NewNsqClient(o *NsqOptions, redisPool *redis.Pool, channel *channel.NsqMqttChannel, logger *zap.Logger) *NsqClient {
	logger.Info("Topics", zap.String("Topics", o.Topics))
	topics := strings.Split(o.Topics, ",")
	for _, topic := range topics {
		channelName := fmt.Sprintf("%s.%s", topic, o.ChnanelName)
		logger.Info("channelName", zap.String("channelName", channelName))
		err := initConsumer(topic, channelName, o.Broker, channel, logger)
		if err != nil {
			logger.Error("dispatcher, InitConsumer Error ", zap.Error(err), zap.String("topic", topic))
			return nil
		}
	}

	logger.Info("启动Nsq消费者 ==> Subscribe Topics 成功", zap.String("Topics", o.Topics))

	p, err := initProducer(o.ProducerAddr)
	if err != nil {
		logger.Error("init Producer error:", zap.Error(err))
		return nil
	}

	//
	for _, topic := range topics {

		//目的是创建topic
		if err := p.Publish(topic, []byte("a")); err != nil {
			logger.Error("创建topic错误", zap.String("topic", topic), zap.Error(err))
		} else {
			logger.Info("创建topic成功", zap.String("topic", topic))
		}

	}

	logger.Info("启动Nsq生产者成功")

	return &NsqClient{
		o:              o,
		topics:         topics,
		nsqMqttChannel: channel,
		// consumer:       c,
		Producer:  p,
		logger:    logger.With(zap.String("type", "nsqclient")),
		redisPool: redisPool,
	}
}

func (nc *NsqClient) Application(name string) {
	nc.app = name
}

//启动Nsq实例
func (nc *NsqClient) Start() error {
	//尝试读取redis
	redisConn := nc.redisPool.Get()
	defer redisConn.Close()

	if bar, err := redis.String(redisConn.Do("GET", "bar")); err == nil {
		nc.logger.Info("redisConn GET ", zap.String("bar", bar))
	}

	//Go程
	go nc.ProcessProduceChan()

	go func() {
		run := true

		sigchan := make(chan os.Signal, 1)
		signal.Notify(sigchan, syscall.SIGINT, syscall.SIGTERM)

		for run == true {
			select {
			case sig := <-sigchan:
				nc.logger.Info("Caught signal terminating")
				_ = sig
				run = false

			}
		}

		nc.logger.Info("Closing nsqclient")
		// nc.consumer.Close()
	}()

	return nil
}

// 处理生产者数据，这些数据来源于mqtt的订阅消费，向后端场景服务程序发布
func (nc *NsqClient) ProcessProduceChan() {
	run := true
	sigchan := make(chan os.Signal, 1)
	signal.Notify(sigchan, syscall.SIGINT, syscall.SIGTERM)

	for run == true {
		select {
		case sig := <-sigchan:
			nc.logger.Info("Caught signal terminating")
			_ = sig
			run = false
		case msg := <-nc.nsqMqttChannel.NsqChan: //从NsqChan通道里读取数据
			//Target字段存储的是业务模块名称: "auth.Backend", "users.Backend" ...
			topic := msg.GetTarget()
			msgID := msg.GetID()
			nc.logger.Debug(fmt.Sprintf("从NsqChan通道里读取数据, 并且向后端 %s 发送", topic), zap.String("msgID", msgID))

			rawData, err := json.Marshal(msg)
			if err != nil {
				nc.logger.Error("json.Marshal error", zap.Error(err))
				continue
			}

			if nc.Producer == nil {
				nc.logger.Error("严重错误: nc.Producer == nil")
				continue
			}

			err = nc.Producer.Public(topic, rawData)
			if err != nil {
				nc.logger.Error("nc.Producer.Public error", zap.Error(err))
				continue
			}
		}
	}
}

//发布消息
func (np *nsqProducer) Public(topic string, data []byte) error {
	err := np.Publish(topic, data)
	if err != nil {
		return err
	}
	return nil
}

func (nc *NsqClient) Stop() error {
	nc.Producer.Stop()
	// for _, consumer := range consumers {
	// 	consumer.Stop()
	// }	
	return nil
}

var ProviderSet = wire.NewSet(NewNsqOptions, NewNsqClient)
