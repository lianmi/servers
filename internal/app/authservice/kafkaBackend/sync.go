/*
本文件是处理业务号是同步模块，分别有
6-1 发起同步请求
    同步处理map，分别处理 myInfoAt, friendsAt, friendUsersAt, teamsAt, systemMsgAt, watchAt

6-2 同步请求完成

客户端触发:
1-3 同步当前用户资料事件

*/
package kafkaBackend

import (
	// "encoding/hex"
	"fmt"
	// "strings"
	"github.com/pkg/errors"
	"net/http"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/gomodule/redigo/redis"
	Friends "github.com/lianmi/servers/api/proto/friends"
	Global "github.com/lianmi/servers/api/proto/global"
	Order "github.com/lianmi/servers/api/proto/order"
	Sync "github.com/lianmi/servers/api/proto/syn"
	Team "github.com/lianmi/servers/api/proto/team"
	User "github.com/lianmi/servers/api/proto/user"
	"github.com/lianmi/servers/internal/common"
	"github.com/lianmi/servers/internal/pkg/models"
	"github.com/lianmi/servers/util/array"
	"go.uber.org/zap"
)

//处理myInfoAt
func (kc *KafkaClient) SyncMyInfoAt(username, token, deviceID string, req Sync.SyncEventReq, ch chan int) error {
	var err error
	errorCode := 200
	var errorMsg string
	var cur_myInfoAt uint64

	redisConn := kc.redisPool.Get()
	defer redisConn.Close()

	//req里的成员
	myInfoAt := req.GetMyInfoAt()
	syncKey := fmt.Sprintf("sync:%s", username)

	cur_myInfoAt, err = redis.Uint64(redisConn.Do("HGET", syncKey, "myInfoAt"))
	if err != nil {
		cur_myInfoAt = uint64(time.Now().UnixNano() / 1e6)
		redisConn.Do("HSET", syncKey, "myInfoAt", cur_myInfoAt)
	}

	kc.logger.Debug("SyncMyInfoAt",
		zap.Uint64("cur_myInfoAt", cur_myInfoAt),
		zap.Uint64("myInfoAt", myInfoAt),
		zap.String("username", username),
	)

	//服务端的时间戳大于客户端上报的时间戳
	if cur_myInfoAt > myInfoAt {
		//构造SyncUserProfileEventRsp
		//先从Redis里读取
		userData := new(models.User)
		userKey := fmt.Sprintf("userData:%s", username)
		if result, err := redis.Values(redisConn.Do("HGETALL", userKey)); err == nil {
			if err := redis.ScanStruct(result, userData); err != nil {
				kc.logger.Error("错误：ScanStruct", zap.Error(err))
				errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
				errorMsg = fmt.Sprintf("错误：ScanStructt error[userKey=%s]", userKey)
				goto COMPLETE

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
				targetMsg.SetBusinessType(uint32(Global.BusinessType_User))                   // 1
				targetMsg.SetBusinessSubType(uint32(Global.UserSubType_SyncUserProfileEvent)) // 3
				targetMsg.BuildHeader("AuthService", time.Now().UnixNano()/1e6)
				targetMsg.FillBody(data) //网络包的body，承载真正的业务数据
				targetMsg.SetCode(200)   //成功的状态码

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
					zap.Int64("Now", time.Now().UnixNano()/1e6))
			}

		}

	}

COMPLETE:
	//完成
	ch <- 1
	if errorCode == 200 {
		//只需返回200
		return nil
	} else {
		return errors.Wrap(err, errorMsg)
	}

}

