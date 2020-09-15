package kafkaBackend

import (
	"encoding/json"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/gomodule/redigo/redis"
	"github.com/google/wire"
	"github.com/jinzhu/gorm"
	"github.com/pkg/errors"

	// uuid "github.com/satori/go.uuid"
	"github.com/spf13/viper"
	"go.uber.org/zap"

	"github.com/lianmi/servers/internal/pkg/models"
	"github.com/lianmi/servers/util/randtool"

	"gopkg.in/confluentinc/confluent-kafka-go.v1/kafka"
)

type KafkaOptions struct {
	Broker string
	Group  string
	Topics string //以逗号隔开: Auth.Frontend Users.Frontend ... etc.
}

type KafkaClient struct {
	o                 *KafkaOptions
	app               string
	topics            []string
	recvFromFrontChan chan *models.Message //接收到payload
	consumer          *kafka.Consumer
	producer          *kafka.Producer
	logger            *zap.Logger
	db                *gorm.DB
	redisPool         *redis.Pool
	//定义key=cmdid的处理func，当收到消息后，从此map里查出对应的处理方法
	handleFuncMap map[uint32]func(payload *models.Message) error
}

func NewKafkaOptions(v *viper.Viper) (*KafkaOptions, error) {
	var (
		err error
		o   = new(KafkaOptions)
	)

	if err = v.UnmarshalKey("kafka", o); err != nil {
		return nil, err
	}

	return o, err
}

