package kafkaBackend

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gomodule/redigo/redis"
	Global "github.com/lianmi/servers/api/proto/global"
	Msg "github.com/lianmi/servers/api/proto/msg"
	Team "github.com/lianmi/servers/api/proto/team"
	"github.com/lianmi/servers/internal/pkg/models"
	"github.com/lianmi/servers/util/sts"
	"google.golang.org/protobuf/proto"

	"go.uber.org/zap"
	simpleJson "github.com/bitly/go-simplejson"
)

var (
	TempKey    = "LTAI4G3o4sECdSBsD7rGLmCs"
	TempSecret = "0XmB9tLOBLhmjIcM6CrBv2PHfnoDa8"
	RoleAcs    = "acs:ram::1230446857465673:role/ipfsuploader"
)

/*
1. 先从redis里读取 哈希表 userData:{username} 里的元素，如果无法读取，则直接从MySQL里读取
2. 注意，更新资料后，也需要同步更新 哈希表 userData:{username}
哈希表 userData:{username} 的元素就是User的各个字段
*/
func (kc *KafkaClient) HandleRecvMsg(msg *models.Message) error {
	var err error
	var toUser, teamID string
	errorCode := 200
	var errorMsg string
	rsp := &Msg.SendMsgRsp{}

	var newSeq uint64
	var data []byte

	redisConn := kc.redisPool.Get()
	defer redisConn.Close()

	username := msg.GetUserName()
	// token := msg.GetJwtToken()
	deviceID := msg.GetDeviceID()

	kc.logger.Info("HandleRecvMsg start...",
		zap.String("username", username),
		zap.String("DeviceId", deviceID))

	//取出当前设备的os， clientType， logonAt
	curDeviceHashKey := fmt.Sprintf("devices:%s:%s", username, deviceID)
	isMaster, _ := redis.Bool(redisConn.Do("HGET", curDeviceHashKey, "ismaster"))
	curOs, _ := redis.String(redisConn.Do("HGET", curDeviceHashKey, "os"))
	curClientType, _ := redis.Int(redisConn.Do("HGET", curDeviceHashKey, "clientType"))
	curLogonAt, _ := redis.Uint64(redisConn.Do("HGET", curDeviceHashKey, "logonAt"))

	kc.logger.Debug("RecvMsg",
		zap.Bool("isMaster", isMaster),
		zap.String("username", username),
		zap.String("deviceID", deviceID),
		zap.String("curOs", curOs),
		zap.Int("curClientType", curClientType),
		zap.Uint64("curLogonAt", curLogonAt))

	//打开msg里的负载， 获取请求参数
	body := msg.GetContent()
	//解包body
	var req Msg.SendMsgReq
	if err := proto.Unmarshal(body, &req); err != nil {
		errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
		errorMsg = fmt.Sprintf("Protobuf Unmarshal Error: %s", err.Error())
		kc.logger.Error("Protobuf Unmarshal Error", zap.Error(err))
		goto COMPLETE

	} else {
		kc.logger.Debug("RecvMsg  payload",
			zap.Int32("Scene", int32(req.GetScene())),
			zap.Int32("Type", int32(req.GetType())),
			zap.String("To", req.GetTo()),
			zap.String("Uuid", req.GetUuid()),
			zap.Uint64("SendAt", req.GetSendAt()),
		)

		//根据场景判断消息是个人消息、群聊消息
		switch req.GetScene() {
		case Msg.MessageScene_MsgScene_C2C: //个人消息
			toUser = req.GetTo()
			kc.logger.Debug("MessageScene_MsgScene_C2C", zap.String("toUser", req.GetTo()))
			//判断toUser的合法性以及是否封禁等
			userData := new(models.User)
			userKey := fmt.Sprintf("userData:%s", toUser)
			if result, err := redis.Values(redisConn.Do("HGETALL", userKey)); err == nil {
				if err := redis.ScanStruct(result, userData); err != nil {

					kc.logger.Error("错误：ScanStruct", zap.Error(err))
					errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
					errorMsg = fmt.Sprintf("ScanStruct Error[Username=%s]", toUser)
					goto COMPLETE

				}
			}

			if userData.State == 2 {
				kc.logger.Warn("此用户已被封号", zap.String("toUser", req.GetTo()))
				errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
				errorMsg = fmt.Sprintf("User is blocked[Username=%s]", toUser)
				goto COMPLETE
			}

			//查出接收人对此用户消息接收的设定，黑名单，屏蔽等
			if reply, err := redisConn.Do("ZRANK", fmt.Sprintf("BlackList:%s", toUser), username); err == nil {
				if reply != nil {
					kc.logger.Warn("用户已被对方拉黑", zap.String("toUser", req.GetTo()))
					errorCode = http.StatusNotFound //错误码， 200是正常，其它是错误
					errorMsg = fmt.Sprintf("User is blocked[Username=%s]", toUser)
					goto COMPLETE
				}
			}

			//构造转发消息数据
			if newSeq, err = redis.Uint64(redisConn.Do("INCR", fmt.Sprintf("userSeq:%s", toUser))); err != nil {
				kc.logger.Error("redisConn INCR userSeq Error", zap.Error(err))
				errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
				errorMsg = "INCR Error"
				goto COMPLETE
			}

			eRsp := &Msg.RecvMsgEventRsp{
				Scene:        req.GetScene(), //传输场景
				Type:         req.GetType(),  //消息类型
				Body:         req.GetBody(),  //不拆包，直接透传body给接收者
				FromDeviceId: deviceID,
				ServerMsgId:  msg.GetID(),                        //服务器分配的消息ID
				Seq:          newSeq,                             //消息序号，单个会话内自然递增, 这里是对targetUsername这个用户的通知序号
				Uuid:         fmt.Sprintf("%d", msg.GetTaskID()), //客户端分配的消息ID，SDK生成的消息id，这里返回TaskID
				Time:         uint64(time.Now().Unix()),
			}

			//转发消息
			go kc.SendMsgToUser(eRsp, username, deviceID, toUser)

		case Msg.MessageScene_MsgScene_C2G: //群聊消息
			teamID = req.GetTo()
			kc.logger.Debug("MessageScene_MsgScene_C2G", zap.String("toTeamID", req.GetTo()))
			//判断toTeamID的合法性以及是否封禁等

			//获取到群信息
			key := fmt.Sprintf("TeamInfo:%s", teamID)
			teamInfo := new(models.Team)
			if result, err := redis.Values(redisConn.Do("HGETALL", key)); err == nil {
				if err := redis.ScanStruct(result, teamInfo); err != nil {
					kc.logger.Error("错误：ScanStruct", zap.Error(err))
					errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
					errorMsg = fmt.Sprintf("Team is not exists[teamID=%s]", teamID)
					goto COMPLETE
				}
			}

			//此群是否是正常的
			if teamInfo.Status != 2 {
				kc.logger.Warn("Team status is not normal", zap.Int("Status", teamInfo.Status))
				errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
				errorMsg = fmt.Sprintf("Team status is not normal")
				goto COMPLETE
			}

			//for..range 群成员
			teamMembers, _ := redis.Strings(redisConn.Do("ZRANGEBYSCORE", fmt.Sprintf("TeamUsers:%s", teamID), "-inf", "+inf"))
			for _, teamMember := range teamMembers {
				toUser = teamMember

				//判断toUser的合法性以及是否封禁等
				userData := new(models.User)
				userKey := fmt.Sprintf("userData:%s", toUser)
				if result, err := redis.Values(redisConn.Do("HGETALL", userKey)); err == nil {
					if err := redis.ScanStruct(result, userData); err != nil {
						kc.logger.Error("错误：ScanStruct", zap.Error(err))
						continue
					}
				}

				if userData.State == 2 {
					kc.logger.Warn("此用户已被封号", zap.String("toUser", req.GetTo()))
					continue
				}

				//查出此用户对此群消息接收的设定，如果允许接收，就发
				toUserKey := fmt.Sprintf("TeamUser:%s:%s", teamID, toUser)
				notifyType, _ := redis.Int(redisConn.Do("HGET", toUserKey, "NotifyType"))
				switch notifyType {
				case 1: //群全部消息提醒
				case 2: //管理员消息提醒
					if teamInfo.GetType() == Team.TeamMemberType_Tmt_Manager || teamInfo.GetType() == Team.TeamMemberType_Tmt_Owner {
						//pass
					} else {
						kc.logger.Warn("此用户设置了管理员信息提醒", zap.String("toUser", req.GetTo()))
						continue
					}
				case 3: //联系人提醒
				case 4: //所有消息均不提醒
					kc.logger.Warn("此用户设置了所有消息均不提醒", zap.String("toUser", req.GetTo()))
					continue
				}

				//构造转发消息数据
				if newSeq, err = redis.Uint64(redisConn.Do("INCR", fmt.Sprintf("userSeq:%s", toUser))); err != nil {
					kc.logger.Warn("redisConn INCR userSeq Error", zap.Error(err))
					continue
				}
				body := &Msg.MessageNotificationBody{
					HandledAccount: toUser,
					HandledMsg:     "",
					Status:         1,  //TODO, 消息状态  存储
					Text:           "", // 附带的文本 该系统消息的文本
					To:             toUser,
				}
				bodyData, _ := proto.Marshal(body)
				eRsp := &Msg.RecvMsgEventRsp{
					Scene:        Msg.MessageScene_MsgScene_C2C, //个人消息
					Type:         req.GetType(),                 //消息类型
					Body:         bodyData,
					FromDeviceId: deviceID,
					ServerMsgId:  msg.GetID(),                        //服务器分配的消息ID
					Seq:          newSeq,                             //消息序号，单个会话内自然递增, 这里是对targetUsername这个用户的通知序号
					Uuid:         fmt.Sprintf("%d", msg.GetTaskID()), //客户端分配的消息ID，SDK生成的消息id，这里返回TaskID
					Time:         uint64(time.Now().Unix()),
				}

				//转发消息
				go kc.SendMsgToUser(eRsp, username, deviceID, toUser)
			}

		case Msg.MessageScene_MsgScene_P2P: //用户设备之间传输文件或消息
			toUser = username //必须是用户自己
			toDeviceID := req.GetToDeviceId()

			kc.logger.Debug("MessageScene_MsgScene_P2P", zap.String("toUser", toUser))
			//判断toUser的合法性以及是否封禁等
			userData := new(models.User)
			userKey := fmt.Sprintf("userData:%s", toUser)
			if result, err := redis.Values(redisConn.Do("HGETALL", userKey)); err == nil {
				if err := redis.ScanStruct(result, userData); err != nil {
					kc.logger.Error("错误：ScanStruct", zap.Error(err))
					errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
					errorMsg = fmt.Sprintf("ScanStruct Error[Username=%s]", toUser)
					goto COMPLETE

				}
			}

			if userData.State == 2 {
				kc.logger.Warn("此用户已被封号", zap.String("toUser", toUser))
				errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
				errorMsg = fmt.Sprintf("User is blocked[Username=%s]", toUser)
				goto COMPLETE
			}

			//构造转发消息数据
			if newSeq, err = redis.Uint64(redisConn.Do("INCR", fmt.Sprintf("userSeq:%s", toUser))); err != nil {
				kc.logger.Error("redisConn INCR userSeq Error", zap.Error(err))
				errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
				errorMsg = "INCR Error"
				goto COMPLETE
			}

			eRsp := &Msg.RecvMsgEventRsp{
				Scene:        req.GetScene(), //传输场景
				Type:         req.GetType(),  //消息类型
				Body:         req.GetBody(),  //不拆包，直接透传body给接收者
				FromDeviceId: deviceID,
				ServerMsgId:  msg.GetID(),                        //服务器分配的消息ID
				Seq:          newSeq,                             //消息序号，单个会话内自然递增, 这里是对targetUsername这个用户的通知序号
				Uuid:         fmt.Sprintf("%d", msg.GetTaskID()), //客户端分配的消息ID，SDK生成的消息id，这里返回TaskID
				Time:         uint64(time.Now().Unix()),
			}
			data, _ := proto.Marshal(eRsp)

			//转发透传消息
			targetMsg := &models.Message{}
			curDeviceKey := fmt.Sprintf("DeviceJwtToken:%s", toDeviceID)
			curJwtToken, _ := redis.String(redisConn.Do("GET", curDeviceKey)) //每个设备都有自己的token
			kc.logger.Debug("Redis GET ", zap.String("curDeviceKey", curDeviceKey), zap.String("curJwtToken", curJwtToken))

			targetMsg.UpdateID()
			//构建消息路由, 第一个参数是要处理的业务类型，后端服务器处理完成后，需要用此来拼接topic: {businessTypeName.Frontend}
			targetMsg.BuildRouter("Msg", "", "Msg.Frontend")
			targetMsg.SetJwtToken(curJwtToken)
			targetMsg.SetUserName(toUser) //发给自己
			targetMsg.SetDeviceID(toDeviceID)
			// kickMsg.SetTaskID(uint32(taskId))
			targetMsg.SetBusinessTypeName("Msg")
			targetMsg.SetBusinessType(uint32(Global.BusinessType_Msg))           //消息模块
			targetMsg.SetBusinessSubType(uint32(Global.MsgSubType_RecvMsgEvent)) //接收消息事件
			targetMsg.BuildHeader("ChatService", time.Now().UnixNano()/1e6)
			targetMsg.FillBody(data) //网络包的body，承载真正的业务数据
			targetMsg.SetCode(200)   //成功的状态码

			//构建数据完成，向dispatcher发送
			topic := "Msg.Frontend"
			go kc.Produce(topic, targetMsg)

			kc.logger.Info("HandleRecvMsg Succeed",
				zap.String("Username:", username))
		}
	}

COMPLETE:
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
	if err := kc.Produce(topic, msg); err == nil {
		kc.logger.Info("SendMsgRsp message succeed send to ProduceChannel", zap.String("topic", topic))
	} else {
		kc.logger.Error("Failed to send SendMsgRsp message to ProduceChannel", zap.Error(err))
	}
	_ = err
	return nil
}

