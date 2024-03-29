/*
本文件是处理业务号是同步模块，分别有
6-1 发起同步请求
	同步处理map，分别处理:
	myInfoAt, friendsAt, friendUsersAt, teamsAt, tagsAt, watchAt, systemMsgAt
	取消 productAt

6-2 同步请求完成

客户端触发:
1-3 同步当前用户资料事件

*/
package nsqMq

import (
	"encoding/json"
	"fmt"

	"strings"
	"time"

	"github.com/pkg/errors"

	"github.com/golang/protobuf/proto"
	"github.com/gomodule/redigo/redis"
	Friends "github.com/lianmi/servers/api/proto/friends"
	Global "github.com/lianmi/servers/api/proto/global"
	MSG "github.com/lianmi/servers/api/proto/msg"
	Order "github.com/lianmi/servers/api/proto/order"
	Sync "github.com/lianmi/servers/api/proto/syn"
	Team "github.com/lianmi/servers/api/proto/team"
	User "github.com/lianmi/servers/api/proto/user"
	LMCommon "github.com/lianmi/servers/internal/common"
	LMCError "github.com/lianmi/servers/internal/pkg/lmcerror"

	"github.com/lianmi/servers/internal/common"
	"github.com/lianmi/servers/internal/pkg/models"

	// "github.com/lianmi/servers/util/array"
	"go.uber.org/zap"
)

//处理myInfoAt
func (nc *NsqClient) SyncMyInfoAt(username, token, deviceID string, req Sync.SyncEventReq) error {
	var err error
	errorCode := 200

	var cur_myInfoAt uint64

	redisConn := nc.redisPool.Get()
	defer redisConn.Close()

	//req里的成员
	myInfoAt := req.GetMyInfoAt()
	syncKey := fmt.Sprintf("sync:%s", username)

	cur_myInfoAt, err = redis.Uint64(redisConn.Do("HGET", syncKey, "myInfoAt"))
	if err != nil {
		cur_myInfoAt = uint64(time.Now().UnixNano() / 1e6)
		redisConn.Do("HSET", syncKey, "myInfoAt", cur_myInfoAt)
	}

	nc.logger.Debug("SyncMyInfoAt",
		zap.Uint64("cur_myInfoAt", cur_myInfoAt),
		zap.Uint64("myInfoAt", myInfoAt),
		zap.String("username", username),
	)

	//服务端的时间戳大于客户端上报的时间戳
	if cur_myInfoAt > myInfoAt {
		//构造SyncUserProfileEventRsp
		//先从Redis里读取
		userData := new(models.UserBase)
		userKey := fmt.Sprintf("userData:%s", username)

		if result, err := redis.Values(redisConn.Do("HGETALL", userKey)); err == nil {
			if err := redis.ScanStruct(result, userData); err != nil {
				nc.logger.Error("错误：ScanStruct", zap.Error(err))
				errorCode = LMCError.ProtobufUnmarshalError
				goto COMPLETE

			} else {

				userInfo := &User.User{}
				if userRsp, err := nc.service.GetUser(username); err != nil {
					nc.logger.Error("获取用户个人资料错误", zap.Error(err))
				} else {
					userInfo = userRsp.User
				}

				//假如用户是商户
				storeInfo := &User.Store{}
				if userData.UserType == int(User.UserType_Ut_Business) {
					if storeInfo, err = nc.service.GetStore(userData.Username); err != nil {
						nc.logger.Error("获取店铺资料错误", zap.Error(err))
					}
				}

				rsp := &User.SyncUserProfileEventRsp{
					TimeTag:   uint64(time.Now().UnixNano() / 1e6),
					UInfo:     userInfo,
					StoreInfo: storeInfo,
				}
				data, _ := proto.Marshal(rsp)

				//向客户端响应SyncUserProfileEvent事件

				targetMsg := &models.Message{}

				targetMsg.UpdateID()
				//构建消息路由, 第一个参数是要处理的业务类型，后端服务器处理完成后，需要用此来拼接topic: {businessTypeName.Frontend}
				targetMsg.BuildRouter("Sync", "", "Sync.Frontend")
				targetMsg.SetJwtToken(token)
				targetMsg.SetUserName(username)
				targetMsg.SetDeviceID(deviceID)
				// kickMsg.SetTaskID(uint32(taskId))
				targetMsg.SetBusinessTypeName("User")
				targetMsg.SetBusinessType(uint32(Global.BusinessType_User))                   // 1
				targetMsg.SetBusinessSubType(uint32(Global.UserSubType_SyncUserProfileEvent)) // 3
				targetMsg.BuildHeader("Dispatcher", time.Now().UnixNano()/1e6)
				targetMsg.FillBody(data) //网络包的body，承载真正的业务数据
				targetMsg.SetCode(200)   //成功的状态码

				//构建数据完成，向dispatcher发送
				topic := "Sync.Frontend"
				rawData, _ := json.Marshal(targetMsg)
				if err := nc.Producer.Public(topic, rawData); err == nil {
					nc.logger.Info("message succeed send to ProduceChannel", zap.String("topic", topic))
				} else {
					nc.logger.Error(" failed to send message to ProduceChannel", zap.Error(err))
				}

				nc.logger.Info("Sync myInfoAt Succeed",
					zap.String("Username:", username),
					zap.String("DeviceID:", deviceID),
					zap.Int64("Now", time.Now().UnixNano()/1e6))
			}

		}

	}

COMPLETE:
	//完成
	if errorCode == 200 {
		//只需返回200
		return nil
	} else {
		errorMsg := LMCError.ErrorMsg(errorCode) //错误描述
		return errors.Wrap(err, errorMsg)
	}

}

