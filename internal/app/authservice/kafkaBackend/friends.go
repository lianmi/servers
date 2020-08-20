/*
本文件是处理业务号是好友模块，分别有
3-1 好友请求发起与处理 FriendRequest 未完成
3-2 好友关系变更事件 FriendChangeEvent 未完成
3-3 好友列表同步事件 未完成
3-4 好友资料同步事件 未完成
3-5 移除好友 未完成
3-6 刷新好友资料 未完成
3-7 主从设备好友资料同步事件 未完成
3-8 增量同步好友列表 未完成
*/
package kafkaBackend

import (
	"time"
	// "encoding/hex"
	"fmt"
	// "strings"

	"github.com/golang/protobuf/proto"
	"github.com/gomodule/redigo/redis"

	// User "github.com/lianmi/servers/api/proto/user"
	Friends "github.com/lianmi/servers/api/proto/friends"
	"github.com/lianmi/servers/internal/common"
	"github.com/lianmi/servers/internal/pkg/models"

	uuid "github.com/satori/go.uuid"
	"go.uber.org/zap"
)

/*
3-1 好友请求发起与处理
注意：
1. Alice加Bob, 先判断Bob是否允许加好友
2. 服务端利用redis的哈希表，保存Alice加Bob的状态，当Bob同意或拒绝后，才进行入库及更新Alice的好友表
3. 要考虑到多端的环境，交互的动作可以在任一端进行，结果需要同步给其他端
4. 以有序集合存储之间的系统通知， 当已经有了最终结果后，这个有序集合就会只保留最后一个结果，
   如果长时间离线再重新上线的其他端，会收到最后一个结果，而不会重现整个交互流程。
*/
func (kc *KafkaClient) HandleFriendRequest(msg *models.Message) error {
	var err error
	var errorMsg string
	var data []byte

	var isAhaveB, isBhaveA bool //A好友列表里有B， B好友列表里有A
	var allowType int

	redisConn := kc.redisPool.Get()
	defer redisConn.Close()

	username := msg.GetUserName() //用户自己的账号
	// token := msg.GetJwtToken()
	deviceID := msg.GetDeviceID()

	kc.logger.Info("HandleFriendRequest start...",
		zap.String("username", username),
		zap.String("deviceId", deviceID))

	//取出当前设备的os， clientType， logonAt
	curDeviceHashKey := fmt.Sprintf("devices:%s:%s", username, deviceID)
	isMaster, _ := redis.Bool(redisConn.Do("HGET", curDeviceHashKey, "ismaster"))
	curOs, _ := redis.String(redisConn.Do("HGET", curDeviceHashKey, "os"))
	curClientType, _ := redis.Int(redisConn.Do("HGET", curDeviceHashKey, "clientType"))
	curLogonAt, _ := redis.Uint64(redisConn.Do("HGET", curDeviceHashKey, "logonAt"))

	kc.logger.Debug("FriendRequest",
		zap.Bool("isMaster", isMaster),
		zap.String("username", username),
		zap.String("deviceID", deviceID),
		zap.String("curOs", curOs),
		zap.Int("curClientType", curClientType),
		zap.Uint64("curLogonAt", curLogonAt))

	//打开msg里的负载， 获取请求参数
	body := msg.GetContent()

	//解包body
	req := &Friends.FriendRequestReq{}
	if err := proto.Unmarshal(body, req); err != nil {
		kc.logger.Error("Protobuf Unmarshal Error", zap.Error(err))
		errorMsg = fmt.Sprintf("Protobuf Unmarshal Error: %s", err.Error())
		goto COMPLETE

	} else {
		kc.logger.Debug("FriendRequest body",
			zap.String("Username", req.GetUsername()),
			zap.String("Ps", req.GetPs()),
			zap.String("Source", req.GetSource()),
			zap.Int("Type", int(req.GetType())))

		targetUser := req.GetUsername() //要加好友的对方的用户账号

		//查出 targetUser 有效性，是否已经是好友，好友增加的设置等信息

		targetKey := fmt.Sprintf("userData:%s", targetUser)
		targetUserData := new(models.User)

		//检测目标用户是否存在及添加好友的设定
		isExists, _ := redis.Bool(redisConn.Do("EXISTS", targetKey))
		if !isExists {
			if err = kc.db.Model(targetUserData).Where("username = ?", targetUser).First(targetUserData).Error; err != nil {
				kc.logger.Error("MySQL里读取错误", zap.Error(err))
				errorMsg = fmt.Sprintf("Query user error[username=%s]", targetUser)
				goto COMPLETE
			}
			if _, err := redisConn.Do("HMSET", redis.Args{}.Add(targetKey).AddFlat(targetUserData)...); err != nil {
				kc.logger.Error("错误：HMSET", zap.Error(err))
			} else {
				kc.logger.Debug("刷新Redis的用户数据成功", zap.String("targetUser", targetUser))
			}

			allowType = targetUserData.AllowType

		} else {
			allowType, _ = redis.Int(redisConn.Do("HGET", targetKey, "AllowType"))
		}

		if reply, err := redisConn.Do("ZRANK", fmt.Sprintf("Friend:%s:1", username), targetUserData.Username); err == nil {
			if reply == nil {
				//A好友列表中没有B
				isAhaveB = false
			} else {
				isAhaveB = true
			}

		}
		if reply, err := redisConn.Do("ZRANK", fmt.Sprintf("Friend:%s:1", targetUserData.Username), username); err == nil {
			if reply == nil {
				//B好友列表中没有A
				isBhaveA = false
			} else {
				isBhaveA = true
			}

		}

		rsp := &Friends.FriendRequestRsp{}

		switch Friends.OptType(req.GetType()) {
		case Friends.OptType_Fr_ApplyFriend: //发起好友验证
			{
				//拒绝任何人添加好友
				if allowType == common.DenyAny {

					rsp.Status = Friends.OpStatusType_Ost_RejectFriendApply

				} else if allowType == common.AllowAny {

					rsp.Status = Friends.OpStatusType_Ost_ApplySucceed

					//在A的预审核好友列表里删除B ZREM
					if _, err = redisConn.Do("ZREM", fmt.Sprintf("Friend:%s:0", username), targetUserData.Username); err != nil {
						kc.logger.Error("ZREM Error", zap.Error(err))
					}

					//在A的移除好友列表里删除B ZREM
					if _, err = redisConn.Do("ZREM", fmt.Sprintf("Friend:%s:2", username), targetUserData.Username); err != nil {
						kc.logger.Error("ZREM Error", zap.Error(err))
					}

					//在B的移除好友列表里删除A ZREM
					if _, err = redisConn.Do("ZREM", fmt.Sprintf("Friend:%s:2", targetUserData.Username), username); err != nil {
						kc.logger.Error("ZREM Error", zap.Error(err))
					}

					//直接让双方成为好友
					if _, err = redisConn.Do("ZADD", fmt.Sprintf("Friend:%s:1", username), time.Now().Unix(), targetUserData.Username); err != nil {
						kc.logger.Error("ZADD Error", zap.Error(err))
					}
					if _, err = redisConn.Do("ZADD", fmt.Sprintf("Friend:%s:1", targetUserData.Username), time.Now().Unix(), username); err != nil {
						kc.logger.Error("ZADD Error", zap.Error(err))
					}
					//增加A的好友B的信息哈希表
					//HMSET FriendInfo:{A}:{B} username {username} nick {nick} source {source} ex {ex} createAt {createAt} updateAt {updateAt}
					nick, _ := redis.String(redisConn.Do("HGET", fmt.Sprintf("userData:%s", targetUserData.Username), "Nick"))
					_, err = redisConn.Do("HMSET",
						fmt.Sprintf("FriendInfo:%s:%s", username, targetUserData.Username),
						"Username", targetUserData.Username,
						"Nick", nick,
						"Source", req.GetSource(),
						"Ex", "", //TODO
						"CreateAt", uint64(time.Now().UnixNano()/1e6),
						"UpdateAt", uint64(time.Now().UnixNano()/1e6),
					)

					//增加B的好友A的信息哈希表
					//HMSET FriendInfo:{B}:{A} username {username} nick {nick} source {source} ex {ex} createAt {createAt} updateAt {updateAt}
					nick, _ = redis.String(redisConn.Do("HGET", fmt.Sprintf("userData:%s", username), "Nick"))
					_, err = redisConn.Do("HMSET",
						fmt.Sprintf("FriendInfo:%s:%s", targetUserData.Username, username),
						"Username", username,
						"Nick", nick,
						"Source", req.GetSource(),
						"Ex", "", //TODO
						"CreateAt", uint64(time.Now().UnixNano()/1e6),
						"UpdateAt", uint64(time.Now().UnixNano()/1e6),
					)

					//写入MySQL, 需要增加两条记录
					{

						userID, _ := redis.Int(redisConn.Do("HGET", fmt.Sprintf("userData:%s", username), "ID"))

						pFriendA := new(models.Friend)
						pFriendA.UserID = uint64(userID)
						pFriendA.FriendUserID = targetUserData.ID
						pFriendA.FriendUsername = targetUserData.Username
						if err := kc.SaveAddFriend(pFriendA); err != nil {
							kc.logger.Error("Save Add Friend Error", zap.Error(err))
							errorMsg = "无法保存到数据库"
							goto COMPLETE
						}

						userID, _ = redis.Int(redisConn.Do("HGET", fmt.Sprintf("userData:%s", targetUserData.Username), "ID"))

						pFriendB := new(models.Friend)
						pFriendB.UserID = uint64(userID)
						pFriendB.FriendUserID = targetUserData.ID
						pFriendB.FriendUsername = targetUserData.Username
						if err := kc.SaveAddFriend(pFriendB); err != nil {
							kc.logger.Error("Save Add Friend Error", zap.Error(err))
							errorMsg = "无法保存到数据库"
							goto COMPLETE
						}

					}

					//下发通知给A所有端(From是B，To是A)
					{
						//构造回包里的数据
						eRsp := &Friends.FriendChangeEventRsp{
							Uuid:   uuid.NewV4().String(),
							Type:   Friends.FriendChangeType_Fc_PassFriendApply, //通过加好友请求
							From:   targetUserData.Username,
							To:     username,
							Ps:     req.GetPs(),
							Source: req.GetSource(),
							TimeAt: uint64(time.Now().Unix()),
						}

						data, _ = proto.Marshal(eRsp)
						go kc.BroadcastMsgToAllDevices(data, username)
					}

					//下发通知给B所有端(From是A，To是B)
					{
						//构造回包里的数据
						eRsp := &Friends.FriendChangeEventRsp{
							Uuid:   uuid.NewV4().String(),
							Type:   Friends.FriendChangeType_Fc_PassFriendApply, //通过加好友请求
							From:   username,
							To:     targetUserData.Username,
							Ps:     req.GetPs(),
							Source: req.GetSource(),
							TimeAt: uint64(time.Now().Unix()),
						}

						data, _ = proto.Marshal(eRsp)
						go kc.BroadcastMsgToAllDevices(data, targetUserData.Username)
					}

				} else if allowType == common.NeedConfirm {
					//redis里增加A的预审核好友列表
					if _, err = redisConn.Do("ZADD", fmt.Sprintf("Friend:%s:0", username), time.Now().Unix(), targetUserData.Username); err != nil {
						kc.logger.Error("ZADD Error", zap.Error(err))
					}

					rsp.Status = Friends.OpStatusType_Ost_WaitConfirm

					//构造回包里的数据
					eRsp := &Friends.FriendChangeEventRsp{
						Uuid:   uuid.NewV4().String(),
						Type:   Friends.FriendChangeType_Fc_ApplyFriend,
						From:   username,
						To:     targetUserData.Username,
						Ps:     req.GetPs(),
						Source: req.GetSource(),
						TimeAt: uint64(time.Now().Unix()),
					}

					data, _ = proto.Marshal(eRsp)
					//A和B互相不为好友，B所有终端均会收到该消息。
					if !isAhaveB && !isBhaveA {
						//Go程，下发系统通知给B
						go kc.BroadcastMsgToAllDevices(data, targetUserData.Username)
					}

					//A好友列表中有B，B好友列表没有A，A发起好友申请，B所有终端均会接收该消息，并且B可以选择同意、拒绝
					if isAhaveB && !isBhaveA {
						//Go程，下发系统通知给B
						go kc.BroadcastMsgToAllDevices(data, targetUserData.Username)
					}

				}

			}
		case Friends.OptType_Fr_PassFriendApply: //对方同意加你为好友
			{
				rsp.Status = Friends.OpStatusType_Ost_ApplySucceed

				//在A的预审核好友列表里删除B ZREM
				if _, err = redisConn.Do("ZREM", fmt.Sprintf("Friend:%s:0", username), targetUserData.Username); err != nil {
					kc.logger.Error("ZREM Error", zap.Error(err))
				}

				//在A的移除好友列表里删除B ZREM
				if _, err = redisConn.Do("ZREM", fmt.Sprintf("Friend:%s:2", username), targetUserData.Username); err != nil {
					kc.logger.Error("ZREM Error", zap.Error(err))
				}

				//在B的移除好友列表里删除A ZREM
				if _, err = redisConn.Do("ZREM", fmt.Sprintf("Friend:%s:2", targetUserData.Username), username); err != nil {
					kc.logger.Error("ZREM Error", zap.Error(err))
				}

				//让双方成为好友
				if _, err = redisConn.Do("ZADD", fmt.Sprintf("Friend:%s:1", username), time.Now().Unix(), targetUserData.Username); err != nil {
					kc.logger.Error("ZADD Error", zap.Error(err))
				}
				if _, err = redisConn.Do("ZADD", fmt.Sprintf("Friend:%s:1", targetUserData.Username), time.Now().Unix(), username); err != nil {
					kc.logger.Error("ZADD Error", zap.Error(err))
				}

				//增加A的好友B的信息哈希表
				//HMSET FriendInfo:{A}:{B} username {username} nick {nick} source {source} ex {ex} createAt {createAt} updateAt {updateAt}
				nick, _ := redis.Int(redisConn.Do("HGET", fmt.Sprintf("userData:%s", targetUserData.Username), "Nick"))
				_, err = redisConn.Do("HMSET",
					fmt.Sprintf("FriendInfo:%s:%s", username, targetUserData.Username),
					"Username", targetUserData.Username,
					"Nick", nick,
					"Source", req.GetSource(),
					"Ex", "", //TODO
					"CreateAt", uint64(time.Now().UnixNano()/1e6),
					"UpdateAt", uint64(time.Now().UnixNano()/1e6),
				)

				//增加B的好友A的信息哈希表
				//HMSET FriendInfo:{B}:{A} username {username} nick {nick} source {source} ex {ex} createAt {createAt} updateAt {updateAt}
				nick, _ = redis.Int(redisConn.Do("HGET", fmt.Sprintf("userData:%s", username), "Nick"))
				_, err = redisConn.Do("HMSET",
					fmt.Sprintf("FriendInfo:%s:%s", targetUserData.Username, username),
					"Username", username,
					"Nick", nick,
					"Source", req.GetSource(),
					"Ex", "", //TODO
					"CreateAt", uint64(time.Now().UnixNano()/1e6),
					"UpdateAt", uint64(time.Now().UnixNano()/1e6),
				)
				//写入数据库，增加两条记录
				{

					userKey := fmt.Sprintf("userData:%s", username)
					userID, _ := redis.Int(redisConn.Do("HGET", userKey, "ID"))

					pFriendA := new(models.Friend)
					pFriendA.UserID = uint64(userID)
					pFriendA.FriendUserID = targetUserData.ID
					pFriendA.FriendUsername = targetUserData.Username
					if err := kc.SaveAddFriend(pFriendA); err != nil {
						kc.logger.Error("Save Add Friend Error", zap.Error(err))
						errorMsg = "无法保存到数据库"
						goto COMPLETE
					}

					userKey = fmt.Sprintf("userData:%s", targetUserData.Username)
					userID, _ = redis.Int(redisConn.Do("HGET", userKey, "ID"))

					pFriendB := new(models.Friend)
					pFriendB.UserID = uint64(userID)
					pFriendB.FriendUserID = targetUserData.ID
					pFriendB.FriendUsername = targetUserData.Username
					if err := kc.SaveAddFriend(pFriendB); err != nil {
						kc.logger.Error("Save Add Friend Error", zap.Error(err))
						errorMsg = "无法保存到数据库"
						goto COMPLETE
					}

				}

				//A发起好友申请, A好友列表没有B, B好友列表中有A, B会收到A好友通过系统通知，A不接收好友申请系统通知。
				//下发通知给A所有端(From是B，To是A)
				if !(!isAhaveB && isBhaveA) {
					//构造回包里的数据
					eRsp := &Friends.FriendChangeEventRsp{
						Uuid:   uuid.NewV4().String(),
						Type:   Friends.FriendChangeType_Fc_PassFriendApply, //通过加好友请求
						From:   targetUserData.Username,
						To:     username,
						Ps:     req.GetPs(),
						Source: req.GetSource(),
						TimeAt: uint64(time.Now().Unix()),
					}

					data, _ = proto.Marshal(eRsp)
					go kc.BroadcastMsgToAllDevices(data, username)
				}

				//下发通知给B所有端(From是A，To是B)
				{
					//构造回包里的数据
					eRsp := &Friends.FriendChangeEventRsp{
						Uuid:   uuid.NewV4().String(),
						Type:   Friends.FriendChangeType_Fc_PassFriendApply, //通过加好友请求
						From:   username,
						To:     targetUserData.Username,
						Ps:     req.GetPs(),
						Source: req.GetSource(),
						TimeAt: uint64(time.Now().Unix()),
					}

					data, _ = proto.Marshal(eRsp)
					go kc.BroadcastMsgToAllDevices(data, targetUserData.Username)
				}

			}
		case Friends.OptType_Fr_RejectFriendApply: //对方拒绝添加好友
			{
				rsp.Status = Friends.OpStatusType_Ost_RejectFriendApply

				//在A的预审核好友列表里删除B ZREM
				if _, err = redisConn.Do("ZREM", fmt.Sprintf("Friend:%s:0", username), targetUserData.Username); err != nil {
					kc.logger.Error("ZREM Error", zap.Error(err))
				}

				//下发通知给A所有端(From是B，To是A)
				{
					//构造回包里的数据
					eRsp := &Friends.FriendChangeEventRsp{
						Uuid:   uuid.NewV4().String(),
						Type:   Friends.FriendChangeType_Fc_RejectFriendApply, //拒绝加好友请求
						From:   targetUserData.Username,
						To:     username,
						Ps:     req.GetPs(),
						Source: req.GetSource(),
						TimeAt: uint64(time.Now().Unix()),
					}

					data, _ = proto.Marshal(eRsp)
					go kc.BroadcastMsgToAllDevices(data, username)
				}

				//下发通知给B所有端(From是A，To是B)
				{
					//构造回包里的数据
					eRsp := &Friends.FriendChangeEventRsp{
						Uuid:   uuid.NewV4().String(),
						Type:   Friends.FriendChangeType_Fc_RejectFriendApply, //拒绝加好友请求
						From:   username,
						To:     targetUserData.Username,
						Ps:     req.GetPs(),
						Source: req.GetSource(),
						TimeAt: uint64(time.Now().Unix()),
					}

					data, _ = proto.Marshal(eRsp)
					go kc.BroadcastMsgToAllDevices(data, targetUserData.Username)
				}
			}
		}

		data, _ = proto.Marshal(rsp)

	}

COMPLETE:
	if err != nil {
		msg.SetCode(400)                  //状态码
		msg.SetErrorMsg([]byte(errorMsg)) //错误提示
		msg.FillBody(nil)

	} else {
		msg.SetCode(200) //状态码
		msg.FillBody(data)
	}

	//处理完成，向dispatcher发送
	topic := msg.GetSource() + ".Frontend"
	if err := kc.Produce(topic, msg); err == nil {
		kc.logger.Info("Succeed succeed send message to ProduceChannel", zap.String("topic", topic))
	} else {
		kc.logger.Error("Failed to send  message to ProduceChannel", zap.Error(err))
	}
	_ = err
	return nil

}

