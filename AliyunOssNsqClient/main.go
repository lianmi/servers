/*
本程序是Nsq aliyunoss client端, 用于接收阿里云的copyObject信息并实现
*/
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"log"
	"time"
	// "github.com/lianmi/servers/internal/pkg/log"

	// "github.com/lianmi/servers/util/randtool"
	// "go.uber.org/zap"

	"github.com/lianmi/servers/internal/pkg/models"

	"github.com/nsqio/go-nsq"
)

type NsqOptions struct {
	Broker       string //127.0.0.1:4161
	ProducerAddr string //127.0.0.1:4150
	Topics       string //以逗号隔开
	ChnanelName  string //channel
}

type nsqHandler struct {
	nsqConsumer      *nsq.Consumer
	messagesReceived int
	// logger           *zap.Logger
}

type nsqProducer struct {
	*nsq.Producer
}

type NsqClient struct {
	o                 *NsqOptions
	app               string
	recvFromFrontChan chan *models.Message //接收到payload

	topics    []string
	Producer  *nsqProducer    // 生产者
	consumers []*nsq.Consumer // 消费者

	// logger *zap.Logger
	//定义key=cmdid的处理func，当收到消息后，从此map里查出对应的处理方法
	handleFuncMap map[uint32]func(payload *models.Message) error
}

var (
	msgFromDispatcherChan = make(chan *models.Message, 10)
)

//初始化消费者
func initConsumer(topic, channelName, addr string) (*nsq.Consumer, error) {
	cfg := nsq.NewConfig()

	//设置轮询时间间隔，最小10ms， 最大 5m， 默认60s
	cfg.LookupdPollInterval = 3 * time.Second

	c, err := nsq.NewConsumer(topic, channelName, cfg)
	if err != nil {
		return nil, err
	}
	c.SetLoggerLevel(nsq.LogLevelWarning) // 设置警告级别

	handler := &nsqHandler{
		nsqConsumer: c,
		// logger:      logger,
	}
	c.AddHandler(handler)

	err = c.ConnectToNSQLookupd(addr)
	if err != nil {
		return nil, err
	}
	return c, nil
}

//处理消息
func (nh *nsqHandler) HandleMessage(msg *nsq.Message) error {
	nh.messagesReceived++
	// nh.logger.Debug(fmt.Sprintf("receive ID: %s, addr: %s", msg.ID, msg.NSQDAddress))

	var kfaPayload models.Message
	if string(msg.Body) == "a" {
		// 创建topic
	} else {
		if err := json.Unmarshal(msg.Body, &kfaPayload); err == nil {

			msgFromDispatcherChan <- &kfaPayload //将来自dispatcher的数据压入本地通道

		} else {
			// nh.logger.Error("HandleMessage, json.Unmarshal Error", zap.Error(err))
			log.Printf("HandleMessage, json.Unmarshal Error:%v\n", err)
			return err
		}
	}

	return nil
}

//初始化生产者
func initProducer(addr string) (*nsqProducer, error) {
	producer, err := nsq.NewProducer(addr, nsq.NewConfig())
	if err != nil {
		return nil, err
	}
	return &nsqProducer{producer}, nil
}

func NewNsqClient(o *NsqOptions) *NsqClient {

	p, err := initProducer(o.ProducerAddr)
	if err != nil {
		log.Printf("init Producer error: %v\n", err)
		return nil
	}

	log.Println("启动Nsq生产者成功")

	nsqClient := &NsqClient{
		o:             o,
		Producer:      p,
		handleFuncMap: make(map[uint32]func(payload *models.Message) error),
	}

	//注册每个业务子类型的处理方法
	// nsqClient.handleFuncMap[randtool.UnionUint16ToUint32(5, 1)] = nsqClient.HandleRecvMsg       //5-1 收到消息的处理程序
	// nsqClient.handleFuncMap[randtool.UnionUint16ToUint32(5, 4)] = nsqClient.HandleMsgAck        //5-4 确认消息送达
	// nsqClient.handleFuncMap[randtool.UnionUint16ToUint32(5, 6)] = nsqClient.HandleSendCancelMsg //5-9 发送撤销消息
	// nsqClient.handleFuncMap[randtool.UnionUint16ToUint32(5, 12)] = nsqClient.HandleGetOssToken  //5-12 获取阿里云OSS上传Token

	return nsqClient
}

//启动Nsq实例
func (nc *NsqClient) Start() error {
	log.Println("Dispatcher NsqClient Start()")
	log.Println("Topics: ", nc.o.Topics)
	nc.topics = strings.Split(nc.o.Topics, ",")
	for _, topic := range nc.topics {
		channelName := fmt.Sprintf("%s.%s", topic, nc.o.ChnanelName)
		log.Println("channelName", channelName)
		consumer, err := initConsumer(topic, channelName, nc.o.Broker)
		if err != nil {
			log.Printf("dispatcher, InitConsumer Error: %v\n", err)
			return nil
		}
		nc.consumers = append(nc.consumers, consumer)
	}

	log.Println("启动Nsq消费者 ==> Subscribe Topics 成功")

	for _, topic := range nc.topics {

		//目的是创建topic
		if err := nc.Producer.Publish(topic, []byte("a")); err != nil {
			log.Printf("创建topic错误: %v\n", err)
		} else {
			log.Printf("创建topicv: %s\n", topic)
		}

	}

	go func() {
		run := true

		sigchan := make(chan os.Signal, 1)
		signal.Notify(sigchan, syscall.SIGINT, syscall.SIGTERM)

		for run == true {
			select {
			case sig := <-sigchan:
				log.Println("Caught signal terminating")
				_ = sig
				run = false

			}
		}

		log.Println("Closing dispatcher nsqclient")
	}()

	return nil
}

func main() {
	o := &NsqOptions{
		Broker:       "127.0.0.1:4161",
		ProducerAddr: "127.0.0.1:4150",
		Topics:       "AliyunOss",
		ChnanelName:  "im",
	}

	nsqClient := NewNsqClient(o)
	if nsqClient != nil {
		log.Println("nsqClient ok")
	} else {
		return
	}

	sigchan := make(chan os.Signal, 1)
	<-sigchan

}