//处理friendsAt
func (nc *NsqClient) SyncFriendsAt(username, token, deviceID string, req Sync.SyncEventReq) error {
	var err error
	errorCode := 200

	var cur_friendsAt uint64

	redisConn := nc.redisPool.Get()
	defer redisConn.Close()

	//req里的成员
	friendsAt := req.GetFriendsAt()
	syncKey := fmt.Sprintf("sync:%s", username)

	cur_friendsAt, err = redis.Uint64(redisConn.Do("HGET", syncKey, "friendsAt"))
	if err != nil {
		cur_friendsAt = uint64(time.Now().UnixNano() / 1e6)
		redisConn.Do("HSET", syncKey, "friendsAt", cur_friendsAt)
	}

	nc.logger.Debug("SyncFriendsAt",
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
			RemovedAccounts: make([]string, 0),
		}

		//从redis的有序集合查询出score大于req.GetTimeTag()的成员
		friends, _ := redis.Strings(redisConn.Do("ZRANGEBYSCORE", fmt.Sprintf("Friend:%s:1", username), friendsAt, "+inf"))
		for _, friendUsername := range friends {

			nick, _ := redis.String(redisConn.Do("HGET", fmt.Sprintf("FriendInfo:%s:%s", username, friendUsername), "Nick"))
			source, _ := redis.String(redisConn.Do("HGET", fmt.Sprintf("FriendInfo:%s:%s", username, friendUsername), "Source"))
			ex, _ := redis.String(redisConn.Do("HGET", fmt.Sprintf("FriendInfo:%s:%s", username, friendUsername), "Ex"))

			avatar, _ := redis.String(redisConn.Do("HGET", fmt.Sprintf("userData:%s", friendUsername), "Avatar"))
			if err != nil {
				nc.logger.Error("HGET Avatar error", zap.Error(err))
			}
			if strings.HasPrefix(avatar, "http") || strings.HasPrefix(avatar, "https") {
				//
			} else {

				avatar = LMCommon.OSSUploadPicPrefix + avatar + "?x-oss-process=image/resize,w_50/quality,q_50"
			}

			rsp.Friends = append(rsp.Friends, &Friends.Friend{
				Username: friendUsername,
				Nick:     nick,
				Avatar:   avatar,
				Source:   source,
				Ex:       ex,
			})
		}
		//从redis里读取username的删除的好友列表
		RemoveFriends, _ := redis.Strings(redisConn.Do("ZRANGEBYSCORE", fmt.Sprintf("Friend:%s:2", username), friendsAt, "+inf"))
		for _, friendUsername := range RemoveFriends {
			rsp.RemovedAccounts = append(rsp.RemovedAccounts, friendUsername)
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
		targetMsg.BuildHeader("Dispatcher", time.Now().UnixNano()/1e6)
		targetMsg.FillBody(data) //网络包的body，承载真正的业务数据
		targetMsg.SetCode(200)   //成功的状态码

		//构建数据完成，向dispatcher发送
		topic := "Auth.Frontend"
		rawData, _ := json.Marshal(targetMsg)
		if err := nc.Producer.Public(topic, rawData); err == nil {
			nc.logger.Info("message succeed send to ProduceChannel", zap.String("topic", topic))
		} else {
			nc.logger.Error(" failed to send message to ProduceChannel", zap.Error(err))
		}

		nc.logger.Info("Sync FriendsAt Succeed",
			zap.String("Username:", username),
			zap.String("DeviceID:", deviceID),
			zap.Int64("Now", time.Now().UnixNano()/1e6))
	}

	// COMPLETE:
	//完成
	if errorCode == 200 {
		//只需返回200
		return nil
	} else {
		errorMsg := LMCError.ErrorMsg(errorCode) //错误描述
		return errors.Wrap(err, errorMsg)
	}
}

