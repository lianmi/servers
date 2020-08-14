package mqtt

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"os"
	"os/signal"
	"strconv"

	"encoding/hex"
	"strings"
	"syscall"
	"time"

	"github.com/gomodule/redigo/redis"
	"github.com/google/wire"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
	"go.uber.org/zap"

	"github.com/eclipse/paho.golang/paho" //支持v5.0

	Global "github.com/lianmi/servers/api/proto/global"
	"github.com/lianmi/servers/internal/app/channel"
	"github.com/lianmi/servers/internal/pkg/models"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/lianmi/servers/internal/common"
)

const (
	MQTT_CLIENT_DISCONNECTED = "Disconneted"
	MQTT_CLIENT_CONNECTED    = "Conneted"
	retryCount               = 1200
	cloudAccessSleep         = 5 * time.Second
)

type MQTTOptions struct {
	Addr          string `yaml:"addr"`          // broker addr, 127.0.0.1:1883
	User          string `yaml:"user"`          // broker auth user
	Passwd        string `yaml:"passwd"`        // broker auth password
	ClientID      string `yaml:"clientid"`      // broker node name
	TopicPrefix   string `yaml:"topicprefix"`   //topic prefix to client
	ResponseTopic string `yaml:"responseTopic"` //topic for response
}

type MQTTClient struct {
	o                   *MQTTOptions
	app                 string
	Addr                string
	User, Passwd        string
	ClientID            string
	CleanSession        bool
	Order               bool
	KeepAliveInterval   time.Duration
	PingTimeout         time.Duration
	MessageChannelDepth uint
	//0: QOSAtMostOnce, 1: QOSAtLeastOnce, 2: QOSExactlyOnce.
	QOS byte
	//if the flag set true, server will store the message and
	//can be delivered to future subscribers.
	Retain bool
	//the state of client.
	State string

	// optinal as below.
	// OnConnect     pahomqtt.OnConnectHandler
	// OnLost        pahomqtt.ConnectionLostHandler
	FileStorePath string
	//Will message, optional
	WillTopic    string
	WillMessage  string
	WillQOS      byte
	WillRetained bool
	// tls config
	TLSConfig        *tls.Config
	client           *paho.Client //支持v5.0
	kafkamqttChannel *channel.KafkaMqttChannel
	logger           *zap.Logger
	redisPool        *redis.Pool
}

func NewMQTTOptions(v *viper.Viper) (*MQTTOptions, error) {
	var (
		err error
		o   = new(MQTTOptions)
	)

	if err = v.UnmarshalKey("mqtt", o); err != nil {
		fmt.Println("UnmarshalKey mqtt error:", err.Error())
		return nil, err
	}
	fmt.Println("addr:", o.Addr)
	fmt.Println("user:", o.User)
	fmt.Println("passwd:", o.Passwd)
	fmt.Println("clientid:", o.ClientID)
	fmt.Println("topicPrefix:", o.TopicPrefix)
	fmt.Println("responseTopic:", o.ResponseTopic)

	return o, err
}

func NewMQTTClient(o *MQTTOptions, redisPool *redis.Pool, channel *channel.KafkaMqttChannel, logger *zap.Logger) *MQTTClient {
	mc := &MQTTClient{
		o:                   o,
		Addr:                o.Addr,
		User:                o.User,
		Passwd:              o.Passwd,
		ClientID:            o.ClientID,
		Order:               false,
		KeepAliveInterval:   120 * time.Second,
		PingTimeout:         120 * time.Second,
		MessageChannelDepth: 100,
		QOS:                 byte(2),
		Retain:              false,
		CleanSession:        true,
		FileStorePath:       "memory",
		WillTopic:           "", //no will topic.
		kafkamqttChannel:    channel,
		logger:              logger.With(zap.String("type", "mqtt.Client")),
		redisPool:           redisPool,
	}
	// mc.OnConnect = mc.OnMQTTConnect
	// mc.OnLost = mc.OnMQTTLost
	return mc
}

func (mc *MQTTClient) OnMQTTConnect(client paho.Client) {
	mc.logger.Info("Client connected ", zap.String("ClientID", mc.ClientID))

	mc.State = MQTT_CLIENT_CONNECTED
}

func (mc *MQTTClient) OnMQTTLost(client paho.Client, err error) {
	mc.logger.Error("Client disconnected with error ", zap.Error(err))

	mc.State = MQTT_CLIENT_DISCONNECTED

}

