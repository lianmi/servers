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
package nsqBackend

import (
	"encoding/json"
	"fmt"
	// "strings"
	"time"

	"net/http"

	"github.com/pkg/errors"

	"github.com/golang/protobuf/proto"
	"github.com/gomodule/redigo/redis"
	Auth "github.com/lianmi/servers/api/proto/auth"
	"github.com/lianmi/servers/internal/pkg/models"
	"github.com/lianmi/servers/util/randtool"
	"go.uber.org/zap"
)

/*
2-2 登出
1. 主设备登出，需要删除从设备一切数据，踢出从设备
2. 从设备登出，只删除自己的数据，并刷新此用户的在线设备列表
*/
func (kc *NsqClient) HandleSignOut(msg *models.Message) error {
	var err error
	// var errorMsg string

	//TODO 将此设备从在线列表里删除，然后更新对应用户的在线列表。
	redisConn := kc.redisPool.Get()
	defer redisConn.Close()

	username := msg.GetUserName()
	// token := msg.GetJwtToken()
	deviceID := msg.GetDeviceID()
	kc.logger.Info("HandleSignOut start...",
		zap.String("username", username),
		zap.String("DeviceId", deviceID))

	//取出当前旧的设备的os， clientType， logonAt
	curDeviceHashKey := fmt.Sprintf("devices:%s:%s", username, deviceID)
	isMaster, _ := redis.Bool(redisConn.Do("HGET", curDeviceHashKey, "ismaster"))
	curOs, _ := redis.String(redisConn.Do("HGET", curDeviceHashKey, "os"))
	curClientType, _ := redis.Int(redisConn.Do("HGET", curDeviceHashKey, "clientType"))
	curLogonAt, _ := redis.Uint64(redisConn.Do("HGET", curDeviceHashKey, "logonAt"))

	kc.logger.Debug("SignOut",
		zap.Bool("isMaster", isMaster),
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
			deviceKey := fmt.Sprintf("DeviceJwtToken:%s", eDeviceID)
			jwtToken, _ := redis.String(redisConn.Do("GET", deviceKey))
			kc.logger.Debug("Redis GET ", zap.String("deviceKey", deviceKey), zap.String("jwtToken", jwtToken))

			businessType := 2
			businessSubType := 5 //KickedEvent

			businessTypeName := "Auth"
			nsqTopic := businessTypeName + ".Frontend"
			backendService := businessTypeName + "Service"

			//向当前主设备及从设备发出踢下线
			kickMsg := &models.Message{}
			now := time.Now().UnixNano() / 1e6 //毫秒
			kickMsg.UpdateID()
			//构建消息路由, 第一个参数是要处理的业务类型，后端服务器处理完成后，需要用此来拼接topic: {businessTypeName.Frontend}
			kickMsg.BuildRouter(businessTypeName, "", nsqTopic)

			kickMsg.SetJwtToken(jwtToken)
			kickMsg.SetUserName(username)
			kickMsg.SetDeviceID(string(eDeviceID))
			// kickMsg.SetTaskID(uint32(taskId))
			kickMsg.SetBusinessTypeName(businessTypeName)
			kickMsg.SetBusinessType(uint32(businessType))
			kickMsg.SetBusinessSubType(uint32(businessSubType))

			kickMsg.BuildHeader(backendService, now)

			//构造负载数据
			resp := &Auth.KickedEventRsp{
				ClientType: 0,
				Reason:     Auth.KickReason_SamePlatformKick,
				TimeTag:    uint64(time.Now().UnixNano() / 1e6), //毫秒
			}
			data, _ := proto.Marshal(resp)
			kickMsg.FillBody(data) //网络包的body，承载真正的业务数据

			kickMsg.SetCode(200) //成功的状态码

			//构建数据完成，向dispatcher发送
			topic := "Auth.Frontend"
			rawData, _ := json.Marshal(kickMsg)
			if err := kc.Producer.Public(topic, rawData); err == nil {
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
		deviceKey := fmt.Sprintf("DeviceJwtToken:%s", deviceID)
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
			if deviceID == eDeviceID {
				continue
			}

			targetMsg := &models.Message{}
			curDeviceKey := fmt.Sprintf("DeviceJwtToken:%s", eDeviceID)
			curJwtToken, _ := redis.String(redisConn.Do("GET", curDeviceKey))
			kc.logger.Debug("Redis GET ", zap.String("curDeviceKey", curDeviceKey), zap.String("curJwtToken", curJwtToken))

			now := time.Now().UnixNano() / 1e6 //毫秒
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
			rawData, _ := json.Marshal(targetMsg)
			if err := kc.Producer.Public(topic, rawData); err == nil {
				kc.logger.Info("Succeed to send message to ProduceChannel", zap.String("topic", topic))
			} else {
				kc.logger.Error("Failed to send message to ProduceChannel", zap.Error(err))
			}

		}

	}

	//向dispatcher发送
	topic := msg.GetSource() + ".Frontend"
	msg.SetCode(200) //成功的状态码
	rawData, _ := json.Marshal(msg)
	if err := kc.Producer.Public(topic, rawData); err == nil {
		kc.logger.Info("Succeed to send message to ProduceChannel", zap.String("topic", topic))
	} else {
		kc.logger.Error("Failed to send message to ProduceChannel", zap.Error(err))
	}

	_ = err

	kc.logger.Debug("登出成功", zap.Bool("isMaster", isMaster),
		zap.String("username", username),
		zap.String("deviceID", deviceID),
		zap.String("curOs", curOs),
		zap.Int("curClientType", curClientType),
		zap.Uint64("curLogonAt", curLogonAt))

	return nil
}