/*
3-5 移除好友
*/
func (kc *KafkaClient) HandleDeleteFriend(msg *models.Message) error {
	var err error
	var errorMsg string
	var data []byte

	redisConn := kc.redisPool.Get()
	defer redisConn.Close()

	username := msg.GetUserName() //用户自己的账号
	// token := msg.GetJwtToken()
	deviceID := msg.GetDeviceID()

	kc.logger.Info("HandleFriendRequest start...",
		zap.String("username", username),
		zap.String("deviceId", deviceID))

	//取出当前设备的os， clientType， logonAt
	curDeviceHashKey := fmt.Sprintf("devices:%s:%s", username, deviceID)
	isMaster, _ := redis.Bool(redisConn.Do("HGET", curDeviceHashKey, "ismaster"))
	curOs, _ := redis.String(redisConn.Do("HGET", curDeviceHashKey, "os"))
	curClientType, _ := redis.Int(redisConn.Do("HGET", curDeviceHashKey, "clientType"))
	curLogonAt, _ := redis.Uint64(redisConn.Do("HGET", curDeviceHashKey, "logonAt"))

	kc.logger.Debug("FriendRequest",
		zap.Bool("isMaster", isMaster),
		zap.String("username", username),
		zap.String("deviceID", deviceID),
		zap.String("curOs", curOs),
		zap.Int("curClientType", curClientType),
		zap.Uint64("curLogonAt", curLogonAt))

	//打开msg里的负载， 获取请求参数
	body := msg.GetContent()

	//解包body
	req := &Friends.DeleteFriendReq{}
	if err := proto.Unmarshal(body, req); err != nil {
		kc.logger.Error("Protobuf Unmarshal Error", zap.Error(err))
		errorMsg = fmt.Sprintf("Protobuf Unmarshal Error: %s", err.Error())
		goto COMPLETE

	} else {
		targetUsername := req.GetUsername() //对方的用户账号
		kc.logger.Debug("FriendRequest body",
			zap.String("Username", targetUsername))

		//本地好友表，删除双方的好友关系

		//GO程，通知对方的多个端，每个端都删除这个username
		rsp := &Friends.DeleteFriendRsp{}
		data, _ = proto.Marshal(rsp)

	}

COMPLETE:
	if err != nil {
		msg.SetCode(400)                  //状态码
		msg.SetErrorMsg([]byte(errorMsg)) //错误提示
		msg.FillBody(nil)

	} else {
		msg.SetCode(200) //状态码
		msg.FillBody(data)
	}

	//处理完成，向dispatcher发送
	topic := msg.GetSource() + ".Frontend"
	if err := kc.Produce(topic, msg); err == nil {
		kc.logger.Info("Succeed succeed send message to ProduceChannel", zap.String("topic", topic))
	} else {
		kc.logger.Error("Failed to send  message to ProduceChannel", zap.Error(err))
	}
	_ = err
	return nil

}

