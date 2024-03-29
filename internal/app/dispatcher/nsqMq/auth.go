/*
本文件是处理业务号是授权鉴权模块，分别有
2-1 登录 /login
2-2 登出 SignOut
2-3 同步其它设备登录状态事件 MultiLoginEvent
2-4 踢出其它在线设备 Kick
2-5 在线设备被踢下线事件 KickedEvent
2-6 添加从设备 AddSlaveDevice
2-7 从设备申请授权码 AuthorizeCode
2-8 从设备被授权登录事件 SlaveDeviceAuthEvent
2-9 获取所有主从设备  GetAllDevices
*/
package nsqMq

import (
	"encoding/json"
	"fmt"

	// "strings"
	"time"

	// "net/http"

	"github.com/golang/protobuf/proto"
	"github.com/gomodule/redigo/redis"
	Auth "github.com/lianmi/servers/api/proto/auth"
	LMCommon "github.com/lianmi/servers/internal/common"
	LMCError "github.com/lianmi/servers/internal/pkg/lmcerror"
	"github.com/lianmi/servers/internal/pkg/models"

	// "github.com/lianmi/servers/util/randtool"
	"go.uber.org/zap"
)

/*
2-2 登出
1. 主设备登出，需要删除从设备一切数据，踢出从设备
2. 从设备登出，只删除自己的数据，并刷新此用户的在线设备列表
*/
func (nc *NsqClient) HandleSignOut(msg *models.Message) error {
	var err error

	//将此设备从在线列表里删除，然后更新对应用户的在线列表。
	redisConn := nc.redisPool.Get()
	defer redisConn.Close()

	username := msg.GetUserName()
	// token := msg.GetJwtToken()
	deviceID := msg.GetDeviceID()
	nc.logger.Info("HandleSignOut start...",
		zap.String("username", username),
		zap.String("DeviceId", deviceID))

	//取出当前旧的设备的os，  logonAt
	curDeviceHashKey := fmt.Sprintf("devices:%s:%s", username, deviceID)
	isMaster, _ := redis.Bool(redisConn.Do("HGET", curDeviceHashKey, "ismaster"))
	curOs, _ := redis.String(redisConn.Do("HGET", curDeviceHashKey, "os"))
	curLogonAt, _ := redis.Uint64(redisConn.Do("HGET", curDeviceHashKey, "logonAt"))

	nc.logger.Debug("SignOut",
		zap.Bool("isMaster", isMaster),
		zap.String("username", username),
		zap.String("deviceID", deviceID),
		zap.String("curOs", curOs),
		zap.Uint64("curLogonAt", curLogonAt))

	redisConn.Do("DEL", fmt.Sprintf("devices:%s", username))
	redisConn.Do("DEL", fmt.Sprintf("devices:%s:%s", username, deviceID)) //删除deviceHashKey
	redisConn.Do("DEL", fmt.Sprintf("DeviceJwtToken:%s", deviceID))

	//向dispatcher发送
	topic := msg.GetSource() + ".Frontend"
	msg.SetCode(200) //成功的状态码
	rawData, _ := json.Marshal(msg)
	if err := nc.Producer.Public(topic, rawData); err == nil {
		nc.logger.Info("Succeed to send message to ProduceChannel", zap.String("topic", topic))
	} else {
		nc.logger.Error("Failed to send message to ProduceChannel", zap.Error(err))
	}

	_ = err

	nc.logger.Debug("登出成功",
		zap.Bool("isMaster", isMaster),
		zap.String("username", username),
		zap.String("deviceID", deviceID),
		zap.String("curOs", curOs),
		zap.Uint64("curLogonAt", curLogonAt))

	return nil
}

