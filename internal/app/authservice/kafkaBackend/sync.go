/*
本文件是处理业务号是同步模块，分别有
6-1 发起同步请求
    同步处理map，分别处理 myInfoAt, friendsAt, friendUsersAt, teamsAt, conversationAckAt, systemMsgAt

6-2 同步请求完成

客户端触发:
1-3 同步当前用户资料事件



*/
package kafkaBackend

import (
	// "encoding/hex"
	"fmt"
	// "strings"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/gomodule/redigo/redis"
	Friends "github.com/lianmi/servers/api/proto/friends"
	Sync "github.com/lianmi/servers/api/proto/syn"
	User "github.com/lianmi/servers/api/proto/user"
	"github.com/lianmi/servers/internal/pkg/models"
	"go.uber.org/zap"
)

//处理myInfoAt
func (kc *KafkaClient) SyncMyInfoAt(username, token, deviceID string, req Sync.SyncEventReq, ch chan int) error {
	// var err error
	// var errorMsg string

	redisConn := kc.redisPool.Get()
	defer redisConn.Close()

	//req里的成员
	myInfoAt := req.GetMyInfoAt()
	myInfoAtKey := fmt.Sprintf("sync:%s", username)

	cur_myInfoAt, _ := redis.Uint64(redisConn.Do("HGET", myInfoAtKey, "myInfoAt"))

	//服务端的时间戳大于客户端上报的时间戳
	if cur_myInfoAt > myInfoAt {
		//构造SyncUserProfileEventRsp
		//先从Redis里读取
		userData := new(models.User)
		userKey := fmt.Sprintf("userData:%s", username)
		if result, err := redis.Values(redisConn.Do("HGETALL", userKey)); err == nil {
			if err := redis.ScanStruct(result, userData); err != nil {

				kc.logger.Error("错误：ScanStruct", zap.Error(err))

			} else {
				rsp := &User.SyncUserProfileEventRsp{
					TimeTag: uint64(time.Now().UnixNano() / 1e6),
					UInfo: &User.User{
						Username:          username,
						Gender:            User.Gender(userData.Gender),
						Nick:              userData.Nick,
						Avatar:            userData.Avatar,
						Label:             userData.Label,
						Mobile:            userData.Mobile,
						Email:             userData.Email,
						UserType:          User.UserType(userData.UserType),
						Extend:            userData.Extend,
						ContactPerson:     userData.ContactPerson,
						Introductory:      userData.Introductory,
						Province:          userData.Province,
						City:              userData.City,
						County:            userData.County,
						Street:            userData.Street,
						Address:           userData.Address,
						Branchesname:      userData.Branchesname,
						LegalPerson:       userData.LegalPerson,
						LegalIdentityCard: userData.LegalIdentityCard,
					},
				}
				data, _ := proto.Marshal(rsp)

				//向客户端响应SyncUserProfileEvent事件

				targetMsg := &models.Message{}

				targetMsg.UpdateID()
				//构建消息路由, 第一个参数是要处理的业务类型，后端服务器处理完成后，需要用此来拼接topic: {businessTypeName.Frontend}
				targetMsg.BuildRouter("Auth", "", "Auth.Frontend")

				targetMsg.SetJwtToken(token)
				targetMsg.SetUserName(username)
				targetMsg.SetDeviceID(deviceID)
				// kickMsg.SetTaskID(uint32(taskId))
				targetMsg.SetBusinessTypeName("User")
				targetMsg.SetBusinessType(uint32(1))
				targetMsg.SetBusinessSubType(uint32(3)) //SyncUserProfileEvent = 3

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
					zap.String("Username:", username),
					zap.String("DeviceID:", deviceID),
					zap.Int64("Now", time.Now().Unix()))
			}

		}

	}

	//完成
	ch <- 1

	return nil
}