//5-4 确认消息送达
func (kc *KafkaClient) HandleMsgAck(msg *models.Message) error {
	var err error
	errorCode := 200
	var errorMsg string

	redisConn := kc.redisPool.Get()
	defer redisConn.Close()

	username := msg.GetUserName()
	// token := msg.GetJwtToken()
	deviceID := msg.GetDeviceID()

	kc.logger.Info("HandleMsgAck start...",
		zap.String("username", username),
		zap.String("DeviceId", deviceID))

	//取出当前设备的os， clientType， logonAt
	curDeviceHashKey := fmt.Sprintf("devices:%s:%s", username, deviceID)
	isMaster, _ := redis.Bool(redisConn.Do("HGET", curDeviceHashKey, "ismaster"))
	curOs, _ := redis.String(redisConn.Do("HGET", curDeviceHashKey, "os"))
	curClientType, _ := redis.Int(redisConn.Do("HGET", curDeviceHashKey, "clientType"))
	curLogonAt, _ := redis.Uint64(redisConn.Do("HGET", curDeviceHashKey, "logonAt"))

	kc.logger.Debug("MsgAck",
		zap.Bool("isMaster", isMaster),
		zap.String("username", username),
		zap.String("deviceID", deviceID),
		zap.String("curOs", curOs),
		zap.Int("curClientType", curClientType),
		zap.Uint64("curLogonAt", curLogonAt))

	//打开msg里的负载， 获取请求参数
	body := msg.GetContent()
	//解包body
	var req Msg.MsgAckReq
	if err := proto.Unmarshal(body, &req); err != nil {
		errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
		errorMsg = fmt.Sprintf("Protobuf Unmarshal Error: %s", err.Error())
		kc.logger.Error("Protobuf Unmarshal Error", zap.Error(err))
		goto COMPLETE

	} else {
		kc.logger.Debug("MsgAck payload",
			zap.Int32("Scene", int32(req.GetScene())),
			zap.Int32("Type", int32(req.GetType())),
			zap.String("ServerMsgId", req.GetServerMsgId()),
			zap.Uint64("Seq", req.GetSeq()),
		)

		//从Redis缓存的消息队列里删除seq
		if Msg.MessageType_MsgType_Notification == req.GetType() { //通知类型
			if _, err = redisConn.Do("ZREM", fmt.Sprintf("systemMsgAt:%s", username), req.GetServerMsgId()); err != nil {
				kc.logger.Error("ZREM Error", zap.Error(err))
			} else {
				key := fmt.Sprintf("systemMsg:%s:%s", username, req.GetServerMsgId())
				_, err = redisConn.Do("DEL", key)
			}
		} else { //其它消息类型
			if _, err = redisConn.Do("ZREM", fmt.Sprintf("offLineMsgList:%s", username), req.GetServerMsgId()); err != nil {
				kc.logger.Error("ZREM Error", zap.Error(err))
			} else {
				key := fmt.Sprintf("offLineMsg:%s:%s", username, req.GetServerMsgId())
				_, err = redisConn.Do("DEL", key)
			}
		}
	}

COMPLETE:
	msg.SetCode(int32(errorCode)) //状态码
	if errorCode == 200 {
		//
	} else {
		msg.SetErrorMsg([]byte(errorMsg)) //错误提示
		msg.FillBody(nil)
	}

	//处理完成，向dispatcher发送
	topic := msg.GetSource() + ".Frontend"
	if err := kc.Produce(topic, msg); err == nil {
		kc.logger.Info("MsgAck message succeed send to ProduceChannel", zap.String("topic", topic))
	} else {
		kc.logger.Error("Failed to send MsgAck message to ProduceChannel", zap.Error(err))
	}
	_ = err
	return nil

}

