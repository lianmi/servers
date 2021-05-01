package nsqMq

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/gomodule/redigo/redis"
	Global "github.com/lianmi/servers/api/proto/global"
	Msg "github.com/lianmi/servers/api/proto/msg"
	Order "github.com/lianmi/servers/api/proto/order"
	Team "github.com/lianmi/servers/api/proto/team"
	LMCommon "github.com/lianmi/servers/internal/common"
	LMCError "github.com/lianmi/servers/internal/pkg/lmcerror"
	"github.com/lianmi/servers/internal/pkg/models"
	"github.com/lianmi/servers/internal/pkg/sts"
	"google.golang.org/protobuf/proto"

	simpleJson "github.com/bitly/go-simplejson"
	"go.uber.org/zap"
)

/*
处理SDK发来的sendmsg

5-1 发送消息
1. 消息发送接口，包括单聊、群聊、系统通知
2. 服务器会处理接收方的主从设备的消息分发
3. 除了商户及客服账号外，不支持陌生人聊天
*/
func (nc *NsqClient) HandleRecvMsg(msg *models.Message) error {
	var err error
	var toUser, teamID string
	errorCode := 200
	var isCustomerService bool //toUser是否是客服账号
	rsp := &Msg.SendMsgRsp{}

	var newSeq uint64
	var data []byte

	redisConn := nc.redisPool.Get()
	defer redisConn.Close()

	username := msg.GetUserName()
	// token := msg.GetJwtToken()
	deviceID := msg.GetDeviceID()

	nc.logger.Info("HandleRecvMsg start...",
		zap.String("username", username),
		zap.String("DeviceId", deviceID))

	//取出当前设备的os logonAt
	curDeviceHashKey := fmt.Sprintf("devices:%s:%s", username, deviceID)
	isMaster, _ := redis.Bool(redisConn.Do("HGET", curDeviceHashKey, "ismaster"))
	curOs, _ := redis.String(redisConn.Do("HGET", curDeviceHashKey, "os"))
	curLogonAt, _ := redis.Uint64(redisConn.Do("HGET", curDeviceHashKey, "logonAt"))

	nc.logger.Debug("RecvMsg",
		zap.Bool("isMaster", isMaster),
		zap.String("username", username),
		zap.String("deviceID", deviceID),
		zap.String("curOs", curOs),
		zap.Uint64("curLogonAt", curLogonAt))

	//打开msg里的负载， 获取请求参数
	body := msg.GetContent()
	//解包body
	var req Msg.SendMsgReq
	if err := proto.Unmarshal(body, &req); err != nil {
		nc.logger.Error("Protobuf Unmarshal Error", zap.Error(err))
		errorCode = LMCError.ProtobufUnmarshalError
		goto COMPLETE

	} else {
		nc.logger.Debug("RecvMsg  payload",
			zap.Int32("Scene", int32(req.GetScene())),
			zap.Int32("Type", int32(req.GetType())),
			zap.String("To", req.GetTo()),
			zap.String("Uuid", req.GetUuid()),
			zap.Uint64("SendAt", req.GetSendAt()),
		)

		//根据场景判断消息是个人消息、群聊消息
		switch req.GetScene() {
		case Msg.MessageScene_MsgScene_C2C: //个人消息
			toUser = req.GetTo() //给谁发消息
			nc.logger.Debug("MessageScene_MsgScene_C2C",
				zap.String("toUser", req.GetTo()),
				zap.Int("Type", int(req.GetType())),
			)

			//判断消息类型
			switch req.GetType() {
			case Msg.MessageType_MsgType_Text, //Text(0)-文本
				Msg.MessageType_MsgType_Attach, //附件类型
				Msg.MessageType_MsgType_Bin,    // 二进制
				Msg.MessageType_MsgType_Secret: //加密类型

				userKey := fmt.Sprintf("userData:%s", username)
				userState, _ := redis.Int(redisConn.Do("HGET", userKey, "State"))
				userType, _ := redis.Int(redisConn.Do("HGET", userKey, "UserType"))

				if userState == 2 {
					nc.logger.Warn("此用户已被封号", zap.String("Username", req.GetTo()))
					errorCode = LMCError.UserIsBlockedError
					goto COMPLETE
				}

				//判断toUser的合法性以及是否封禁等
				toUserKey := fmt.Sprintf("userData:%s", toUser)
				toUserState, _ := redis.Int(redisConn.Do("HGET", toUserKey, "State"))
				toUserType, _ := redis.Int(redisConn.Do("HGET", toUserKey, "UserType"))

				if toUserState == 2 {
					nc.logger.Warn("此用户已被封号", zap.String("toUser", req.GetTo()))
					errorCode = LMCError.UserIsBlockedError
					goto COMPLETE
				}

				// 商户(或客服)与买家之间的私聊
				if userType == 2 || toUserType == 2 || userType == 4 || toUserType == 4 {
					//允许私聊
					nc.logger.Debug("商户(或客服)与用户的会话, 放行... ", zap.String("username", username), zap.String("toUser", req.GetTo()))
				} else {
					// 除了商户及客服账号外，不支持陌生人聊天

					//判断 本次会话是不是 客服 与 用户 的会话， 如果是则放行
					isCustomerService = nc.CheckIsCustomerService(username, toUser)

					if isCustomerService {
						//pass
						nc.logger.Info("客服与用户的会话, 放行... ", zap.String("username", username), zap.String("toUser", req.GetTo()))

					} else {

						//判断两者是不是好友， 单向好友也不能发消息
						var isAhaveB, isBhaveA bool //A好友列表里有B， B好友列表里有A
						if reply, err := redisConn.Do("ZRANK", fmt.Sprintf("Friend:%s:1", username), toUser); err == nil {
							if reply == nil {
								//A好友列表中没有B
								isAhaveB = false
							} else {
								isAhaveB = true
							}

						}
						if reply, err := redisConn.Do("ZRANK", fmt.Sprintf("Friend:%s:1", toUser), username); err == nil {
							if reply == nil {
								//B好友列表中没有A
								isBhaveA = false
							} else {
								isBhaveA = true
							}

						}

						if isAhaveB && isBhaveA {
							//pass
						} else {
							nc.logger.Warn("对方用户不是当前用户的好友", zap.String("toUser", req.GetTo()))
							errorCode = LMCError.IsNotFriendError
							goto COMPLETE
						}

						//查出接收人对此用户消息接收的设定，黑名单，屏蔽等
						if reply, err := redisConn.Do("ZRANK", fmt.Sprintf("BlackList:%s:1", toUser), username); err == nil {
							if reply != nil {
								nc.logger.Warn("用户已被对方拉黑", zap.String("toUser", req.GetTo()))
								errorCode = LMCError.IsNotFriendError
								goto COMPLETE
							}
						}

					}
				}

				//构造转发消息数据
				if newSeq, err = redis.Uint64(redisConn.Do("INCR", fmt.Sprintf("userSeq:%s", toUser))); err != nil {
					nc.logger.Error("redisConn INCR userSeq Error", zap.Error(err))
					errorCode = LMCError.RedisError
					goto COMPLETE
				}

				eRsp := &Msg.RecvMsgEventRsp{
					Scene:        req.GetScene(), //传输场景
					Type:         req.GetType(),  //消息类型
					Body:         req.GetBody(),  //不拆包，直接透传body给接收者
					From:         username,       //谁发的
					FromDeviceId: deviceID,       //哪个设备发的
					Recv:         toUser,         //个人消息接收方
					ServerMsgId:  msg.GetID(),    //服务器分配的消息ID
					Seq:          newSeq,         //消息序号，单个会话内自然递增, 这里是对targetUsername这个用户的通知序号
					Uuid:         req.GetUuid(),  //客户端分配的消息ID，SDK生成的消息id
					Time:         uint64(time.Now().UnixNano() / 1e6),
				}

				//发送消息到目标用户
				go nc.SendMsgToUser(eRsp, username, deviceID, toUser)

			case Msg.MessageType_MsgType_Order: //订单
				//将消息转发到OrderService
				nc.logger.Debug("MessageType_MsgType_Order, 收到订单, 将消息转发到Order.Backend")
				msg.BuildRouter("Order", "", "Order.Backend")
				topic := "Order.Backend"

				rawData, _ := json.Marshal(msg)
				go nc.Producer.Public(topic, rawData)

				//转发后，不需要发回包，直接退出
				return nil

			case Msg.MessageType_MsgType_Roof: //吸顶式群消息 只能是系统、群主或管理员发送，此消息会吸附在群会话的最上面，适合一些倒计时、股价、币价、比分、赔率等

			case Msg.MessageType_MSgType_Customer: //自定义消息
				nc.logger.Debug("收到自定义消息, 将消息转发", zap.String("toUser", req.GetTo()))
				//构造转发消息数据
				if newSeq, err = redis.Uint64(redisConn.Do("INCR", fmt.Sprintf("userSeq:%s", req.GetTo()))); err != nil {
					nc.logger.Error("redisConn INCR userSeq Error", zap.Error(err))
					errorCode = LMCError.RedisError
					goto COMPLETE
				}

				eRsp := &Msg.RecvMsgEventRsp{
					Scene:        req.GetScene(), //传输场景
					Type:         req.GetType(),  //消息类型
					Body:         req.GetBody(),  //不拆包，直接透传body给接收者
					From:         username,       //谁发的
					FromDeviceId: deviceID,       //哪个设备发的
					Recv:         req.GetTo(),    //个人消息接收方
					ServerMsgId:  msg.GetID(),    //服务器分配的消息ID
					Seq:          newSeq,         //消息序号，单个会话内自然递增, 这里是对targetUsername这个用户的通知序号
					Uuid:         req.GetUuid(),  //客户端分配的消息ID，SDK生成的消息id
					Time:         uint64(time.Now().UnixNano() / 1e6),
				}

				//发送消息到目标用户
				go nc.SendMsgToUser(eRsp, username, deviceID, req.GetTo())
			}

		case Msg.MessageScene_MsgScene_C2G: //群聊消息
			teamID = req.GetTo()
			nc.logger.Debug("MessageScene_MsgScene_C2G", zap.String("toTeamID", req.GetTo()))
			//判断toTeamID的合法性以及是否封禁等

			//获取到群信息
			key := fmt.Sprintf("TeamInfo:%s", teamID)
			teamInfoStatus, _ := redis.Int(redisConn.Do("HGET", key, "Status"))
			teamInfoType, _ := redis.Int(redisConn.Do("HGET", key, "Type"))

			//此群是否是正常的
			if teamInfoStatus != int(Team.TeamStatus_Status_Normal) {
				nc.logger.Warn("Team status is not normal", zap.Int("Status", teamInfoStatus))
				errorCode = LMCError.TeamStatusError
				goto COMPLETE
			}

			//for..range 群成员
			teamMembers, _ := redis.Strings(redisConn.Do("ZRANGEBYSCORE", fmt.Sprintf("TeamUsers:%s", teamID), "-inf", "+inf"))
			for _, teamMember := range teamMembers {
				toUser = teamMember

				//判断toUser的合法性以及是否封禁等
				userKey := fmt.Sprintf("userData:%s", toUser)
				userState, _ := redis.Int(redisConn.Do("HGET", userKey, "State"))

				if userState == 2 {
					nc.logger.Warn("此用户已被封号", zap.String("toUser", req.GetTo()))
					continue
				}

				//查出此用户对此群消息接收的设定，如果允许接收，就发
				toUserKey := fmt.Sprintf("TeamUser:%s:%s", teamID, toUser)
				notifyType, _ := redis.Int(redisConn.Do("HGET", toUserKey, "NotifyType"))
				switch notifyType {
				case 1: //群全部消息提醒
				case 2: //管理员消息提醒
					if teamInfoType == int(Team.TeamMemberType_Tmt_Manager) || teamInfoType == int(Team.TeamMemberType_Tmt_Owner) {
						nc.logger.Warn("此用户设置了管理员信息提醒", zap.String("toUser", req.GetTo()))
						//pass

					} else {

						continue
					}
				case 3: //联系人提醒
				case 4: //所有消息均不提醒
					nc.logger.Warn("此用户设置了所有消息均不提醒", zap.String("toUser", req.GetTo()))
					continue
				}

				//构造转发消息数据
				if newSeq, err = redis.Uint64(redisConn.Do("INCR", fmt.Sprintf("userSeq:%s", toUser))); err != nil {
					nc.logger.Warn("redisConn INCR userSeq Error", zap.Error(err))
					continue
				}

				eRsp := &Msg.RecvMsgEventRsp{
					Scene:        req.GetScene(), //传输场景
					Type:         req.GetType(),  //消息类型
					Body:         req.GetBody(),  //不拆包，直接透传body给接收者
					From:         username,       //谁发的
					FromDeviceId: deviceID,       //哪个设备发的
					Recv:         teamID,         //接收方， 群id
					ServerMsgId:  msg.GetID(),    //服务器分配的消息ID
					Seq:          newSeq,         //消息序号，单个会话内自然递增, 这里是对targetUsername这个用户的通知序号
					Uuid:         req.GetUuid(),  //客户端分配的消息ID，SDK生成的消息id
					Time:         uint64(time.Now().UnixNano() / 1e6),
				}

				//发送消息到目标用户
				go nc.SendMsgToUser(eRsp, username, deviceID, toUser)
			}

		case Msg.MessageScene_MsgScene_P2P: //用户设备之间传输文件或消息
			toUser = username //必须是用户自己
			toDeviceID := req.GetToDeviceId()

			nc.logger.Debug("MessageScene_MsgScene_P2P", zap.String("toUser", toUser))

			//判断toUser的合法性以及是否封禁等
			userKey := fmt.Sprintf("userData:%s", toUser)
			userState, _ := redis.Int(redisConn.Do("HGET", userKey, "State"))

			if userState == 2 {
				nc.logger.Warn("此用户已被封号", zap.String("toUser", toUser))
				errorCode = LMCError.UserIsBlockedError
				goto COMPLETE
			}

			//构造转发消息数据
			if newSeq, err = redis.Uint64(redisConn.Do("INCR", fmt.Sprintf("userSeq:%s", toUser))); err != nil {
				nc.logger.Error("redisConn INCR userSeq Error", zap.Error(err))
				errorCode = LMCError.RedisError
				goto COMPLETE
			}

			eRsp := &Msg.RecvMsgEventRsp{
				Scene:        req.GetScene(), //传输场景
				Type:         req.GetType(),  //消息类型
				Body:         req.GetBody(),  //不拆包，直接透传body给接收者
				From:         username,       //谁发的消息
				FromDeviceId: deviceID,       //哪个设备发的
				Recv:         toUser,         //接收方
				ServerMsgId:  msg.GetID(),    //服务器分配的消息ID
				Seq:          newSeq,         //消息序号，单个会话内自然递增, 这里是对targetUsername这个用户的通知序号
				Uuid:         req.GetUuid(),  //客户端分配的消息ID，SDK生成的消息id
				Time:         uint64(time.Now().UnixNano() / 1e6),
			}
			data, _ := proto.Marshal(eRsp)

			//转发透传消息
			targetMsg := &models.Message{}
			curDeviceKey := fmt.Sprintf("DeviceJwtToken:%s", toDeviceID)
			curJwtToken, _ := redis.String(redisConn.Do("GET", curDeviceKey)) //每个设备都有自己的token
			nc.logger.Debug("Redis GET ", zap.String("curDeviceKey", curDeviceKey), zap.String("curJwtToken", curJwtToken))

			targetMsg.UpdateID()
			//构建消息路由, 第一个参数是要处理的业务类型，后端服务器处理完成后，需要用此来拼接topic: {businessTypeName.Frontend}
			targetMsg.BuildRouter("Msg", "", "Msg.Frontend")
			targetMsg.SetJwtToken(curJwtToken)
			targetMsg.SetUserName(toUser)     //发给自己
			targetMsg.SetDeviceID(toDeviceID) //发给哪个设备
			targetMsg.SetBusinessTypeName("Msg")
			targetMsg.SetBusinessType(uint32(Global.BusinessType_Msg))           //消息模块
			targetMsg.SetBusinessSubType(uint32(Global.MsgSubType_RecvMsgEvent)) //接收消息事件
			targetMsg.BuildHeader("ChatService", time.Now().UnixNano()/1e6)
			targetMsg.FillBody(data) //网络包的body，承载真正的业务数据
			targetMsg.SetCode(200)   //成功的状态码

			//构建数据完成，向dispatcher发送
			topic := "Msg.Frontend"
			rawData, _ := json.Marshal(targetMsg)
			go nc.Producer.Public(topic, rawData)

			nc.logger.Info("HandleRecvMsg Succeed",
				zap.String("Username:", username))
		}
	}

COMPLETE:
	msg.SetCode(int32(errorCode)) //状态码
	if errorCode == 200 {
		//构造回包消息数据
		if curSeq, err := redis.Uint64(redisConn.Do("INCR", fmt.Sprintf("userSeq:%s", username))); err != nil {
			nc.logger.Error("redisConn INCR userSeq Error", zap.Error(err))
			errorCode = LMCError.RedisError
			msg.SetCode(int32(errorCode))            //状态码
			errorMsg := LMCError.ErrorMsg(errorCode) //错误描述
			msg.SetErrorMsg([]byte(errorMsg))        //错误提示
			msg.FillBody(nil)
		} else {
			rsp = &Msg.SendMsgRsp{
				Uuid:        req.GetUuid(),
				ServerMsgId: msg.GetID(),
				Seq:         curSeq,
				Time:        uint64(time.Now().UnixNano() / 1e6),
			}
			data, _ = proto.Marshal(rsp)
			msg.FillBody(data)
		}

	} else {
		errorMsg := LMCError.ErrorMsg(errorCode) //错误描述
		msg.SetErrorMsg([]byte(errorMsg))        //错误提示
		msg.FillBody(nil)
	}

	//处理完成，向dispatcher发送
	topic := msg.GetSource() + ".Frontend"
	rawData, _ := json.Marshal(msg)
	if err := nc.Producer.Public(topic, rawData); err == nil {
		nc.logger.Info("SendMsgRsp message succeed send to ProduceChannel", zap.String("topic", topic))
	} else {
		nc.logger.Error("Failed to send SendMsgRsp message to ProduceChannel", zap.Error(err))
	}
	_ = err
	return nil
}

