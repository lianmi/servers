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
	// "encoding/hex"
	"fmt"
	// "strings"

	"github.com/golang/protobuf/proto"
	"github.com/gomodule/redigo/redis"

	// User "github.com/lianmi/servers/api/proto/user"
	Friends "github.com/lianmi/servers/api/proto/friends"
	"github.com/lianmi/servers/internal/common"
	"github.com/lianmi/servers/internal/pkg/models"

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
		userData := new(models.User)
		var allowType int

		isExists, _ := redis.Bool(redisConn.Do("EXISTS", targetKey))
		if !isExists {
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

			allowType = userData.AllowType

		} else {
			allowType, _ = redis.Int(redisConn.Do("HGET", targetKey, "AllowType"))
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

					//Go程，直接让双方成为好友

				} else if allowType == common.NeedConfirm {
					rsp.Status = Friends.OpStatusType_Ost_WaitConfirm
					//Go程，下发系统通知
					//A和B互相不为好友，B所有终端均会收到该消息。

					//A好友列表中有B，B好友列表没有A，A发起好友申请，B所有终端均会接收该消息，并且B可以选择同意、拒绝

					//A好友列表中有B，B好友列表没有A，B发起好友申请，A会收到B好友通过系统通知，B不接收好友申请系统通知。
				}

			}
		case Friends.OptType_Fr_PassFriendApply: //对方同意加你为好友
			{
				rsp.Status = Friends.OpStatusType_Ost_ApplySucceed
			}
		case Friends.OptType_Fr_RejectFriendApply: //对方拒绝添加好友
			{
				rsp.Status = Friends.OpStatusType_Ost_RejectFriendApply
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
		kc.logger.Error("Failed to send GetUsersResp message to ProduceChannel", zap.Error(err))
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
		kc.logger.Error("Failed to send GetUsersResp message to ProduceChannel", zap.Error(err))
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
		kc.logger.Error("Failed to send GetUsersResp message to ProduceChannel", zap.Error(err))
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
		kc.logger.Error("Failed to send GetUsersResp message to ProduceChannel", zap.Error(err))
	}
	_ = err
	return nil

}