//处理friendsAt
func (kc *KafkaClient) SyncFriendsAt(username, token, deviceID string, req Sync.SyncEventReq, ch chan int) error {
	var err error
	errorCode := 200
	var errorMsg string
	var cur_friendsAt uint64

	redisConn := kc.redisPool.Get()
	defer redisConn.Close()

	//req里的成员
	friendsAt := req.GetFriendsAt()
	syncKey := fmt.Sprintf("sync:%s", username)

	cur_friendsAt, err = redis.Uint64(redisConn.Do("HGET", syncKey, "friendsAt"))
	if err != nil {
		cur_friendsAt = uint64(time.Now().UnixNano() / 1e6)
		redisConn.Do("HSET", syncKey, "friendsAt", cur_friendsAt)
	}

	kc.logger.Debug("SyncFriendsAt",
		zap.Uint64("cur_friendsAt", cur_friendsAt),
		zap.Uint64("friendsAt", friendsAt),
		zap.String("username", username),
	)

	//服务端的时间戳大于客户端上报的时间戳
	if cur_friendsAt > friendsAt {
		//构造SyncFriendsEventRsp
		rsp := &Friends.SyncFriendsEventRsp{
			TimeTag:         uint64(time.Now().UnixNano() / 1e6), //毫秒
			Friends:         make([]*Friends.Friend, 0),
			RemovedAccounts: make([]*Friends.Friend, 0),
		}

		//从redis里读取username的好友列表
		friends, err := redis.Strings(redisConn.Do("ZRANGEBYSCORE", fmt.Sprintf("Friend:%s:1", username), "-inf", "+inf"))
		if err != nil {
			kc.logger.Error("ZRANGEBYSCORE", zap.Error(err))
			errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
			errorMsg = fmt.Sprintf("错误: ZRANGEBYSCORE error[key=Friend:%s:1]", username)
			goto COMPLETE
		}

		for _, friendUsername := range friends {

			nick, err := redis.String(redisConn.Do("HGET", fmt.Sprintf("FriendInfo:%s:%s", username, friendUsername), "Nick"))
			if err != nil {
				kc.logger.Error("HGET error", zap.Error(err))
				continue
			}
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
		RemoveFriends, err := redis.Strings(redisConn.Do("ZRANGEBYSCORE", fmt.Sprintf("Friend:%s:2", username), "-inf", "+inf"))
		if err != nil {
			kc.logger.Error("ZRANGEBYSCORE", zap.Error(err))
			errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
			errorMsg = fmt.Sprintf("错误: ZRANGEBYSCORE error[key=Friend:%s:2]", username)
			goto COMPLETE
		}

		for _, friendUsername := range RemoveFriends {
			nick, err := redis.String(redisConn.Do("HGET", fmt.Sprintf("FriendInfo:%s:%s", username, friendUsername), "Nick"))
			if err != nil {
				kc.logger.Error("HGET error", zap.Error(err))
				continue
			}
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
		targetMsg.SetBusinessTypeName("Friends")
		targetMsg.SetBusinessType(uint32(Global.BusinessType_Friends))              //3
		targetMsg.SetBusinessSubType(uint32(Global.FriendSubType_SyncFriendsEvent)) //3
		targetMsg.BuildHeader("AuthService", time.Now().UnixNano()/1e6)
		targetMsg.FillBody(data) //网络包的body，承载真正的业务数据
		targetMsg.SetCode(200)   //成功的状态码

		//构建数据完成，向dispatcher发送
		topic := "Auth.Frontend"
		if err := kc.Produce(topic, targetMsg); err == nil {
			kc.logger.Info("message succeed send to ProduceChannel", zap.String("topic", topic))
		} else {
			kc.logger.Error(" failed to send message to ProduceChannel", zap.Error(err))
		}

		kc.logger.Info("Sync FriendsAt Succeed",
			zap.String("Username:", username),
			zap.String("DeviceID:", deviceID),
			zap.Int64("Now", time.Now().UnixNano()/1e6))
	}

COMPLETE:
	//完成
	ch <- 1
	if errorCode == 200 {
		//只需返回200
		return nil
	} else {
		return errors.Wrap(err, errorMsg)
	}
}

//处理 friendUsersAt
func (kc *KafkaClient) SyncFriendUsersAt(username, token, deviceID string, req Sync.SyncEventReq, ch chan int) error {
	var err error
	errorCode := 200
	var errorMsg string
	var cur_friendUsersAt uint64

	redisConn := kc.redisPool.Get()
	defer redisConn.Close()

	//req里的成员
	friendUsersAt := req.GetFriendUsersAt()
	syncKey := fmt.Sprintf("sync:%s", username)

	cur_friendUsersAt, err = redis.Uint64(redisConn.Do("HGET", syncKey, "friendUsersAt"))
	if err != nil {
		cur_friendUsersAt = uint64(time.Now().UnixNano() / 1e6)
		redisConn.Do("HSET", syncKey, "friendUsersAt", cur_friendUsersAt)
	}

	kc.logger.Debug("SyncFriendUsersAt",
		zap.Uint64("cur_friendUsersAt", cur_friendUsersAt),
		zap.Uint64("friendUsersAt", friendUsersAt),
		zap.String("username", username),
	)

	//服务端的时间戳大于客户端上报的时间戳
	if cur_friendUsersAt > friendUsersAt {
		//构造SyncFriendsEventRsp
		rsp := &Friends.SyncFriendUsersEventRsp{
			TimeTag: uint64(time.Now().UnixNano() / 1e6),
			UInfos:  make([]*User.User, 0),
		}

		//从redis里读取username的好友列表
		fUsers, err := redis.Strings(redisConn.Do("ZRANGEBYSCORE", fmt.Sprintf("Friend:%s:1", username), "-inf", "+inf"))
		if err != nil {
			kc.logger.Error("ZRANGEBYSCORE", zap.Error(err))
			errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
			errorMsg = fmt.Sprintf("错误: ZRANGEBYSCORE error[key=Friend:%s:1]", username)
			goto COMPLETE
		}
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
		targetMsg.SetBusinessTypeName("Friends")
		targetMsg.SetBusinessType(uint32(Global.BusinessType_Friends))                  // 3
		targetMsg.SetBusinessSubType(uint32(Global.FriendSubType_SyncFriendUsersEvent)) // 4
		targetMsg.BuildHeader("AuthService", time.Now().UnixNano()/1e6)
		targetMsg.FillBody(data) //网络包的body，承载真正的业务数据
		targetMsg.SetCode(200)   //成功的状态码

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
			zap.Int64("Now", time.Now().UnixNano()/1e6))
	}

COMPLETE:
	//完成
	ch <- 1
	if errorCode == 200 {
		//只需返回200
		return nil
	} else {
		return errors.Wrap(err, errorMsg)
	}
}

//处理 TeamsAt
func (kc *KafkaClient) SyncTeamsAt(username, token, deviceID string, req Sync.SyncEventReq, ch chan int) error {
	var err error
	errorCode := 200
	var errorMsg string
	var cur_teamsAt uint64

	redisConn := kc.redisPool.Get()
	defer redisConn.Close()

	//req里的成员
	teamsAt := req.GetTeamsAt()
	syncKey := fmt.Sprintf("sync:%s", username)

	cur_teamsAt, err = redis.Uint64(redisConn.Do("HGET", syncKey, "teamsAt"))
	if err != nil {
		cur_teamsAt = uint64(time.Now().UnixNano() / 1e6)
		redisConn.Do("HSET", syncKey, "teamsAt", cur_teamsAt)

	}

	kc.logger.Debug("SyncTeamsAt",
		zap.Uint64("cur_teamsAt", cur_teamsAt),
		zap.Uint64("teamsAt", teamsAt),
		zap.String("username", username),
	)

	//服务端的时间戳大于客户端上报的时间戳
	if cur_teamsAt > teamsAt {
		//构造
		rsp := &Team.SyncMyTeamsEventRsp{ //4-17
			TimeAt:       uint64(time.Now().UnixNano() / 1e6),
			Teams:        make([]*Team.TeamInfo, 0),
			RemovedTeams: make([]string, 0),
		}

		//从redis里读取username的群列表
		teamIDs, err := redis.Strings(redisConn.Do("ZRANGEBYSCORE", fmt.Sprintf("Team:%s", username), "-inf", "+inf"))
		if err != nil {
			kc.logger.Error("ZRANGEBYSCORE", zap.Error(err))
			errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
			errorMsg = fmt.Sprintf("错误: ZRANGEBYSCORE error[key=Team:%s]", username)
			goto COMPLETE
		}

		for _, teamID := range teamIDs {
			kc.logger.Debug("for..range teamIDs", zap.String("teamID", teamID))
			teamInfo := new(models.Team)
			if result, err := redis.Values(redisConn.Do("HGETALL", fmt.Sprintf("TeamInfo:%s", teamID))); err == nil {
				if err := redis.ScanStruct(result, teamInfo); err != nil {
					kc.logger.Error("错误：ScanStruct", zap.Error(err), zap.String("key", fmt.Sprintf("TeamInfo:%s", teamID)))
					continue
				} else {
					//计算群成员数量。
					var count int
					if count, err = redis.Int(redisConn.Do("ZCARD", fmt.Sprintf("TeamUsers:%s", teamID))); err != nil {
						kc.logger.Error("ZCARD Error", zap.Error(err))
						continue
					}
					kc.logger.Debug("ZCARD 群成功总数", zap.Int("count", count))

					rsp.Teams = append(rsp.Teams, &Team.TeamInfo{
						TeamId:       teamInfo.TeamID,
						Name:         teamInfo.Teamname,
						Icon:         teamInfo.Icon,
						Announcement: teamInfo.Announcement,
						Introduce:    teamInfo.Introductory,
						Owner:        teamInfo.Owner,
						Type:         Team.TeamType(teamInfo.Type),
						VerifyType:   Team.VerifyType(teamInfo.VerifyType),
						MemberLimit:  int32(common.PerTeamMembersLimit),
						MemberNum:    int32(count),
						Status:       Team.Status(teamInfo.Status),
						MuteType:     Team.MuteMode(teamInfo.MuteType),
						InviteMode:   Team.InviteMode(teamInfo.InviteMode),
						Ex:           teamInfo.Extend,
						CreateAt:     uint64(teamInfo.CreatedAt),
						UpdateAt:     uint64(teamInfo.UpdatedAt),
					})
				}
			}
		}

		//用户自己的退群列表
		removeTeamIDs, _ := redis.Strings(redisConn.Do("ZRANGEBYSCORE", fmt.Sprintf("RemoveTeam:%s", username), "-inf", "+inf"))
		for _, removeTeamID := range removeTeamIDs {
			rsp.RemovedTeams = append(rsp.RemovedTeams, removeTeamID)
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
		targetMsg.SetBusinessTypeName("Team")
		targetMsg.SetBusinessType(uint32(Global.BusinessType_Team))               // 4
		targetMsg.SetBusinessSubType(uint32(Global.TeamSubType_SyncMyTeamsEvent)) // 17

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

		kc.logger.Info("Sync MyTeamsEvent  Succeed",
			zap.String("Username:", username),
			zap.String("DeviceID:", deviceID),
			zap.Int64("Now", time.Now().UnixNano()/1e6))
	}

COMPLETE:
	//完成
	ch <- 1
	if errorCode == 200 {
		//只需返回200
		return nil
	} else {
		return errors.Wrap(err, errorMsg)
	}
}

func (kc *KafkaClient) SendOffLineMsg(toUser, token, deviceID string, data []byte) error {

	targetMsg := &models.Message{}

	targetMsg.UpdateID()

	//构建消息路由, 第一个参数是要处理的业务类型，后端服务器处理完成后，需要用此来拼接topic: {businessTypeName.Frontend}
	targetMsg.BuildRouter("Auth", "", "Auth.Frontend")

	targetMsg.SetJwtToken(token)
	targetMsg.SetUserName(toUser)
	targetMsg.SetDeviceID(deviceID)
	// kickMsg.SetTaskID(uint32(taskId))
	targetMsg.SetBusinessTypeName("Msg")
	targetMsg.SetBusinessType(uint32(Global.BusinessType_Msg))                      //消息模块
	targetMsg.SetBusinessSubType(uint32(Global.MsgSubType_SyncOfflineSysMsgsEvent)) //同步系统离线消息

	targetMsg.BuildHeader("AuthService", time.Now().UnixNano()/1e6)

	targetMsg.FillBody(data) //网络包的body，承载真正的业务数据

	targetMsg.SetCode(200) //成功的状态码

	//构建数据完成，向dispatcher发送
	topic := "Auth.Frontend"
	go kc.Produce(topic, targetMsg)

	kc.logger.Info("SyncOfflineSysMsgsEvent Succeed",
		zap.String("Username:", toUser),
		zap.String("DeviceID:", deviceID),
		zap.Int64("Now", time.Now().UnixNano()/1e6))

	return nil
}

//处理 systemMsgAt
func (kc *KafkaClient) SyncSystemMsgAt(username, token, deviceID string, req Sync.SyncEventReq, ch chan int) error {
	var err error
	errorCode := 200
	var errorMsg string
	var cur_systemMsgAt uint64

	redisConn := kc.redisPool.Get()
	defer redisConn.Close()

	//req里的成员
	systemMsgAt := req.GetSystemMsgAt()
	syncKey := fmt.Sprintf("sync:%s", username)

	cur_systemMsgAt, err = redis.Uint64(redisConn.Do("HGET", syncKey, "systemMsgAt"))
	if err != nil {
		cur_systemMsgAt = uint64(time.Now().UnixNano() / 1e6)
		redisConn.Do("HSET", syncKey, "systemMsgAt", cur_systemMsgAt)

	}

	//服务端的时间戳大于客户端上报的时间戳
	if cur_systemMsgAt > systemMsgAt {
		nTime := time.Now()
		yesTime := nTime.AddDate(0, 0, -7).UnixNano() / 1e6

		//移除时间少于yesTime离线通知
		_, err = redisConn.Do("ZREMRANGEBYSCORE", fmt.Sprintf("systemMsgAt:%s", username), "-inf", yesTime)
		if err != nil {
			kc.logger.Error("ZRANGEBYSCORE", zap.Error(err))
			errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
			errorMsg = fmt.Sprintf("错误: ZRANGEBYSCORE error[key=%s]", fmt.Sprintf("systemMsgAt:%s", username))
			goto COMPLETE
		} else {
			kc.logger.Debug("移除时间少于yesTime离线通知", zap.Int64("yesTime", yesTime))
		}

		msgIDs, _ := redis.Strings(redisConn.Do("ZRANGEBYSCORE", fmt.Sprintf("systemMsgAt:%s", username), yesTime, "+inf"))
		for _, msgID := range msgIDs {
			key := fmt.Sprintf("systemMsg:%s:%s", username, msgID)
			data, _ := redis.Bytes(redisConn.Do("HGET", key, "Data"))
			err = kc.SendOffLineMsg(username, token, deviceID, data)
			if err == nil {
				kc.logger.Debug("成功发送离线通知", zap.String("msgID", msgID))
			}
		}

		//同步离线期间的个人，群聊，订单，商品推送的消息 消息只保留最后10条，而且只能缓存7天
		//从大到小递减获取, 只获取最大的10条
		msgIDArray, err := redis.ByteSlices(redisConn.Do("ZREVRANGEBYSCORE", fmt.Sprintf("offLineMsgList:%s", username), "+inf", "-inf", "LIMIT", 0, 10))
		if err != nil {
			kc.logger.Error("ZREVRANGEBYSCORE Error", zap.Error(err))
		} else {
			if len(msgIDArray) > 0 {
				//反转，数组变为从小到大排序
				array.ReverseBytes(msgIDArray)
				var maxSeq int
				for seq, msgID := range msgIDArray {
					if seq > maxSeq {
						maxSeq = seq
					}
					kc.logger.Debug("同步离线期间最大的10条离线消息",
						zap.Int("seq", seq),
						zap.String("msgID", string(msgID)),
					)
					key := fmt.Sprintf("offLineMsg:%s:%s", username, string(msgID))
					data, _ := redis.Bytes(redisConn.Do("HGET", key, "Data"))

					err = kc.SendOffLineMsg(username, token, deviceID, data)
					if err == nil {
						kc.logger.Debug("成功发送离线消息",
							zap.Int("seq", seq),
							zap.String("msgID", string(msgID)),
							zap.String("username", username),
						)
					} else {
						kc.logger.Error("发送离线消息失败，Error", zap.Error(err))
					}

				}

				//移除序号小于maxSeq的离线消息
				_, err = redisConn.Do("ZREMRANGEBYSCORE", fmt.Sprintf("offLineMsgList:%s", username), "-inf", maxSeq)

			}
		}
	}

COMPLETE:
	//完成
	ch <- 1
	if errorCode == 200 {
		//只需返回200
		return nil
	} else {
		return errors.Wrap(err, errorMsg)
	}
	return nil
}

//处理watchAt 7-8 同步关注的商户事件
func (kc *KafkaClient) SyncWatchAt(username, token, deviceID string, req Sync.SyncEventReq, ch chan int) error {
	var err error
	errorCode := 200
	var errorMsg string
	var cur_watchAt uint64

	rsp := &Order.SyncWatchEventRsp{
		TimeAt:              uint64(time.Now().UnixNano() / 1e6),
		WatchingUsers:       make([]string, 0), //用户关注的商户
		CancelWatchingUsers: make([]string, 0), //用户取消关注的商户
	}
	redisConn := kc.redisPool.Get()
	defer redisConn.Close()

	//req里的成员
	watchAt := req.GetWatchAt()
	syncKey := fmt.Sprintf("sync:%s", username)

	cur_watchAt, err = redis.Uint64(redisConn.Do("HGET", syncKey, "watchAt"))
	if err != nil {
		cur_watchAt = uint64(time.Now().UnixNano() / 1e6)
		redisConn.Do("HSET", syncKey, "watchAt", cur_watchAt)

	}

	kc.logger.Debug("SyncWatchAt",
		zap.Uint64("cur_watchAt", cur_watchAt),
		zap.Uint64("watchAt", watchAt),
		zap.String("username", username),
	)

	//服务端的时间戳大于客户端上报的时间戳
	if cur_watchAt > watchAt {
		watchingUsers, err := redis.Strings(redisConn.Do("ZRANGEBYSCORE", fmt.Sprintf("Watching:%s", username), watchAt, "+inf"))
		if err != nil {

			kc.logger.Error("ZRANGEBYSCORE", zap.Error(err))
			errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
			errorMsg = fmt.Sprintf("ZRANGEBYSCORE error[Watching:%s]", username)
			goto COMPLETE
		}

		for _, watchingUser := range watchingUsers {
			rsp.WatchingUsers = append(rsp.WatchingUsers, watchingUser)
		}
		cancelWatchingUsers, _ := redis.Strings(redisConn.Do("ZRANGEBYSCORE", fmt.Sprintf("CancelWatching:%s", username), watchAt, "+inf"))
		for _, cancelWatchingUser := range cancelWatchingUsers {
			rsp.CancelWatchingUsers = append(rsp.CancelWatchingUsers, cancelWatchingUser)
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
		targetMsg.SetBusinessTypeName("Friends")
		targetMsg.SetBusinessType(uint32(Global.BusinessType_Friends))            // 3
		targetMsg.SetBusinessSubType(uint32(Global.FriendSubType_SyncWatchEvent)) // 11

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

		kc.logger.Info("SyncWatchEvent Succeed",
			zap.String("Username:", username),
			zap.String("DeviceID:", deviceID),
			zap.Int64("Now", time.Now().UnixNano()/1e6))
	}

COMPLETE:
	//完成
	ch <- 1
	if errorCode == 200 {
		//只需返回200
		return nil
	} else {
		return errors.Wrap(err, errorMsg)
	}
}

/*
注意： syncCount 是所有需要同步的数量，最终是6个
*/
func (kc *KafkaClient) HandleSync(msg *models.Message) error {
	var err error

	errorCode := 200
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
		errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
		errorMsg = fmt.Sprintf("Protobuf Unmarshal Error: %s", err.Error())
		kc.logger.Error("Protobuf Unmarshal Error", zap.Error(err))
		goto COMPLETE

	} else {
		kc.logger.Debug("Sync  payload",
			zap.Uint64("MyInfoAt", req.MyInfoAt),
			zap.Uint64("FriendsAt", req.FriendsAt),
			zap.Uint64("FriendUsersAt", req.FriendUsersAt),
			zap.Uint64("TeamsAt", req.TeamsAt),
			zap.Uint64("TagsAt", req.TagsAt),
			zap.Uint64("SystemMsgAt", req.SystemMsgAt),
			zap.Uint64("WatchAt", req.WatchAt),
		)

		//所有同步的时间戳数量
		chs := make([]chan int, common.TotalSyncCount) //6 个
		for i := 0; i < common.TotalSyncCount; i++ {
			chs[i] = make(chan int)
		}

		//异步
		go func() {
			if err := kc.SyncMyInfoAt(username, token, deviceID, req, chs[0]); err != nil {
				kc.logger.Error("SyncMyInfoAt 失败，Error", zap.Error(err))
			} else {
				kc.logger.Debug("SyncMyInfoAt is done")
			}
		}()

		go func() {
			if err := kc.SyncFriendsAt(username, token, deviceID, req, chs[1]); err != nil {
				kc.logger.Error("SyncFriendsAt 失败，Error", zap.Error(err))
			} else {
				kc.logger.Debug("SyncFriendsAt is done")
			}
		}()

		go func() {
			if err := kc.SyncFriendUsersAt(username, token, deviceID, req, chs[2]); err != nil {
				kc.logger.Error("SyncFriendUsersAt 失败，Error", zap.Error(err))
			} else {
				kc.logger.Debug("SyncFriendUsersAt is done")
			}
		}()

		go func() {
			if err := kc.SyncTeamsAt(username, token, deviceID, req, chs[3]); err != nil {
				kc.logger.Error("SyncTeamsAt 失败，Error", zap.Error(err))
			} else {
				kc.logger.Debug("SyncTeamsAt is done")
			}
		}()

		go func() {
			if err := kc.SyncSystemMsgAt(username, token, deviceID, req, chs[4]); err != nil {
				kc.logger.Error("SyncSystemMsgAt 失败，Error", zap.Error(err))
			} else {
				kc.logger.Debug("SyncSystemMsgAt is done")
			}
		}()

		go func() {
			if err := kc.SyncWatchAt(username, token, deviceID, req, chs[5]); err != nil {
				kc.logger.Error("SyncWatchAt 失败，Error", zap.Error(err))
			} else {
				kc.logger.Debug("SyncWatchAt is done")
			}
		}()

		for i := 0; i < common.TotalSyncCount; i++ {
			<-chs[i]
		}

		//发送SyncDoneEvent
		kc.SendSyncDoneEventToUser(username, deviceID, token)
		kc.logger.Debug("All Sync done")

	}

COMPLETE:
	msg.SetCode(int32(errorCode)) //状态码
	if errorCode == 200 {
		//只需返回200
	} else {
		msg.SetErrorMsg([]byte(errorMsg)) //错误提示
		msg.FillBody(nil)
	}

	//处理完成，向dispatcher发送
	topic := msg.GetSource() + ".Frontend"
	if err := kc.Produce(topic, msg); err == nil {
		kc.logger.Info("Sync message succeed send to ProduceChannel", zap.String("topic", topic))
	} else {
		kc.logger.Error("Failed to send Sync message to ProduceChannel", zap.Error(err))
	}
	_ = err
	return nil

}

func (kc *KafkaClient) SendSyncDoneEventToUser(toUser, deviceID, token string) error {
	rsp := &Sync.SyncDoneEventRsp{
		TimeTag: uint64(time.Now().UnixNano() / 1e6),
	}
	data, _ := proto.Marshal(rsp)

	targetMsg := &models.Message{}
	targetMsg.UpdateID()
	//构建消息路由, 第一个参数是要处理的业务类型，后端服务器处理完成后，需要用此来拼接topic: {businessTypeName.Frontend}
	targetMsg.BuildRouter("Auth", "", "Auth.Frontend")
	targetMsg.SetJwtToken(token)
	targetMsg.SetUserName(toUser)
	targetMsg.SetDeviceID(deviceID)
	// kickMsg.SetTaskID(uint32(taskId))
	targetMsg.SetBusinessTypeName("Sync")
	targetMsg.SetBusinessType(uint32(Global.BusinessType_Sync))            //sync模块
	targetMsg.SetBusinessSubType(uint32(Global.SyncSubType_SyncDoneEvent)) // 同步完成事件
	targetMsg.BuildHeader("authservice", time.Now().UnixNano()/1e6)
	targetMsg.FillBody(data) //网络包的body，承载真正的业务数据
	targetMsg.SetCode(200)   //成功的状态码

	//构建数据完成，向dispatcher发送
	topic := "Sync.Frontend"
	if err := kc.Produce(topic, targetMsg); err == nil {
		kc.logger.Info("SendSyncDoneEventToUser, Msg message succeed send to ProduceChannel",
			zap.String("topic", topic),
			zap.String("to", toUser),
			zap.String("toDeviceID:", deviceID),
			zap.String("msgID:", targetMsg.GetID()),
		)
	} else {
		kc.logger.Error("SendSyncDoneEventToUser, Failed to send message to ProduceChannel",
			zap.String("topic", topic),
			zap.String("to", toUser),
			zap.String("toDeviceID:", deviceID),
			zap.String("msgID:", targetMsg.GetID()),
			zap.Error(err),
		)
		return err
	}

	return nil
}
