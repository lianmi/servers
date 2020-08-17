package kafkaBackend

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"time"

	"github.com/golang/protobuf/proto"
	"github.com/gomodule/redigo/redis"
	"github.com/google/wire"
	"github.com/jinzhu/gorm"
	"github.com/pkg/errors"

	// uuid "github.com/satori/go.uuid"
	"github.com/spf13/viper"
	"go.uber.org/zap"

	Auth "github.com/lianmi/servers/api/proto/auth"
	User "github.com/lianmi/servers/api/proto/user"
	"github.com/lianmi/servers/internal/pkg/models"

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
	kClient.handleFuncMap[UnionUint16ToUint32(2, 2)] = kClient.HandleSignOut  //登出处理程序
	kClient.handleFuncMap[UnionUint16ToUint32(2, 4)] = kClient.HandleKick     //Kick处理程序
	kClient.handleFuncMap[UnionUint16ToUint32(1, 1)] = kClient.HandleGetUsers //根据用户标示返回用户信息

	return kClient
}

func (kc *KafkaClient) Application(name string) {
	kc.app = name
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

	//尝试读取redis
	// redisConn := kc.redisPool.Get()
	// defer redisConn.Close()
	// vkey := fmt.Sprintf("verificationCode:%s", email)

	// if bar, err := redis.String(redisConn.Do("GET", "bar")); err == nil {
	// 	kc.logger.Info("redisConn GET ", zap.String("bar", bar))
	// }

	//Go程，处理dispatcher发来的业务数据
	go kc.ProcessRecvPayload()

	go func() {
		run := true

		sigchan := make(chan os.Signal, 1)
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
			taskId := msg.GetTaskID()
			businessType := uint16(msg.GetBusinessType())
			businessSubType := uint16(msg.GetBusinessSubType())
			businessTypeName := msg.GetBusinessTypeName()

			//根据目标target,  组装数据包， 写入processChan
			kc.logger.Info("kfaPayload",
				// zap.String("Topic:", payload.Topic),
				zap.Uint32("taskId:", taskId),                     //taskId
				zap.String("BusinessTypeName:", businessTypeName), //业务名称
				zap.Uint16("businessType:", businessType),         // 业务类型
				zap.Uint16("businessSubType:", businessSubType),   // 业务子类型
				zap.String("Source:", msg.GetSource()),            // 业务数据发送者, 这里是businessTypeName
				zap.String("Target:", msg.GetTarget()),            // 接收者, 这里是自己，authService
			)

			//根据businessType以及businessSubType进行处理, func
			// var ok bool
			if handleFunc, ok := kc.handleFuncMap[UnionUint16ToUint32(businessType, businessSubType)]; !ok {
				kc.logger.Warn("Can not process this businessType", zap.Uint16("businessType:", businessType), zap.Uint16("businessSubType:", businessSubType))
				continue
			} else {
				if err := handleFunc(msg); err != nil {

					msg.SetCode(500) //异常出错
					msg.SetErrorMsg([]byte("Internal Server Error"))

					//处理完成，向dispatcher发送
					topic := msg.GetSource() + ".Frontend"
					kc.Produce(topic, msg)
				}
			}

		}
	}
}