func NewKafkaClient(o *KafkaOptions, db *gorm.DB, redisPool *redis.Pool, logger *zap.Logger) *KafkaClient {
	topicArray := strings.Split(o.Topics, ",")
	topics := make([]string, 0)
	for _, topic := range topicArray {
		topics = append(topics, topic)
		logger.Debug("NewKafkaClient增加topic", zap.String("topic", topic))
	}

	c, err := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers":               o.Broker,
		"group.id":                        o.Group,
		"session.timeout.ms":              6000,
		"go.events.channel.enable":        true,
		"go.application.rebalance.enable": true,
		// Enable generation of PartitionEOF when the
		// end of a partition is reached.
		"enable.partition.eof": true,
		"auto.offset.reset":    "earliest"})

	if err != nil {
		logger.Error("Failed to create consumer, retry... ", zap.Error(err))
		return nil
	}
	p, err := kafka.NewProducer(&kafka.ConfigMap{"bootstrap.servers": o.Broker})

	if err != nil {
		logger.Error("Failed to create producer: ", zap.Error(err))
		return nil
	}
	kClient := &KafkaClient{
		o:                 o,
		topics:            topics,
		recvFromFrontChan: make(chan *models.Message, 10),
		consumer:          c,
		producer:          p,
		logger:            logger.With(zap.String("type", "kafka.Client")),
		db:                db,
		redisPool:         redisPool,
		handleFuncMap:     make(map[uint32]func(payload *models.Message) error),
	}

	//注册每个业务子类型的处理方法
	kClient.handleFuncMap[randtool.UnionUint16ToUint32(1, 1)] = kClient.HandleGetUsers          //1-1 获取用户资料
	kClient.handleFuncMap[randtool.UnionUint16ToUint32(1, 2)] = kClient.HandleUpdateUserProfile //1-2 修改用户资料
	kClient.handleFuncMap[randtool.UnionUint16ToUint32(1, 5)] = kClient.HandleMarkTag           //1-5 打标签

	kClient.handleFuncMap[randtool.UnionUint16ToUint32(2, 2)] = kClient.HandleSignOut        //登出处理程序
	kClient.handleFuncMap[randtool.UnionUint16ToUint32(2, 4)] = kClient.HandleKick           //Kick处理程序
	kClient.handleFuncMap[randtool.UnionUint16ToUint32(2, 6)] = kClient.HandleAddSlaveDevice //Kick处理程序
	kClient.handleFuncMap[randtool.UnionUint16ToUint32(2, 7)] = kClient.HandleAuthorizeCode  //2-7 从设备申请授权码
	kClient.handleFuncMap[randtool.UnionUint16ToUint32(2, 10)] = kClient.HandleGetAllDevices //向服务端查询所有主从设备列表

	kClient.handleFuncMap[randtool.UnionUint16ToUint32(6, 1)] = kClient.HandleSync //6-1 发起同步请求

	kClient.handleFuncMap[randtool.UnionUint16ToUint32(3, 1)] = kClient.HandleFriendRequest       //3-1 好友请求发起与处理
	kClient.handleFuncMap[randtool.UnionUint16ToUint32(3, 5)] = kClient.HandleDeleteFriend        //3-5 好友请求发起与处理
	kClient.handleFuncMap[randtool.UnionUint16ToUint32(3, 6)] = kClient.HandleUpdateFriend        //3-6 刷新好友资料
	kClient.handleFuncMap[randtool.UnionUint16ToUint32(3, 8)] = kClient.HandleGetFriends          //3-8 增量同步好友列表
	kClient.handleFuncMap[randtool.UnionUint16ToUint32(3, 9)] = kClient.HandleWatchRequest        //3-9 关注商户
	kClient.handleFuncMap[randtool.UnionUint16ToUint32(3, 10)] = kClient.HandleCancelWatchRequest //3-11 取消关注商户

	kClient.handleFuncMap[randtool.UnionUint16ToUint32(4, 1)] = kClient.HandleCreateTeam          //4-1 创建群组
	kClient.handleFuncMap[randtool.UnionUint16ToUint32(4, 2)] = kClient.HandleGetTeamMembers      //4-2 获取群组成员
	kClient.handleFuncMap[randtool.UnionUint16ToUint32(4, 3)] = kClient.HandleGetTeam             //4-3 查询群信息
	kClient.handleFuncMap[randtool.UnionUint16ToUint32(4, 4)] = kClient.HandleInviteTeamMembers   //4-4 邀请用户加群
	kClient.handleFuncMap[randtool.UnionUint16ToUint32(4, 5)] = kClient.HandleRemoveTeamMembers   //4-5 删除群组成员
	kClient.handleFuncMap[randtool.UnionUint16ToUint32(4, 6)] = kClient.HandleAcceptTeamInvite    //4-6 接受群邀请
	kClient.handleFuncMap[randtool.UnionUint16ToUint32(4, 7)] = kClient.HandleRejectTeamInvitee   //4-7 拒绝群邀请
	kClient.handleFuncMap[randtool.UnionUint16ToUint32(4, 8)] = kClient.HandleApplyTeam           //4-8 主动申请加群
	kClient.handleFuncMap[randtool.UnionUint16ToUint32(4, 9)] = kClient.HandlePassTeamApply       //4-9 批准加群申请
	kClient.handleFuncMap[randtool.UnionUint16ToUint32(4, 10)] = kClient.HandleRejectTeamApply    //4-10 否决加群申请
	kClient.handleFuncMap[randtool.UnionUint16ToUint32(4, 11)] = kClient.HandleUpdateTeam         //4-11 更新群组信息
	kClient.handleFuncMap[randtool.UnionUint16ToUint32(4, 13)] = kClient.HandleLeaveTeam          //4-13 退群
	kClient.handleFuncMap[randtool.UnionUint16ToUint32(4, 14)] = kClient.HandleAddTeamManagers    //4-14 设置群管理员
	kClient.handleFuncMap[randtool.UnionUint16ToUint32(4, 15)] = kClient.HandleRemoveTeamManagers //4-15 撤销群管理员
	kClient.handleFuncMap[randtool.UnionUint16ToUint32(4, 18)] = kClient.HandleMuteTeam           //4-18 设置群禁言模式
	kClient.handleFuncMap[randtool.UnionUint16ToUint32(4, 19)] = kClient.HandleMuteTeamMember     //4-19 设置群成员禁言
	kClient.handleFuncMap[randtool.UnionUint16ToUint32(4, 20)] = kClient.HandleSetNotifyType      //4-20 用户设置群消息通知方式
	kClient.handleFuncMap[randtool.UnionUint16ToUint32(4, 21)] = kClient.HandleUpdateMyInfo       //4-21 用户设置其在群里的资料
	kClient.handleFuncMap[randtool.UnionUint16ToUint32(4, 22)] = kClient.HandleUpdateMemberInfo   //4-22 管理员设置群成员资料
	kClient.handleFuncMap[randtool.UnionUint16ToUint32(4, 24)] = kClient.HandlePullTeamMembers    //4-24 获取指定群组成员
	kClient.handleFuncMap[randtool.UnionUint16ToUint32(4, 25)] = kClient.HandleGetMyTeams         //4-25 增量同步群组信息
	kClient.handleFuncMap[randtool.UnionUint16ToUint32(4, 26)] = kClient.HandleCheckTeamInvite    //4-26 管理员审核用户入群申请
	kClient.handleFuncMap[randtool.UnionUint16ToUint32(4, 27)] = kClient.HandleGetTeamMembersPage //4-27 分页获取群成员信息

	return kClient
}

func (kc *KafkaClient) Application(name string) {
	kc.app = name
}

/*
判断redis是否存在键值，如果没，则创建
*/
func (kc *KafkaClient) RedisInit() {
	var err error

	kc.logger.Info("RedisInit start...")
	redisConn := kc.redisPool.Get()
	defer redisConn.Close()

	isExists, _ := redis.Bool(redisConn.Do("EXISTS", "Teams"))
	if !isExists {
		teamIDs := kc.GetTeams()
		for _, teamID := range teamIDs {
			err = redisConn.Send("ZADD", "Teams", time.Now().UnixNano()/1e6, teamID)
		}
		redisConn.Flush()
		kc.logger.Info("ZADD succeed", zap.Int("length of teamIDs ", len(teamIDs)))
	}

	_ = err

}