/*
5-4 确认消息送达 ACK
如果是系统通知，则需要删除
*/
func (nc *NsqClient) HandleMsgAck(msg *models.Message) error {
	var err error
	errorCode := 200

	var newSeq uint64

	//经过服务端更改状态后的新的OrderProductBody字节流
	var orderProductBodyData []byte

	redisConn := nc.redisPool.Get()
	defer redisConn.Close()

	username := msg.GetUserName()
	// token := msg.GetJwtToken()
	deviceID := msg.GetDeviceID()

	nc.logger.Info("HandleMsgAck start...",
		zap.String("username", username),
		zap.String("DeviceId", deviceID))

	//取出当前设备的os， logonAt
	curDeviceHashKey := fmt.Sprintf("devices:%s:%s", username, deviceID)
	isMaster, _ := redis.Bool(redisConn.Do("HGET", curDeviceHashKey, "ismaster"))
	curOs, _ := redis.String(redisConn.Do("HGET", curDeviceHashKey, "os"))
	curLogonAt, _ := redis.Uint64(redisConn.Do("HGET", curDeviceHashKey, "logonAt"))

	nc.logger.Debug("MsgAck",
		zap.Bool("isMaster", isMaster),
		zap.String("username", username),
		zap.String("deviceID", deviceID),
		zap.String("curOs", curOs),
		zap.Uint64("curLogonAt", curLogonAt))

	//打开msg里的负载， 获取请求参数
	body := msg.GetContent()
	//解包body
	var req Msg.MsgAckReq
	if err := proto.Unmarshal(body, &req); err != nil {
		errorCode = LMCError.ProtobufUnmarshalError
		nc.logger.Error("Protobuf Unmarshal Error", zap.Error(err))
		goto COMPLETE

	} else {

		//系统通知类型, 并且scene是S2C(??)
		//&& req.GetScene() == Msg.MessageScene_MsgScene_S2C
		if req.GetType() == Msg.MessageType_MsgType_Notification {
			nc.logger.Debug("Notification payload",
				zap.Int32("Scene", int32(req.GetScene())),
				zap.Int32("Type", int32(req.GetType())),
				zap.String("ServerMsgId", req.GetServerMsgId()),
				zap.Uint64("Seq", req.GetSeq()),
			)

			//删除缓存的哈希表数据
			systemMsgKey := fmt.Sprintf("systemMsg:%s:%s", username, req.GetServerMsgId())
			if _, err = redisConn.Do("DEL", systemMsgKey); err != nil {
				nc.logger.Error("删除缓存的哈希表数据 Error", zap.String("systemMsgKey", systemMsgKey), zap.Error(err))
			} else {
				nc.logger.Debug("删除缓存的哈希表数据 成功", zap.String("systemMsgKey", systemMsgKey))
			}

			//删除此用户的离线系统通知缓存有序集合里的成员
			offLineMsgListKey := fmt.Sprintf("offLineMsgList:%s", username)
			if _, err = redisConn.Do("ZREM", offLineMsgListKey, req.GetServerMsgId()); err != nil {
				nc.logger.Error("删除此用户的离线缓冲有序集合里的成员 Error", zap.String("offLineMsgListKey", offLineMsgListKey), zap.Error(err))
			} else {
				nc.logger.Debug("删除此用户的离线缓冲有序集合里的成员 成功", zap.String("offLineMsgListKey", offLineMsgListKey), zap.Error(err))
			}

		} else if req.GetType() == Msg.MessageType_MsgType_Order {
			nc.logger.Debug("Order订单类型消息的ACK",
				zap.Int32("Scene", int32(req.GetScene())),
				zap.Int32("Type", int32(req.GetType())),
				zap.String("ServerMsgId", req.GetServerMsgId()),
				zap.Uint64("Seq", req.GetSeq()))

			//从Redis里取出此 ServerMsgId 对应的订单信息及状态
			orderProductBodyKey := fmt.Sprintf("OrderProductBody:%s", req.GetServerMsgId())
			state, err := redis.Int(redisConn.Do("HGET", orderProductBodyKey, "State"))
			orderID, err := redis.String(redisConn.Do("HGET", orderProductBodyKey, "OrderID"))
			// productID, err := redis.String(redisConn.Do("HGET", orderProductBodyKey, "ProductID"))
			buyUser, err := redis.String(redisConn.Do("HGET", orderProductBodyKey, "BuyUser"))
			// opkBuyUser, err := redis.String(redisConn.Do("HGET", orderProductBodyKey, "OpkBuyUser"))
			// businessUser, err := redis.String(redisConn.Do("HGET", orderProductBodyKey, "BusinessUser"))
			// opkBusinessUser, err := redis.String(redisConn.Do("HGET", orderProductBodyKey, "OpkBusinessUser"))
			// orderTotalAmount, err := redis.Float64(redisConn.Do("HGET", orderProductBodyKey, "OrderTotalAmount"))
			// attach, err := redis.String(redisConn.Do("HGET", orderProductBodyKey, "Attach"))

			if err != nil {
				nc.logger.Error("从Redis里取出此 ServerMsgId 对应的订单信息及状态 Error", zap.String("orderProductBodyKey", orderProductBodyKey), zap.Error(err))
			} else {
				if state == 0 {
					nc.logger.Warn("从Redis里取出此 ServerMsgId 对应的订单信息及状态 state==0", zap.String("orderProductBodyKey", orderProductBodyKey), zap.Error(err))
					goto COMPLETE
				}

				//商户的订单状态为OrderState_OS_SendOK，向下单用户发送订单送达通知: OrderState_OS_RecvOK
				if Global.OrderState(state) == Global.OrderState_OS_SendOK {
					var orderProductBody = new(Order.OrderProductBody)
					orderProductBody.State = Global.OrderState_OS_RecvOK
					orderProductBody.OrderID = orderID
					//将redis里的订单信息哈希表状态字段设置为OS_RecvOK

					if newSeq, err = redis.Uint64(redisConn.Do("INCR", fmt.Sprintf("userSeq:%s", buyUser))); err != nil {
						nc.logger.Error("redisConn INCR userSeq Error", zap.Error(err))
						errorCode = LMCError.RedisError
						goto COMPLETE
					}

					orderProductBodyData, _ = proto.Marshal(orderProductBody)
					eRsp := &Msg.RecvMsgEventRsp{
						Scene:        Msg.MessageScene_MsgScene_S2C,      //系统消息
						Type:         Msg.MessageType_MsgType_Order,      //类型-订单消息
						Body:         orderProductBodyData,               //订单载体 OrderProductBody
						From:         username,                           //谁发的
						FromDeviceId: deviceID,                           //哪个设备发的
						Recv:         buyUser,                            //用户账户id
						ServerMsgId:  msg.GetID(),                        //服务器分配的消息ID
						Seq:          newSeq,                             //消息序号，单个会话内自然递增, 这里是对targetUsername这个用户的通知序号
						Uuid:         fmt.Sprintf("%d", msg.GetTaskID()), //客户端分配的消息ID，SDK生成的消息id，这里返回TaskID
						Time:         uint64(time.Now().UnixNano() / 1e6),
					}
					
					go nc.BroadcastSystemMsgToAllDevices(eRsp, buyUser) //向用户推送订单消息

				}
			}

		} else {
			nc.logger.Debug("MsgAck payload",
				zap.Int32("Scene", int32(req.GetScene())),
				zap.Int32("Type", int32(req.GetType())),
				zap.String("ServerMsgId", req.GetServerMsgId()),
				zap.Uint64("Seq", req.GetSeq()),
			)
		}

	}

COMPLETE:
	msg.SetCode(int32(errorCode)) //状态码
	if errorCode == 200 {
		//
		msg.FillBody(nil)
	} else {
		errorMsg := LMCError.ErrorMsg(errorCode) //错误描述
		msg.SetErrorMsg([]byte(errorMsg))        //错误提示
		msg.FillBody(nil)
	}

	//处理完成，向dispatcher发送
	topic := msg.GetSource() + ".Frontend"
	rawData, _ := json.Marshal(msg)
	if err := nc.Producer.Public(topic, rawData); err == nil {
		nc.logger.Info("MsgAck message succeed send to ProduceChannel", zap.String("topic", topic))
	} else {
		nc.logger.Error("Failed to send MsgAck message to ProduceChannel", zap.Error(err))
	}
	_ = err
	return nil

}