/*
2-4 踢出其它在线设备 Kick
1. 主设备才能踢出从设备, 被踢的从设备收到被踢下线事件
2. 从设备被踢后，只删除自己的数据，并发出多端登录状态变化事件
*/
func (kc *NsqClient) HandleKick(msg *models.Message) error {
	var err error
	kickRsp := &Auth.KickRsp{}
	errorCode := 200 //错误码， 200是正常，其它是错误
	var errorMsg string
	var data []byte

	//TODO 将此设备从在线列表里删除，然后更新对应用户的在线列表。
	redisConn := kc.redisPool.Get()
	defer redisConn.Close()

	username := msg.GetUserName()
	// token := msg.GetJwtToken()
	deviceID := msg.GetDeviceID()

	kc.logger.Info("HandleKick start...",
		zap.String("username", username),
		zap.String("DeviceId", deviceID))

	//取出当前设备的os， clientType， logonAt
	curDeviceHashKey := fmt.Sprintf("devices:%s:%s", username, deviceID)
	isMaster, _ := redis.Bool(redisConn.Do("HGET", curDeviceHashKey, "ismaster"))
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
			errorCode = http.StatusInternalServerError //500
			errorMsg = "Protobuf Unmarshal Error"
			goto COMPLETE
		} else {
			deviceiIDs = kickReq.GetDeviceIds()

			for _, did := range deviceiIDs {
				kc.logger.Debug("To be kick ...", zap.String("DeviceId", did))
				//删除token
				deviceKey := fmt.Sprintf("DeviceJwtToken:%s", did)
				_, err = redisConn.Do("DEL", deviceKey)

				//删除有序集合里的元素
				//移除单个元素 ZREM deviceListKey {设备id}
				_, err = redisConn.Do("ZREM", deviceListKey, did)

				//删除在线设备哈希表
				deviceHashKey := fmt.Sprintf("devices:%s:%s", username, did)
				_, err = redisConn.Do("DEL", deviceHashKey)

				//向被踢的从设备发出被踢下线的通知
				{
					beKickedMsg := &models.Message{}

					beKickedMsg.UpdateID()
					//构建消息路由, 第一个参数是要处理的业务类型，后端服务器处理完成后，需要用此来拼接topic: {businessTypeName.Frontend}
					beKickedMsg.BuildRouter("Auth", "", "Auth.Frontend")

					beKickedMsg.SetJwtToken("kicked")
					beKickedMsg.SetUserName(username)
					beKickedMsg.SetDeviceID(did)
					// kickMsg.SetTaskID(uint32(taskId))
					beKickedMsg.SetBusinessTypeName("Auth")
					beKickedMsg.SetBusinessType(uint32(2))
					beKickedMsg.SetBusinessSubType(uint32(5)) // KickedEvent = 5

					beKickedMsg.BuildHeader("AuthService", time.Now().UnixNano()/1e6) //毫秒

					//构造负载数据
					kickedEventRsp := &Auth.KickedEventRsp{
						ClientType: Auth.ClientType(curClientType),
						Reason:     Auth.KickReason_OtherPlatformKick, //被主设备踢下线
					}
					data, _ := proto.Marshal(kickedEventRsp)
					beKickedMsg.FillBody(data) //网络包的body，承载真正的业务数据
					beKickedMsg.SetCode(200)   //成功的状态码
					//构建数据完成，向dispatcher发送
					topic := "Auth.Frontend"
					rawData, _ := json.Marshal(beKickedMsg)
					if err := kc.Producer.Public(topic, rawData); err == nil {
						kc.logger.Info("Succeed send message to ProduceChannel", zap.String("topic", topic))
					} else {
						kc.logger.Error("Failed to send message to ProduceChannel", zap.Error(err))
					}

				}

				//多端登录状态变化事件
				//向其它端发送此从设备离线的事件
				deviceIDSliceNew, _ := redis.Strings(redisConn.Do("ZRANGEBYSCORE", deviceListKey, "-inf", "+inf"))
				//查询出当前在线所有主从设备
				for _, eDeviceID := range deviceIDSliceNew {
					if deviceID == eDeviceID {
						continue
					}
					targetMsg := &models.Message{}
					curDeviceKey := fmt.Sprintf("DeviceJwtToken:%s", eDeviceID)
					curJwtToken, _ := redis.String(redisConn.Do("GET", curDeviceKey))
					kc.logger.Debug("Redis GET ", zap.String("curDeviceKey", curDeviceKey), zap.String("curJwtToken", curJwtToken))

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

					targetMsg.BuildHeader("AuthService", time.Now().UnixNano()/1e6) //毫秒

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
					rawData, _ := json.Marshal(targetMsg)
					if err := kc.Producer.Public(topic, rawData); err == nil {
						kc.logger.Info("message succeed send to ProduceChannel", zap.String("topic", topic))
					} else {
						kc.logger.Error(" failed to send message to ProduceChannel", zap.Error(err))
					}

				}

			}

			//响应客户端的OnKick
			kickRsp.DeviceIds = deviceiIDs

			kc.logger.Info("Kick Succeed",
				zap.String("Username:", username),
				zap.Int("length", len(data)))

		}

	} else {
		//从设备无权踢出其它设备
		err = errors.Wrapf(err, "Slave service no right to kick other device")
		errorMsg = "Slave service no right to kick other device"
	}