//5-9 发送撤销消息
func (kc *KafkaClient) HandleSendCancelMsg(msg *models.Message) error {
	var err error
	errorCode := 200
	var errorMsg string
	var isExists bool
	var newSeq uint64

	redisConn := kc.redisPool.Get()
	defer redisConn.Close()

	username := msg.GetUserName()
	// token := msg.GetJwtToken()
	deviceID := msg.GetDeviceID()

	kc.logger.Info("HandleSendCancelMsg start...",
		zap.String("username", username),
		zap.String("DeviceId", deviceID))

	//取出当前设备的os， clientType， logonAt
	curDeviceHashKey := fmt.Sprintf("devices:%s:%s", username, deviceID)
	isMaster, _ := redis.Bool(redisConn.Do("HGET", curDeviceHashKey, "ismaster"))
	curOs, _ := redis.String(redisConn.Do("HGET", curDeviceHashKey, "os"))
	curClientType, _ := redis.Int(redisConn.Do("HGET", curDeviceHashKey, "clientType"))
	curLogonAt, _ := redis.Uint64(redisConn.Do("HGET", curDeviceHashKey, "logonAt"))

	kc.logger.Debug("SendCancelMsg",
		zap.Bool("isMaster", isMaster),
		zap.String("username", username),
		zap.String("deviceID", deviceID),
		zap.String("curOs", curOs),
		zap.Int("curClientType", curClientType),
		zap.Uint64("curLogonAt", curLogonAt))

	//打开msg里的负载， 获取请求参数
	body := msg.GetContent()
	//解包body
	var req Msg.SendCancelMsgReq
	if err := proto.Unmarshal(body, &req); err != nil {
		errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
		errorMsg = fmt.Sprintf("Protobuf Unmarshal Error: %s", err.Error())
		kc.logger.Error("Protobuf Unmarshal Error", zap.Error(err))
		goto COMPLETE

	} else {
		kc.logger.Debug("SendCancelMsg payload",
			zap.Int32("Scene", int32(req.GetScene())),
			zap.Int32("Type", int32(req.GetType())),
			zap.String("ServerMsgId", req.GetServerMsgId()),
		)

		//查询出谁接收了此消息，如果超过1分钟，则无法撤销
		recvKey := fmt.Sprintf("recvMsg:%s", req.GetServerMsgId())
		if isExists, err = redis.Bool(redisConn.Do("EXISTS", recvKey)); err != nil {
			kc.logger.Error("EXISTS Error", zap.Error(err))
		}

		if isExists {
			recvUsers, err := redis.Strings(redisConn.Do("ZRANGEBYSCORE", recvKey, "-inf", "+inf"))

			//查询出用户的当前在线所有主从设备
			for _, recvUser := range recvUsers {
				//此消息的接收人, 需要将消息从接收人的缓存队列删除，并下发撤销通知, 如果是群消息，会有很多接收人
				//向接收此消息的用户发出撤销消息 RecvCancelMsgEvent
				if newSeq, err = redis.Uint64(redisConn.Do("INCR", fmt.Sprintf("userSeq:%s", recvUser))); err != nil {
					kc.logger.Error("redisConn INCR userSeq Error", zap.Error(err))
				}

				body := &Msg.MessageNotificationBody{
					Type:           Msg.MessageNotificationType_MNT_RejectFriendApply, //对方拒绝添加好友
					HandledAccount: username,
					HandledMsg:     "",
					Status:         1,  //TODO, 消息状态  存储
					Text:           "", // 附带的文本 该系统消息的文本
					To:             recvUser,
				}
				bodyData, _ := proto.Marshal(body)

				eRsp := &Msg.RecvMsgEventRsp{
					Scene:        req.GetScene(), //传输场景
					Type:         req.GetType(),  //消息类型
					Body:         bodyData,
					FromDeviceId: deviceID,
					ServerMsgId:  msg.GetID(),                        //服务器分配的消息ID
					Seq:          newSeq,                             //消息序号，单个会话内自然递增, 这里是对targetUsername这个用户的通知序号
					Uuid:         fmt.Sprintf("%d", msg.GetTaskID()), //客户端分配的消息ID，SDK生成的消息id，这里返回TaskID
					Time:         uint64(time.Now().Unix()),
				}

				//从Redis里删除recvUser缓存的消息队列
				if _, err = redisConn.Do("ZREM", fmt.Sprintf("offLineMsgList:%s", recvUser), req.GetServerMsgId()); err != nil {
					kc.logger.Error("ZREM Error", zap.Error(err))
				} else {
					key := fmt.Sprintf("offLineMsg:%s:%s", recvUser, req.GetServerMsgId())
					_, err = redisConn.Do("DEL", key)
				}

				go kc.SendMsgToUser(eRsp, username, deviceID, recvUser)

			}

		}

		//向用户自己的其它端发送撤销消息
		selfDeviceListKey := fmt.Sprintf("devices:%s", username)
		selfDeviceIDSliceNew, _ := redis.Strings(redisConn.Do("ZRANGEBYSCORE", selfDeviceListKey, "-inf", "+inf"))

		//查询出用户的当前在线所有主从设备
		for _, selfDeviceID := range selfDeviceIDSliceNew {
			if selfDeviceID == deviceID {
				continue
			}
			if newSeq, err = redis.Uint64(redisConn.Do("INCR", fmt.Sprintf("userSeq:%s", username))); err != nil {
				kc.logger.Error("redisConn INCR userSeq Error", zap.Error(err))
				errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
				errorMsg = "INCR Error"
				goto COMPLETE
			}
			rsp := &Msg.RecvMsgEventRsp{
				Scene: req.GetScene(), //系统消息
				Type:  req.GetType(),  //通知类型
				// Body:         bodyData, //此时没有body
				FromDeviceId: deviceID,
				ServerMsgId:  req.GetServerMsgId(),               //服务器分配的消息ID
				Seq:          newSeq,                             //消息序号，单个会话内自然递增, 这里是对targetUsername这个用户的通知序号
				Uuid:         fmt.Sprintf("%d", msg.GetTaskID()), //客户端分配的消息ID，SDK生成的消息id，这里返回TaskID
				Time:         uint64(time.Now().Unix()),
			}
			data, _ := proto.Marshal(rsp)

			targetMsg := &models.Message{}
			curDeviceKey := fmt.Sprintf("DeviceJwtToken:%s", selfDeviceID)
			curJwtToken, _ := redis.String(redisConn.Do("GET", curDeviceKey)) //每个设备都有自己的token
			targetMsg.UpdateID()
			//构建消息路由, 第一个参数是要处理的业务类型，后端服务器处理完成后，需要用此来拼接topic: {businessTypeName.Frontend}
			targetMsg.BuildRouter("Msg", "", "Msg.Frontend")
			targetMsg.SetJwtToken(curJwtToken)
			targetMsg.SetUserName(username) //发给自己
			targetMsg.SetDeviceID(selfDeviceID)
			// kickMsg.SetTaskID(uint32(taskId))
			targetMsg.SetBusinessTypeName("Msg")
			targetMsg.SetBusinessType(uint32(Global.BusinessType_Msg))                     //消息模块
			targetMsg.SetBusinessSubType(uint32(Global.MsgSubType_SyncSendCancelMsgEvent)) //自己的设备同步发送撤销消息事件
			targetMsg.BuildHeader("ChatService", time.Now().UnixNano()/1e6)
			targetMsg.FillBody(data) //网络包的body，承载真正的业务数据
			targetMsg.SetCode(200)   //成功的状态码

			//构建数据完成，向dispatcher发送
			topic := "Msg.Frontend"
			go kc.Produce(topic, targetMsg)

			kc.logger.Info("Msg message send to ProduceChannel",
				zap.String("topic", topic),
				zap.String("to", username),
				zap.String("toDeviceID:", curDeviceKey),
				zap.String("msgID:", targetMsg.GetID()),
				zap.Uint64("seq", newSeq),
			)
		}
	}

COMPLETE:
	msg.SetCode(int32(errorCode)) //状态码
	if errorCode == 200 {
		//
	} else {
		msg.SetErrorMsg([]byte(errorMsg)) //错误提示
		msg.FillBody(nil)
	}

	//处理完成，向dispatcher发送
	topic := msg.GetSource() + ".Frontend"
	if err := kc.Produce(topic, msg); err == nil {
		kc.logger.Info("SendCancelMsg message succeed send to ProduceChannel", zap.String("topic", topic))
	} else {
		kc.logger.Error("Failed to send SendCancelMsg message to ProduceChannel", zap.Error(err))
	}
	_ = err
	return nil
}