/*
登出
1. 主设备登出，需要删除从设备一切数据，踢出从设备
2. 从设备登出，只删除自己的数据，并刷新此用户的在线设备列表
*/
func (kc *KafkaClient) HandleSignOut(msg *models.Message) error {
	kc.logger.Info("HandleSignOut star...", zap.String("DeviceId", msg.GetDeviceID()))

	//TODO 将此设备从在线列表里删除，然后更新对应用户的在线列表。
	redisConn := kc.redisPool.Get()
	defer redisConn.Close()

	username := msg.GetUserName()
	// token := msg.GetJwtToken()
	deviceID := msg.GetDeviceID()
	var err error

	//取出当前旧的设备的os， clientType， logonAt
	curDeviceHashKey := fmt.Sprintf("devices:%s:%s", username, deviceID)
	isMaster, _ := redis.Bool(redisConn.Do("HGET", curDeviceHashKey))
	curOs, _ := redis.String(redisConn.Do("HGET", curDeviceHashKey, "os"))
	curClientType, _ := redis.Int(redisConn.Do("HGET", curDeviceHashKey, "clientType"))
	curLogonAt, _ := redis.Uint64(redisConn.Do("HGET", curDeviceHashKey, "logonAt"))

	kc.logger.Debug("SignOut", zap.Bool("isMaster", isMaster),
		zap.String("username", username),
		zap.String("deviceID", deviceID),
		zap.String("curOs", curOs),
		zap.Int("curClientType", curClientType),
		zap.Uint64("curLogonAt", curLogonAt))

	deviceListKey := fmt.Sprintf("devices:%s", username)

	if isMaster { //如果是主设备
		//查询出所有主从设备
		deviceIDSlice, _ := redis.Strings(redisConn.Do("ZRANGEBYSCORE", deviceListKey, "-inf", "+inf"))
		for index, eDeviceID := range deviceIDSlice {
			kc.logger.Debug("查询出所有主从设备", zap.Int("index", index), zap.String("eDeviceID", eDeviceID))
			deviceKey := fmt.Sprintf("userDeviceID:%s", eDeviceID)
			jwtToken, _ := redis.String(redisConn.Do("GET", deviceKey))
			kc.logger.Debug("Redis GET ", zap.String("deviceKey", deviceKey), zap.String("jwtToken", jwtToken))

			businessType := 2
			businessSubType := 5 //KickedEvent

			businessTypeName := "Auth"
			kafkaTopic := businessTypeName + ".Frontend"
			backendService := businessTypeName + "Service"

			//向当前主设备及从设备发出踢下线
			kickMsg := &models.Message{}
			now := time.Now().UnixNano() / 1e6
			kickMsg.UpdateID()
			//构建消息路由, 第一个参数是要处理的业务类型，后端服务器处理完成后，需要用此来拼接topic: {businessTypeName.Frontend}
			kickMsg.BuildRouter(businessTypeName, "", kafkaTopic)

			kickMsg.SetJwtToken(jwtToken)
			kickMsg.SetUserName(username)
			kickMsg.SetDeviceID(string(eDeviceID))
			// kickMsg.SetTaskID(uint32(taskId))
			kickMsg.SetBusinessTypeName(businessTypeName)
			kickMsg.SetBusinessType(uint32(businessType))
			kickMsg.SetBusinessSubType(uint32(businessSubType))

			kickMsg.BuildHeader(backendService, now)

			//构造负载数据
			resp := &Auth.KickEventRsp{
				ClientType: 0,
				Reason:     Auth.KickReason_SamePlatformKick,
				TimeTag:    uint64(time.Now().UnixNano() / 1e6),
			}
			data, _ := proto.Marshal(resp)
			kickMsg.FillBody(data) //网络包的body，承载真正的业务数据

			kickMsg.SetCode(200) //成功的状态码

			//构建数据完成，向dispatcher发送
			topic := "Auth.Frontend"
			if err := kc.Produce(topic, kickMsg); err == nil {
				kc.logger.Info("message succeed send to ProduceChannel", zap.String("topic", topic))
			} else {
				kc.logger.Error(" failed to send message to ProduceChannel", zap.Error(err))
			}

			_, err = redisConn.Do("DEL", deviceKey) //删除deviceKey

			deviceHashKey := fmt.Sprintf("devices:%s:%s", username, eDeviceID)
			_, err = redisConn.Do("DEL", deviceHashKey) //删除deviceHashKey

		}

		//删除所有与之相关的key
		_, err = redisConn.Do("DEL", deviceListKey) //删除deviceListKey

	} else { //如果是从设备

		//删除token
		deviceKey := fmt.Sprintf("userDeviceID:%s", deviceID)
		_, err = redisConn.Do("DEL", deviceKey)

		//删除有序集合里的元素
		//移除单个元素 ZREM deviceListKey {设备id}
		_, err = redisConn.Do("ZREM", deviceListKey, deviceID)

		//删除哈希
		deviceHashKey := fmt.Sprintf("devices:%s:%s", username, deviceID)
		_, err = redisConn.Do("DEL", deviceHashKey)

		//多端登录状态变化事件
		//向其它端发送此从设备离线的事件
		deviceIDSliceNew, _ := redis.Strings(redisConn.Do("ZRANGEBYSCORE", deviceListKey, "-inf", "+inf"))
		//查询出当前在线所有主从设备
		for _, eDeviceID := range deviceIDSliceNew {
			targetMsg := &models.Message{}
			curDeviceKey := fmt.Sprintf("userDeviceID:%s", eDeviceID)
			curJwtToken, _ := redis.String(redisConn.Do("GET", curDeviceKey))
			kc.logger.Debug("Redis GET ", zap.String("curDeviceKey", curDeviceKey), zap.String("curJwtToken", curJwtToken))

			now := time.Now().UnixNano() / 1e6
			targetMsg.UpdateID()
			//构建消息路由, 第一个参数是要处理的业务类型，后端服务器处理完成后，需要用此来拼接topic: {businessTypeName.Frontend}
			targetMsg.BuildRouter("Auth", "", "Auth.Frontend")

			targetMsg.SetJwtToken(curJwtToken)
			targetMsg.SetUserName(username)
			targetMsg.SetDeviceID(eDeviceID)
			// kickMsg.SetTaskID(uint32(taskId))
			targetMsg.SetBusinessTypeName("Auth")
			targetMsg.SetBusinessType(uint32(2))
			targetMsg.SetBusinessSubType(uint32(3)) //MultiLoginEvent = 3

			targetMsg.BuildHeader("AuthService", now)

			//构造负载数据
			clients := make([]*Auth.DeviceInfo, 0)
			deviceInfo := &Auth.DeviceInfo{
				Username:     username,
				ConnectionId: "",
				DeviceId:     deviceID,
				DeviceIndex:  0,
				IsMaster:     false,
				Os:           curOs,
				ClientType:   Auth.ClientType(curClientType),
				LogonAt:      curLogonAt,
			}

			clients = append(clients, deviceInfo)

			resp := &Auth.MultiLoginEventRsp{
				State:   false,
				Clients: clients,
			}

			data, _ := proto.Marshal(resp)
			targetMsg.FillBody(data) //网络包的body，承载真正的业务数据

			targetMsg.SetCode(200) //成功的状态码
			//构建数据完成，向dispatcher发送
			topic := "Auth.Frontend"
			if err := kc.Produce(topic, targetMsg); err == nil {
				kc.logger.Info("message succeed send to ProduceChannel", zap.String("topic", topic))
			} else {
				kc.logger.Error(" failed to send message to ProduceChannel", zap.Error(err))
			}
		}

	}

	_ = err

	kc.logger.Debug("登出成功")
	return nil
}