COMPLETE:
	msg.SetCode(int32(errorCode)) //状态码
	if errorCode == 200 {
		data, _ = proto.Marshal(kickRsp)
		msg.FillBody(data) //网络包的body，承载真正的业务数据
	} else {
		msg.SetErrorMsg([]byte(errorMsg)) //错误提示
		msg.FillBody(nil)
	}

	//处理完成，向dispatcher发送
	topic := msg.GetSource() + ".Frontend"
	rawData, _ := json.Marshal(msg)
	if err := kc.Producer.Public(topic, rawData); err == nil {
		kc.logger.Info("KickRsp message succeed send to ProduceChannel", zap.String("topic", topic))
	} else {
		kc.logger.Error("Failed to send KickRsp message to ProduceChannel", zap.Error(err))
	}

	_ = err
	return nil
}

/*
2-9 获取所有主从设备
*/
func (kc *NsqClient) HandleGetAllDevices(msg *models.Message) error {
	var err error

	redisConn := kc.redisPool.Get()
	defer redisConn.Close()

	username := msg.GetUserName()
	// token := msg.GetJwtToken()
	deviceID := msg.GetDeviceID()

	kc.logger.Info("HandleGetAllDevices start...",
		zap.String("username", username),
		zap.String("DeviceId", deviceID))

	//取出当前设备的os， clientType， logonAt
	curDeviceHashKey := fmt.Sprintf("devices:%s:%s", username, deviceID)
	isMaster, _ := redis.Bool(redisConn.Do("HGET", curDeviceHashKey, "ismaster"))
	curOs, _ := redis.String(redisConn.Do("HGET", curDeviceHashKey, "os"))
	curClientType, _ := redis.Int(redisConn.Do("HGET", curDeviceHashKey, "clientType"))
	curLogonAt, _ := redis.Uint64(redisConn.Do("HGET", curDeviceHashKey, "logonAt"))

	kc.logger.Debug("GetAllDevices",
		zap.Bool("isMaster", isMaster),
		zap.String("username", username),
		zap.String("deviceID", deviceID),
		zap.String("curOs", curOs),
		zap.Int("curClientType", curClientType),
		zap.Uint64("curLogonAt", curLogonAt))

	deviceListKey := fmt.Sprintf("devices:%s", username)
	rsp := &Auth.GetAllDevicesRsp{}
	rsp.OnlineDevices = make([]*Auth.DeviceInfo, 0)
	rsp.OfflineDevices = make([]*Auth.DeviceInfo, 0)

	deviceIDSliceNew, _ := redis.Strings(redisConn.Do("ZRANGEBYSCORE", deviceListKey, "-inf", "+inf"))
	//查询出当前在线所有主从设备
	for index, eDeviceID := range deviceIDSliceNew {

		curDeviceKey := fmt.Sprintf("DeviceJwtToken:%s", eDeviceID)
		curJwtToken, _ := redis.String(redisConn.Do("GET", curDeviceKey))
		kc.logger.Debug("Redis GET ", zap.String("curDeviceKey", curDeviceKey), zap.String("curJwtToken", curJwtToken))

		curIsMaster := isMaster
		if eDeviceID != deviceID {
			curIsMaster = false
		}
		//构造负载数据
		deviceInfo := &Auth.DeviceInfo{
			Username:     username,
			ConnectionId: "",
			DeviceId:     eDeviceID,
			DeviceIndex:  int32(index + 1), //从1开始
			IsMaster:     curIsMaster,
			Os:           curOs,
			ClientType:   Auth.ClientType(curClientType),
			LogonAt:      curLogonAt,
		}

		rsp.OnlineDevices = append(rsp.OnlineDevices, deviceInfo)

	}

	//响应客户端

	msg.SetCode(200) //状态码

	data, _ := proto.Marshal(rsp)

	kc.logger.Info("GetAllDevices Succeed",
		zap.String("Username:", username),
		zap.Int("length", len(data)),
	)

	msg.FillBody(data) //网络包的body，承载真正的业务数据

	//处理完成，向dispatcher发送
	topic := msg.GetSource() + ".Frontend"
	rawData, _ := json.Marshal(msg)
	if err := kc.Producer.Public(topic, rawData); err == nil {
		kc.logger.Info("Succeed to send message to ProduceChannel", zap.String("topic", topic))
	} else {
		kc.logger.Error("Failed to send message to ProduceChannel", zap.Error(err))
	}
	_ = err
	return nil
}