//5-12 获取阿里云OSS上传Token
func (kc *KafkaClient) HandleGetOssToken(msg *models.Message) error {
	var err error
	errorCode := 200
	var errorMsg string
	rsp := &Msg.GetOssTokenRsp{}
	// var isExists bool
	// var newSeq uint64

	redisConn := kc.redisPool.Get()
	defer redisConn.Close()

	username := msg.GetUserName()
	// token := msg.GetJwtToken()
	deviceID := msg.GetDeviceID()

	kc.logger.Info("HandleGetOssToken start...",
		zap.String("username", username),
		zap.String("DeviceId", deviceID))

	//取出当前设备的os， clientType， logonAt
	curDeviceHashKey := fmt.Sprintf("devices:%s:%s", username, deviceID)
	isMaster, _ := redis.Bool(redisConn.Do("HGET", curDeviceHashKey, "ismaster"))
	curOs, _ := redis.String(redisConn.Do("HGET", curDeviceHashKey, "os"))
	curClientType, _ := redis.Int(redisConn.Do("HGET", curDeviceHashKey, "clientType"))
	curLogonAt, _ := redis.Uint64(redisConn.Do("HGET", curDeviceHashKey, "logonAt"))

	kc.logger.Debug("GetOssToken",
		zap.Bool("isMaster", isMaster),
		zap.String("username", username),
		zap.String("deviceID", deviceID),
		zap.String("curOs", curOs),
		zap.Int("curClientType", curClientType),
		zap.Uint64("curLogonAt", curLogonAt))

	//打开msg里的负载， 获取请求参数
	body := msg.GetContent()
	//解包body
	var req Msg.GetOssTokenReq
	if err := proto.Unmarshal(body, &req); err != nil {
		kc.logger.Error("Protobuf Unmarshal Error", zap.Error(err))
		errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
		errorMsg = fmt.Sprintf("Protobuf Unmarshal Error: %s", err.Error())
		goto COMPLETE

	} else {
		kc.logger.Debug("GetOssToken payload")

		//生成阿里云oss临时sts
		client := sts.NewStsClient(TempKey, TempSecret, RoleAcs)

		//阿里云规定，最低expire为1500秒  
		url, err := client.GenerateSignatureUrl("client", "1500") 
		if err != nil {
			kc.logger.Error("GenerateSignatureUrl Error", zap.Error(err))
			errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
			errorMsg = fmt.Sprintf("GenerateSignatureUrl Error: %s", err.Error())
			goto COMPLETE
		}

		data, err := client.GetStsResponse(url)
		if err != nil {
			kc.logger.Error("阿里云oss GetStsResponse Error", zap.Error(err))
			errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
			errorMsg = fmt.Sprintf("GetOssToken Error: %s", err.Error())
			goto COMPLETE
		}

		/*
			{
				"RequestId": "1468D6F6-DA5D-4DAB-8B6C-383C22524E32",
				"AssumedRoleUser": {
					"Arn": "acs:ram::1230446857465673:role/ipfsuploader/client",
					"AssumedRoleId": "359775758821401491:client"
				},
				"Credentials": {
					"SecurityToken": "CAIS8QF1q6Ft5B2yfSjIr5bBGYnNhK5Aw4S9W3bUoGgeOc5Pi/39hDz2IH1Fe3ZtBu0Wvv42mGhR6vcblq94T55IQ1CchGnlLT8Ro22beIPkl5Gfz95t0e+IewW6Dxr8w7WhAYHQR8/cffGAck3NkjQJr5LxaTSlWS7OU/TL8+kFCO4aRQ6ldzFLKc5LLw950q8gOGDWKOymP2yB4AOSLjIx4FEk1T8hufngnpPBtEWFtjCglL9J/baWC4O/csxhMK14V9qIx+FsfsLDqnUNukcVqfgr3PweoGuf543MWkM14g2IKPfM9tpmIAJjdgmMmRj3JgeWGoABp6cWPupoRAUG7ohyuVM1+vZlTObY1ZRkp40wHpvxFLs6i2UZILNQ7+Myf0ZDGcc2MMpxIu+Vl5Kdy4YEMBDokmzzRCZGXqAQzzkSm+9jvbaIKwjj90wjmellGbwIQ4zZ4NbZ4HeYDIyUgBuQ1bwdn6UHZvtG7NC2g9sawbPtKZ0=",
					"AccessKeyId": "STS.NUtR3yiqatFVYPeDhK5Acd8Vf",
					"AccessKeySecret": "9baJcUJEHbmC1rdEzRg8DviWrSMuSPzM1QRBMt12Y3cC",
					"Expiration": "2020-08-28T14:36:15Z"
				}
			}
		*/

		// log.Println("result:", string(data))
		sjson, err := simpleJson.NewJson(data)
		if err != nil {
			kc.logger.Warn("simplejson.NewJson Error", zap.Error(err))
			errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
			errorMsg = fmt.Sprintf("GetOssToken Error: %s", err.Error())
			goto COMPLETE
		}

		kc.logger.Debug("收到阿里云OSS服务端的消息", 
			zap.String("RequestId", sjson.Get("RequestId").MustString()),
			zap.String("AccessKeyId", sjson.Get("Credentials").Get("AccessKeyId").MustString()),
			zap.String("AccessKeySecret", sjson.Get("Credentials").Get("AccessKeySecret").MustString()),
			zap.String("SecurityToken", sjson.Get("Credentials").Get("SecurityToken").MustString()),
			zap.String("Expiration", sjson.Get("Credentials").Get("Expiration").MustString()),
		)
		rsp = &Msg.GetOssTokenRsp{
			EndPoint: "https://oss-cn-hangzhou.aliyuncs.com",
			BucketName: "lianmi-ipfs",
			AccessKeyId: sjson.Get("Credentials").Get("AccessKeyId").MustString(),
			AccessKeySecret: sjson.Get("Credentials").Get("AccessKeySecret").MustString(),
			SecurityToken: sjson.Get("Credentials").Get("SecurityToken").MustString(),
			Directory: time.Now().Format("2006/01/02/"),
			Expire: 1500, //阿里云规定，最低为1500秒
			Callback: "", //不填

		}

		// if sjson.Get("head").Get("method").MustString() == "price.subscribe" {
	}

COMPLETE:
	msg.SetCode(int32(errorCode)) //状态码
	if errorCode == 200 {
		data, _ := proto.Marshal(rsp)
		msg.FillBody(data) //网络包的body，承载真正的业务数据
	} else {
		msg.SetErrorMsg([]byte(errorMsg)) //错误提示
		msg.FillBody(nil)
	}

	//处理完成，向dispatcher发送
	topic := msg.GetSource() + ".Frontend"
	if err := kc.Produce(topic, msg); err == nil {
		kc.logger.Info("SendCancelMsg message succeed send to ProduceChannel", zap.String("topic", topic))
	} else {
		kc.logger.Error("Failed to send SendCancelMsg message to ProduceChannel", zap.Error(err))
	}
	_ = err
	return nil
}