//处理 friendUsersAt
func (nc *NsqClient) SyncFriendUsersAt(username, token, deviceID string, req Sync.SyncEventReq) error {
	var err error
	errorCode := 200

	var cur_friendUsersAt uint64

	redisConn := nc.redisPool.Get()
	defer redisConn.Close()

	//req里的成员
	friendUsersAt := req.GetFriendUsersAt()
	syncKey := fmt.Sprintf("sync:%s", username)

	cur_friendUsersAt, err = redis.Uint64(redisConn.Do("HGET", syncKey, "friendUsersAt"))
	if err != nil {
		cur_friendUsersAt = uint64(time.Now().UnixNano() / 1e6)
		redisConn.Do("HSET", syncKey, "friendUsersAt", cur_friendUsersAt)
	}

	nc.logger.Debug("SyncFriendUsersAt",
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
		fUsers, err := redis.Strings(redisConn.Do("ZRANGEBYSCORE", fmt.Sprintf("Friend:%s:1", username), friendUsersAt, "+inf"))
		if err != nil {
			nc.logger.Error("ZRANGEBYSCORE", zap.Error(err))
			errorCode = LMCError.RedisError
			goto COMPLETE
		}
		for index, fuser := range fUsers {
			nc.logger.Info("fuser", zap.Int("index", index), zap.String("fuser", fuser))
			fUserData := new(models.UserBase)
			if result, err := redis.Values(redisConn.Do("HGETALL", fmt.Sprintf("userData:%s", fuser))); err == nil {
				if err := redis.ScanStruct(result, fUserData); err != nil {

					nc.logger.Error("错误：ScanStruct", zap.Error(err))

				} else {
					var avatar string

					if fUserData.Avatar != "" {
						if strings.HasPrefix(fUserData.Avatar, "https") {
							avatar = fUserData.Avatar + "?x-oss-process=image/resize,w_50/quality,q_50"
						} else {

							avatar = LMCommon.OSSUploadPicPrefix + fUserData.Avatar + "?x-oss-process=image/resize,w_50/quality,q_50"
						}

					}

					rsp.UInfos = append(rsp.UInfos, &User.User{
						Username: fuser, // 好友注册id
						Gender:   User.Gender(fUserData.Gender),
						Nick:     fUserData.Nick,
						Avatar:   avatar,
						Label:    fUserData.Label,
						// Mobile:       fUserData.Mobile, 隐私
						// Email:        fUserData.Email,
						UserType: User.UserType(fUserData.UserType),
						Extend:   fUserData.Extend,
						// TrueName:     fUserData.TrueName,
						// IdentityCard: fUserData.IdentityCard,
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
		targetMsg.BuildHeader("Dispatcher", time.Now().UnixNano()/1e6)
		targetMsg.FillBody(data) //网络包的body，承载真正的业务数据
		targetMsg.SetCode(200)   //成功的状态码

		//构建数据完成，向dispatcher发送
		topic := "Auth.Frontend"
		rawData, _ := json.Marshal(targetMsg)
		if err := nc.Producer.Public(topic, rawData); err == nil {
			nc.logger.Info("message succeed send to ProduceChannel", zap.String("topic", topic))
		} else {
			nc.logger.Error(" failed to send message to ProduceChannel", zap.Error(err))
		}

		nc.logger.Info("Sync FriendUsers Event Succeed",
			zap.String("Username:", username),
			zap.String("DeviceID:", deviceID),
			zap.Int64("Now", time.Now().UnixNano()/1e6))
	}

COMPLETE:
	//完成
	if errorCode == 200 {
		//只需返回200
		return nil
	} else {
		errorMsg := LMCError.ErrorMsg(errorCode) //错误描述
		return errors.Wrap(err, errorMsg)
	}
}

//处理 TeamsAt
func (nc *NsqClient) SyncTeamsAt(username, token, deviceID string, req Sync.SyncEventReq) error {
	var err error
	errorCode := 200

	var cur_teamsAt uint64

	redisConn := nc.redisPool.Get()
	defer redisConn.Close()

	//req里的成员
	teamsAt := req.GetTeamsAt()
	syncKey := fmt.Sprintf("sync:%s", username)

	cur_teamsAt, err = redis.Uint64(redisConn.Do("HGET", syncKey, "teamsAt"))
	if err != nil {
		cur_teamsAt = uint64(time.Now().UnixNano() / 1e6)
		redisConn.Do("HSET", syncKey, "teamsAt", cur_teamsAt)

	}

	nc.logger.Debug("SyncTeamsAt",
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
			nc.logger.Error("ZRANGEBYSCORE", zap.Error(err))
			errorCode = LMCError.RedisError
			goto COMPLETE
		}

		for _, teamID := range teamIDs {
			nc.logger.Debug("for..range teamIDs", zap.String("teamID", teamID))
			teamInfo := new(models.TeamInfo)
			if result, err := redis.Values(redisConn.Do("HGETALL", fmt.Sprintf("TeamInfo:%s", teamID))); err == nil {
				if err := redis.ScanStruct(result, teamInfo); err != nil {
					nc.logger.Error("错误：ScanStruct", zap.Error(err), zap.String("key", fmt.Sprintf("TeamInfo:%s", teamID)))
					continue
				} else {
					//计算群成员数量。
					var count int
					if count, err = redis.Int(redisConn.Do("ZCARD", fmt.Sprintf("TeamUsers:%s", teamID))); err != nil {
						nc.logger.Error("ZCARD Error", zap.Error(err))
						continue
					}
					nc.logger.Debug("ZCARD 群成功总数", zap.Int("count", count))

					_, err = redisConn.Do("ZREM", fmt.Sprintf("RemoveTeam:%s", username), teamID)
					if err != nil {
						nc.logger.Error("删除redis里的RemoveTeam, ZREM 出错", zap.Error(err))
					}

					rsp.Teams = append(rsp.Teams, &Team.TeamInfo{
						TeamId:       teamInfo.TeamID,
						TeamName:     teamInfo.Teamname,
						Icon:         teamInfo.Icon,
						Announcement: teamInfo.Announcement,
						Introduce:    teamInfo.Introductory,
						Owner:        teamInfo.Owner,
						Type:         Team.TeamType(teamInfo.Type),
						VerifyType:   Team.VerifyType(teamInfo.VerifyType),
						MemberLimit:  int32(common.PerTeamMembersLimit),
						MemberNum:    int32(count),
						Status:       Team.TeamStatus(teamInfo.Status),
						MuteType:     Team.MuteMode(teamInfo.MuteType),
						InviteMode:   Team.InviteMode(teamInfo.InviteMode),
						Ex:           teamInfo.Extend,
						IsMute:       teamInfo.IsMute,
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

		targetMsg.BuildHeader("Dispatcher", time.Now().UnixNano()/1e6)

		targetMsg.FillBody(data) //网络包的body，承载真正的业务数据

		targetMsg.SetCode(200) //成功的状态码

		//构建数据完成，向dispatcher发送
		topic := "Auth.Frontend"
		rawData, _ := json.Marshal(targetMsg)
		if err := nc.Producer.Public(topic, rawData); err == nil {
			nc.logger.Info("message succeed send to ProduceChannel", zap.String("topic", topic))
		} else {
			nc.logger.Error(" failed to send message to ProduceChannel", zap.Error(err))
		}

		nc.logger.Info("Sync MyTeamsEvent  Succeed",
			zap.String("Username:", username),
			zap.String("DeviceID:", deviceID),
			zap.Int64("Now", time.Now().UnixNano()/1e6))
	}

COMPLETE:
	//完成
	if errorCode == 200 {
		//只需返回200
		return nil
	} else {
		errorMsg := LMCError.ErrorMsg(errorCode) //错误描述
		return errors.Wrap(err, errorMsg)
	}
}

//1-7 同步用户标签列表 处理 TagsAt
func (nc *NsqClient) SyncTagsAt(username, token, deviceID string, req Sync.SyncEventReq) error {
	var err error
	errorCode := 200

	var cur_tagsAt uint64

	redisConn := nc.redisPool.Get()
	defer redisConn.Close()

	//req里的成员
	tagsAt := req.GetTagsAt()
	syncKey := fmt.Sprintf("sync:%s", username)

	cur_tagsAt, err = redis.Uint64(redisConn.Do("HGET", syncKey, "tagsAt"))
	if err != nil {
		cur_tagsAt = uint64(time.Now().UnixNano() / 1e6)
		redisConn.Do("HSET", syncKey, "tagsAt", cur_tagsAt)

	}

	nc.logger.Debug("SyncTagsAt",
		zap.Uint64("cur_tagsAt", cur_tagsAt),
		zap.Uint64("tagsAt", tagsAt),
		zap.String("username", username),
	)
	//1613367673452
	//{"level":"debug","ts":1613370824.9488163,"caller":"nsqMq/sync.go:571","msg":"SyncTagsAt","type":"dispatcher.nsq","cur_tagsAt":1613367673452,"tagsAt":1613370516932,"username":"id1"}

	//服务端的时间戳大于客户端上报的时间戳
	if cur_tagsAt > tagsAt {
		//构造
		rsp := &User.SyncTagsEventRsp{ //1-7 同步用户标签列表
			TimeTag:     uint64(time.Now().UnixNano() / 1e6),
			AddTags:     make([]*User.Tag, 0),
			RemovedTags: make([]*User.Tag, 0),
		}

		//遍历时间戳大于客户端上报的时间戳的黑名单
		blackUsers, err := redis.Strings(redisConn.Do("ZRANGEBYSCORE", fmt.Sprintf("BlackList:%s:1", username), tagsAt, "+inf"))
		if err != nil {
			nc.logger.Error("遍历时间戳大于客户端上报的时间戳的黑名单 ZRANGEBYSCORE", zap.Error(err))
			errorCode = LMCError.RedisError
			goto COMPLETE
		}

		for _, blackUser := range blackUsers {
			rsp.AddTags = append(rsp.AddTags, &User.Tag{
				Username: blackUser,
				Type:     User.MarkTagType_Mtt_Blocked,
			})
		}

		//遍历时间戳大于客户端上报的时间戳的免打扰
		mutedUsers, err := redis.Strings(redisConn.Do("ZRANGEBYSCORE", fmt.Sprintf("MutedList:%s:1", username), tagsAt, "+inf"))
		if err != nil {
			nc.logger.Error("遍历时间戳大于客户端上报的时间戳的免打扰 ZRANGEBYSCORE", zap.Error(err))
			errorCode = LMCError.RedisError
			goto COMPLETE
		}

		for _, mutedUser := range mutedUsers {
			rsp.AddTags = append(rsp.AddTags, &User.Tag{
				Username: mutedUser,
				Type:     User.MarkTagType_Mtt_Muted,
			})
		}

		//遍历时间戳大于客户端上报的时间戳的置顶
		stickyUsers, err := redis.Strings(redisConn.Do("ZRANGEBYSCORE", fmt.Sprintf("StickyList:%s:1", username), tagsAt, "+inf"))
		if err != nil {
			nc.logger.Error("遍历时间戳大于客户端上报的时间戳的置顶  ZRANGEBYSCORE", zap.Error(err))
			errorCode = LMCError.RedisError
			goto COMPLETE
		}

		for _, stickyUser := range stickyUsers {
			rsp.AddTags = append(rsp.AddTags, &User.Tag{
				Username: stickyUser,
				Type:     User.MarkTagType_Mtt_Sticky,
			})
		}

		//遍历时间戳大于客户端上报的时间戳的移除黑名单
		removeBlackUsers, err := redis.Strings(redisConn.Do("ZRANGEBYSCORE", fmt.Sprintf("BlackList:%s:2", username), tagsAt, "+inf"))
		if err != nil {
			nc.logger.Error("遍历时间戳大于客户端上报的时间戳的移除黑名单 ZRANGEBYSCORE", zap.Error(err))
			errorCode = LMCError.RedisError
			goto COMPLETE
		}

		for _, removeBlackUser := range removeBlackUsers {
			rsp.RemovedTags = append(rsp.RemovedTags, &User.Tag{
				Username: removeBlackUser,
				Type:     User.MarkTagType_Mtt_Blocked,
			})
		}

		//遍历时间戳大于客户端上报的时间戳的移除免打扰
		removeMutedUsers, err := redis.Strings(redisConn.Do("ZRANGEBYSCORE", fmt.Sprintf("MutedList:%s:2", username), tagsAt, "+inf"))
		if err != nil {
			nc.logger.Error("遍历时间戳大于客户端上报的时间戳的移除免打扰 ZRANGEBYSCORE", zap.Error(err))
			errorCode = LMCError.RedisError
			goto COMPLETE
		}

		for _, removeMutedUser := range removeMutedUsers {
			rsp.RemovedTags = append(rsp.RemovedTags, &User.Tag{
				Username: removeMutedUser,
				Type:     User.MarkTagType_Mtt_Muted,
			})
		}

		//遍历时间戳大于客户端上报的时间戳的移除置顶
		removeStickyUsers, err := redis.Strings(redisConn.Do("ZRANGEBYSCORE", fmt.Sprintf("StickyList:%s:2", username), tagsAt, "+inf"))
		if err != nil {

			nc.logger.Error("遍历时间戳大于客户端上报的时间戳的移除置顶 ZRANGEBYSCORE", zap.Error(err))
			errorCode = LMCError.RedisError
			goto COMPLETE
		}

		for _, removeStickyUser := range removeStickyUsers {
			rsp.RemovedTags = append(rsp.RemovedTags, &User.Tag{
				Username: removeStickyUser,
				Type:     User.MarkTagType_Mtt_Sticky,
			})
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
		targetMsg.SetBusinessType(uint32(Global.BusinessType_User))            // 1
		targetMsg.SetBusinessSubType(uint32(Global.UserSubType_SyncTagsEvent)) // 7

		targetMsg.BuildHeader("Dispatcher", time.Now().UnixNano()/1e6)

		targetMsg.FillBody(data) //网络包的body，承载真正的业务数据

		targetMsg.SetCode(200) //成功的状态码

		//构建数据完成，向dispatcher发送
		topic := "Auth.Frontend"
		rawData, _ := json.Marshal(targetMsg)
		if err := nc.Producer.Public(topic, rawData); err == nil {
			nc.logger.Info("message succeed send to ProduceChannel", zap.String("topic", topic))
		} else {
			nc.logger.Error(" failed to send message to ProduceChannel", zap.Error(err))
		}

		nc.logger.Info("Sync SyncTagsAt Succeed",
			zap.String("Username:", username),
			zap.String("DeviceID:", deviceID),
			zap.Int64("Now", time.Now().UnixNano()/1e6))
	} else {
		nc.logger.Info("Sync SyncTagsAt 不需要同步")
	}

COMPLETE:
	//完成
	if errorCode == 200 {
		//只需返回200
		return nil
	} else {
		errorMsg := LMCError.ErrorMsg(errorCode) //错误描述
		return errors.Wrap(err, errorMsg)
	}
}

//发送离线系统通知 systemMsg:%s:%s
func (nc *NsqClient) SendOffLineMsg(toUser, token, deviceID string, data []byte) error {

	targetMsg := &models.Message{}

	targetMsg.UpdateID()

	//构建消息路由, 第一个参数是要处理的业务类型，后端服务器处理完成后，需要用此来拼接topic: {businessTypeName.Frontend}
	targetMsg.BuildRouter("Msg", "", "Msg.Frontend")

	targetMsg.SetJwtToken(token)
	targetMsg.SetUserName(toUser)
	targetMsg.SetDeviceID(deviceID)
	// kickMsg.SetTaskID(uint32(taskId))
	targetMsg.SetBusinessTypeName("Msg")
	targetMsg.SetBusinessType(uint32(Global.BusinessType_Msg))                      //消息模块
	targetMsg.SetBusinessSubType(uint32(Global.MsgSubType_SyncOfflineSysMsgsEvent)) //同步系统离线消息

	targetMsg.BuildHeader("Dispatcher", time.Now().UnixNano()/1e6)

	targetMsg.FillBody(data) //网络包的body，承载真正的业务数据

	targetMsg.SetCode(200) //成功的状态码

	//构建数据完成，向dispatcher发送
	topic := "Msg.Frontend"
	rawData, _ := json.Marshal(targetMsg)

	nc.Producer.Public(topic, rawData)

	nc.logger.Info("SendOffLineMsg Succeed",
		zap.String("Username:", toUser),
		zap.String("DeviceID:", deviceID),
		zap.Int64("Now", time.Now().UnixNano()/1e6))

	return nil
}

//处理watchAt 7-8 同步关注的商户事件
func (nc *NsqClient) SyncWatchAt(username, token, deviceID string, req Sync.SyncEventReq) error {
	var err error
	errorCode := 200

	var cur_watchAt uint64

	rsp := &Order.SyncWatchEventRsp{
		TimeAt:              uint64(time.Now().UnixNano() / 1e6),
		WatchingUsers:       make([]string, 0), //用户关注的商户
		CancelWatchingUsers: make([]string, 0), //用户取消关注的商户
	}
	redisConn := nc.redisPool.Get()
	defer redisConn.Close()

	//req里的成员
	watchAt := req.GetWatchAt()
	syncKey := fmt.Sprintf("sync:%s", username)

	cur_watchAt, err = redis.Uint64(redisConn.Do("HGET", syncKey, "watchAt"))
	if err != nil {
		cur_watchAt = uint64(time.Now().UnixNano() / 1e6)
		redisConn.Do("HSET", syncKey, "watchAt", cur_watchAt)

	}

	nc.logger.Debug("SyncWatchAt",
		zap.Uint64("cur_watchAt", cur_watchAt),
		zap.Uint64("watchAt", watchAt),
		zap.String("username", username),
	)

	//服务端的时间戳大于客户端上报的时间戳
	if cur_watchAt > watchAt {
		watchingUsers, err := redis.Strings(redisConn.Do("ZRANGEBYSCORE", fmt.Sprintf("Watching:%s", username), watchAt, "+inf"))
		if err != nil {

			nc.logger.Error("ZRANGEBYSCORE", zap.Error(err))
			errorCode = LMCError.RedisError
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

		targetMsg.BuildHeader("Dispatcher", time.Now().UnixNano()/1e6)

		targetMsg.FillBody(data) //网络包的body，承载真正的业务数据

		targetMsg.SetCode(200) //成功的状态码

		//构建数据完成，向dispatcher发送
		topic := "Auth.Frontend"
		rawData, _ := json.Marshal(targetMsg)
		if err := nc.Producer.Public(topic, rawData); err == nil {
			nc.logger.Info("message succeed send to ProduceChannel", zap.String("topic", topic))
		} else {
			nc.logger.Error(" failed to send message to ProduceChannel", zap.Error(err))
		}

		nc.logger.Info("SyncWatchEvent Succeed",
			zap.String("Username:", username),
			zap.String("DeviceID:", deviceID),
			zap.Int64("Now", time.Now().UnixNano()/1e6))
	}

COMPLETE:
	//完成
	if errorCode == 200 {
		//只需返回200
		return nil
	} else {
		errorMsg := LMCError.ErrorMsg(errorCode) //错误描述
		return errors.Wrap(err, errorMsg)
	}
}

//处理SystemMsgAt 5-9 同步系统公告
func (nc *NsqClient) SyncSystemMsgAt(username, token, deviceID string, req Sync.SyncEventReq) error {
	var err error
	errorCode := 200

	redisConn := nc.redisPool.Get()
	defer redisConn.Close()

	//req里的成员
	systemMsgAt := req.GetSysmtemMsgAt()

	//构造数据
	systemMsgs, err := nc.service.GetSystemMsgs(systemMsgAt)
	if err != nil {
		nc.logger.Error(err.Error())
		return err
	}
	rsp := &MSG.SyncSystemMsgRsp{
		TimeTag: uint64(time.Now().UnixNano() / 1e6),
	}
	for _, systemMsg := range systemMsgs {
		rsp.SystemMsgs = append(rsp.SystemMsgs, &MSG.SystemMsgBase{
			Id:        uint64(systemMsg.ID),
			Level:     int32(systemMsg.Level),
			Title:     systemMsg.Title,
			Content:   systemMsg.Content,
			CreatedAt: uint64(systemMsg.CreatedAt),
		})

	}

	data, _ := proto.Marshal(rsp)

	//向客户端响应 SyncSystemMsgEvent 事件
	targetMsg := &models.Message{}

	targetMsg.UpdateID()
	//构建消息路由, 第一个参数是要处理的业务类型，后端服务器处理完成后，需要用此来拼接topic: {businessTypeName.Frontend}
	targetMsg.BuildRouter("Auth", "", "Auth.Frontend")

	targetMsg.SetJwtToken(token)
	targetMsg.SetUserName(username)
	targetMsg.SetDeviceID(deviceID)
	// kickMsg.SetTaskID(uint32(taskId))
	targetMsg.SetBusinessTypeName("MSG")
	targetMsg.SetBusinessType(uint32(Global.BusinessType_Msg))                 // 5
	targetMsg.SetBusinessSubType(uint32(Global.MsgSubType_SyncSystemMsgEvent)) // 9

	targetMsg.BuildHeader("Dispatcher", time.Now().UnixNano()/1e6)

	targetMsg.FillBody(data) //网络包的body，承载真正的业务数据

	targetMsg.SetCode(200) //成功的状态码

	//构建数据完成，向dispatcher发送
	topic := "Auth.Frontend"
	rawData, _ := json.Marshal(targetMsg)
	if err := nc.Producer.Public(topic, rawData); err == nil {
		nc.logger.Info("Message succeed send to ProduceChannel", zap.String("topic", topic))
	} else {
		nc.logger.Error("Failed to send message to ProduceChannel", zap.Error(err))
	}

	//完成
	if errorCode == 200 {
		//只需返回200
		return nil
	} else {
		errorMsg := LMCError.ErrorMsg(errorCode) //错误描述
		return errors.Wrap(err, errorMsg)
	}
}

/*
6-1 发起同步请求
由于SDK在断线重连后也需要增量同步一次，因为，将broker的 在线用户列表放在这里

*/
func (nc *NsqClient) HandleSync(msg *models.Message) error {
	var err error

	errorCode := 200

	redisConn := nc.redisPool.Get()
	defer redisConn.Close()

	username := msg.GetUserName()
	token := msg.GetJwtToken()
	deviceID := msg.GetDeviceID()

	nc.logger.Info("HandleSync start...",
		zap.String("username", username),
		zap.String("DeviceId", deviceID))

	//取出当前设备的os，  logonAt
	curDeviceHashKey := fmt.Sprintf("devices:%s:%s", username, deviceID)
	isMaster, _ := redis.Bool(redisConn.Do("HGET", curDeviceHashKey, "ismaster"))
	curOs, _ := redis.String(redisConn.Do("HGET", curDeviceHashKey, "os"))
	curLogonAt, _ := redis.Uint64(redisConn.Do("HGET", curDeviceHashKey, "logonAt"))

	nc.logger.Debug("HandleSync",
		zap.Bool("isMaster", isMaster),
		zap.String("username", username),
		zap.String("deviceID", deviceID),
		zap.String("curOs", curOs),
		zap.Uint64("curLogonAt", curLogonAt))

	//打开msg里的负载， 获取请求参数
	body := msg.GetContent()
	//解包body
	var req Sync.SyncEventReq
	if err := proto.Unmarshal(body, &req); err != nil {
		errorCode = LMCError.ProtobufUnmarshalError
		nc.logger.Error("Protobuf Unmarshal Error", zap.Error(err))
		goto COMPLETE

	} else {
		if err := nc.AddOnlineUsers(deviceID); err != nil {
			nc.logger.Error("AddOnlineUsers Error", zap.Error(err))
		} else {
			nc.logger.Debug("AddOnlineUsers OK", zap.String("deviceID", deviceID))
		}

		nc.logger.Debug("Sync payload",
			zap.Uint64("MyInfoAt", req.MyInfoAt),
			zap.Uint64("FriendsAt", req.FriendsAt),
			zap.Uint64("FriendUsersAt", req.FriendUsersAt),
			zap.Uint64("TeamsAt", req.TeamsAt),
			zap.Uint64("TagsAt", req.TagsAt),
			zap.Uint64("WatchAt", req.WatchAt),
			zap.Uint64("SysmtemMsgAt", req.SysmtemMsgAt),
		)

		//延时200ms下发
		go func() {
			time.Sleep(200 * time.Millisecond)

			if err := nc.SyncMyInfoAt(username, token, deviceID, req); err != nil {
				nc.logger.Error("SyncMyInfoAt 失败，Error", zap.Error(err))
			} else {
				nc.logger.Debug("SyncMyInfoAt is done")
			}

			if err := nc.SyncFriendsAt(username, token, deviceID, req); err != nil {
				nc.logger.Error("SyncFriendsAt 失败，Error", zap.Error(err))
			} else {
				nc.logger.Debug("SyncFriendsAt is done")
			}

			if err := nc.SyncFriendUsersAt(username, token, deviceID, req); err != nil {
				nc.logger.Error("SyncFriendUsersAt 失败，Error", zap.Error(err))
			} else {
				nc.logger.Debug("SyncFriendUsersAt is done")
			}

			if err := nc.SyncTeamsAt(username, token, deviceID, req); err != nil {
				nc.logger.Error("SyncTeamsAt 失败，Error", zap.Error(err))
			} else {
				nc.logger.Debug("SyncTeamsAt is done")
			}

			if err := nc.SyncTagsAt(username, token, deviceID, req); err != nil {
				nc.logger.Error("SyncTagsAt 失败，Error", zap.Error(err))
			} else {
				nc.logger.Debug("SyncTagsAt is done")
			}

			if err := nc.SyncWatchAt(username, token, deviceID, req); err != nil {
				nc.logger.Error("SyncWatchAt 失败，Error", zap.Error(err))
			} else {
				nc.logger.Debug("SyncWatchAt is done")
			}

			//同步系统公告
			if err := nc.SyncSystemMsgAt(username, token, deviceID, req); err != nil {
				nc.logger.Error("SyncSystemMsgAt 失败，Error", zap.Error(err))
			} else {
				nc.logger.Debug("SyncSystemMsgAt is done")
			}

			//发送SyncDoneEvent
			nc.SendSyncDoneEventToUser(username, deviceID, token)
			nc.logger.Debug("All Sync done")
		}()

	}

COMPLETE:
	msg.SetCode(int32(errorCode)) //状态码
	if errorCode == 200 {
		//只需返回200
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
		nc.logger.Info("Sync message succeed send to ProduceChannel", zap.String("topic", topic))
	} else {
		nc.logger.Error("Failed to send Sync message to ProduceChannel", zap.Error(err))
	}
	_ = err
	_ = token
	return nil

}

//redis 的 OnlineUsers
func (nc *NsqClient) AddOnlineUsers(deviceId string) error {
	redisConn := nc.redisPool.Get()
	defer redisConn.Close()

	if _, err := redisConn.Do("ZADD", "OnlineUsers", time.Now().UnixNano()/1e6, deviceId); err != nil {
		nc.logger.Error("ZADD Error", zap.Error(err))
		return err
	}

	return nil
}

func (nc *NsqClient) SendSyncDoneEventToUser(toUser, deviceID, token string) error {
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
	targetMsg.BuildHeader("Dispatcher", time.Now().UnixNano()/1e6)
	targetMsg.FillBody(data) //网络包的body，承载真正的业务数据
	targetMsg.SetCode(200)   //成功的状态码

	//构建数据完成，向dispatcher发送
	topic := "Sync.Frontend"
	rawData, _ := json.Marshal(targetMsg)
	if err := nc.Producer.Public(topic, rawData); err == nil {
		nc.logger.Info("SendSyncDoneEventToUser, Msg message succeed send to ProduceChannel",
			zap.String("topic", topic),
			zap.String("to", toUser),
			zap.String("toDeviceID:", deviceID),
			zap.String("msgID:", targetMsg.GetID()),
		)
	} else {
		nc.logger.Error("SendSyncDoneEventToUser, Failed to send message to ProduceChannel",
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