/*
2-4 踢出其它在线设备 Kick
1. 主设备才能踢出从设备, 被踢的从设备收到被踢下线事件
2. 从设备被踢后，只删除自己的数据，并发出多端登录状态变化事件
*/
func (nc *NsqClient) HandleKick(msg *models.Message) error {
	var err error
	errorCode := 200 //错误码， 200是正常，其它是错误

	//将此设备从在线列表里删除，然后更新对应用户的在线列表。
	redisConn := nc.redisPool.Get()
	defer redisConn.Close()

	username := msg.GetUserName()
	// token := msg.GetJwtToken()
	deviceID := msg.GetDeviceID()

	nc.logger.Info("HandleKick start...",
		zap.String("username", username),
		zap.String("DeviceId", deviceID))

	//取出当前设备的os，  logonAt
	curDeviceHashKey := fmt.Sprintf("devices:%s:%s", username, deviceID)
	isMaster, _ := redis.Bool(redisConn.Do("HGET", curDeviceHashKey, "ismaster"))
	curOs, _ := redis.String(redisConn.Do("HGET", curDeviceHashKey, "os"))
	curLogonAt, _ := redis.Uint64(redisConn.Do("HGET", curDeviceHashKey, "logonAt"))

	nc.logger.Debug("Kick",
		zap.Bool("isMaster", isMaster),
		zap.String("username", username),
		zap.String("deviceID", deviceID),
		zap.String("curOs", curOs),
		zap.Uint64("curLogonAt", curLogonAt))

	msg.SetCode(int32(errorCode)) //状态码
	if errorCode == 200 {
		msg.FillBody(nil) //网络包的body，承载真正的业务数据
	} else {
		// errorMsg := LMCError.ErrorMsg(errorCode)
		msg.FillBody(nil)
	}

	//处理完成，向dispatcher发送
	topic := msg.GetSource() + ".Frontend"
	rawData, _ := json.Marshal(msg)
	if err := nc.Producer.Public(topic, rawData); err == nil {
		nc.logger.Info("KickRsp message succeed send to ProduceChannel", zap.String("topic", topic))
	} else {
		nc.logger.Error("Failed to send KickRsp message to ProduceChannel", zap.Error(err))
	}

	_ = err
	return nil
}

/*
2-9 获取所有主从设备
*/
func (nc *NsqClient) HandleGetAllDevices(msg *models.Message) error {
	var err error

	redisConn := nc.redisPool.Get()
	defer redisConn.Close()

	username := msg.GetUserName()
	// token := msg.GetJwtToken()
	deviceID := msg.GetDeviceID()

	nc.logger.Info("HandleGetAllDevices start...",
		zap.String("username", username),
		zap.String("DeviceId", deviceID))

	//取出当前设备的os，  logonAt
	curDeviceHashKey := fmt.Sprintf("devices:%s:%s", username, deviceID)
	isMaster, _ := redis.Bool(redisConn.Do("HGET", curDeviceHashKey, "ismaster"))
	curOs, _ := redis.String(redisConn.Do("HGET", curDeviceHashKey, "os"))
	curLogonAt, _ := redis.Uint64(redisConn.Do("HGET", curDeviceHashKey, "logonAt"))

	nc.logger.Debug("GetAllDevices",
		zap.Bool("isMaster", isMaster),
		zap.String("username", username),
		zap.String("deviceID", deviceID),
		zap.String("curOs", curOs),
		zap.Uint64("curLogonAt", curLogonAt))

	deviceListKey := fmt.Sprintf("devices:%s", username)
	eDeviceID, _ := redis.String(redisConn.Do("GET", deviceListKey))
	rsp := &Auth.GetAllDevicesRsp{}
	rsp.OnlineDevices = make([]*Auth.DeviceInfo, 0)
	rsp.OfflineDevices = make([]*Auth.DeviceInfo, 0)

	curDeviceKey := fmt.Sprintf("DeviceJwtToken:%s", eDeviceID)
	curJwtToken, _ := redis.String(redisConn.Do("GET", curDeviceKey))
	nc.logger.Debug("Redis GET ", zap.String("curDeviceKey", curDeviceKey), zap.String("curJwtToken", curJwtToken))

	//构造负载数据
	deviceInfo := &Auth.DeviceInfo{
		Username:     username,
		ConnectionId: "",
		DeviceId:     eDeviceID,
		DeviceIndex:  int32(1), //从1开始
		IsMaster:     true,
		Os:           curOs,
		LogonAt:      curLogonAt,
	}

	rsp.OnlineDevices = append(rsp.OnlineDevices, deviceInfo)

	//响应客户端

	msg.SetCode(200) //状态码

	data, _ := proto.Marshal(rsp)

	nc.logger.Info("GetAllDevices Succeed",
		zap.String("Username:", username),
		zap.Int("length", len(data)),
	)

	msg.FillBody(data) //网络包的body，承载真正的业务数据

	//处理完成，向dispatcher发送
	topic := msg.GetSource() + ".Frontend"
	rawData, _ := json.Marshal(msg)
	if err := nc.Producer.Public(topic, rawData); err == nil {
		nc.logger.Info("Succeed to send message to ProduceChannel", zap.String("topic", topic))
	} else {
		nc.logger.Error("Failed to send message to ProduceChannel", zap.Error(err))
	}
	_ = err
	return nil
}