func (kc *KafkaClient) SendMsgToUser(rsp *Msg.RecvMsgEventRsp, fromUser, fromDeviceID, toUser string) error {
	data, _ := proto.Marshal(rsp)

	redisConn := kc.redisPool.Get()
	defer redisConn.Close()

	//Redis里缓存此消息,目的是用户从离线状态恢复到上线状态后同步这些消息给用户
	msgAt := time.Now().Unix()
	if _, err := redisConn.Do("ZADD", fmt.Sprintf("offLineMsgList:%s", toUser), rsp.Seq, rsp.GetServerMsgId()); err != nil {
		kc.logger.Error("ZADD Error", zap.Error(err))
	}

	//离线消息具体内容
	key := fmt.Sprintf("offLineMsg:%s:%s", toUser, rsp.GetServerMsgId())

	_, err := redisConn.Do("HMSET",
		key,
		"Scene", rsp.GetScene(),
		"Type", rsp.GetType(),
		"Username", toUser,
		"MsgAt", msgAt,
		"Seq", rsp.Seq,
		"Data", data,
	)

	_, err = redisConn.Do("EXPIRE", key, 7*24*3600) //设置有效期为7天

	//存储谁接收了此消息，以便撤销
	if _, err := redisConn.Do("ZADD", fmt.Sprintf("recvMsgList:%s", rsp.GetServerMsgId()), msgAt, toUser); err != nil {
		kc.logger.Error("ZADD Error", zap.Error(err))
	}
	_, err = redisConn.Do("EXPIRE", fmt.Sprintf("recvMsgList:%s", toUser), 60) //设置有效期为60秒

	//向用户的其它端发送 SyncSendMsgEvent
	selfDeviceListKey := fmt.Sprintf("devices:%s", fromUser)
	selfDeviceIDSliceNew, _ := redis.Strings(redisConn.Do("ZRANGEBYSCORE", selfDeviceListKey, "-inf", "+inf"))

	//查询出当前在线所有主从设备
	for _, selfDeviceID := range selfDeviceIDSliceNew {
		if selfDeviceID == fromDeviceID {
			continue
		}
		targetMsg := &models.Message{}
		curDeviceKey := fmt.Sprintf("DeviceJwtToken:%s", selfDeviceID)
		curJwtToken, _ := redis.String(redisConn.Do("GET", curDeviceKey)) //每个设备都有自己的token
		targetMsg.UpdateID()
		//构建消息路由, 第一个参数是要处理的业务类型，后端服务器处理完成后，需要用此来拼接topic: {businessTypeName.Frontend}
		targetMsg.BuildRouter("Msg", "", "Msg.Frontend")
		targetMsg.SetJwtToken(curJwtToken)
		targetMsg.SetUserName(fromUser) //发给自己
		targetMsg.SetDeviceID(selfDeviceID)
		// kickMsg.SetTaskID(uint32(taskId))
		targetMsg.SetBusinessTypeName("Msg")
		targetMsg.SetBusinessType(uint32(Global.BusinessType_Msg))           //消息模块
		targetMsg.SetBusinessSubType(uint32(Global.MsgSubType_RecvMsgEvent)) //接收消息事件
		targetMsg.BuildHeader("ChatService", time.Now().UnixNano()/1e6)
		targetMsg.FillBody(data) //网络包的body，承载真正的业务数据
		targetMsg.SetCode(200)   //成功的状态码

		//构建数据完成，向dispatcher发送
		topic := "Msg.Frontend"
		if err := kc.Produce(topic, targetMsg); err == nil {
			kc.logger.Info("Msg message succeed send to ProduceChannel",
				zap.String("topic", topic),
				zap.String("to", fromUser),
				zap.String("toDeviceID:", curDeviceKey),
				zap.String("msgID:", rsp.GetServerMsgId()),
				zap.Uint64("seq", rsp.Seq),
			)
		} else {
			kc.logger.Error("Failed to send message to ProduceChannel",
				zap.String("topic", topic),
				zap.String("to", fromUser),
				zap.String("toDeviceID:", curDeviceKey),
				zap.String("msgID:", rsp.GetServerMsgId()),
				zap.Uint64("seq", rsp.Seq),
				zap.Error(err),
			)
		}
	}

	//向toUser所有端发送
	deviceListKey := fmt.Sprintf("devices:%s", toUser)
	deviceIDSliceNew, _ := redis.Strings(redisConn.Do("ZRANGEBYSCORE", deviceListKey, "-inf", "+inf"))

	//查询出toUser当前在线所有主从设备
	for _, eDeviceID := range deviceIDSliceNew {
		targetMsg := &models.Message{}
		curDeviceKey := fmt.Sprintf("DeviceJwtToken:%s", eDeviceID)
		curJwtToken, _ := redis.String(redisConn.Do("GET", curDeviceKey))
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
		targetMsg.SetCode(200)   //成功的状态码

		//构建数据完成，向dispatcher发送
		topic := "Msg.Frontend"
		if err := kc.Produce(topic, targetMsg); err == nil {
			kc.logger.Info("Msg message succeed send to ProduceChannel",
				zap.String("topic", topic),
				zap.String("toUser", toUser),
				zap.String("toDeviceID:", curDeviceKey),
				zap.String("msgID:", rsp.GetServerMsgId()),
				zap.Uint64("seq", rsp.Seq),
			)
		} else {
			kc.logger.Error("Failed to send message to ProduceChannel",
				zap.String("topic", topic),
				zap.String("toUser", toUser),
				zap.String("toDeviceID:", curDeviceKey),
				zap.String("msgID:", rsp.GetServerMsgId()),
				zap.Uint64("seq", rsp.Seq),
				zap.Error(err),
			)
		}
	}
	_ = err
	return nil
}