//5-9 发送撤销消息 的处理
func (nc *NsqClient) HandleSendCancelMsg(msg *models.Message) error {
	var err error
	var data []byte
	errorCode := 200

	var isExists bool
	// var newSeq uint64

	redisConn := nc.redisPool.Get()
	defer redisConn.Close()

	username := msg.GetUserName()
	// token := msg.GetJwtToken()
	deviceID := msg.GetDeviceID()

	nc.logger.Info("HandleSendCancelMsg start...",
		zap.String("username", username),
		zap.String("DeviceId", deviceID))

	//取出当前设备的os， logonAt
	curDeviceHashKey := fmt.Sprintf("devices:%s:%s", username, deviceID)
	isMaster, _ := redis.Bool(redisConn.Do("HGET", curDeviceHashKey, "ismaster"))
	curOs, _ := redis.String(redisConn.Do("HGET", curDeviceHashKey, "os"))
	curLogonAt, _ := redis.Uint64(redisConn.Do("HGET", curDeviceHashKey, "logonAt"))

	nc.logger.Debug("SendCancelMsg",
		zap.Bool("isMaster", isMaster),
		zap.String("username", username),
		zap.String("deviceID", deviceID),
		zap.String("curOs", curOs),
		zap.Uint64("curLogonAt", curLogonAt))

	//打开msg里的负载， 获取请求参数
	body := msg.GetContent()
	//解包body
	var req Msg.SendCancelMsgReq
	if err := proto.Unmarshal(body, &req); err != nil {
		errorCode = LMCError.ProtobufUnmarshalError
		nc.logger.Error("Protobuf Unmarshal Error", zap.Error(err))
		goto COMPLETE

	} else {
		nc.logger.Debug("SendCancelMsg payload",
			zap.Int32("Scene", int32(req.GetScene())),
			zap.Int32("Type", int32(req.GetType())),
			zap.String("From", req.GetFrom()),
			zap.String("To", req.GetTo()),
			zap.String("ServerMsgId", req.GetServerMsgId()),
		)

		//查询出谁接收了此消息，如果超过1分钟，则无法撤销
		recvKey := fmt.Sprintf("recvMsgList:%s", req.GetServerMsgId())
		if isExists, err = redis.Bool(redisConn.Do("EXISTS", recvKey)); err != nil {
			nc.logger.Error("EXISTS Error", zap.Error(err))
			errorCode = LMCError.RedisError
			goto COMPLETE
		}

		if isExists {
			recvUsers, _ := redis.Strings(redisConn.Do("ZRANGEBYSCORE", recvKey, "-inf", "+inf"))
			eRsp := &Msg.RecvCancelMsgEventRsp{
				Scene:       req.GetScene(),
				Type:        req.GetType(),
				From:        req.GetFrom(),
				To:          req.GetTo(),
				ServerMsgId: req.GetServerMsgId(), //要撤销的消息ID
			}
			data, _ = proto.Marshal(eRsp)
			//查询出用户的当前在线所有主从设备
			for _, recvUser := range recvUsers {
				//此消息的接收人, 需要将消息从接收人的缓存队列删除，并下发撤销通知, 如果是群消息，会有很多接收人
				//向接收此消息的用户发出撤销消息 RecvCancelMsgEvent
				// if newSeq, err = redis.Uint64(redisConn.Do("INCR", fmt.Sprintf("userSeq:%s", recvUser))); err != nil {
				// 	nc.logger.Error("redisConn INCR userSeq Error", zap.Error(err))
				// }

				go nc.SendDataToUserDevices(
					data,
					recvUser,
					uint32(Global.BusinessType_Msg), //消息模块
					uint32(Global.MsgSubType_RecvCancelMsgEvent), //接收撤销消息事件
				)

			}

		}

		//删除recvKey
		_, err = redisConn.Do("DEL", recvKey)

	}

COMPLETE:
	msg.SetCode(int32(errorCode)) //状态码
	if errorCode == 200 {
		//
		msg.FillBody(nil)
	} else {
		errorMsg := LMCError.ErrorMsg(errorCode) //错误描述
		msg.SetErrorMsg([]byte(errorMsg))        //错误提示
		msg.FillBody(nil)
	}

	//处理完成，向dispatcher发送
	topic := msg.GetSource() + ".Frontend"
	rawData, _ := json.Marshal(msg)
	if err := nc.Producer.Public(topic, rawData); err == nil {
		nc.logger.Info("SendCancelMsg message succeed send to ProduceChannel", zap.String("topic", topic))
	} else {
		nc.logger.Error("Failed to send SendCancelMsg message to ProduceChannel", zap.Error(err))
	}
	_ = err
	return nil
}