/*
2-6 添加从设备
1. 此接口是主设备批准从设备的授权
*/
func (nc *NsqClient) HandleAddSlaveDevice(msg *models.Message) error {
	var err error
	errorCode := 200 //错误码， 200是正常，其它是错误

	username := msg.GetUserName() //主设备用户账号
	deviceID := msg.GetDeviceID() //主设备设备id

	nc.logger.Info("HandleAddSlaveDevice start...",
		zap.String("username", username),
		zap.String("DeviceId", deviceID))

	redisConn := nc.redisPool.Get()
	defer redisConn.Close()

	//打开msg里的负载， 获取里面的参赛
	body := msg.GetContent()

	//解包body
	var req Auth.AddSlaveDeviceReq
	if err := proto.Unmarshal(body, &req); err != nil {
		nc.logger.Error("Protobuf Unmarshal Error", zap.Error(err))
		return err
	}
	nc.logger.Debug("AddSlaveDeviceReq",
		zap.Int32("authCode", req.GetAuthCode()),
	)

	tempKey := fmt.Sprintf("SlaveTemporaryIdentity:%d", req.GetAuthCode())

	slaveDeviceID, err := redis.String(redisConn.Do("GET", tempKey))
	if err != nil {
		nc.logger.Error("Failed to GET slaveDeviceID", zap.Error(err))
		errorCode = LMCError.RedisError
		goto COMPLETE

	} else {

		//查询出主设备的手机号
		userKey := fmt.Sprintf("userData:%s", username)
		mobile, _ := redis.String(redisConn.Do("HGET", userKey, "Mobile"))
		if mobile == "" {
			nc.logger.Error("获取手机失败", zap.Error(err))
			errorCode = LMCError.RedisError
			goto COMPLETE
		}

		//生成 smscode
		key := fmt.Sprintf("smscode:%s", mobile)

		_, err = redisConn.Do("DEL", key) //删除key

		//TODO 调用短信接口发送  暂时固定为123456

		err = redisConn.Send("SET", key, "123456") //增加key

		err = redisConn.Send("EXPIRE", key, LMCommon.SMSEXPIRE) //设置有效期

		//一次性写入到Redis
		if err := redisConn.Flush(); err != nil {
			nc.logger.Error("写入redis失败", zap.Error(err))
			errorCode = LMCError.RedisError
			goto COMPLETE
		}
		nc.logger.Debug("GenerateSmsCode, 写入redis成功")

		//向从设备下发消息
		{
			targetMsg := &models.Message{}

			targetMsg.UpdateID()
			//构建消息路由, 第一个参数是要处理的业务类型，后端服务器处理完成后，需要用此来拼接topic: {businessTypeName.Frontend}
			targetMsg.BuildRouter("Auth", "", "Auth.Frontend")

			targetMsg.SetJwtToken(fmt.Sprintf("%d", req.GetAuthCode()))
			targetMsg.SetUserName(username)
			targetMsg.SetDeviceID(slaveDeviceID)
			// kickMsg.SetTaskID(uint32(taskId))
			targetMsg.SetBusinessTypeName("Auth")
			targetMsg.SetBusinessType(uint32(2))
			targetMsg.SetBusinessSubType(uint32(9)) //SlaveDeviceAuthEvent = 9

			targetMsg.BuildHeader("Dispatcher", time.Now().UnixNano()/1e6) //毫秒

			//构造负载数据
			resp := &Auth.SlaveDeviceAuthEventRsp{
				Username: username,
				Smscode:  "123456", //暂时
			}

			targetData, _ := proto.Marshal(resp)
			targetMsg.FillBody(targetData) //网络包的body，承载真正的业务数据

			targetMsg.SetCode(200) //成功的状态码
			//构建数据完成，向dispatcher发送
			topic := "Auth.Frontend"
			rawData, _ := json.Marshal(targetMsg)
			if err := nc.Producer.Public(topic, rawData); err == nil {
				nc.logger.Info("message succeed send to ProduceChannel", zap.String("topic", topic))
			} else {
				nc.logger.Error(" failed to send message to ProduceChannel", zap.Error(err))
			}
		}

		nc.logger.Info("AddSlaveDevice Succeed",
			zap.String("deviceID:", slaveDeviceID),
			zap.Int32("Code", req.GetAuthCode()))

	}

COMPLETE:
	msg.SetCode(int32(errorCode)) //状态码
	if errorCode == 200 {
		//这里不需要发送body，直接向主设备发送200即可
		msg.FillBody(nil)
	} else {
		// errorMsg := LMCError.ErrorMsg(errorCode)
		msg.FillBody(nil)
	}

	//处理完成，向dispatcher发送
	topic := msg.GetSource() + ".Frontend"
	rawData, _ := json.Marshal(msg)
	if err := nc.Producer.Public(topic, rawData); err == nil {
		nc.logger.Info("Succeed send AddSlaveDevice message to ProduceChannel", zap.String("topic", topic))
	} else {
		nc.logger.Error("Failed to send AddSlaveDevice message to ProduceChannel", zap.Error(err))
	}
	_ = err
	return nil
}

