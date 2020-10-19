package nsqBackend

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
	"github.com/jinzhu/gorm"
	"github.com/spf13/viper"
	"go.uber.org/zap"

	"github.com/golang/protobuf/proto"
	Global "github.com/lianmi/servers/api/proto/global"
	Msg "github.com/lianmi/servers/api/proto/msg"

	"github.com/lianmi/servers/internal/pkg/models"
	"github.com/lianmi/servers/util/randtool"

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
	o   *NsqOptions
	app string

	topics    []string
	Producer  *nsqProducer    // 生产者
	consumers []*nsq.Consumer // 消费者

	logger        *zap.Logger
	db            *gorm.DB
	redisPool     *redis.Pool
	handleFuncMap map[uint32]func(payload *models.Message) error //定义key=cmdid的处理func，当收到消息后，从此map里查出对应的处理方法
}

var (
	msgFromDispatcherChan = make(chan *models.Message, 10)
)

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

func NewNsqClient(o *NsqOptions, db *gorm.DB, redisPool *redis.Pool, logger *zap.Logger) *NsqClient {

	p, err := initProducer(o.ProducerAddr)
	if err != nil {
		logger.Error("init Producer error:", zap.Error(err))
		return nil
	}

	logger.Info("启动Nsq生产者成功", zap.String("addr", o.ProducerAddr))

	nsqClient := &NsqClient{
		o:             o,
		Producer:      p,
		logger:        logger.With(zap.String("type", "AuthService")),
		db:            db,
		redisPool:     redisPool,
		handleFuncMap: make(map[uint32]func(payload *models.Message) error),
	}

	//注册每个业务子类型的处理方法
	nsqClient.handleFuncMap[randtool.UnionUint16ToUint32(1, 1)] = nsqClient.HandleGetUsers          //1-1 获取用户资料
	nsqClient.handleFuncMap[randtool.UnionUint16ToUint32(1, 2)] = nsqClient.HandleUpdateUserProfile //1-2 修改用户资料
	nsqClient.handleFuncMap[randtool.UnionUint16ToUint32(1, 5)] = nsqClient.HandleMarkTag           //1-5 打标签

	nsqClient.handleFuncMap[randtool.UnionUint16ToUint32(2, 2)] = nsqClient.HandleSignOut        //登出处理程序
	nsqClient.handleFuncMap[randtool.UnionUint16ToUint32(2, 4)] = nsqClient.HandleKick           //Kick处理程序
	nsqClient.handleFuncMap[randtool.UnionUint16ToUint32(2, 6)] = nsqClient.HandleAddSlaveDevice //Kick处理程序
	nsqClient.handleFuncMap[randtool.UnionUint16ToUint32(2, 7)] = nsqClient.HandleAuthorizeCode  //2-7 从设备申请授权码
	nsqClient.handleFuncMap[randtool.UnionUint16ToUint32(2, 10)] = nsqClient.HandleGetAllDevices //向服务端查询所有主从设备列表

	nsqClient.handleFuncMap[randtool.UnionUint16ToUint32(6, 1)] = nsqClient.HandleSync //6-1 发起同步请求

	nsqClient.handleFuncMap[randtool.UnionUint16ToUint32(3, 1)] = nsqClient.HandleFriendRequest       //3-1 好友请求发起与处理
	nsqClient.handleFuncMap[randtool.UnionUint16ToUint32(3, 5)] = nsqClient.HandleDeleteFriend        //3-5 好友请求发起与处理
	nsqClient.handleFuncMap[randtool.UnionUint16ToUint32(3, 6)] = nsqClient.HandleUpdateFriend        //3-6 刷新好友资料
	nsqClient.handleFuncMap[randtool.UnionUint16ToUint32(3, 8)] = nsqClient.HandleGetFriends          //3-8 增量同步好友列表
	nsqClient.handleFuncMap[randtool.UnionUint16ToUint32(3, 9)] = nsqClient.HandleWatchRequest        //3-9 关注商户
	nsqClient.handleFuncMap[randtool.UnionUint16ToUint32(3, 10)] = nsqClient.HandleCancelWatchRequest //3-11 取消关注商户

	nsqClient.handleFuncMap[randtool.UnionUint16ToUint32(4, 1)] = nsqClient.HandleCreateTeam          //4-1 创建群组
	nsqClient.handleFuncMap[randtool.UnionUint16ToUint32(4, 2)] = nsqClient.HandleGetTeamMembers      //4-2 获取群组成员
	nsqClient.handleFuncMap[randtool.UnionUint16ToUint32(4, 3)] = nsqClient.HandleGetTeam             //4-3 查询群信息
	nsqClient.handleFuncMap[randtool.UnionUint16ToUint32(4, 4)] = nsqClient.HandleInviteTeamMembers   //4-4 邀请用户加群
	nsqClient.handleFuncMap[randtool.UnionUint16ToUint32(4, 5)] = nsqClient.HandleRemoveTeamMembers   //4-5 删除群组成员
	nsqClient.handleFuncMap[randtool.UnionUint16ToUint32(4, 6)] = nsqClient.HandleAcceptTeamInvite    //4-6 接受群邀请
	nsqClient.handleFuncMap[randtool.UnionUint16ToUint32(4, 7)] = nsqClient.HandleRejectTeamInvitee   //4-7 拒绝群邀请
	nsqClient.handleFuncMap[randtool.UnionUint16ToUint32(4, 8)] = nsqClient.HandleApplyTeam           //4-8 主动申请加群
	nsqClient.handleFuncMap[randtool.UnionUint16ToUint32(4, 9)] = nsqClient.HandlePassTeamApply       //4-9 批准加群申请
	nsqClient.handleFuncMap[randtool.UnionUint16ToUint32(4, 10)] = nsqClient.HandleRejectTeamApply    //4-10 否决加群申请
	nsqClient.handleFuncMap[randtool.UnionUint16ToUint32(4, 11)] = nsqClient.HandleUpdateTeam         //4-11 更新群组信息
	nsqClient.handleFuncMap[randtool.UnionUint16ToUint32(4, 13)] = nsqClient.HandleLeaveTeam          //4-13 退群
	nsqClient.handleFuncMap[randtool.UnionUint16ToUint32(4, 14)] = nsqClient.HandleAddTeamManagers    //4-14 设置群管理员
	nsqClient.handleFuncMap[randtool.UnionUint16ToUint32(4, 15)] = nsqClient.HandleRemoveTeamManagers //4-15 撤销群管理员
	nsqClient.handleFuncMap[randtool.UnionUint16ToUint32(4, 18)] = nsqClient.HandleMuteTeam           //4-18 设置群禁言模式
	nsqClient.handleFuncMap[randtool.UnionUint16ToUint32(4, 19)] = nsqClient.HandleMuteTeamMember     //4-19 设置群成员禁言
	nsqClient.handleFuncMap[randtool.UnionUint16ToUint32(4, 20)] = nsqClient.HandleSetNotifyType      //4-20 用户设置群消息通知方式
	nsqClient.handleFuncMap[randtool.UnionUint16ToUint32(4, 21)] = nsqClient.HandleUpdateMyInfo       //4-21 用户设置其在群里的资料
	nsqClient.handleFuncMap[randtool.UnionUint16ToUint32(4, 22)] = nsqClient.HandleUpdateMemberInfo   //4-22 管理员设置群成员资料
	nsqClient.handleFuncMap[randtool.UnionUint16ToUint32(4, 24)] = nsqClient.HandlePullTeamMembers    //4-24 获取指定群组成员
	nsqClient.handleFuncMap[randtool.UnionUint16ToUint32(4, 25)] = nsqClient.HandleGetMyTeams         //4-25 增量同步群组信息
	nsqClient.handleFuncMap[randtool.UnionUint16ToUint32(4, 26)] = nsqClient.HandleCheckTeamInvite    //4-26 管理员审核用户入群申请
	nsqClient.handleFuncMap[randtool.UnionUint16ToUint32(4, 27)] = nsqClient.HandleGetTeamMembersPage //4-27 分页获取群成员信息

	return nsqClient
}

