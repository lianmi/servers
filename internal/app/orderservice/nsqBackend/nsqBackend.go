/*
处理商品及订单的业务
*/
package nsqBackend

import (
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	// "context"

	"time"

	"github.com/gomodule/redigo/redis"
	"github.com/google/wire"
	"github.com/jinzhu/gorm"

	"github.com/lianmi/servers/util/randtool"
	"github.com/spf13/viper"
	"go.uber.org/zap"

	Wallet "github.com/lianmi/servers/api/proto/wallet"
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

	logger    *zap.Logger
	db        *gorm.DB
	redisPool *redis.Pool
	walletSvc Wallet.LianmiWalletClient
	//定义key=cmdid的处理func，当收到消息后，从此map里查出对应的处理方法
	handleFuncMap map[uint32]func(payload *models.Message) error
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

func NewNsqClient(o *NsqOptions, db *gorm.DB, redisPool *redis.Pool, logger *zap.Logger, walletSvc Wallet.LianmiWalletClient) *NsqClient {

	p, err := initProducer(o.ProducerAddr)
	if err != nil {
		logger.Error("init Producer error:", zap.Error(err))
		return nil
	}
	if walletSvc != nil {
		logger.Info("LianmiWalletClient 连接walletservice微服务成功")
	}

	logger.Info("启动Nsq生产者成功")

	nsqClient := &NsqClient{
		o:             o,
		Producer:      p,
		logger:        logger.With(zap.String("type", "OrderService")),
		db:            db,
		redisPool:     redisPool,
		walletSvc:     walletSvc,
		handleFuncMap: make(map[uint32]func(payload *models.Message) error),
	}
	/*
		ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
		transferResp, err := walletSvc.TransferByOrder(ctx, &Wallet.TransferReq{
			OrderID: "test-23432432-00000-kkkkk",
			PayType: int32(1),
		})
		if err != nil {
			logger.Error(" walletSvc.TransferByOrder Error", zap.Error(err))
		} else {
			logger.Debug("transferResp", zap.Int32("ErrCode", transferResp.ErrCode), zap.String("ErrMsg", transferResp.ErrMsg))
		}
	*/

	//注册每个业务子类型的处理方法
	nsqClient.handleFuncMap[randtool.UnionUint16ToUint32(7, 1)] = nsqClient.HandleQueryProducts  //7-1  查询某个商户的所有商品信息
	nsqClient.handleFuncMap[randtool.UnionUint16ToUint32(7, 2)] = nsqClient.HandleAddProduct     //7-2  查询某个商户的所有商品信息
	nsqClient.handleFuncMap[randtool.UnionUint16ToUint32(7, 3)] = nsqClient.HandleUpdateProduct  //7-3  商品编辑更新
	nsqClient.handleFuncMap[randtool.UnionUint16ToUint32(7, 4)] = nsqClient.HandleSoldoutProduct //7-4 商品下架

	nsqClient.handleFuncMap[randtool.UnionUint16ToUint32(9, 1)] = nsqClient.HandleRegisterPreKeys  //9-1 商户上传订单DH加密公钥
	nsqClient.handleFuncMap[randtool.UnionUint16ToUint32(9, 2)] = nsqClient.HandleGetPreKeyOrderID //9-2 获取网点OPK公钥及订单ID

	//9-3 处理chatService转发过来的订单消息
	nsqClient.handleFuncMap[randtool.UnionUint16ToUint32(5, 1)] = nsqClient.HandleOrderMsg

	nsqClient.handleFuncMap[randtool.UnionUint16ToUint32(9, 5)] = nsqClient.HandleChangeOrderState //9-5 对订单进行状态更改
	nsqClient.handleFuncMap[randtool.UnionUint16ToUint32(9, 6)] = nsqClient.HandleGetPreKeysCount  //9-8 商户获取OPK存量

	return nsqClient
}

func (nc *NsqClient) Application(name string) {
	nc.app = name
}

//启动Nsq实例
func (nc *NsqClient) Start() error {
	nc.logger.Info("OrderService NsqClient Start()")

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

	//Go程，处理dispatcher发来的业务数据
	go nc.ProcessRecvPayload()

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
				zap.String("Target:", msg.GetTarget()),            // 接收者, 这里是自己，authService
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

var ProviderSet = wire.NewSet(NewNsqOptions, NewNsqClient)
