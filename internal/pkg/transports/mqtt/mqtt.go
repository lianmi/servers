package mqtt

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/gomodule/redigo/redis"
	"github.com/google/wire"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"

	"github.com/eclipse/paho.golang/paho" //支持v5.0

	Global "github.com/lianmi/servers/api/proto/global"
	Log "github.com/lianmi/servers/api/proto/log"
	"github.com/lianmi/servers/internal/pkg/channel"
	"github.com/lianmi/servers/internal/pkg/models"

	jwt "github.com/dgrijalva/jwt-go"
	LMCommon "github.com/lianmi/servers/internal/common"
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
	CaPath        string `yaml:"caPath"`        //ca path
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
	TLSConfig      *tls.Config
	client         *paho.Client //支持v5.0
	nsqMqttChannel *channel.NsqMqttChannel
	logger         *zap.Logger
	redisPool      *redis.Pool
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
	fmt.Println("caPath:", o.CaPath)

	return o, err
}

func NewMQTTClient(o *MQTTOptions, redisPool *redis.Pool, channel *channel.NsqMqttChannel, logger *zap.Logger) *MQTTClient {
	if LMCommon.IsUseCa {
		certpool := x509.NewCertPool()
		ca, err := ioutil.ReadFile(o.CaPath + "/ca.crt")
		if err != nil {
			log.Fatalln(err.Error())
		}
		certpool.AppendCertsFromPEM(ca)
		// Import client certificate/key pair
		clientKeyPair, err := tls.LoadX509KeyPair(o.CaPath+"/mqtt.lianmi.cloud.crt", o.CaPath+"/mqtt.lianmi.cloud.key")
		if err != nil {
			panic(err)
		}

		tlsConfig := &tls.Config{
			RootCAs:            certpool,
			ClientAuth:         tls.NoClientCert,
			ClientCAs:          nil,
			InsecureSkipVerify: true,
			Certificates:       []tls.Certificate{clientKeyPair},
		}

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
			QOS:                 byte(2), //1: QOSAtLeastOnce, 2: QOSExactlyOnce.
			Retain:              false,
			CleanSession:        true,
			FileStorePath:       "memory",
			WillTopic:           "", //no will topic.
			TLSConfig:           tlsConfig,
			nsqMqttChannel:      channel,
			logger:              logger.With(zap.String("type", "mqtt.Client")),
			redisPool:           redisPool,
		}
		return mc
	} else {
		return &MQTTClient{
			o:                   o,
			Addr:                o.Addr,
			User:                o.User,
			Passwd:              o.Passwd,
			ClientID:            o.ClientID,
			Order:               false,
			KeepAliveInterval:   120 * time.Second,
			PingTimeout:         120 * time.Second,
			MessageChannelDepth: 100,
			QOS:                 byte(2), //1: QOSAtLeastOnce, 2: QOSExactlyOnce.
			Retain:              false,
			CleanSession:        true,
			FileStorePath:       "memory",
			WillTopic:           "", //no will topic.
			TLSConfig:           nil,
			nsqMqttChannel:      channel,
			logger:              logger.With(zap.String("type", "mqtt.Client")),
			redisPool:           redisPool,
		}
	}

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

	if LMCommon.IsUseCa { //利用TLS协议连接broker
		conn, err := tls.Dial("tcp", mc.Addr, mc.TLSConfig)

		if err != nil {
			mc.logger.Error("tls.Dial error ", zap.String("BrokerServer", mc.Addr), zap.Error(err))
			return errors.New("tls.Dial error")
		}
		if conn == nil {
			return errors.New("tls.Dial error, conn is nil")
		}

		// Create paho client.
		mc.client = paho.NewClient(paho.ClientConfig{
			Router: paho.NewSingleHandlerRouter(func(m *paho.Publish) {
				topic := m.Topic
				jwtToken := m.Properties.User["jwtToken"] // Add by lishijia  for flutter mqtt
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

				//从设备申请授权码，此时还没有令牌
				if businessType == 2 && businessSubType == 7 {
					mc.logger.Debug("从设备申请授权码，此时还没有令牌")

				} else if businessType == int(Global.BusinessType_Log) && (businessSubType == 1) {
					mc.logger.Debug("=====日志======",
						zap.ByteString("log", m.Payload),
					)

					mc.SendLogMsg(m.Payload)

				} else {
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
				}

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

				nsqTopic := businessTypeName + ".Backend"
				backendService := businessTypeName + "Service"

				//重要! 构造在后端传输的消息，包括：消息头，消息路由，业务负载
				backendMsg := &models.Message{}
				now := time.Now().UnixNano() / 1e6
				backendMsg.UpdateID()
				//构建消息路由, 第一个参数是要处理的业务类型，后端服务器处理完成后，需要用此来拼接topic: {businessTypeName.Frontend}
				backendMsg.BuildRouter(businessTypeName, "", nsqTopic)

				backendMsg.SetJwtToken(jwtToken)
				backendMsg.SetUserName(userName)
				backendMsg.SetDeviceID(string(deviceId))
				backendMsg.SetTaskID(uint32(taskId))
				backendMsg.SetBusinessTypeName(businessTypeName)
				backendMsg.SetBusinessType(uint32(businessType))
				backendMsg.SetBusinessSubType(uint32(businessSubType))

				backendMsg.BuildHeader(backendService, now)
				backendMsg.FillBody(m.Payload) //承载真正的业务数据

				//分发
				switch Global.BusinessType(businessType) {
				case Global.BusinessType_User,
					Global.BusinessType_Auth,
					Global.BusinessType_Friends,
					Global.BusinessType_Team,
					Global.BusinessType_Sync,
					Global.BusinessType_Msg,
					Global.BusinessType_Product,
					Global.BusinessType_Order,
					Global.BusinessType_Wallet:

					//发送到Nsq
					mc.nsqMqttChannel.NsqChan <- backendMsg
					mc.logger.Info("Message发送到Nsq通道",
						zap.String("nsqTopic", nsqTopic),
						zap.String("backendService", backendService),
						zap.Int("businessType", businessType),
						zap.Int("businessSubType", businessSubType),
						zap.String("msgID", backendMsg.GetID()),
					)

				case Global.BusinessType_Log: //日志, 转发到日志订阅服务器

				case Global.BusinessType_Custom: //自定义服务， 一般用于测试

				default: //default case
					mc.logger.Warn("Incorrect business type", zap.Int("businessType", businessType), zap.String("m.Payload", string(m.Payload)))
					return
				}

			}),
			Conn: conn,
		})
	} else { //利用TCP协议连接broker
		conn, err := net.Dial("tcp", mc.Addr)

		if err != nil {
			mc.logger.Error("net.Dial error ", zap.String("BrokerServer", mc.Addr), zap.Error(err))
			return errors.New("net.Dial error")
		}
		if conn == nil {
			return errors.New("net.Dial error, conn is nil")
		}

		// Create paho client.
		mc.client = paho.NewClient(paho.ClientConfig{
			Router: paho.NewSingleHandlerRouter(func(m *paho.Publish) {
				topic := m.Topic
				jwtToken := m.Properties.User["jwtToken"] // Add by lishijia  for flutter mqtt
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
				mc.logger.Debug("=====是否是必须经过授权的请求包======",
					zap.Int("BusinessType", businessType),       // 业务类型
					zap.Int("BusinessSubType", businessSubType), // 业务子类型
				)

				//从设备申请授权码，此时还没有令牌
				if businessType == 2 && businessSubType == 7 {
					mc.logger.Debug("从设备申请授权码，此时还没有令牌")

				} else {
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
				}

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

				nsqTopic := businessTypeName + ".Backend"
				backendService := businessTypeName + "Service"

				//重要! 构造在后端传输的消息，包括：消息头，消息路由，业务负载
				backendMsg := &models.Message{}
				now := time.Now().UnixNano() / 1e6
				backendMsg.UpdateID()
				//构建消息路由, 第一个参数是要处理的业务类型，后端服务器处理完成后，需要用此来拼接topic: {businessTypeName.Frontend}
				backendMsg.BuildRouter(businessTypeName, "", nsqTopic)

				backendMsg.SetJwtToken(jwtToken)
				backendMsg.SetUserName(userName)
				backendMsg.SetDeviceID(string(deviceId))
				backendMsg.SetTaskID(uint32(taskId))
				backendMsg.SetBusinessTypeName(businessTypeName)
				backendMsg.SetBusinessType(uint32(businessType))
				backendMsg.SetBusinessSubType(uint32(businessSubType))

				backendMsg.BuildHeader(backendService, now)
				backendMsg.FillBody(m.Payload) //承载真正的业务数据

				//分发
				switch Global.BusinessType(businessType) {
				case Global.BusinessType_User,
					Global.BusinessType_Auth,
					Global.BusinessType_Friends,
					Global.BusinessType_Team,
					Global.BusinessType_Sync,
					Global.BusinessType_Msg,
					Global.BusinessType_Product,
					Global.BusinessType_Order,
					Global.BusinessType_Wallet:

					//发送到Nsq
					mc.nsqMqttChannel.NsqChan <- backendMsg
					mc.logger.Info("Message发送到Nsq通道",
						zap.String("nsqTopic", nsqTopic),
						zap.String("backendService", backendService),
						zap.Int("businessType", businessType),
						zap.Int("businessSubType", businessSubType),
						zap.String("msgID", backendMsg.GetID()),
					)
					break //与C语言等规则相反，Go语言不需要用break来明确退出一个case；

				case Global.BusinessType_Custom: //自定义服务， 一般用于测试

				default: //default case
					mc.logger.Warn("Incorrect business type", zap.Int("businessType", businessType), zap.String("m.Payload", string(m.Payload)))
					return
				}

			}),
			Conn: conn,
		})
	}

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
	subTopic := mc.o.ResponseTopic
	for i := 0; i < retryCount; i++ {
		ca, err := mc.client.Connect(context.Background(), cp)
		if err == nil {
			if ca.ReasonCode == 0 {
				mc.logger.Info("Connected to broker server successful.", zap.String("BrokerServer", mc.Addr))
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

//主循环，从MTChan读取消息，并发送到imsdk的某个设备
func (mc *MQTTClient) Run() {
	var run bool
	sigchan := make(chan os.Signal, 1)
	signal.Notify(sigchan, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	run = true
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
		case msg := <-mc.nsqMqttChannel.MTChan: //从MTChan读取数据
			if msg != nil {
				//向MQTT Broker发送，加入SDK订阅了此topic，则会收到
				jwtToken := msg.GetJwtToken()
				topic := mc.o.TopicPrefix + msg.GetDeviceID()
				// topic := "lianmi/cloud/device/" + msg.GetDeviceID()
				businessTypeStr := fmt.Sprintf("%d", msg.GetBusinessType())
				businessSubTypeStr := fmt.Sprintf("%d", msg.GetBusinessSubType())
				taskIdStr := fmt.Sprintf("%d", msg.GetTaskID())
				codeStr := fmt.Sprintf("%d", msg.GetCode())
				mc.logger.Info("Consume backend nsq message for send to mqtt broker",
					zap.String("topic", topic),
					zap.String("deviceId", msg.GetDeviceID()),
					zap.String("businessType", businessTypeStr),
					zap.String("businessSubType", businessSubTypeStr),
					zap.String("taskId", taskIdStr),
					zap.String("code", codeStr))

				pb := &paho.Publish{
					Topic:   topic,
					QoS:     byte(1),
					Payload: msg.Content,
					Properties: &paho.PublishProperties{
						ResponseTopic: mc.o.ResponseTopic, //"lianmi/cloud/dispatcher",
						User: map[string]string{
							"jwtToken":        jwtToken,
							"deviceId":        msg.GetDeviceID(),
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
					mc.logger.Error("Failed to Publish to broker ", zap.Error(err))
				} else {
					mc.logger.Info("Succeed Publish to broker", zap.String("topic", topic))
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

func (mc *MQTTClient) SendLogMsg(body []byte) error {
	var err error

	//解包body
	req := &Log.SendLogReq{}
	if err = proto.Unmarshal(body, req); err != nil {
		mc.logger.Error("Protobuf Unmarshal Error", zap.Error(err))
		return err

	} else {
		mc.logger.Debug("SendLogReq  payload",
			zap.String("Username", req.Username),
			zap.String("Content", req.Content),
		)
	}

	//向MQTT Broker发送，加入SDK订阅了Log topic，则会收到
	jwtToken := ""
	topic := "lianmi/cloud/sdklogs"
	businessTypeStr := fmt.Sprintf("%d", int(Global.BusinessType_Log))
	businessSubTypeStr := fmt.Sprintf("%d", 1)
	taskIdStr := fmt.Sprintf("%d", 0)
	codeStr := fmt.Sprintf("%d", 200)

	mc.logger.Info("Send LogMsg to mqtt broker",
		zap.String("topic", topic),
		zap.String("businessType", businessTypeStr),
		zap.String("code", codeStr))

	pb := &paho.Publish{
		Topic:   topic,
		QoS:     byte(2),
		Payload: []byte(req.Content),
		Properties: &paho.PublishProperties{
			User: map[string]string{
				"jwtToken":        jwtToken,
				"deviceId":        "",
				"businessType":    businessTypeStr,
				"businessSubType": businessSubTypeStr,
				"taskId":          taskIdStr,
				"code":            codeStr,
				"errormsg":        "",
			},
		},
	}

	if _, err := mc.client.Publish(context.Background(), pb); err != nil {
		// log.Println(err)
		mc.logger.Error("Failed to Publish to broker ", zap.Error(err))
	} else {
		mc.logger.Info("Succeed Publish to broker", zap.String("topic", topic))
	}

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
func (mc *MQTTClient) MakeSureAuthed(jwtToken, deviceID string, businessType, businessSubType, taskID int) (string, bool, error) {
	mc.logger.Info("MakeSureAuthed start...",
		zap.Int("businessType:", businessType),
		zap.Int("businessSubType:", businessSubType),
		zap.Int("taskID:", taskID),
		zap.String("jwtToken:", jwtToken))

	var isAuthed bool = false
	var jwtUserName string
	var tokenInRedis string
	var jwtDeviceID string
	var err error

	if jwtToken != "" {
		//TODO redis里查找EXISTS('jwtToken', jwtToken)
		redisConn := mc.redisPool.Get()
		defer redisConn.Close()
		deviceKey := fmt.Sprintf("DeviceJwtToken:%s", deviceID)
		if tokenInRedis, err = redis.String(redisConn.Do("GET", deviceKey)); err != nil {
			mc.logger.Error("redisConn GET JWT Error", zap.String("deviceKey", deviceKey), zap.Error(err))
			isAuthed = false
		} else {
			mc.logger.Info("redisConn GET JWT ok ", zap.String("tokenInRedis", tokenInRedis))

			if tokenInRedis == jwtToken {
				//验证token是否有效
				claims, err := ParseToken(jwtToken, []byte(LMCommon.SecretKey))
				if nil != err {
					mc.logger.Error("ParseToken Error", zap.Error(err))
				}

				//jwt令牌里的用户名
				jwtUserName = claims.(jwt.MapClaims)[LMCommon.IdentityKey].(string)
				jwtDeviceID = claims.(jwt.MapClaims)["deviceID"].(string)
				mc.logger.Debug("jwt令牌", zap.String("jwtUserName", jwtUserName), zap.String("jwtDeviceID", jwtDeviceID))

				if deviceID == jwtDeviceID {

					deviceKey := fmt.Sprintf("DeviceJwtToken:%s", jwtDeviceID)
					//判断deviceKey是否存在这个key，如果设备登出后，这个key就会删除，如果非法获取token，也无法使用
					if isExistDeviceID, err := redis.Bool(redisConn.Do("EXISTS", deviceKey)); err != nil {
						mc.logger.Error("redisConn GET DeviceID Error", zap.String("deviceKey", deviceKey), zap.Error(err))
						isAuthed = false
					} else {
						if isExistDeviceID {
							isBlocked := true
							//TODO 检测userName是否在封号名单上，如果是，则不授权
							if isBlocked, err = redis.Bool(redisConn.Do("SISMEMBER", "BlockedSet", jwtUserName)); err != nil {
								mc.logger.Error("redisConn SISMEMBER Error", zap.Error(err))
							} else {
								if isBlocked {
									mc.logger.Debug("此用户被封禁了", zap.String("jwtUserName", jwtUserName))
									isAuthed = false
								} else {
									isAuthed = true //授权通过
								}
							}
						}
					}

				} else {
					mc.logger.Warn("警告, 令牌里的设备id和消息携带的设备id不相同，需要重新登录授权", zap.String("deviceID", deviceID), zap.String("jwtDeviceID", jwtDeviceID))
				}
			}
		}
	}

	if isAuthed {
		mc.logger.Debug("MakeSureAuthed, 授权通过", zap.Bool("isAuthed", isAuthed))
		return jwtUserName, isAuthed, nil

	} else {
		mc.logger.Warn("MakeSureAuthed, 授权被拒绝", zap.Bool("isAuthed", isAuthed))
		// topic := "lianmi/cloud/device/" + deviceId

		topic := mc.o.TopicPrefix + deviceID
		businessTypeStr := fmt.Sprintf("%d", businessType)
		businessSUbTypeStr := fmt.Sprintf("%d", businessSubType)
		taskIDStr := fmt.Sprintf("%d", taskID)

		pb := &paho.Publish{
			Topic:   topic,
			QoS:     mc.QOS,
			Payload: []byte{},
			Properties: &paho.PublishProperties{
				ResponseTopic: mc.o.ResponseTopic,
				User: map[string]string{
					"jwtToken":        "none",
					"deviceId":        deviceID,
					"businessType":    businessTypeStr,
					"businessSubType": businessSUbTypeStr,
					"taskId":          taskIDStr,
					"code":            "403", //无授权
					"errormsg":        "Without authorization",
				},
			},
		}

		if _, err := mc.client.Publish(context.Background(), pb); err != nil {
			// log.Println(err)
			mc.logger.Error("Failed to Publish ", zap.Error(err))
		}
		return jwtUserName, false, errors.New("DeviceId is not authorized")
	}

}

var ProviderSet = wire.NewSet(NewMQTTOptions, NewMQTTClient)