/*
3-6 刷新好友资料
*/
func (kc *KafkaClient) HandleUpdateFriend(msg *models.Message) error {
	var err error
	var errorMsg string
	var data []byte

	redisConn := kc.redisPool.Get()
	defer redisConn.Close()

	username := msg.GetUserName() //用户自己的账号
	// token := msg.GetJwtToken()
	deviceID := msg.GetDeviceID()

	kc.logger.Info("HandleFriendRequest start...",
		zap.String("username", username),
		zap.String("deviceId", deviceID))

	//取出当前设备的os， clientType， logonAt
	curDeviceHashKey := fmt.Sprintf("devices:%s:%s", username, deviceID)
	isMaster, _ := redis.Bool(redisConn.Do("HGET", curDeviceHashKey, "ismaster"))
	curOs, _ := redis.String(redisConn.Do("HGET", curDeviceHashKey, "os"))
	curClientType, _ := redis.Int(redisConn.Do("HGET", curDeviceHashKey, "clientType"))
	curLogonAt, _ := redis.Uint64(redisConn.Do("HGET", curDeviceHashKey, "logonAt"))

	kc.logger.Debug("FriendRequest",
		zap.Bool("isMaster", isMaster),
		zap.String("username", username),
		zap.String("deviceID", deviceID),
		zap.String("curOs", curOs),
		zap.Int("curClientType", curClientType),
		zap.Uint64("curLogonAt", curLogonAt))

	//打开msg里的负载， 获取请求参数
	body := msg.GetContent()

	//解包body
	req := &Friends.UpdateFriendReq{}
	if err := proto.Unmarshal(body, req); err != nil {
		kc.logger.Error("Protobuf Unmarshal Error", zap.Error(err))
		errorMsg = fmt.Sprintf("Protobuf Unmarshal Error: %s", err.Error())
		goto COMPLETE

	} else {
		kc.logger.Debug("FriendRequest body",
			zap.String("Username", req.GetUsername()))

		targetUser := req.GetUsername() //对方的用户账号

		//查出 targetUser 有效性，是否已经是好友，好友增加的设置等信息

		targetKey := fmt.Sprintf("userData:%s", targetUser)
		userData := new(models.User)

		isExists, _ := redis.Bool(redisConn.Do("EXISTS", targetKey))
		if !isExists {
			// kc.logger.Error("targetKey is not exists", zap.String("targetKey", targetKey))
			// errorMsg = fmt.Sprintf("target user is not exists: %s", targetKey)
			// goto COMPLETE

			if err = kc.db.Model(userData).Where("username = ?", targetUser).First(userData).Error; err != nil {
				kc.logger.Error("MySQL里读取错误", zap.Error(err))
				errorMsg = fmt.Sprintf("Query user error[username=%s]", targetUser)
				goto COMPLETE
			}
			if _, err := redisConn.Do("HMSET", redis.Args{}.Add(targetKey).AddFlat(userData)...); err != nil {
				kc.logger.Error("错误：HMSET", zap.Error(err))
			} else {
				kc.logger.Debug("刷新Redis的用户数据成功", zap.String("targetUser", targetUser))
			}

		} else {
		}
		rsp := &Friends.UpdateFriendRsp{}

		data, _ = proto.Marshal(rsp)

	}

COMPLETE:
	if err != nil {
		msg.SetCode(400)                  //状态码
		msg.SetErrorMsg([]byte(errorMsg)) //错误提示
		msg.FillBody(nil)

	} else {
		msg.SetCode(200) //状态码
		msg.FillBody(data)
	}

	//处理完成，向dispatcher发送
	topic := msg.GetSource() + ".Frontend"
	if err := kc.Produce(topic, msg); err == nil {
		kc.logger.Info("Succeed succeed send message to ProduceChannel", zap.String("topic", topic))
	} else {
		kc.logger.Error("Failed to send  message to ProduceChannel", zap.Error(err))
	}
	_ = err
	return nil

}