//5-12 获取阿里云OSS上传Token
func (nc *NsqClient) HandleGetOssToken(msg *models.Message) error {
	var err error
	errorCode := 200

	rsp := &Msg.GetOssTokenRsp{}

	redisConn := nc.redisPool.Get()
	defer redisConn.Close()

	username := msg.GetUserName()
	deviceID := msg.GetDeviceID()

	nc.logger.Info("HandleGetOssToken start...",
		zap.String("username", username),
		zap.String("DeviceId", deviceID))

	//取出当前设备的os， logonAt
	curDeviceHashKey := fmt.Sprintf("devices:%s:%s", username, deviceID)
	isMaster, _ := redis.Bool(redisConn.Do("HGET", curDeviceHashKey, "ismaster"))
	curOs, _ := redis.String(redisConn.Do("HGET", curDeviceHashKey, "os"))
	curLogonAt, _ := redis.Uint64(redisConn.Do("HGET", curDeviceHashKey, "logonAt"))

	nc.logger.Debug("GetOssToken",
		zap.Bool("isMaster", isMaster),
		zap.String("username", username),
		zap.String("deviceID", deviceID),
		zap.String("curOs", curOs),
		zap.Uint64("curLogonAt", curLogonAt))

	//打开msg里的负载， 获取请求参数
	body := msg.GetContent()
	//解包body
	var req Msg.GetOssTokenReq
	if err := proto.Unmarshal(body, &req); err != nil {
		nc.logger.Error("Protobuf Unmarshal Error", zap.Error(err))
		errorCode = LMCError.ProtobufUnmarshalError
		goto COMPLETE

	} else {
		nc.logger.Debug("GetOssToken payload", zap.Bool("IsPrivate", req.IsPrivate))

		var client *sts.AliyunStsClient
		var url string

		client = sts.NewStsClient(LMCommon.AccessID, LMCommon.AccessKey, LMCommon.RoleAcs)
		/*

			//仅允许用户向lianmi-ipfs这个Bucket上传类似users/{username}/的文件
			acs := fmt.Sprintf("acs:oss:*:*:lianmi-ipfs/users/%s/*", username)
			nc.logger.Debug("acs", zap.String("acs", acs))

			policy := sts.Policy{
				Version: "1",
				Statement: []sts.StatementBase{sts.StatementBase{
					Effect: "Allow",
					// Action:   []string{"oss:GetObject", "oss:ListObjects", "oss:PutObject", "oss:AbortMultipartUpload"},
					Action:   []string{"oss:ListObjects", "oss:PutObject", "oss:AbortMultipartUpload"},
					Resource: []string{acs},
				}},
			}

			//300秒
			url, err = client.GenerateSignatureUrl("client", fmt.Sprintf("%d", LMCommon.PrivateEXPIRESECONDS), policy.ToJson())
			if err != nil {
				nc.logger.Error("GenerateSignatureUrl Error", zap.Error(err))
				errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
				errorMsg = fmt.Sprintf("GenerateSignatureUrl Error: %s", err.Error())
				goto COMPLETE
			}

			nc.logger.Debug("url", zap.String("url", url))


		*/

		//生成阿里云oss临时sts, Policy是对lianmi-ipfs这个bucket下的 avatars, generalavatars, msg, products, orders, stores, teamicons, 目录有可读写权限
		acsAvatars := fmt.Sprintf("acs:oss:*:*:lianmi-ipfs/avatars/%s/*", username)
		acsGeneralavatars := fmt.Sprintf("acs:oss:*:*:lianmi-ipfs/generalavatars/%s/*", username)
		acsMsg := fmt.Sprintf("acs:oss:*:*:lianmi-ipfs/msg/%s/*", username)
		acsMsgs := fmt.Sprintf("acs:oss:*:*:lianmi-ipfs/msgs/%s/*", username)
		acsProducts := fmt.Sprintf("acs:oss:*:*:lianmi-ipfs/products/%s/*", username)
		acsStores := fmt.Sprintf("acs:oss:*:*:lianmi-ipfs/stores/%s/*", username)
		acsOrders := fmt.Sprintf("acs:oss:*:*:lianmi-ipfs/orders/%s/*", username)
		acsTeamIcons := fmt.Sprintf("acs:oss:*:*:lianmi-ipfs/teamicons/%s/*", username)
		acsUsers := fmt.Sprintf("acs:oss:*:*:lianmi-ipfs/users/%s/*", username)

		// Policy是对lianmi-ipfs这个bucket下的user目录有可读写权限
		policy := sts.Policy{
			Version: "1",
			Statement: []sts.StatementBase{sts.StatementBase{
				Effect: "Allow",
				Action: []string{"oss:GetObject", "oss:ListObjects", "oss:PutObject", "oss:PutObjectRequest", "oss:AbortMultipartUpload"},
				// Action: []string{"oss:*"},  // 开放所有权限
				Resource: []string{acsAvatars, acsGeneralavatars, acsMsg, acsMsgs, acsProducts, acsStores, acsOrders, acsTeamIcons, acsUsers},
				// Resource: []string{"acs:oss:*:*:lianmi-ipfs", "acs:oss:*:*:lianmi-ipfs/*"},
			}},
		}

		url, err = client.GenerateSignatureUrl("client", fmt.Sprintf("%d", LMCommon.EXPIRESECONDS), policy.ToJson())
		if err != nil {
			nc.logger.Error("GenerateSignatureUrl Error", zap.Error(err))
			errorCode = LMCError.GenerateSignatureUrlError
			goto COMPLETE
		}

		data, err := client.GetStsResponse(url)
		if err != nil {
			nc.logger.Error("阿里云oss GetStsResponse Error", zap.Error(err))
			errorCode = LMCError.GenerateSignatureUrlError
			goto COMPLETE
		}

		// log.Println("result:", string(data))
		sjson, err := simpleJson.NewJson(data)
		if err != nil {
			nc.logger.Warn("simplejson.NewJson Error", zap.Error(err))
			errorCode = LMCError.GenerateSignatureUrlError
			goto COMPLETE
		}

		nc.logger.Debug("收到阿里云OSS服务端的回包",
			zap.String("RequestId", sjson.Get("RequestId").MustString()),
			zap.String("AccessKeyId", sjson.Get("Credentials").Get("AccessKeyId").MustString()),
			zap.String("AccessKeySecret", sjson.Get("Credentials").Get("AccessKeySecret").MustString()),
			zap.String("SecurityToken", sjson.Get("Credentials").Get("SecurityToken").MustString()),
			zap.String("Expiration", sjson.Get("Credentials").Get("Expiration").MustString()),
		)

		//计算出Expire
		dt, _ := time.Parse("2006-01-02T15:04:05Z", sjson.Get("Credentials").Get("Expiration").MustString())
		format := "2006-01-02T15:04:05Z"
		now, _ := time.Parse(format, time.Now().Format(format))
		expire := uint64(dt.Unix()-now.UTC().Unix()+8*3600) * 1000

		rsp = &Msg.GetOssTokenRsp{
			EndPoint:        LMCommon.Endpoint,
			BucketName:      LMCommon.BucketName,
			AccessKeyId:     sjson.Get("Credentials").Get("AccessKeyId").MustString(),
			AccessKeySecret: sjson.Get("Credentials").Get("AccessKeySecret").MustString(),
			SecurityToken:   sjson.Get("Credentials").Get("SecurityToken").MustString(),
			Expiration:      sjson.Get("Credentials").Get("Expiration").MustString(),
			Directory:       time.Now().Format("2006/01/02/"),
			Expire:          expire, //毫秒
			Callback:        "",     //不填
		}
	}

COMPLETE:
	msg.SetCode(int32(errorCode)) //状态码
	if errorCode == 200 {
		data, _ := proto.Marshal(rsp)
		msg.FillBody(data) //网络包的body，承载真正的业务数据
	} else {
		errorMsg := LMCError.ErrorMsg(errorCode) //错误描述
		msg.SetErrorMsg([]byte(errorMsg))        //错误提示
		msg.FillBody(nil)
	}

	//处理完成，向dispatcher发送
	topic := msg.GetSource() + ".Frontend"
	rawData, _ := json.Marshal(msg)
	if err := nc.Producer.Public(topic, rawData); err == nil {
		nc.logger.Info("SendCancelMsg message succeed send to ProduceChannel", zap.String("topic", topic))
	} else {
		nc.logger.Error("Failed to send SendCancelMsg message to ProduceChannel", zap.Error(err))
	}
	_ = err
	return nil
}