/*
踢出其它终端
1. 主设备才能踢出从设备
2. 从设备被踢后，只删除自己的数据，并发出多端登录状态变化事件
*/
func (kc *KafkaClient) HandleKick(msg *models.Message) error {
	var err error
	kc.logger.Info("HandleKick start...", zap.String("DeviceId", msg.GetDeviceID()))

	//TODO 将此设备从在线列表里删除，然后更新对应用户的在线列表。
	redisConn := kc.redisPool.Get()
	defer redisConn.Close()

	username := msg.GetUserName()
	// token := msg.GetJwtToken()
	deviceID := msg.GetDeviceID()

	//取出当前设备的os， clientType， logonAt
	curDeviceHashKey := fmt.Sprintf("devices:%s:%s", username, deviceID)
	isMaster, _ := redis.Bool(redisConn.Do("HGET", curDeviceHashKey))
	curOs, _ := redis.String(redisConn.Do("HGET", curDeviceHashKey, "os"))
	curClientType, _ := redis.Int(redisConn.Do("HGET", curDeviceHashKey, "clientType"))
	curLogonAt, _ := redis.Uint64(redisConn.Do("HGET", curDeviceHashKey, "logonAt"))

	kc.logger.Debug("Kick",
		zap.Bool("isMaster", isMaster),
		zap.String("username", username),
		zap.String("deviceID", deviceID),
		zap.String("curOs", curOs),
		zap.Int("curClientType", curClientType),
		zap.Uint64("curLogonAt", curLogonAt))

	deviceListKey := fmt.Sprintf("devices:%s", username)

	var deviceiIDs []string
	if isMaster { //如果是主设备

		//打开msg里的负载， 获取即将被踢的设备列表
		body := msg.GetContent()
		//解包body
		var kickReq Auth.KickReq
		if err := proto.Unmarshal(body, &kickReq); err != nil {
			kc.logger.Error("Protobuf Unmarshal Error", zap.Error(err))
			return err
		}
		deviceiIDs = kickReq.GetDeviceIds()

		for _, did := range deviceiIDs {
			kc.logger.Debug("To be kick ...", zap.String("DeviceId", did))
			//删除token
			deviceKey := fmt.Sprintf("userDeviceID:%s", did)
			_, err = redisConn.Do("DEL", deviceKey)

			//删除有序集合里的元素
			//移除单个元素 ZREM deviceListKey {设备id}
			_, err = redisConn.Do("ZREM", deviceListKey, did)

			//删除哈希
			deviceHashKey := fmt.Sprintf("devices:%s:%s", username, did)
			_, err = redisConn.Do("DEL", deviceHashKey)

			//多端登录状态变化事件
			//向其它端发送此从设备离线的事件
			deviceIDSliceNew, _ := redis.Strings(redisConn.Do("ZRANGEBYSCORE", deviceListKey, "-inf", "+inf"))
			//查询出当前在线所有主从设备
			for _, eDeviceID := range deviceIDSliceNew {
				targetMsg := &models.Message{}
				curDeviceKey := fmt.Sprintf("userDeviceID:%s", eDeviceID)
				curJwtToken, _ := redis.String(redisConn.Do("GET", curDeviceKey))
				kc.logger.Debug("Redis GET ", zap.String("curDeviceKey", curDeviceKey), zap.String("curJwtToken", curJwtToken))

				now := time.Now().UnixNano() / 1e6
				targetMsg.UpdateID()
				//构建消息路由, 第一个参数是要处理的业务类型，后端服务器处理完成后，需要用此来拼接topic: {businessTypeName.Frontend}
				targetMsg.BuildRouter("Auth", "", "Auth.Frontend")

				targetMsg.SetJwtToken(curJwtToken)
				targetMsg.SetUserName(username)
				targetMsg.SetDeviceID(eDeviceID)
				// kickMsg.SetTaskID(uint32(taskId))
				targetMsg.SetBusinessTypeName("Auth")
				targetMsg.SetBusinessType(uint32(2))
				targetMsg.SetBusinessSubType(uint32(3)) //MultiLoginEvent = 3

				targetMsg.BuildHeader("AuthService", now)

				//构造负载数据
				clients := make([]*Auth.DeviceInfo, 0)
				deviceInfo := &Auth.DeviceInfo{
					Username:     username,
					ConnectionId: "",
					DeviceId:     did,
					DeviceIndex:  0,
					IsMaster:     false,
					Os:           curOs,
					ClientType:   Auth.ClientType(curClientType),
					LogonAt:      curLogonAt,
				}

				clients = append(clients, deviceInfo)

				resp := &Auth.MultiLoginEventRsp{
					State:   false,
					Clients: clients,
				}

				data, _ := proto.Marshal(resp)
				targetMsg.FillBody(data) //网络包的body，承载真正的业务数据

				targetMsg.SetCode(200) //成功的状态码
				//构建数据完成，向dispatcher发送
				topic := "Auth.Frontend"
				if err := kc.Produce(topic, targetMsg); err == nil {
					kc.logger.Info("message succeed send to ProduceChannel", zap.String("topic", topic))
				} else {
					kc.logger.Error(" failed to send message to ProduceChannel", zap.Error(err))
				}
			}

		}

	} else {
		//从设备无权踢出其它设备
	}

	//响应客户端的OnKick

	kickRsp := &Auth.KickRsp{
		DeviceIds: deviceiIDs,
	}
	msg.SetCode(200) //登录成功的状态码

	data, _ := proto.Marshal(kickRsp)
	rspHex := strings.ToUpper(hex.EncodeToString(data))

	kc.logger.Info("Kick Succeed",
		zap.String("Username:", username),
		zap.Int("length", len(data)),
		zap.String("rspHex", rspHex))

	msg.FillBody(data) //网络包的body，承载真正的业务数据

	//处理完成，向dispatcher发送
	topic := msg.GetSource() + ".Frontend"
	if err := kc.Produce(topic, msg); err == nil {
		kc.logger.Info("KickRsp message succeed send to ProduceChannel", zap.String("topic", topic))
	} else {
		kc.logger.Error("Failed to send KickRsp message to ProduceChannel", zap.Error(err))
	}
	_ = err
	return nil
}

