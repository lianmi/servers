package nsqMq

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
	"github.com/jinzhu/gorm"
	"github.com/spf13/viper"
	"go.uber.org/zap"

	"github.com/lianmi/servers/internal/app/channel"
	"github.com/lianmi/servers/internal/pkg/models"

	"github.com/nsqio/go-nsq"

	"github.com/golang/protobuf/proto"
	Global "github.com/lianmi/servers/api/proto/global"
	Msg "github.com/lianmi/servers/api/proto/msg"
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

	Producer  *nsqProducer    // 生产者
	consumers []*nsq.Consumer // 消费者

	logger    *zap.Logger
	redisPool *redis.Pool
	db        *gorm.DB
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
func initConsumer(topic, channelName, addr string, nqChannel *channel.NsqMqttChannel, logger *zap.Logger) (*nsq.Consumer, error) {
	cfg := nsq.NewConfig()

	//设置轮询时间间隔，最小10ms， 最大 5m， 默认60s
	cfg.LookupdPollInterval = 3 * time.Second

	c, err := nsq.NewConsumer(topic, channelName, cfg)
	if err != nil {
		return nil, err
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
		return nil, err
	}
	return c, nil
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
		nh.logger.Info("Receive message from backend service",
			zap.Uint32("taskId:", taskId),
			zap.String("BusinessTypeName:", businessTypeName), //业务
			zap.Uint32("businessType:", businessType),         // 业务类型
			zap.Uint32("businessSubType:", businessSubType),   // 业务子类型
			zap.Int32("code:", code),                          // 状态码
			zap.String("Source:", backendMessage.GetSource()), //发送者
			zap.String("Target:", backendMessage.GetTarget()), //接收者
		)

		//向MTChan通道写入数据, 从而实现向mqtt客户端发送数据
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

// func NewNsqClient(o *NsqOptions, redisPool *redis.Pool, channel *channel.NsqMqttChannel, logger *zap.Logger) *NsqClient {
func NewNsqClient(o *NsqOptions, db *gorm.DB, redisPool *redis.Pool, channel *channel.NsqMqttChannel, logger *zap.Logger) *NsqClient {

	p, err := initProducer(o.ProducerAddr)
	if err != nil {
		logger.Error("init Producer error:", zap.Error(err), zap.String("ProducerAddr", o.ProducerAddr))
		return nil
	}

	logger.Info("启动Nsq生产者成功")

	return &NsqClient{
		o:              o,
		nsqMqttChannel: channel,
		Producer:       p,
		consumers:      make([]*nsq.Consumer, 0),
		logger:         logger.With(zap.String("type", "nsqclient")),
		redisPool:      redisPool,
		db:             db,
	}

}

func (nc *NsqClient) Application(name string) {
	nc.app = name
}

//启动Nsq实例
func (nc *NsqClient) Start() error {
	nc.logger.Info("Topics", zap.String("Topics", nc.o.Topics))
	nc.topics = strings.Split(nc.o.Topics, ",")
	for _, topic := range nc.topics {
		channelName := fmt.Sprintf("%s.%s", topic, nc.o.ChnanelName)
		nc.logger.Info("channelName", zap.String("channelName", channelName))
		consumer, err := initConsumer(topic, channelName, nc.o.Broker, nc.nsqMqttChannel, nc.logger)
		if err != nil {
			nc.logger.Error("dispatcher, InitConsumer Error ", zap.Error(err), zap.String("topic", topic))
			return nil
		}
		nc.consumers = append(nc.consumers, consumer)
	}

	nc.logger.Info("启动Nsq消费者 ==> Subscribe Topics 成功", zap.Strings("Topics", nc.topics))

	for _, topic := range nc.topics {

		//目的是创建topic
		if err := nc.Producer.Publish(topic, []byte("a")); err != nil {
			nc.logger.Error("创建topic错误", zap.String("topic", topic), zap.Error(err))
		} else {
			nc.logger.Info("创建topic成功", zap.String("topic", topic))
		}

	}

	//尝试读取redis
	redisConn := nc.redisPool.Get()
	defer redisConn.Close()

	if bar, err := redis.String(redisConn.Do("GET", "bar")); err == nil {
		nc.logger.Info("redisConn GET", zap.String("bar", bar))
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

		nc.logger.Info("Closing dispatcher nsqclient")
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

/*
向目标用户账号的所有端推送系统通知
业务号： BusinessType_Msg(5)
业务子号： MsgSubType_RecvMsgEvent(2)
系统通知，Scene的值是 S2C,其它的场景不需要处理
*/
func (nc *NsqClient) BroadcastSystemMsgToAllDevices(rsp *Msg.RecvMsgEventRsp, toUser string, exceptDeviceIDs ...string) error {

	data, _ := proto.Marshal(rsp)

	redisConn := nc.redisPool.Get()
	defer redisConn.Close()

	//删除7天前的缓存系统消息
	nTime := time.Now()
	yesTime := nTime.AddDate(0, 0, -7).Unix()
	offLineMsgListKey := fmt.Sprintf("offLineMsgList:%s", toUser)

	_, err := redisConn.Do("ZREMRANGEBYSCORE", offLineMsgListKey, "-inf", yesTime)

	//Redis里缓存此系统消息,目的是6-1同步接口里的 systemmsgAt, 然后同步给用户
	systemMsgAt := time.Now().UnixNano() / 1e6
	if _, err := redisConn.Do("ZADD", offLineMsgListKey, systemMsgAt, rsp.GetServerMsgId()); err != nil {
		nc.logger.Error("ZADD Error", zap.Error(err))
	}

	//系统消息具体内容
	systemMsgKey := fmt.Sprintf("systemMsg:%s:%s", toUser, rsp.GetServerMsgId())

	_, err = redisConn.Do("HMSET",
		systemMsgKey,
		"Username", toUser,
		"SystemMsgAt", systemMsgAt,
		"Seq", rsp.Seq,
		"Data", data,
	)

	_, err = redisConn.Do("EXPIRE", systemMsgKey, 7*24*3600) //设置有效期为7天

	//向toUser所有端发送
	deviceListKey := fmt.Sprintf("devices:%s", toUser)
	deviceIDSliceNew, _ := redis.Strings(redisConn.Do("ZRANGEBYSCORE", deviceListKey, "-inf", "+inf"))
	//查询出当前在线所有主从设备
	for _, eDeviceID := range deviceIDSliceNew {
		if inArray(eDeviceID, exceptDeviceIDs) == eDeviceID {
			continue
		}
		targetMsg := &models.Message{}
		curDeviceKey := fmt.Sprintf("DeviceJwtToken:%s", eDeviceID)
		curJwtToken, _ := redis.String(redisConn.Do("GET", curDeviceKey))
		nc.logger.Debug("Redis GET ", zap.String("curDeviceKey", curDeviceKey), zap.String("curJwtToken", curJwtToken))

		targetMsg.UpdateID()
		//构建消息路由, 第一个参数是要处理的业务类型，后端服务器处理完成后，需要用此来拼接topic: {businessTypeName.Frontend}
		targetMsg.BuildRouter("Msg", "", "Msg.Frontend")

		targetMsg.SetJwtToken(curJwtToken)
		targetMsg.SetUserName(toUser)
		targetMsg.SetDeviceID(eDeviceID)
		// kickMsg.SetTaskID(uint32(taskId))
		targetMsg.SetBusinessTypeName("Msg")
		targetMsg.SetBusinessType(uint32(Global.BusinessType_Msg))           //消息模块
		targetMsg.SetBusinessSubType(uint32(Global.MsgSubType_RecvMsgEvent)) //接收消息事件

		targetMsg.BuildHeader("ChatService", time.Now().UnixNano()/1e6)

		targetMsg.FillBody(data) //网络包的body，承载真正的业务数据

		targetMsg.SetCode(200) //成功的状态码

		//构建数据完成，向dispatcher发送
		topic := "Msg.Frontend"
		rawData, _ := json.Marshal(targetMsg)
		if err := nc.Producer.Public(topic, rawData); err == nil {
			nc.logger.Info("message succeed send to ProduceChannel", zap.String("topic", topic))
		} else {
			nc.logger.Error("Failed to send message to ProduceChannel", zap.Error(err))
		}

		nc.logger.Info("Broadcast Msg To AllDevices Succeed",
			zap.String("Username:", toUser),
			zap.String("DeviceID:", curDeviceKey),
			zap.Int64("Now", time.Now().UnixNano()/1e6))

		_ = err

	}

	return nil
}

func (nc *NsqClient) Stop() error {
	nc.Producer.Stop()
	for _, consumer := range nc.consumers {
		consumer.Stop()
	}
	return nil
}

//判断in是否在设备列表里，如果在，则返回in，如果不在，则返回 空
func inArray(in string, exceptDeviceIDs []string) string {
	if len(exceptDeviceIDs) > 0 {
		for _, exceptDeviceID := range exceptDeviceIDs {
			if in == exceptDeviceID {
				return in
			}
		}
	}
	return ""
}

var ProviderSet = wire.NewSet(NewNsqOptions, NewNsqClient)