/*
2-6 添加从设备
1. 此接口是主设备批准从设备的授权
*/
func (kc *NsqClient) HandleAddSlaveDevice(msg *models.Message) error {
	var err error
	errorCode := 200 //错误码， 200是正常，其它是错误
	var errorMsg string

	username := msg.GetUserName() //主设备用户账号
	deviceID := msg.GetDeviceID() //主设备设备id

	kc.logger.Info("HandleAddSlaveDevice start...",
		zap.String("username", username),
		zap.String("DeviceId", deviceID))

	redisConn := kc.redisPool.Get()
	defer redisConn.Close()

	//打开msg里的负载， 获取里面的参赛
	body := msg.GetContent()

	//解包body
	var req Auth.AddSlaveDeviceReq
	if err := proto.Unmarshal(body, &req); err != nil {
		kc.logger.Error("Protobuf Unmarshal Error", zap.Error(err))
		return err
	}
	kc.logger.Debug("AddSlaveDeviceReq",
		zap.Int32("authCode", req.GetAuthCode()),
	)

	tempKey := fmt.Sprintf("SlaveTemporaryIdentity:%d", req.GetAuthCode())

	slaveDeviceID, err := redis.String(redisConn.Do("GET", tempKey))
	if err != nil {
		kc.logger.Error("Failed to GET slaveDeviceID", zap.Error(err))
		errorCode = http.StatusInternalServerError //500
		errorMsg = fmt.Sprintf("Protobuf Unmarshal Error: %s", err.Error())
		goto COMPLETE

	} else {

		//查询出主设备的手机号
		userKey := fmt.Sprintf("userData:%s", username)
		mobile, _ := redis.String(redisConn.Do("HGET", userKey, "Mobile"))
		if mobile == "" {
			kc.logger.Error("获取手机失败", zap.Error(err))
			errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
			errorMsg = fmt.Sprintf("Get Mobile from redis error")
			goto COMPLETE
		}

		//生成 smscode
		key := fmt.Sprintf("smscode:%s", mobile)

		_, err = redisConn.Do("DEL", key) //删除key

		//TODO 调用短信接口发送  暂时固定为123456

		err = redisConn.Send("SET", key, "123456") //增加key

		err = redisConn.Send("EXPIRE", key, 300) //设置有效期为300秒

		//一次性写入到Redis
		if err := redisConn.Flush(); err != nil {
			kc.logger.Error("写入redis失败", zap.Error(err))
			errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
			errorMsg = fmt.Sprintf("写入redis失败: %s", err.Error())
			goto COMPLETE
		}
		kc.logger.Debug("GenerateSmsCode, 写入redis成功")

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

			targetMsg.BuildHeader("AuthService", time.Now().UnixNano()/1e6) //毫秒

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
			if err := kc.Producer.Public(topic, rawData); err == nil {
				kc.logger.Info("message succeed send to ProduceChannel", zap.String("topic", topic))
			} else {
				kc.logger.Error(" failed to send message to ProduceChannel", zap.Error(err))
			}
		}

		kc.logger.Info("AddSlaveDevice Succeed",
			zap.String("deviceID:", slaveDeviceID),
			zap.Int32("Code", req.GetAuthCode()))

	}