//启动Kafka实例
func (kc *KafkaClient) Start() error {
	kc.logger.Info("==> Subscribe Topics ", zap.Strings("Topics", kc.topics))

	err := kc.consumer.SubscribeTopics(kc.topics, nil)
	// err := kc.consumer.SubscribeTopics([]string{"auth.Backend"}, nil)
	if err != nil {
		kc.logger.Error("Failed to SubscribeTopics: ", zap.Error(err))
		return err
	}

	//redis初始化
	go kc.RedisInit()

	//Go程，启动定时任务
	go kc.RunCron()

	//Go程，处理dispatcher发来的业务数据
	go kc.ProcessRecvPayload()

	var (
		sigchan chan os.Signal
		run     bool = true
	)

	go func() {

		sigchan = make(chan os.Signal, 1)
		signal.Notify(sigchan, syscall.SIGINT, syscall.SIGTERM)

		for run == true {
			select {
			case sig := <-sigchan:
				kc.logger.Info("Caught signal terminating")
				_ = sig
				run = false

			case ev := <-kc.consumer.Events():
				switch e := ev.(type) {
				case kafka.AssignedPartitions:
					// kc.logger.Info("AssignedPartitions: ", zap.String("e:", e))
					kc.consumer.Assign(e.Partitions)
				case kafka.RevokedPartitions:
					// kc.logger.Info("RevokedPartitions: ", zap.String("e:", e))
					kc.consumer.Unassign()
				case *kafka.Message:
					kc.logger.Info("Message on: ", zap.String("TopicPartition:", e.TopicPartition.String()))

					kfaPayload := new(models.Message)

					if err := json.Unmarshal(e.Value, kfaPayload); err == nil {
						kc.recvFromFrontChan <- kfaPayload //将来自dispatcher的数据压入本地通道

					} else {
						kc.logger.Error("json.Unmarshal Error", zap.Error(err))
						continue
					}

				case kafka.PartitionEOF:
					kc.logger.Info("kafka.PartitionEOF")

				case kafka.Error:
					// Errors should generally be considered as informational, the client will try to automatically recover
					kc.logger.Info("kafka.Error")
				}
			}
		}

		kc.logger.Info("Closing consumer")
		kc.consumer.Close()
	}()

	return nil
}

// 处理dispatcher发来的业务数据
func (kc *KafkaClient) ProcessRecvPayload() {
	run := true
	sigchan := make(chan os.Signal, 1)
	signal.Notify(sigchan, syscall.SIGINT, syscall.SIGTERM)

	for run == true {
		select {
		case sig := <-sigchan:
			kc.logger.Info("Caught signal terminating")
			_ = sig
			run = false
		case msg := <-kc.recvFromFrontChan: //读取来着dispatcher的数据
			taskID := msg.GetTaskID()
			businessType := uint16(msg.GetBusinessType())
			businessSubType := uint16(msg.GetBusinessSubType())
			businessTypeName := msg.GetBusinessTypeName()

			//根据目标target,  组装数据包， 写入processChan
			kc.logger.Info("kfaPayload",
				// zap.String("Topic:", payload.Topic),
				zap.Uint32("taskId:", taskID),                     //SDK的任务ID
				zap.String("BusinessTypeName:", businessTypeName), //业务名称
				zap.Uint16("businessType:", businessType),         // 业务类型
				zap.Uint16("businessSubType:", businessSubType),   // 业务子类型
				zap.String("Source:", msg.GetSource()),            // 业务数据发送者, 这里是businessTypeName
				zap.String("Target:", msg.GetTarget()),            // 接收者, 这里是自己，authService
			)

			//根据businessType以及businessSubType进行处理, func
			if handleFunc, ok := kc.handleFuncMap[randtool.UnionUint16ToUint32(businessType, businessSubType)]; !ok {
				kc.logger.Warn("Can not process this businessType", zap.Uint16("businessType:", businessType), zap.Uint16("businessSubType:", businessSubType))
				//向SDK回包，业务号及业务子号无法匹配 404
				go kc.SendErrorMsg(msg)

				continue
			} else {
				//启动Go程
				go handleFunc(msg)
			}

		}
	}
}

//SendErrorMsg
func (kc *KafkaClient) SendErrorMsg(msg *models.Message) {
	msg.SetCode(int32(404))                                                          //状态码
	msg.SetErrorMsg([]byte("Can not process this businessType and businessSubType")) //错误提示
	//向dispatcher发送
	topic := msg.GetSource() + ".Frontend"
	msg.FillBody(nil)
	kc.Produce(topic, msg)
}

//Produce
func (kc *KafkaClient) Produce(topic string, msg *models.Message) error {
	if msg == nil {
		return errors.New("msg is nil")
	}

	if kc.producer == nil {
		return errors.New("nil Producer")
	}

	//需要序化后才能传输
	rawData, _ := json.Marshal(msg)

	kc.producer.ProduceChannel() <- &kafka.Message{TopicPartition: kafka.TopicPartition{Topic: &topic, Partition: kafka.PartitionAny}, Value: rawData}

	return nil
}

func (kc *KafkaClient) Stop() error {
	kc.producer.Close()
	kc.consumer.Close()
	return nil
}

var ProviderSet = wire.NewSet(NewKafkaOptions, NewKafkaClient)