func (kc *KafkaClient) HandleGetUsers(msg *models.Message) error {
	var err error
	kc.logger.Info("HandleGetUsers start...", zap.String("DeviceId", msg.GetDeviceID()))

	//TODO 将此设备从在线列表里删除，然后更新对应用户的在线列表。
	redisConn := kc.redisPool.Get()
	defer redisConn.Close()

	username := msg.GetUserName()
	// token := msg.GetJwtToken()
	deviceID := msg.GetDeviceID()

	//取出当前设备的os， clientType， logonAt
	curDeviceHashKey := fmt.Sprintf("devices:%s:%s", username, deviceID)
	isMaster, _ := redis.Bool(redisConn.Do("HGET", curDeviceHashKey))
	curOs, _ := redis.String(redisConn.Do("HGET", curDeviceHashKey, "os"))
	curClientType, _ := redis.Int(redisConn.Do("HGET", curDeviceHashKey, "clientType"))
	curLogonAt, _ := redis.Uint64(redisConn.Do("HGET", curDeviceHashKey, "logonAt"))

	kc.logger.Debug("GetUsers",
		zap.Bool("isMaster", isMaster),
		zap.String("username", username),
		zap.String("deviceID", deviceID),
		zap.String("curOs", curOs),
		zap.Int("curClientType", curClientType),
		zap.Uint64("curLogonAt", curLogonAt))

	// deviceListKey := fmt.Sprintf("devices:%s", username)

	//打开msg里的负载， 获取即将被踢的设备列表
	body := msg.GetContent()
	//解包body
	var getUsersReq User.GetUsersReq
	if err := proto.Unmarshal(body, &getUsersReq); err != nil {
		kc.logger.Error("Protobuf Unmarshal Error", zap.Error(err))
		return err
	}
	getUsersResp := &User.GetUsersResp{}
	getUsersResp.Users = make([]*User.User, 0)

	for _, username := range getUsersReq.GetUsernames() {
		p := new(models.User)
		if err = kc.db.Model(p).Where("username = ?", username).First(p).Error; err != nil {
			return errors.Wrapf(err, "Get user error[username=%s]", username)
		}
		user := &User.User{
			Username:     p.Username,
			Gender:       p.Gender,
			Nick:         p.Nick,
			Avatar:       p.Avatar,
			Label:        p.Label,
			Introductory: p.Introductory,
			Province:     p.Province,
			City:         p.City,
			County:       p.County,
			Street:       p.Street,
			Address:      p.Address,
			Branchesname: p.Branchesname,
			LegalPerson:  p.LegalPerson,
		}
		getUsersResp.Users = append(getUsersResp.Users, user)
	}
	msg.SetCode(200) //登录成功的状态码

	data, _ := proto.Marshal(getUsersResp)
	rspHex := strings.ToUpper(hex.EncodeToString(data))

	kc.logger.Info("GetUsers Succeed",
		zap.String("Username:", username),
		zap.Int("length", len(data)),
		zap.String("rspHex", rspHex))

	msg.FillBody(data) //网络包的body，承载真正的业务数据

	//处理完成，向dispatcher发送
	topic := msg.GetSource() + ".Frontend"
	if err := kc.Produce(topic, msg); err == nil {
		kc.logger.Info("GetUsersResp message succeed send to ProduceChannel", zap.String("topic", topic))
	} else {
		kc.logger.Error("Failed to send GetUsersResp message to ProduceChannel", zap.Error(err))
	}
	_ = err
	return nil

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

func Uint16ToBytes(n uint16) []byte {
	return []byte{
		byte(n),
		byte(n >> 8),
	}
}

func BytesToUint32(buf []byte) uint32 {
	b_buf := bytes.NewBuffer(buf)
	var x uint32
	binary.Read(b_buf, binary.BigEndian, &x)
	return x
}

//将两个uint16的数字合并为一个uint32
func UnionUint16ToUint32(a uint16, b uint16) uint32 {
	a_buf := Uint16ToBytes(a)
	b_buf := Uint16ToBytes(b)

	a_buf = append(a_buf, b_buf[:]...)
	return BytesToUint32(a_buf)
}

var ProviderSet = wire.NewSet(NewKafkaOptions, NewKafkaClient)
