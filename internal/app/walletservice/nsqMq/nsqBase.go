/*
处理 钱包相关的业务
*/
package nsqMq

import (
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"time"

	"github.com/gomodule/redigo/redis"
	"github.com/google/wire"
	"github.com/lianmi/servers/internal/app/walletservice/repositories"
	"github.com/lianmi/servers/internal/pkg/blockchain"
	"github.com/lianmi/servers/internal/pkg/models"
	"github.com/lianmi/servers/util/randtool"
	"github.com/nsqio/go-nsq"
	"github.com/spf13/viper"
	"go.uber.org/zap"
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
	logger           *zap.Logger
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

	logger *zap.Logger

	Repository repositories.WalletRepository

	redisPool  *redis.Pool
	ethService *blockchain.Service //连接以太坊geth的websocket
	//定义key=cmdid的处理func，当收到消息后，从此map里查出对应的处理方法
	handleFuncMap map[uint32]func(payload *models.Message) error
}

var (
	msgFromDispatcherChan = make(chan *models.Message, 10)
)

//初始化消费者
func initConsumer(topic, channelName, addr string, logger *zap.Logger) (*nsq.Consumer, error) {
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
		logger:      logger,
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
	nh.logger.Debug(fmt.Sprintf("receive ID: %s, addr: %s", msg.ID, msg.NSQDAddress))

	var kfaPayload models.Message
	if string(msg.Body) == "a" {
		// 创建topic
	} else {
		if err := json.Unmarshal(msg.Body, &kfaPayload); err == nil {

			msgFromDispatcherChan <- &kfaPayload //将来自dispatcher的数据压入本地通道

		} else {
			nh.logger.Error("HandleMessage, json.Unmarshal Error", zap.Error(err))
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

func NewNsqClient(o *NsqOptions, Repository repositories.WalletRepository, redisPool *redis.Pool, logger *zap.Logger, ethService *blockchain.Service) *NsqClient {

	p, err := initProducer(o.ProducerAddr)
	if err != nil {
		logger.Error("init Producer error:", zap.Error(err))
		return nil
	}

	logger.Info("启动Nsq生产者成功")

	nsqClient := &NsqClient{
		o:             o,
		Producer:      p,
		logger:        logger.With(zap.String("type", "WalletService")),
		Repository:    Repository,
		redisPool:     redisPool,
		ethService:    ethService,
		handleFuncMap: make(map[uint32]func(payload *models.Message) error),
	}

	//注册每个业务子类型的处理方法, BusinessType = 10
	nsqClient.handleFuncMap[randtool.UnionUint16ToUint32(10, 1)] = nsqClient.HandleRegisterWallet            //10-1 钱包账号注册
	nsqClient.handleFuncMap[randtool.UnionUint16ToUint32(10, 2)] = nsqClient.HandleDeposit                   //10-2 充值
	nsqClient.handleFuncMap[randtool.UnionUint16ToUint32(10, 3)] = nsqClient.HandlePreTransfer               //10-3 发起转账
	nsqClient.handleFuncMap[randtool.UnionUint16ToUint32(10, 4)] = nsqClient.HandleConfirmTransfer           //10-4 确认转账
	nsqClient.handleFuncMap[randtool.UnionUint16ToUint32(10, 5)] = nsqClient.HandleBalance                   //10-5 查询账号余额
	nsqClient.handleFuncMap[randtool.UnionUint16ToUint32(10, 6)] = nsqClient.HandlePreWithDraw               //10-6 发起提现预审核
	nsqClient.handleFuncMap[randtool.UnionUint16ToUint32(10, 7)] = nsqClient.HandleWithDraw                  //10-7 确认提现
	nsqClient.handleFuncMap[randtool.UnionUint16ToUint32(10, 9)] = nsqClient.HandleSyncCollectionHistoryPage //10-9 同步收款历史
	nsqClient.handleFuncMap[randtool.UnionUint16ToUint32(10, 10)] = nsqClient.HandleSyncDepositHistoryPage   //10-10 同步充值历史
	nsqClient.handleFuncMap[randtool.UnionUint16ToUint32(10, 11)] = nsqClient.HandleSyncWithdrawHistoryPage  //10-11 同步提现历史
	nsqClient.handleFuncMap[randtool.UnionUint16ToUint32(10, 12)] = nsqClient.HandleSyncTransferHistoryPage  //10-12 同步转账历史
	nsqClient.handleFuncMap[randtool.UnionUint16ToUint32(10, 13)] = nsqClient.HandleUserSignIn               //10-13 签到
	nsqClient.handleFuncMap[randtool.UnionUint16ToUint32(10, 14)] = nsqClient.HandleTxHashInfo               //10-14查询交易哈希详情

	return nsqClient
}

func (nc *NsqClient) Application(name string) {
	nc.app = name
}

//启动Nsq实例
func (nc *NsqClient) Start() error {
	nc.logger.Info("WalletService NsqClient Start()")

	nc.topics = strings.Split(nc.o.Topics, ",")
	for _, topic := range nc.topics {

		//目的是创建topic
		if err := nc.Producer.Publish(topic, []byte("a")); err != nil {
			nc.logger.Error("创建topic错误", zap.String("topic", topic), zap.Error(err))
		} else {
			nc.logger.Info("创建topic成功", zap.String("topic", topic))
		}

	}

	for _, topic := range nc.topics {
		channelName := fmt.Sprintf("%s.%s", topic, nc.o.ChnanelName)
		nc.logger.Info("channelName", zap.String("channelName", channelName))
		consumer, err := initConsumer(topic, channelName, nc.o.Broker, nc.logger)
		if err != nil {
			nc.logger.Error("InitConsumer Error ", zap.Error(err))
			return nil
		}
		nc.consumers = append(nc.consumers, consumer)
	}

	nc.logger.Info("启动Nsq消费者 ==> Subscribe Topics 成功", zap.Strings("topics", nc.topics))

	nc.RedisInit()

	//Go程，处理dispatcher发来的业务数据
	go nc.ProcessRecvPayload()

	// 测试，创建系统HD钱包
	nc.ethService.CreateHDWallet()

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

		nc.logger.Info("Exiting Start()")
	}()

	return nil
}

// 处理dispatcher发来的业务数据
func (nc *NsqClient) ProcessRecvPayload() {
	run := true
	sigchan := make(chan os.Signal, 1)
	signal.Notify(sigchan, syscall.SIGINT, syscall.SIGTERM)

	for run == true {
		select {
		case sig := <-sigchan:
			nc.logger.Info("Caught signal terminating")
			_ = sig
			run = false
		case msg := <-msgFromDispatcherChan: //读取来着dispatcher的数据
			taskId := msg.GetTaskID()
			businessType := uint16(msg.GetBusinessType())
			businessSubType := uint16(msg.GetBusinessSubType())
			businessTypeName := msg.GetBusinessTypeName()

			nc.logger.Info("msgFromDispatcherChan",
				// zap.String("Topic:", payload.Topic),
				zap.Uint32("taskId:", taskId),                     //taskId
				zap.String("BusinessTypeName:", businessTypeName), //业务名称
				zap.Uint16("businessType:", businessType),         // 业务类型
				zap.Uint16("businessSubType:", businessSubType),   // 业务子类型
				zap.String("Source:", msg.GetSource()),            // 业务数据发送者, 这里是businessTypeName
				zap.String("Target:", msg.GetTarget()),            // 接收者
			)

			//根据businessType以及businessSubType进行处理, func
			if handleFunc, ok := nc.handleFuncMap[randtool.UnionUint16ToUint32(businessType, businessSubType)]; !ok {
				nc.logger.Warn("Can not process this businessType", zap.Uint16("businessType:", businessType), zap.Uint16("businessSubType:", businessSubType))
				msg.SetCode(int32(404))                                                          //状态码
				msg.SetErrorMsg([]byte("Can not process this businessType and businessSubType")) //错误提示
				msg.FillBody(nil)

				rawData, _ := json.Marshal(msg)

				//向dispatcher发送
				topic := msg.GetSource() + ".Frontend"
				err := nc.Producer.Public(topic, rawData)
				if err != nil {
					nc.logger.Error("nc.Producer.Public error", zap.Error(err))
				}

				continue
			} else {
				//启动Go程
				go handleFunc(msg)
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
	for _, consumer := range nc.consumers {
		consumer.Stop()
	}
	return nil
}

/*
判断redis是否存在键值，如果没，则创建
*/
func (nc *NsqClient) RedisInit() (err error) {

	nc.logger.Info("RedisInit start...")
	redisConn := nc.redisPool.Get()
	defer redisConn.Close()

	//判断平台HD钱包的派生叶子索引是否存在，如果不存在，则创建key，value: 1
	isExists, _ := redis.Bool(redisConn.Do("EXISTS", "Bip32Index"))
	if !isExists {
		_, err = redisConn.Do("SET", "Bip32Index", 2) //0， 1,  2 的索引号已经用了
	}

	return nil

}

var ProviderSet = wire.NewSet(NewNsqOptions, NewNsqClient)