//处理friendsAt
func (kc *KafkaClient) SyncFriendsAt(username, token, deviceID string, req Sync.SyncEventReq, ch chan int) error {
	redisConn := kc.redisPool.Get()
	defer redisConn.Close()

	//req里的成员
	friendsAt := req.GetFriendsAt()
	friendsAtKey := fmt.Sprintf("sync:%s", username)

	cur_friendsAt, _ := redis.Uint64(redisConn.Do("HGET", friendsAtKey, "friendsAt"))

	//服务端的时间戳大于客户端上报的时间戳
	if cur_friendsAt > friendsAt {
		//构造SyncFriendsEventRsp
		rsp := &Friends.SyncFriendsEventRsp{
			TimeTag:         uint64(time.Now().UnixNano() / 1e6),
			Friends:         make([]*Friends.Friend, 0),
			RemovedAccounts: make([]*Friends.Friend, 0),
		}

		//从redis里读取username的好友列表
		friends, _ := redis.Strings(redisConn.Do("ZRANGEBYSCORE", fmt.Sprintf("Friend:%s:1", username), "-inf", "+inf"))
		for _, friendUsername := range friends {

			nick, _ := redis.String(redisConn.Do("HGET", fmt.Sprintf("FriendInfo:%s:%s", username, friendUsername), "Nick"))
			source, _ := redis.String(redisConn.Do("HGET", fmt.Sprintf("FriendInfo:%s:%s", username, friendUsername), "Source"))
			ex, _ := redis.String(redisConn.Do("HGET", fmt.Sprintf("FriendInfo:%s:%s", username, friendUsername), "Ex"))
			createAt, _ := redis.Uint64(redisConn.Do("HGET", fmt.Sprintf("FriendInfo:%s:%s", username, friendUsername), "CreateAt"))
			updateAt, _ := redis.Uint64(redisConn.Do("HGET", fmt.Sprintf("FriendInfo:%s:%s", username, friendUsername), "UpdateAt"))

			rsp.Friends = append(rsp.Friends, &Friends.Friend{
				Username: friendUsername,
				Nick:     nick,
				Source:   source,
				Ex:       ex,
				CreateAt: createAt,
				UpdateAt: updateAt,
			})
		}
		//从redis里读取username的删除的好友列表
		RemoveFriends, _ := redis.Strings(redisConn.Do("ZRANGEBYSCORE", fmt.Sprintf("Friend:%s:2", username), "-inf", "+inf"))
		for _, friendUsername := range RemoveFriends {
			nick, _ := redis.String(redisConn.Do("HGET", fmt.Sprintf("FriendInfo:%s:%s", username, friendUsername), "Nick"))
			source, _ := redis.String(redisConn.Do("HGET", fmt.Sprintf("FriendInfo:%s:%s", username, friendUsername), "Source"))
			ex, _ := redis.String(redisConn.Do("HGET", fmt.Sprintf("FriendInfo:%s:%s", username, friendUsername), "Ex"))
			createAt, _ := redis.Uint64(redisConn.Do("HGET", fmt.Sprintf("FriendInfo:%s:%s", username, friendUsername), "CreateAt"))
			updateAt, _ := redis.Uint64(redisConn.Do("HGET", fmt.Sprintf("FriendInfo:%s:%s", username, friendUsername), "UpdateAt"))

			rsp.RemovedAccounts = append(rsp.RemovedAccounts, &Friends.Friend{
				Username: friendUsername,
				Nick:     nick,
				Source:   source,
				Ex:       ex,
				CreateAt: createAt,
				UpdateAt: updateAt,
			})
		}

		data, _ := proto.Marshal(rsp)

		//向客户端响应SyncFriendsEvent事件

		targetMsg := &models.Message{}

		targetMsg.UpdateID()
		//构建消息路由, 第一个参数是要处理的业务类型，后端服务器处理完成后，需要用此来拼接topic: {businessTypeName.Frontend}
		targetMsg.BuildRouter("Auth", "", "Auth.Frontend")

		targetMsg.SetJwtToken(token)
		targetMsg.SetUserName(username)
		targetMsg.SetDeviceID(deviceID)
		// kickMsg.SetTaskID(uint32(taskId))
		targetMsg.SetBusinessTypeName("User")
		targetMsg.SetBusinessType(uint32(1))
		targetMsg.SetBusinessSubType(uint32(3)) //SyncFriendsEvent = 3

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
			zap.String("Username:", username),
			zap.String("DeviceID:", deviceID),
			zap.Int64("Now", time.Now().Unix()))
	}
	//完成
	ch <- 1

	return nil
}

