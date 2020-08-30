package kafkaBackend

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gomodule/redigo/redis"
	Global "github.com/lianmi/servers/api/proto/global"
	Msg "github.com/lianmi/servers/api/proto/msg"
	Team "github.com/lianmi/servers/api/proto/team"

	// Order "github.com/lianmi/servers/api/proto/order"
	"github.com/lianmi/servers/internal/pkg/models"
	"google.golang.org/protobuf/proto"

	"go.uber.org/zap"
)

/*

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
				Scene:        req.GetScene(),                     //传输场景
				Type:         req.GetType(),                      //消息类型
				Body:         req.GetBody(),                      //不拆包，直接透传body给接收者
				From:         username,                           //谁发的
				FromDeviceId: deviceID,                           //哪个设备发的
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
					From:         username,                           //谁发的
					FromDeviceId: deviceID,                           //哪个设备发的
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
				Scene:        req.GetScene(),                     //传输场景
				Type:         req.GetType(),                      //消息类型
				Body:         req.GetBody(),                      //不拆包，直接透传body给接收者
				From:         username,                           //谁发的
				FromDeviceId: deviceID,                           //哪个设备发的
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
			targetMsg.BuildHeader("orderservice", time.Now().UnixNano()/1e6)
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
