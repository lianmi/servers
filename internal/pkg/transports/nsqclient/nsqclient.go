package nsqclient

import (
	// "encoding/hex"
	"encoding/json"
	"strings"

	"os"
	"os/signal"
	"syscall"

	// "time"

	"github.com/gomodule/redigo/redis"
	"github.com/google/wire"
	"github.com/pkg/errors"
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
	Channel      string //channel
}

type nsqHandler struct {
	nsqConsumer      *nsq.Consumer
	messagesReceived int
	nsqMqttChannel *channel.NsqMqttChannel

	logger    *zap.Logger
}

type nsqProducer struct {
	*nsq.Producer
}

type NsqClient struct {
	o        *NsqOptions
	app      string
	topics   []string
	nsqMqttChannel *channel.NsqMqttChannel

	// consumer *nsq.Consumer
	producer *nsq.Producer

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
		nsqConsumer: c
		nsqMqttChannel: nqChannel
		logger: logger
	}
	c.AddHandler(handler)

	err = c.ConnectToNSQLookupd(addr)
	if err != nil {
		return err
	}
	return nil
}


//处理消息
func (nh *nsqHandler) HandleMessage(msg *nsq.Message) error {
	nh.messagesReceived++
	nh.logger.Debug(fmt.Sprintf("receive ID: %s, addr: %s", msg.ID, msg.NSQDAddress))


	var backendMessage models.Message

	if err := json.Unmarshal(msg.Body, &backendMessage); err == nil {

		businessType := backendMessage.GetBusinessType()
		businessSubType := backendMessage.GetBusinessSubType()
		taskId := backendMessage.GetTaskID()
		businessTypeName := backendMessage.GetBusinessTypeName()
		code := backendMessage.GetCode()

		//根据目标target,  组装数据包， 向mqtt的channel写入
		nh.logger.Info("Receive message Ffom backend service",
			zap.String("Topic:", topic),
			// zap.String("Account:", backendMessage.GetAccount()),
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
	topicArray := strings.Split(o.Topics, ",")
	topics := make([]string, 0)
	for _, topic := range topicArray {
		topics = append(topics, topic)
		err := initConsumer(topic, nc.o.Channel, nc.o.Broker, channel, logger)
		if err != nil {
			logger.Error("InitConsumer Error ", zap.Error(err))
			return nil
		}
	}

	logger.Info("启动Nsq消费者 ==> Subscribe Topics 成功", zap.String("Topics", nc.o.Topics))

	p, err := initProducer(o.ProducerAddr)
	if err != nil {
		logger.Error("init Producer error:", zap.Error(err))
		return nil
	}
	logger.Info("启动Nsq生产者成功")

	return &NsqClient{
		o:              o,
		topics:         topics,
		nsqMqttChannel: channel,
		// consumer:       c,
		producer:  p,
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

	defer nc.producer.Stop()

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
		nc.consumer.Close()
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

			if nc.producer == nil {
				nc.logger.Error("严重错误: nc.producer == nil")
				continue
			}

			nc.producer.public(topic, rawData)
		}
	}
}

//发布消息
func (np *nsqProducer) public(topic string, data []byte) error {
	err := np.Publish(topic, data)
	if err != nil {
		log.Println("nsq public error:", err)
		return err
	}
	return nil
}

var ProviderSet = wire.NewSet(NewNsqOptions, NewNsqClient)