//处理 friendUsersAt
func (kc *KafkaClient) SyncFriendUsersAt(username, token, deviceID string, req Sync.SyncEventReq, ch chan int) error {
	redisConn := kc.redisPool.Get()
	defer redisConn.Close()

	//req里的成员
	friendUsersAt := req.GetFriendUsersAt()
	friendUsersAtKey := fmt.Sprintf("sync:%s", username)

	cur_friendUsersAt, _ := redis.Uint64(redisConn.Do("HGET", friendUsersAtKey, "friendUsersAt"))

	//服务端的时间戳大于客户端上报的时间戳
	if cur_friendUsersAt > friendUsersAt {
		//构造SyncFriendsEventRsp
		rsp := &Friends.SyncFriendUsersEventRsp{
			TimeTag: uint64(time.Now().UnixNano() / 1e6),
			UInfos:  make([]*User.User, 0),
		}

		//从redis里读取username的好友列表
		fUsers, _ := redis.Strings(redisConn.Do("ZRANGEBYSCORE", fmt.Sprintf("Friend:%s:1", username), "-inf", "+inf"))
		for _, fuser := range fUsers {

			fUserData := new(models.User)
			if result, err := redis.Values(redisConn.Do("HGETALL", fmt.Sprintf("userData:%s", fuser))); err == nil {
				if err := redis.ScanStruct(result, fUserData); err != nil {

					kc.logger.Error("错误：ScanStruct", zap.Error(err))

				} else {
					rsp.UInfos = append(rsp.UInfos, &User.User{
						Username:          username,
						Gender:            User.Gender(fUserData.Gender),
						Nick:              fUserData.Nick,
						Avatar:            fUserData.Avatar,
						Label:             fUserData.Label,
						Mobile:            fUserData.Mobile,
						Email:             fUserData.Email,
						UserType:          User.UserType(fUserData.UserType),
						Extend:            fUserData.Extend,
						ContactPerson:     fUserData.ContactPerson,
						Introductory:      fUserData.Introductory,
						Province:          fUserData.Province,
						City:              fUserData.City,
						County:            fUserData.County,
						Street:            fUserData.Street,
						Address:           fUserData.Address,
						Branchesname:      fUserData.Branchesname,
						LegalPerson:       fUserData.LegalPerson,
						LegalIdentityCard: fUserData.LegalIdentityCard,
					})
				}
			}

		}

		data, _ := proto.Marshal(rsp)

		//向客户端响应 SyncFriendUsersEvent 事件

		targetMsg := &models.Message{}

		targetMsg.UpdateID()
		//构建消息路由, 第一个参数是要处理的业务类型，后端服务器处理完成后，需要用此来拼接topic: {businessTypeName.Frontend}
		targetMsg.BuildRouter("Auth", "", "Auth.Frontend")

		targetMsg.SetJwtToken(token)
		targetMsg.SetUserName(username)
		targetMsg.SetDeviceID(deviceID)
		// kickMsg.SetTaskID(uint32(taskId))
		targetMsg.SetBusinessTypeName("User")
		targetMsg.SetBusinessType(uint32(3))
		targetMsg.SetBusinessSubType(uint32(4)) //SyncFriendUsersEvent = 4

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

		kc.logger.Info("Sync FriendUsers Event Succeed",
			zap.String("Username:", username),
			zap.String("DeviceID:", deviceID),
			zap.Int64("Now", time.Now().Unix()))
	}
	//完成
	ch <- 1

	return nil
}

/*
注意： syncCount 是所有需要同步的数量，最终是6个，目前只实现了3个
*/
func (kc *KafkaClient) HandleSync(msg *models.Message) error {
	var err error
	var errorMsg string

	redisConn := kc.redisPool.Get()
	defer redisConn.Close()

	username := msg.GetUserName()
	token := msg.GetJwtToken()
	deviceID := msg.GetDeviceID()

	kc.logger.Info("HandleSync start...",
		zap.String("username", username),
		zap.String("DeviceId", deviceID))

	//取出当前设备的os， clientType， logonAt
	curDeviceHashKey := fmt.Sprintf("devices:%s:%s", username, deviceID)
	isMaster, _ := redis.Bool(redisConn.Do("HGET", curDeviceHashKey, "ismaster"))
	curOs, _ := redis.String(redisConn.Do("HGET", curDeviceHashKey, "os"))
	curClientType, _ := redis.Int(redisConn.Do("HGET", curDeviceHashKey, "clientType"))
	curLogonAt, _ := redis.Uint64(redisConn.Do("HGET", curDeviceHashKey, "logonAt"))

	kc.logger.Debug("HandleSync",
		zap.Bool("isMaster", isMaster),
		zap.String("username", username),
		zap.String("deviceID", deviceID),
		zap.String("curOs", curOs),
		zap.Int("curClientType", curClientType),
		zap.Uint64("curLogonAt", curLogonAt))

	//打开msg里的负载， 获取请求参数
	body := msg.GetContent()
	//解包body
	var req Sync.SyncEventReq
	if err := proto.Unmarshal(body, &req); err != nil {
		errorMsg = fmt.Sprintf("Protobuf Unmarshal Error: %s", err.Error())
		kc.logger.Error("Protobuf Unmarshal Error", zap.Error(err))
		goto COMPLETE

	} else {

		//异步
		go func() {

			syncCount := 3 //最终是6, 每增加一个处理，就加1
			chs := make([]chan int, syncCount)

			i := 0
			chs[i] = make(chan int)

			if err := kc.SyncMyInfoAt(username, token, deviceID, req, chs[i]); err == nil {
				kc.logger.Debug("myInfoAt is done")
			}

			i++
			chs[i] = make(chan int)
			if err := kc.SyncFriendsAt(username, token, deviceID, req, chs[i]); err == nil {
				kc.logger.Debug("friendUsersAt is done")
			}

			i++
			chs[i] = make(chan int)
			if err := kc.SyncFriendUsersAt(username, token, deviceID, req, chs[i]); err == nil {
				kc.logger.Debug("myInfoAt is done")
			}

			for _, ch := range chs {
				<-ch
			}
			kc.logger.Debug("All Sync done")
		}()

	}

COMPLETE:
	if err != nil {
		msg.SetCode(400)                  //状态码
		msg.SetErrorMsg([]byte(errorMsg)) //错误提示
		msg.FillBody(nil)

	} else {
		msg.SetCode(200) //状态码
	}

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