func (mc *MQTTClient) Application(name string) {
	mc.app = name
}

func (mc *MQTTClient) Start() error {

	conn, err := net.Dial("tcp", mc.Addr)
	if err != nil {
		mc.logger.Error("Client dial error ", zap.String("BrokerServer", mc.Addr), zap.Error(err))
		return errors.New("BrokerServer dial error")
	}

	// Create paho client.
	mc.client = paho.NewClient(paho.ClientConfig{
		Router: paho.NewSingleHandlerRouter(func(m *paho.Publish) {
			topic := m.Topic
			//在屏幕上输出chat消息
			// log.Printf("%s : %s", m.Properties.User["deviceId"], string(m.Payload))
			jwtToken := string(m.Properties.CorrelationData)
			deviceId := m.Properties.User["deviceId"]
			businessTypeStr := m.Properties.User["businessType"]
			businessSubTypeStr := m.Properties.User["businessSubType"]
			taskIdStr := m.Properties.User["taskId"]

			taskId, _ := strconv.Atoi(taskIdStr)
			businessType, _ := strconv.Atoi(businessTypeStr)
			businessSubType, _ := strconv.Atoi(businessSubTypeStr)

			//是否是必须经过授权的请求包
			isAuthed := false
			userName := ""
			// if !(businessType == 2 && businessSubType == 1) {
			//是否需要有效授权才能传递到后端
			if userName, isAuthed, err = mc.MakeSureAuthed(jwtToken, deviceId, businessType, businessSubType, taskId); err != nil {
				mc.logger.Error("MakeSureAuthed error", zap.String("Error", err.Error()))
				return
			} else {
				if !isAuthed {
					mc.logger.Warn("This message is unauthirized!!!")
					return
				}
			}
			// }

			//输出
			mc.logger.Debug("Incoming mqtt message",
				zap.String("jwtToken", jwtToken),
				zap.String("userName", userName),
				zap.String("Topic", topic),
				zap.String("DeviceId", deviceId),            // 设备id
				zap.Int("TaskID", taskId),                   // 任务id
				zap.Int("BusinessType", businessType),       // 业务类型
				zap.Int("BusinessSubType", businessSubType), // 业务子类型
			)

			businessTypeName := Global.BusinessType_name[int32(businessType)]

			kafkaTopic := businessTypeName + ".Backend"
			backendService := businessTypeName + "Service"

			//重要! 构造在后端传输的消息，包括：消息头，消息路由，业务负载
			backendMsg := &models.Message{}
			now := time.Now().UnixNano() / 1e6
			backendMsg.UpdateID()
			//构建消息路由, 第一个参数是要处理的业务类型，后端服务器处理完成后，需要用此来拼接topic: {businessTypeName.Frontend}
			backendMsg.BuildRouter(businessTypeName, "", kafkaTopic)
			
			backendMsg.SetJwtToken(jwtToken)
			backendMsg.SetUserName(userName)
			backendMsg.SetDeviceId(string(deviceId))
			backendMsg.SetTaskId(uint32(taskId))
			backendMsg.SetBusinessTypeName(businessTypeName)
			backendMsg.SetBusinessType(uint32(businessType))
			backendMsg.SetBusinessSubType(uint32(businessSubType))

			backendMsg.BuildHeader(backendService, now)
			backendMsg.FillBody(m.Payload) //承载真正的业务数据

			bodyReqHex := strings.ToUpper(hex.EncodeToString(m.Payload))
			mc.logger.Info("GetContent ok", zap.String("bodyReqHex", bodyReqHex))

			//分发到后端的各自的服务器
			switch Global.BusinessType(businessType) {
			case Global.BusinessType_Auth: //鉴权授权模块
				mc.kafkamqttChannel.KafkaChan <- backendMsg //发送到kafka
				mc.logger.Info("message has send to KafkaChan", zap.String("kafkaTopic", kafkaTopic), zap.String("backendService", backendService))

			case Global.BusinessType_Users:
				//TODO

			case Global.BusinessType_Friends:
				//TODO

			case Global.BusinessType_Group:
				//TODO

			case Global.BusinessType_Chat:
				//TODO

			case Global.BusinessType_Sync: //同步服务程序
				//TODO

			default: //default case
				mc.logger.Warn("Incorrect business type", zap.Int("businessType", businessType), zap.String("m.Payload", string(m.Payload)))
				return
			}

		}),
		Conn: conn,
	})

	cp := &paho.Connect{
		KeepAlive:  30,
		ClientID:   mc.ClientID,
		CleanStart: true,
		Username:   mc.User,
		Password:   []byte(mc.Passwd),
	}

	if mc.User != "" {
		cp.UsernameFlag = true
	}
	if mc.Passwd != "" {
		cp.PasswordFlag = true
	}

	//订阅的topic， 客户端发送这个topic
	// subTopic := "lianmi/cloud/dispatcher"
	subTopic := mc.o.ResponseTopic
	for i := 0; i < retryCount; i++ {
		ca, err := mc.client.Connect(context.Background(), cp)
		if err == nil {
			if ca.ReasonCode == 0 {
				mc.logger.Info("Connected to broker server successfule.", zap.String("BrokerServer", mc.Addr))
				if _, err := mc.client.Subscribe(context.Background(), &paho.Subscribe{
					Subscriptions: map[string]paho.SubscribeOptions{
						subTopic: paho.SubscribeOptions{QoS: byte(1), NoLocal: true},
					},
				}); err != nil {
					mc.logger.Error("Failed to subscribe", zap.Error(err))
				}
				mc.logger.Info("Subscribed succed", zap.String("subTopic", subTopic))

				//启动主循环Go程, 发送给SDK的订阅回包
				go mc.Run()

				return nil
			}
		} else {
			mc.logger.Error("Failed to connect to broker ", zap.String("BrokerServer", mc.Addr), zap.Uint8("ReasonCode", ca.ReasonCode), zap.String("ReasonString", ca.Properties.ReasonString), zap.Error(err))

		}
		time.Sleep(cloudAccessSleep)
	}

	return errors.New("max retry count reached when connecting to cloud")
}