/*
发送消息到目标用户
redis里用recvMsgList有序集合缓存，当接收方回复ACK后，就要删除这个集合里的对应成员
*/
func (nc *NsqClient) SendMsgToUser(rsp *Msg.RecvMsgEventRsp, fromUser, fromDeviceID, toUser string) error {
	var err error
	data, _ := proto.Marshal(rsp)

	redisConn := nc.redisPool.Get()
	defer redisConn.Close()

	//Redis里缓存此消息,目的是用户从离线状态恢复到上线状态后同步这些消息给用户
	msgAt := time.Now().UnixNano() / 1e6

	//有序集合存储哪些用户接收了此消息，以便撤销
	recvKey := fmt.Sprintf("recvMsgList:%s", rsp.GetServerMsgId())
	if _, err := redisConn.Do("ZADD", recvKey, msgAt, toUser); err != nil {
		nc.logger.Error("ZADD Error", zap.Error(err))
	}
	_, err = redisConn.Do("EXPIRE", recvKey, LMCommon.SMSEXPIRE) //设置有效期

	//向接收者toUser发送
	deviceListKey := fmt.Sprintf("devices:%s", toUser)
	eDeviceID, _ := redis.String(redisConn.Do("GET", deviceListKey))

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
	rawData, _ := json.Marshal(targetMsg)
	if err := nc.Producer.Public(topic, rawData); err == nil {
		nc.logger.Info("Msg message succeed send to ProduceChannel",
			zap.String("topic", topic),
			zap.String("toUser", toUser),
			zap.String("toDeviceID:", curDeviceKey),
			zap.String("msgID:", rsp.GetServerMsgId()),
			zap.Uint64("seq", rsp.Seq),
		)
	} else {
		nc.logger.Error("Failed to send message to ProduceChannel",
			zap.String("topic", topic),
			zap.String("toUser", toUser),
			zap.String("toDeviceID:", curDeviceKey),
			zap.String("msgID:", rsp.GetServerMsgId()),
			zap.Uint64("seq", rsp.Seq),
			zap.Error(err),
		)
	}

	_ = err
	return nil
}