func (nc *NsqClient) Application(name string) {
	nc.app = name
}

/*
判断redis是否存在键值，如果没，则创建
*/
func (nc *NsqClient) RedisInit() {
	var err error

	nc.logger.Info("RedisInit start...")
	redisConn := nc.redisPool.Get()
	defer redisConn.Close()

	isExists, _ := redis.Bool(redisConn.Do("EXISTS", "Teams"))
	if !isExists {
		teamIDs := nc.GetTeams()
		for _, teamID := range teamIDs {
			err = redisConn.Send("ZADD", "Teams", time.Now().UnixNano()/1e6, teamID)
		}
		redisConn.Flush()
		nc.logger.Info("ZADD succeed", zap.Int("length of teamIDs ", len(teamIDs)))
	}

	_ = err

}

//启动Nsq实例
func (nc *NsqClient) Start() error {
	nc.logger.Info("AuthService NsqClient Start()")
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
		consumer, err := initConsumer(topic, channelName, nc.o.Broker, nc.logger)
		if err != nil {
			nc.logger.Error("Init Consumer Error", zap.Error(err), zap.String("channelName", channelName))
			return nil
		}
		nc.consumers = append(nc.consumers, consumer)
	}
	// brokerIP := netutil.GetLocalIP4() //本容器ip

	nc.logger.Info("启动Nsq消费者 ==> Subscribe Topics 成功", zap.Strings("topics", nc.topics), zap.String("Broker", nc.o.Broker))

	//redis初始化
	go nc.RedisInit()

	//Go程，启动定时任务
	go nc.RunCron()

	//Go程，处理dispatcher发来的业务数据
	go nc.ProcessRecvPayload()

	var (
		sigchan chan os.Signal
		run     bool = true
	)

	go func() {

		sigchan = make(chan os.Signal, 1)
		signal.Notify(sigchan, syscall.SIGINT, syscall.SIGTERM)
		tiker := time.NewTicker(time.Second)
		for run == true {
			select {
			case sig := <-sigchan:
				nc.logger.Info("Caught signal terminating")
				_ = sig
				run = false
			case <-tiker.C:

				// for index, consumer := range consumers {
				// 	stats := consumer.Stats()
				// 	if stats.Connections > 0 {

				// 		nc.logger.Info("tiker.C: consumer.Stats",
				// 			zap.String("Topic", topics[index]),
				// 			zap.Int("Connections", stats.Connections),
				// 			zap.Uint64("MessagesReceived", stats.MessagesReceived), // 已接收总数
				// 			zap.Uint64("MessagesFinished", stats.MessagesFinished), // 已完成总数
				// 			zap.Uint64("MessagesRequeued", stats.MessagesRequeued), // 排队总数
				// 		)
				// 	}
				// }

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
			taskID := msg.GetTaskID()
			businessType := uint16(msg.GetBusinessType())
			businessSubType := uint16(msg.GetBusinessSubType())
			businessTypeName := msg.GetBusinessTypeName()

			//根据目标target,  组装数据包， 写入processChan
			nc.logger.Info("msgFromDispatcherChan",
				// zap.String("Topic:", payload.Topic),
				zap.Uint32("taskId:", taskID),                     // SDK的任务ID
				zap.String("BusinessTypeName:", businessTypeName), // 业务名称
				zap.Uint16("businessType:", businessType),         // 业务类型
				zap.Uint16("businessSubType:", businessSubType),   // 业务子类型
				zap.String("Source:", msg.GetSource()),            // 业务数据发送者, 这里是businessTypeName
				zap.String("Target:", msg.GetTarget()),            // 接收者, 这里是自己，authService
			)

			//根据businessType以及businessSubType进行处理, func
			if handleFunc, ok := nc.handleFuncMap[randtool.UnionUint16ToUint32(businessType, businessSubType)]; !ok {
				nc.logger.Warn("Can not process this businessType", zap.Uint16("businessType:", businessType), zap.Uint16("businessSubType:", businessSubType))
				//向SDK回包，业务号及业务子号无法匹配 404
				msg.SetCode(int32(404))                                                          //状态码
				msg.SetErrorMsg([]byte("Can not process this businessType and businessSubType")) //错误提示
				msg.FillBody(nil)

				rawData, _ := json.Marshal(msg)

				topic := msg.GetSource() + ".Frontend"

				//向dispatcher发送
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

func (nc *NsqClient) PrintRedisErr(err error) {
	nc.logger.Error("Redis Error", zap.Error(err))
}

var ProviderSet = wire.NewSet(NewNsqOptions, NewNsqClient)