/*
2-7 从设备申请授权码
1. 此接口是从设备发起，从设备还没有JWT令牌，因此不能拦截
*/
func (nc *NsqClient) HandleAuthorizeCode(msg *models.Message) error {
	// 	var err error
	// 	rsp := &Auth.AuthorizeCodeRsp{}
	// 	errorCode := 200

	// 	nc.logger.Info("HandleAuthorizeCode start...", zap.String("DeviceId", msg.GetDeviceID()))

	// 	deviceID := msg.GetDeviceID()
	// 	redisConn := nc.redisPool.Get()
	// 	defer redisConn.Close()

	// 	//打开msg里的负载， 获取里面的参赛
	// 	body := msg.GetContent()
	// 	//解包body
	// 	var req Auth.AuthorizeCodeReq
	// 	if err := proto.Unmarshal(body, &req); err != nil {
	// 		nc.logger.Error("Protobuf Unmarshal Error", zap.Error(err))
	// 		errorCode = LMCError.ProtobufUnmarshalError
	// 		goto COMPLETE
	// 	} else {
	// 		nc.logger.Debug("AuthorizeCodeReq",
	// 			zap.String("AppKey", req.GetAppKey()),
	// 			zap.Int32("ClientType", int32(req.GetClientType())),
	// 			zap.String("Os", req.GetOs()),
	// 			zap.String("ProtocolVersion", req.GetProtocolVersion()),
	// 			zap.String("SdkVersion", req.GetSdkVersion()),
	// 		)

	// 		//生成一个 100000 - 999999 之间的随机数
	// 		tId := randtool.RangeRand(100000, 999999)

	// 		tempKey := fmt.Sprintf("SlaveTemporaryIdentity:%d", tId)

	// 		_, err = redisConn.Do("SET", tempKey, deviceID)
	// 		if err != nil {
	// 			nc.logger.Error("Failed to SET ", zap.Error(err))
	// 			errorCode = LMCError.RedisError
	// 			goto COMPLETE
	// 		}
	// 		nc.logger.Debug("HandleAuthorizeCode",
	// 			zap.String("tempKey", tempKey),
	// 			zap.Int64("tId", tId),
	// 		)
	// 		//有效期
	// 		_, err = redisConn.Do("EXPIRE", tempKey, LMCommon.SMSEXPIRE)
	// 		if err != nil {
	// 			nc.logger.Error("Failed to EXPIRE ", zap.Error(err))
	// 			errorCode = LMCError.RedisError
	// 			goto COMPLETE
	// 		}

	// 		rsp.Code = fmt.Sprintf("%d", tId)

	// 		nc.logger.Info("AuthorizeCode Succeed",
	// 			zap.String("deviceID:", deviceID),
	// 			zap.String("Code", fmt.Sprintf("%d", tId)))

	// 	}

	// COMPLETE:
	// 	msg.SetJwtToken("*")          //为了欺骗map
	// 	msg.SetCode(int32(errorCode)) //状态码
	// 	if errorCode == 200 {
	// 		data, _ = proto.Marshal(rsp)
	// 		msg.FillBody(data)
	// 	} else {
	// 		errorMsg := LMCError.ErrorMsg(errorCode)
	// 		msg.FillBody(nil)
	// 	}

	// 	//处理完成，向dispatcher发送
	// 	topic := msg.GetSource() + ".Frontend"
	// 	rawData, _ := json.Marshal(msg)
	// 	if err := nc.Producer.Public(topic, rawData); err == nil {
	// 		nc.logger.Info("Succeed to send AuthorizeCode message to ProduceChannel", zap.String("topic", topic))
	// 	} else {
	// 		nc.logger.Error("Failed to send AuthorizeCode message to ProduceChannel", zap.Error(err))
	// 	}
	// 	_ = err
	return nil
}