/*
向目标用户账号的所有端发送消息。
传参：
序化后的 二进制数据: data
目标用户： toUser
业务号： businessType
业务子号： businessSubType
*/
func (nc *NsqClient) SendDataToUserDevices(data []byte, toUser string, businessType, businessSubType uint32) error {

	redisConn := nc.redisPool.Get()
	defer redisConn.Close()

	deviceListKey := fmt.Sprintf("devices:%s", toUser)
	eDeviceID, _ := redis.String(redisConn.Do("GET", deviceListKey))

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
	targetMsg.SetBusinessType(businessType)       //业务号
	targetMsg.SetBusinessSubType(businessSubType) //业务子号

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

	nc.logger.Info("SendDataToUserDevices Succeed",
		zap.String("Username:", toUser),
		zap.String("DeviceID:", curDeviceKey),
		zap.Int64("Now", time.Now().UnixNano()/1e6))

	return nil
}

/*
判断会话双方的任意一方是不是客服，如果是，则返回true，如果两者都不是，则返回false
客服账号，用redis的有序集合存储， CustomerServiceList,
*/
func (nc *NsqClient) CheckIsCustomerService(username, toUser string) bool {
	redisConn := nc.redisPool.Get()
	defer redisConn.Close()

	if reply, err := redisConn.Do("ZRANK", "CustomerServiceList", username); err == nil {
		if reply != nil {
			//客服账号有序集合中有username
			return true
		}

	}
	if reply, err := redisConn.Do("ZRANK", "CustomerServiceList", toUser); err == nil {
		if reply != nil {
			//客服账号有序集合中有toUser
			return true
		}

	}
	return false
}