/*
3-8 增量同步好友列表
*/
func (kc *KafkaClient) HandleGetFriends(msg *models.Message) error {
	var err error
	var errorMsg string
	var data []byte

	redisConn := kc.redisPool.Get()
	defer redisConn.Close()

	username := msg.GetUserName() //用户自己的账号
	// token := msg.GetJwtToken()
	deviceID := msg.GetDeviceID()

	kc.logger.Info("HandleFriendRequest start...",
		zap.String("username", username),
		zap.String("deviceId", deviceID))

	//取出当前设备的os， clientType， logonAt
	curDeviceHashKey := fmt.Sprintf("devices:%s:%s", username, deviceID)
	isMaster, _ := redis.Bool(redisConn.Do("HGET", curDeviceHashKey, "ismaster"))
	curOs, _ := redis.String(redisConn.Do("HGET", curDeviceHashKey, "os"))
	curClientType, _ := redis.Int(redisConn.Do("HGET", curDeviceHashKey, "clientType"))
	curLogonAt, _ := redis.Uint64(redisConn.Do("HGET", curDeviceHashKey, "logonAt"))

	kc.logger.Debug("FriendRequest",
		zap.Bool("isMaster", isMaster),
		zap.String("username", username),
		zap.String("deviceID", deviceID),
		zap.String("curOs", curOs),
		zap.Int("curClientType", curClientType),
		zap.Uint64("curLogonAt", curLogonAt))

	//打开msg里的负载， 获取请求参数
	body := msg.GetContent()

	//解包body
	req := &Friends.GetFriendsReq{}
	if err := proto.Unmarshal(body, req); err != nil {
		kc.logger.Error("Protobuf Unmarshal Error", zap.Error(err))
		errorMsg = fmt.Sprintf("Protobuf Unmarshal Error: %s", err.Error())
		goto COMPLETE

	} else {
		kc.logger.Debug("FriendRequest body",
			zap.Uint64("timeAt", req.GetTimeAt()))

		rsp := &Friends.GetFriendsRsp{}
		data, _ = proto.Marshal(rsp)

	}

COMPLETE:
	if err != nil {
		msg.SetCode(400)                  //状态码
		msg.SetErrorMsg([]byte(errorMsg)) //错误提示
		msg.FillBody(nil)

	} else {
		msg.SetCode(200) //状态码
		msg.FillBody(data)
	}

	//处理完成，向dispatcher发送
	topic := msg.GetSource() + ".Frontend"
	if err := kc.Produce(topic, msg); err == nil {
		kc.logger.Info("Succeed succeed send message to ProduceChannel", zap.String("topic", topic))
	} else {
		kc.logger.Error("Failed to send  message to ProduceChannel", zap.Error(err))
	}
	_ = err
	return nil

}