COMPLETE:
	msg.SetCode(int32(errorCode)) //状态码
	if errorCode == 200 {
		//这里不需要发送body，直接向主设备发送200即可
		msg.FillBody(nil)
	} else {
		msg.SetErrorMsg([]byte(errorMsg)) //错误提示
		msg.FillBody(nil)
	}

	//处理完成，向dispatcher发送
	topic := msg.GetSource() + ".Frontend"
	rawData, _ := json.Marshal(msg)
	if err := kc.Producer.Public(topic, rawData); err == nil {
		kc.logger.Info("Succeed send AddSlaveDevice message to ProduceChannel", zap.String("topic", topic))
	} else {
		kc.logger.Error("Failed to send AddSlaveDevice message to ProduceChannel", zap.Error(err))
	}
	_ = err
	return nil
}

/*
2-7 从设备申请授权码
1. 此接口是从设备发起，从设备还没有JWT令牌，因此不能拦截
*/
func (kc *NsqClient) HandleAuthorizeCode(msg *models.Message) error {
	var err error
	rsp := &Auth.AuthorizeCodeRsp{}
	errorCode := 200
	var errorMsg string
	var data []byte

	kc.logger.Info("HandleAuthorizeCode start...", zap.String("DeviceId", msg.GetDeviceID()))

	deviceID := msg.GetDeviceID()
	redisConn := kc.redisPool.Get()
	defer redisConn.Close()

	//打开msg里的负载， 获取里面的参赛
	body := msg.GetContent()
	//解包body
	var req Auth.AuthorizeCodeReq
	if err := proto.Unmarshal(body, &req); err != nil {
		kc.logger.Error("Protobuf Unmarshal Error", zap.Error(err))
		errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
		errorMsg = "Protobuf Unmarshal Error"
		goto COMPLETE
	} else {
		kc.logger.Debug("AuthorizeCodeReq",
			zap.String("AppKey", req.GetAppKey()),
			zap.Int32("ClientType", int32(req.GetClientType())),
			zap.String("Os", req.GetOs()),
			zap.String("ProtocolVersion", req.GetProtocolVersion()),
			zap.String("SdkVersion", req.GetSdkVersion()),
		)

		//生成一个 100000 - 999999 之间的随机数
		tId := randtool.RangeRand(100000, 999999)

		tempKey := fmt.Sprintf("SlaveTemporaryIdentity:%d", tId)

		_, err = redisConn.Do("SET", tempKey, deviceID)
		if err != nil {
			kc.logger.Error("Failed to SET ", zap.Error(err))
			errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
			errorMsg = fmt.Sprintf("Failed to SET: %s", err.Error())
			goto COMPLETE
		}
		kc.logger.Debug("HandleAuthorizeCode",
			zap.String("tempKey", tempKey),
			zap.Int64("tId", tId),
		)
		//120秒的有效期
		_, err = redisConn.Do("EXPIRE", tempKey, 120)
		if err != nil {
			kc.logger.Error("Failed to EXPIRE ", zap.Error(err))
			errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
			errorMsg = fmt.Sprintf("Failed to EXPIRE: %s", err.Error())
			goto COMPLETE
		}

		rsp.Code = fmt.Sprintf("%d", tId)

		kc.logger.Info("AuthorizeCode Succeed",
			zap.String("deviceID:", deviceID),
			zap.String("Code", fmt.Sprintf("%d", tId)))

	}

COMPLETE:
	msg.SetJwtToken("*")          //为了欺骗map
	msg.SetCode(int32(errorCode)) //状态码
	if errorCode == 200 {
		data, _ = proto.Marshal(rsp)
		msg.FillBody(data)
	} else {
		msg.SetErrorMsg([]byte(errorMsg)) //错误提示
		msg.FillBody(nil)
	}

	//处理完成，向dispatcher发送
	topic := msg.GetSource() + ".Frontend"
	rawData, _ := json.Marshal(msg)
	if err := kc.Producer.Public(topic, rawData); err == nil {
		kc.logger.Info("Succeed to send AuthorizeCode message to ProduceChannel", zap.String("topic", topic))
	} else {
		kc.logger.Error("Failed to send AuthorizeCode message to ProduceChannel", zap.Error(err))
	}
	_ = err
	return nil
}