/*
向目标用户账号的所有端推送消息， 接收端会触发接收消息事件
业务号:  BusinessType_Msg(5)
业务子号:  MsgSubType_RecvMsgEvent(2)
*/
func (nc *NsqClient) BroadcastSystemMsgToAllDevices(rsp *Msg.RecvMsgEventRsp, toUser string) error {
	data, _ := proto.Marshal(rsp)

	redisConn := nc.redisPool.Get()
	defer redisConn.Close()
	/*
		//一次性删除7天前的缓存系统消息
		nTime := time.Now()
		yesTime := nTime.AddDate(0, 0, -7).Unix()

		offLineMsgListKey := fmt.Sprintf("offLineMsgList:%s", toUser)

		_, err := redisConn.Do("ZREMRANGEBYSCORE", offLineMsgListKey, "-inf", yesTime)

		//Redis里缓存此消息,目的是用户从离线状态恢复到上线状态后同步这些系统消息给用户
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
			"Data", data, //系统消息的数据体
		)

		_, err = redisConn.Do("EXPIRE", systemMsgKey, 7*24*3600) //设置有效期为7天
	*/

	//向toUser所有端发送
	deviceListKey := fmt.Sprintf("devices:%s", toUser)
	eDeviceID, _ := redis.String(redisConn.Do("GET", deviceListKey))

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
		nc.logger.Info("Message succeed send to ProduceChannel", zap.String("topic", topic))
	} else {
		nc.logger.Error("Failed to send message to ProduceChannel", zap.Error(err))
	}

	nc.logger.Info("Broadcast Msg To All Devices Succeed",
		zap.String("Username:", toUser),
		zap.String("DeviceID:", curDeviceKey),
		zap.Int64("Now", time.Now().UnixNano()/1e6))

	return nil
}