/*
好友请求，向目标好友 用户账号的所有端推送系统通知
*/
func (kc *KafkaClient) BroadcastMsgToAllDevices(data []byte, toUser string) error {

	redisConn := kc.redisPool.Get()
	defer redisConn.Close()

	//向toUser所有端发送FriendChangeEvent事件
	deviceListKey := fmt.Sprintf("devices:%s", toUser)
	deviceIDSliceNew, _ := redis.Strings(redisConn.Do("ZRANGEBYSCORE", deviceListKey, "-inf", "+inf"))
	//查询出当前在线所有主从设备
	for _, eDeviceID := range deviceIDSliceNew {

		targetMsg := &models.Message{}
		curDeviceKey := fmt.Sprintf("DeviceJwtToken:%s", eDeviceID)
		curJwtToken, _ := redis.String(redisConn.Do("GET", curDeviceKey))
		kc.logger.Debug("Redis GET ", zap.String("curDeviceKey", curDeviceKey), zap.String("curJwtToken", curJwtToken))

		targetMsg.UpdateID()
		//构建消息路由, 第一个参数是要处理的业务类型，后端服务器处理完成后，需要用此来拼接topic: {businessTypeName.Frontend}
		targetMsg.BuildRouter("Auth", "", "Auth.Frontend")

		targetMsg.SetJwtToken(curJwtToken)
		targetMsg.SetUserName(toUser)
		targetMsg.SetDeviceID(curDeviceKey)
		// kickMsg.SetTaskID(uint32(taskId))
		targetMsg.SetBusinessTypeName("User")
		targetMsg.SetBusinessType(uint32(3))
		targetMsg.SetBusinessSubType(uint32(2)) //FriendChangeEvent = 2

		targetMsg.BuildHeader("AuthService", time.Now().UnixNano()/1e6)

		targetMsg.FillBody(data) //网络包的body，承载真正的业务数据

		targetMsg.SetCode(200) //成功的状态码

		//构建数据完成，向dispatcher发送
		topic := "Auth.Frontend"
		if err := kc.Produce(topic, targetMsg); err == nil {
			kc.logger.Info("message succeed send to ProduceChannel", zap.String("topic", topic))
		} else {
			kc.logger.Error(" failed to send message to ProduceChannel", zap.Error(err))
		}

		kc.logger.Info("Sync myInfoAt Succeed",
			zap.String("Username:", toUser),
			zap.String("DeviceID:", curDeviceKey),
			zap.Int64("Now", time.Now().Unix()))

	}

	return nil
}