func (mc *MQTTClient) Close() {
	// mc.client.Disconnect(250)
}

//主循环，从kafka读取消息，并发送到imsdk的某个设备
func (mc *MQTTClient) Run() {
	run := true
	sigchan := make(chan os.Signal, 1)
	signal.Notify(sigchan, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	for run == true {
		select {
		case sig := <-sigchan:
			mc.logger.Info("Caught signal terminating")
			_ = sig
			run = false
			d := &paho.Disconnect{ReasonCode: 0}
			err := mc.client.Disconnect(d)
			if err != nil {
				mc.logger.Error("failed to send Disconnect", zap.Error(err))
			}
		case msg := <-mc.kafkamqttChannel.MTChan: //从kafka读取数据
			if msg != nil {
				//向MQTT Broker发送，加入SDK订阅了此topic，则会收到
				jwtToken := msg.GetJwtToken()
				topic := mc.o.TopicPrefix + msg.GetDeviceId()
				// topic := "lianmi/cloud/device/" + msg.GetDeviceId()
				businessTypeStr := fmt.Sprintf("%d", msg.GetBusinessType())
				businessSubTypeStr := fmt.Sprintf("%d", msg.GetBusinessSubType())
				taskIdStr := fmt.Sprintf("%d", msg.GetTaskId())
				codeStr := fmt.Sprintf("%d", msg.GetCode())
				mc.logger.Info("Consume backend kafak message",
					zap.String("topic", topic),
					zap.String("deviceId", msg.GetDeviceId()),
					zap.String("businessType", businessTypeStr),
					zap.String("businessSubType", businessSubTypeStr),
					zap.String("taskId", taskIdStr),
					zap.String("code", codeStr))

				pb := &paho.Publish{
					Topic:   topic,
					QoS:     byte(1),
					Payload: msg.Content,
					Properties: &paho.PublishProperties{
						CorrelationData: []byte(jwtToken),   //jwt令牌
						ResponseTopic:   mc.o.ResponseTopic, //"lianmi/cloud/dispatcher",
						User: map[string]string{
							"businessType":    businessTypeStr,
							"businessSubType": businessSubTypeStr,
							"taskId":          taskIdStr,
							"code":            codeStr,
							"errormsg":        string(msg.GetErrorMsg()),
						},
					},
				}

				if _, err := mc.client.Publish(context.Background(), pb); err != nil {
					// log.Println(err)
					mc.logger.Error("Failed to Publish ", zap.Error(err))
				} else {
					mc.logger.Info("Succeed Publish  to sdk", zap.String("topic", topic))
				}
			} else {
				mc.logger.Warn("msg is nil")
			}
		}
	}
}

func (mc *MQTTClient) Stop() error {

	// scope.Close()

	return nil
}

func ParseToken(tokenSrt string, SecretKey []byte) (claims jwt.Claims, err error) {
	var token *jwt.Token
	token, err = jwt.Parse(tokenSrt, func(*jwt.Token) (interface{}, error) {
		return SecretKey, nil
	})
	claims = token.Claims
	return
}

/*
jwtToken如果在redis里没找到，原因可能有如下三种：
1. 用户未登录；
2. 登录超时（redis中的数据到期自动删除）；
3. SDK存储的 jwtToken, 被非法修改；

该情况直接返回错误信息通知前端让用户重新登录；
*/
func (mc *MQTTClient) MakeSureAuthed(jwtToken, deviceId string, businessType, businessSubType, taskId int) (string, bool, error) {
	mc.logger.Info("MakeSureAuthed start...", zap.Int("businessType:", businessType), zap.Int("businessSubType:", businessSubType), zap.Int("taskId:", taskId), zap.String("jwtToken:", jwtToken))

	var isAuthed bool = false
	var isExists bool = false
	var userName string
	var err error

	//eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE1OTczNzc2MzksIm9yaWdfaWF0IjoxNTk3Mzc0MDM5LCJ1c2VyTmFtZSI6ImFkbWluIiwidXNlclJvbGVzIjoiW3tcImlkXCI6MSxcInVzZXJfaWRcIjoxLFwidXNlcl9uYW1lXCI6XCJhZG1pblwiLFwidmFsdWVcIjpcImFkbWluXCJ9XSJ9.hmh3a3EFTi53seHnfxETXp7MYfmgW20MB4_6eseEb34
	if jwtToken != "" {
		//TODO redis里查找EXISTS('jwtToken', jwtToken)
		redisConn := mc.redisPool.Get()
		defer redisConn.Close()

		if isExists, err = redis.Bool(redisConn.Do("GET", jwtToken)); err != nil {
			mc.logger.Error("redisConn GET Error", zap.Error(err))
			isAuthed = false
		} else {
			mc.logger.Info("redisConn GET ok ", zap.String("jwtToken", jwtToken))
		}

		if isExists {
			//验证token是否有效
			claims, err := ParseToken(jwtToken, []byte(common.SecretKey))
			if nil != err {
				mc.logger.Error("ParseToken Error", zap.Error(err))
			}

			//jwt令牌里的用户名
			userName = claims.(jwt.MapClaims)[common.IdentityKey].(string)
			mc.logger.Debug("jwt令牌里的用户名", zap.String("userName", userName))

			isBlocked := true
			//TODO 检测userName是否在封号名单上，如果是，则不授权
			if isBlocked, err = redis.Bool(redisConn.Do("SISMEMBER", "BlockedSet", userName)); err != nil {
				mc.logger.Error("redisConn SISMEMBER Error", zap.Error(err))
			} else {
				if isBlocked {
					mc.logger.Debug("此用户被封禁了", zap.String("userName", userName))
					isAuthed = false
				} else {
					isAuthed = true
				}
			}

		} else {
			isAuthed = false
		}

	}

	if isAuthed {
		mc.logger.Debug("MakeSureAuthed, 授权通过", zap.Bool("isAuthed", isAuthed))
		return userName, isAuthed, nil

	} else {

		// topic := "lianmi/cloud/device/" + deviceId

		topic := mc.o.TopicPrefix + deviceId
		businessTypeStr := fmt.Sprintf("%d", businessType)
		businessSUbTypeStr := fmt.Sprintf("%d", businessSubType)
		taskIdStr := fmt.Sprintf("%d", taskId)

		pb := &paho.Publish{
			Topic:   topic,
			QoS:     mc.QOS,
			Payload: []byte{},
			Properties: &paho.PublishProperties{
				CorrelationData: []byte("error"), //jwt令牌，填error，客户端必须重新登录进行认证
				ResponseTopic:   mc.o.ResponseTopic,
				User: map[string]string{
					"businessType":    businessTypeStr,
					"businessSubType": businessSUbTypeStr,
					"taskId":          taskIdStr,
					"code":            "403", //无授权
					"errormsg":        "Without authorization",
				},
			},
		}

		if _, err := mc.client.Publish(context.Background(), pb); err != nil {
			// log.Println(err)
			mc.logger.Error("Failed to Publish ", zap.Error(err))
		}
		return userName, false, errors.New("DeviceId is not authorized")
	}

}

var ProviderSet = wire.NewSet(NewMQTTOptions, NewMQTTClient)
