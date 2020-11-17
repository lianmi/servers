/*
本文件是处理业务号是同步模块，分别有
6-1 发起同步请求
	同步处理map，分别处理:
	myInfoAt, friendsAt, friendUsersAt, teamsAt, tagsAt, systemMsgAt, watchAt, productAt,  generalProductAt

6-2 同步请求完成

客户端触发:
1-3 同步当前用户资料事件

*/
package nsqMq

import (
	"encoding/json"
	"fmt"
	// "strings"
	"github.com/pkg/errors"
	"net/http"
	"sync"
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
	// "github.com/lianmi/servers/util/array"
	"go.uber.org/zap"
)

//处理myInfoAt
func (nc *NsqClient) SyncMyInfoAt(username, token, deviceID string, req Sync.SyncEventReq) error {
	var err error
	errorCode := 200
	var errorMsg string
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
		userData := new(models.User)
		userKey := fmt.Sprintf("userData:%s", username)
		if result, err := redis.Values(redisConn.Do("HGETALL", userKey)); err == nil {
			if err := redis.ScanStruct(result, userData); err != nil {
				nc.logger.Error("错误：ScanStruct", zap.Error(err))
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
		return errors.Wrap(err, errorMsg)
	}

}

//处理friendsAt
func (nc *NsqClient) SyncFriendsAt(username, token, deviceID string, req Sync.SyncEventReq) error {
	var err error
	errorCode := 200
	var errorMsg string
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
			RemovedAccounts: make([]*Friends.Friend, 0),
		}

		//从redis里读取username的好友列表
		friends, err := redis.Strings(redisConn.Do("ZRANGEBYSCORE", fmt.Sprintf("Friend:%s:1", username), "-inf", "+inf"))
		if err != nil {
			nc.logger.Error("ZRANGEBYSCORE", zap.Error(err))
			errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
			errorMsg = fmt.Sprintf("错误: ZRANGEBYSCORE error[key=Friend:%s:1]", username)
			goto COMPLETE
		}

		for _, friendUsername := range friends {

			nick, err := redis.String(redisConn.Do("HGET", fmt.Sprintf("FriendInfo:%s:%s", username, friendUsername), "Nick"))
			if err != nil {
				nc.logger.Error("HGET error", zap.Error(err))
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
			nc.logger.Error("ZRANGEBYSCORE", zap.Error(err))
			errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
			errorMsg = fmt.Sprintf("错误: ZRANGEBYSCORE error[key=Friend:%s:2]", username)
			goto COMPLETE
		}

		for _, friendUsername := range RemoveFriends {
			nick, err := redis.String(redisConn.Do("HGET", fmt.Sprintf("FriendInfo:%s:%s", username, friendUsername), "Nick"))
			if err != nil {
				nc.logger.Error("HGET error", zap.Error(err))
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

COMPLETE:
	//完成
	if errorCode == 200 {
		//只需返回200
		return nil
	} else {
		return errors.Wrap(err, errorMsg)
	}
}

//处理 friendUsersAt
func (nc *NsqClient) SyncFriendUsersAt(username, token, deviceID string, req Sync.SyncEventReq) error {
	var err error
	errorCode := 200
	var errorMsg string
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
		fUsers, err := redis.Strings(redisConn.Do("ZRANGEBYSCORE", fmt.Sprintf("Friend:%s:1", username), "-inf", "+inf"))
		if err != nil {
			nc.logger.Error("ZRANGEBYSCORE", zap.Error(err))
			errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
			errorMsg = fmt.Sprintf("错误: ZRANGEBYSCORE error[key=Friend:%s:1]", username)
			goto COMPLETE
		}
		for _, fuser := range fUsers {

			fUserData := new(models.User)
			if result, err := redis.Values(redisConn.Do("HGETALL", fmt.Sprintf("userData:%s", fuser))); err == nil {
				if err := redis.ScanStruct(result, fUserData); err != nil {

					nc.logger.Error("错误：ScanStruct", zap.Error(err))

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
		return errors.Wrap(err, errorMsg)
	}
}

//处理 TeamsAt
func (nc *NsqClient) SyncTeamsAt(username, token, deviceID string, req Sync.SyncEventReq) error {
	var err error
	errorCode := 200
	var errorMsg string
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
			errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
			errorMsg = fmt.Sprintf("错误: ZRANGEBYSCORE error[key=Team:%s]", username)
			goto COMPLETE
		}

		for _, teamID := range teamIDs {
			nc.logger.Debug("for..range teamIDs", zap.String("teamID", teamID))
			teamInfo := new(models.Team)
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
		return errors.Wrap(err, errorMsg)
	}
}

//1-7 同步用户标签列表 处理 TagsAt
func (nc *NsqClient) SyncTagsAt(username, token, deviceID string, req Sync.SyncEventReq) error {
	var err error
	errorCode := 200
	var errorMsg string
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
			errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
			errorMsg = fmt.Sprintf("ZRANGEBYSCORE error[BlackList:%s:1]", username)
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
			errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
			errorMsg = fmt.Sprintf("ZRANGEBYSCORE error[MutedList:%s:1]", username)
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
			errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
			errorMsg = fmt.Sprintf("ZRANGEBYSCORE error[StickyList:%s:1]", username)
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
			errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
			errorMsg = fmt.Sprintf("ZRANGEBYSCORE error[BlackList:%s:2]", username)
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
			errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
			errorMsg = fmt.Sprintf("ZRANGEBYSCORE error[MutedList:%s:2]", username)
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
			errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
			errorMsg = fmt.Sprintf("ZRANGEBYSCORE error[StickyList:%s:2]", username)
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

		targetMsg.BuildHeader("AuthService", time.Now().UnixNano()/1e6)

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

	targetMsg.BuildHeader("ChatService", time.Now().UnixNano()/1e6)

	targetMsg.FillBody(data) //网络包的body，承载真正的业务数据

	targetMsg.SetCode(200) //成功的状态码

	//构建数据完成，向dispatcher发送
	topic := "Msg.Frontend"
	rawData, _ := json.Marshal(targetMsg)

	go nc.Producer.Public(topic, rawData)

	nc.logger.Info("SyncOfflineSysMsgsEvent Succeed",
		zap.String("Username:", toUser),
		zap.String("DeviceID:", deviceID),
		zap.Int64("Now", time.Now().UnixNano()/1e6))

	return nil
}

//处理离线系统通知 systemMsgAt
func (nc *NsqClient) SyncSystemMsgAt(username, token, deviceID string, req Sync.SyncEventReq) error {
	var err error
	errorCode := 200
	var errorMsg string
	var cur_systemMsgAt uint64

	offLineMsgListKey := fmt.Sprintf("offLineMsgList:%s", username)

	redisConn := nc.redisPool.Get()
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
		// 过去的7天
		yesTime := nTime.AddDate(0, 0, -7).UnixNano() / 1e6

		//移除时间少于yesTime离线通知
		_, err = redisConn.Do("ZREMRANGEBYSCORE", offLineMsgListKey, "-inf", yesTime)
		if err != nil {
			nc.logger.Error("ZRANGEBYSCORE执行错误", zap.Error(err))
			errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
			errorMsg = fmt.Sprintf("错误: ZRANGEBYSCORE error[key=%s]", offLineMsgListKey)
			goto COMPLETE
		} else {
			nc.logger.Debug("移除时间少于yesTime(7天)离线通知", zap.Int64("yesTime", yesTime))
		}

		//将有效时间内的离线消息推送给SDK
		msgIDs, _ := redis.Strings(redisConn.Do("ZRANGEBYSCORE", offLineMsgListKey, yesTime, "+inf"))
		for _, msgID := range msgIDs {
			systemMsgKey := fmt.Sprintf("systemMsg:%s:%s", username, msgID)
			//取出缓存的离线消息体
			if data, err := redis.Bytes(redisConn.Do("HGET", systemMsgKey, "Data")); err != nil {
				nc.logger.Error("HGET读取 错误",
					zap.String("systemMsgKey", systemMsgKey),
					zap.Error(err))
				continue

			} else {
				if err := nc.SendOffLineMsg(username, token, deviceID, data); err != nil {
					nc.logger.Error("发送离线通知错误",
						zap.String("systemMsgKey", systemMsgKey),
						zap.Error(err))

				} else {
					nc.logger.Debug("成功发送离线通知", zap.String("msgID", msgID))
				}
			}

		}

		/*
			//同步离线期间的个人，群聊，订单，商品推送的消息 消息只保留最后10条(common.go里定义)，而且只能缓存7天
			//从大到小递减获取, 只获取最大的10条
			msgIDArray, err := redis.ByteSlices(redisConn.Do("ZREVRANGEBYSCORE", offLineMsgListKey, "+inf", "-inf", "LIMIT", 0, common.OffLineMsgCount))
			if err != nil {
				nc.logger.Error("ZREVRANGEBYSCORE Error", zap.Error(err))
			} else {
				if len(msgIDArray) > 0 {
					//反转，数组变为从小到大排序
					array.ReverseBytes(msgIDArray)

					var maxSeq int
					for seq, msgID := range msgIDArray {
						if seq > maxSeq {
							maxSeq = seq
						}
						nc.logger.Debug("同步离线期间离线消息",
							zap.Int("seq", seq),
							zap.Int("maxSeq", maxSeq),
							zap.String("msgID", string(msgID)),
						)
						key := fmt.Sprintf("systemMsg:%s:%s", username, string(msgID))
						data, err := redis.Bytes(redisConn.Do("HGET", key, "Data"))
						if err != nil {
							nc.logger.Error("HGET Data Error", zap.Error(err))
							continue
						}

						if err := nc.SendOffLineMsg(username, token, deviceID, data); err == nil {
							nc.logger.Debug("成功发送离线消息",
								zap.Int("seq", seq),
								zap.String("msgID", string(msgID)),
								zap.String("username", username),
								zap.String("token", token),
								zap.String("deviceID", deviceID),
							)
						} else {
							nc.logger.Error("发送离线消息失败，Error", zap.Error(err))
						}

					}

				}
			}
		*/
	}

COMPLETE:
	//完成
	if errorCode == 200 {
		//只需返回200
		return nil
	} else {
		return errors.Wrap(err, errorMsg)
	}
	return nil
}

//处理watchAt 7-8 同步关注的商户事件
func (nc *NsqClient) SyncWatchAt(username, token, deviceID string, req Sync.SyncEventReq) error {
	var err error
	errorCode := 200
	var errorMsg string
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
		return errors.Wrap(err, errorMsg)
	}
}

//处理productAt 7-8 同步商品列表
func (nc *NsqClient) SyncProductAt(username, token, deviceID string, req Sync.SyncEventReq) error {
	var err error
	errorCode := 200
	var errorMsg string
	var cur_productAt uint64

	rsp := &Order.SyncProductsEventRsp{
		TimeTag:           uint64(time.Now().UnixNano() / 1e6),
		AddProducts:       make([]*Order.Product, 0), //新上架或更新的商品列表
		RemovedProductIDs: make([]string, 0),         //下架的商品ID列表

	}
	redisConn := nc.redisPool.Get()
	defer redisConn.Close()

	//req里的成员
	productAt := req.GetProductAt()
	syncKey := fmt.Sprintf("sync:%s", username)

	cur_productAt, err = redis.Uint64(redisConn.Do("HGET", syncKey, "productAt"))
	if err != nil {
		cur_productAt = uint64(time.Now().UnixNano() / 1e6)
		redisConn.Do("HSET", syncKey, "productAt", cur_productAt)

	}

	nc.logger.Debug("SyncProductAt",
		zap.Uint64("cur_productAt", cur_productAt),
		zap.Uint64("productAt", productAt),
		zap.String("username", username),
	)

	//服务端的时间戳大于客户端上报的时间戳
	if cur_productAt > productAt {
		//获取此用户关注的商户列表
		watchingUsers, err := redis.Strings(redisConn.Do("ZRANGEBYSCORE", fmt.Sprintf("Watching:%s", username), "-inf", "+inf"))
		if err != nil {

			nc.logger.Error("ZRANGEBYSCORE", zap.Error(err))
			errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
			errorMsg = fmt.Sprintf("ZRANGEBYSCORE error[Watching:%s]", username)
			goto COMPLETE
		}

		for _, watchingUser := range watchingUsers {
			//商户 的商品列表
			productIDs, _ := redis.Strings(redisConn.Do("ZRANGEBYSCORE", fmt.Sprintf("Products:%s", watchingUser), productAt, "+inf"))
			for _, productID := range productIDs {
				productInfo := new(models.Product)
				if result, err := redis.Values(redisConn.Do("HGETALL", fmt.Sprintf("Product:%s", productID))); err == nil {
					if err := redis.ScanStruct(result, productInfo); err != nil {
						nc.logger.Error("错误: ScanStruct", zap.Error(err))
						continue
					}
				}
				rsp.AddProducts = append(rsp.AddProducts, &Order.Product{
					ProductId:         productID,
					Expire:            uint64(productInfo.Expire),
					ProductName:       productInfo.ProductName,
					ProductType:       Global.ProductType(productInfo.ProductType),
					ProductDesc:       productInfo.ProductDesc,
					ProductPic1Small:  productInfo.ProductPic1Small,
					ProductPic1Middle: productInfo.ProductPic1Middle,
					ProductPic1Large:  productInfo.ProductPic1Large,
					ProductPic2Small:  productInfo.ProductPic2Small,
					ProductPic2Middle: productInfo.ProductPic2Middle,
					ProductPic2Large:  productInfo.ProductPic2Large,
					ProductPic3Small:  productInfo.ProductPic3Small,
					ProductPic3Middle: productInfo.ProductPic3Middle,
					ProductPic3Large:  productInfo.ProductPic3Large,
					Thumbnail:         productInfo.Thumbnail,
					ShortVideo:        productInfo.ShortVideo,
					Price:             productInfo.Price,
					LeftCount:         productInfo.LeftCount,
					Discount:          productInfo.Discount,
					DiscountDesc:      productInfo.DiscountDesc,
					DiscountStartTime: uint64(productInfo.DiscountStartTime),
					DiscountEndTime:   uint64(productInfo.DiscountEndTime),
					CreateAt:          uint64(productInfo.CreateAt),
					ModifyAt:          uint64(productInfo.ModifyAt),
					AllowCancel:       productInfo.AllowCancel,
				})
			}

			//下架的商品ID

			removeProductIDs, _ := redis.Strings(redisConn.Do("ZRANGEBYSCORE", fmt.Sprintf("RemoveProducts:%s", watchingUser), productAt, "+inf"))
			for _, removeProductID := range removeProductIDs {
				rsp.RemovedProductIDs = append(rsp.RemovedProductIDs, removeProductID)
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
		targetMsg.SetBusinessTypeName("Order")
		targetMsg.SetBusinessType(uint32(Global.BusinessType_Product))               // 7
		targetMsg.SetBusinessSubType(uint32(Global.ProductSubType_SyncProductEvent)) // 8

		targetMsg.BuildHeader("AuthService", time.Now().UnixNano()/1e6)

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

		nc.logger.Info("SyncProductAt Succeed",
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
		return errors.Wrap(err, errorMsg)
	}
}

//处理productAt 7-9 同步商品列表
func (nc *NsqClient) SyncGeneralProductAt(username, token, deviceID string, req Sync.SyncEventReq) error {
	var err error
	errorCode := 200
	var errorMsg string
	var cur_generalProductAt uint64

	rsp := &Order.SyncGeneralProductsEventRsp{
		TimeTag:           uint64(time.Now().UnixNano() / 1e6),
		AddProducts:       make([]*Order.GeneralProduct, 0), //通用商品列表
		RemovedProductIDs: make([]string, 0),                //删除的通用 商品ID列表

	}
	redisConn := nc.redisPool.Get()
	defer redisConn.Close()

	//req里的成员
	generalProductAt := req.GetGeneralProductAt()
	syncKey := fmt.Sprintf("sync:%s", username)

	cur_generalProductAt, err = redis.Uint64(redisConn.Do("HGET", syncKey, "generalProductAt"))
	if err != nil {
		cur_generalProductAt = uint64(time.Now().UnixNano() / 1e6)
		redisConn.Do("HSET", syncKey, "generalProductAt", cur_generalProductAt)

	}

	nc.logger.Debug("GeneralProductAt",
		zap.Uint64("cur_generalProductAt", cur_generalProductAt),
		zap.Uint64("generalProductAt", generalProductAt),
		zap.String("username", username),
	)

	//服务端的时间戳大于客户端上报的时间戳
	if cur_generalProductAt > generalProductAt {
		//获取redis通用商品列表
		productIDs, _ := redis.Strings(redisConn.Do("ZRANGEBYSCORE", "GeneralProducts", generalProductAt, "+inf"))
		for _, productID := range productIDs {
			productInfo := new(models.GeneralProduct)
			if result, err := redis.Values(redisConn.Do("HGETALL", fmt.Sprintf("GeneralProduct:%s", productID))); err == nil {
				if err := redis.ScanStruct(result, productInfo); err != nil {
					nc.logger.Error("错误: ScanStruct", zap.Error(err))
					continue
				}
			}
			rsp.AddProducts = append(rsp.AddProducts, &Order.GeneralProduct{
				ProductId:         productID,
				ProductName:       productInfo.ProductName,
				ProductType:       Global.ProductType(productInfo.ProductType),
				ProductDesc:       productInfo.ProductDesc,
				ProductPic1Small:  productInfo.ProductPic1Small,
				ProductPic1Middle: productInfo.ProductPic1Middle,
				ProductPic1Large:  productInfo.ProductPic1Large,
				ProductPic2Small:  productInfo.ProductPic2Small,
				ProductPic2Middle: productInfo.ProductPic2Middle,
				ProductPic2Large:  productInfo.ProductPic2Large,
				ProductPic3Small:  productInfo.ProductPic3Small,
				ProductPic3Middle: productInfo.ProductPic3Middle,
				ProductPic3Large:  productInfo.ProductPic3Large,
				Thumbnail:         productInfo.Thumbnail,
				ShortVideo:        productInfo.ShortVideo,
				CreateAt:          uint64(productInfo.CreateAt),
				ModifyAt:          uint64(productInfo.ModifyAt),
				AllowCancel:       productInfo.AllowCancel,
			})
		}

		//删除的通用商品ID列表

		removeProductIDs, _ := redis.Strings(redisConn.Do("ZRANGEBYSCORE", "RemoveGeneralProducts", generalProductAt, "+inf"))
		for _, removeProductID := range removeProductIDs {
			rsp.RemovedProductIDs = append(rsp.RemovedProductIDs, removeProductID)
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
		targetMsg.SetBusinessTypeName("Order")
		targetMsg.SetBusinessType(uint32(Global.BusinessType_Product))                       // 7
		targetMsg.SetBusinessSubType(uint32(Global.ProductSubType_SyncGeneralProductsEvent)) // 9

		targetMsg.BuildHeader("AuthService", time.Now().UnixNano()/1e6)

		targetMsg.FillBody(data) //网络包的body，承载真正的业务数据

		targetMsg.SetCode(200) //成功的状态码

		//构建数据完成，向dispatcher发送
		topic := "Auth.Frontend"
		rawData, _ := json.Marshal(targetMsg)
		if err := nc.Producer.Public(topic, rawData); err == nil {
			nc.logger.Info("Message succeed send to ProduceChannel",
				zap.String("topic", topic),
				zap.String("Username:", username),
				zap.String("DeviceID:", deviceID),
				zap.Int64("Now", time.Now().UnixNano()/1e6))

		} else {
			nc.logger.Error("Failed to send message to ProduceChannel", zap.Error(err))
		}

	}

	//完成
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
func (nc *NsqClient) HandleSync(msg *models.Message) error {
	var err error

	errorCode := 200
	var errorMsg string

	redisConn := nc.redisPool.Get()
	defer redisConn.Close()

	username := msg.GetUserName()
	token := msg.GetJwtToken()
	deviceID := msg.GetDeviceID()

	nc.logger.Info("HandleSync start...",
		zap.String("username", username),
		zap.String("DeviceId", deviceID))

	//取出当前设备的os， clientType， logonAt
	curDeviceHashKey := fmt.Sprintf("devices:%s:%s", username, deviceID)
	isMaster, _ := redis.Bool(redisConn.Do("HGET", curDeviceHashKey, "ismaster"))
	curOs, _ := redis.String(redisConn.Do("HGET", curDeviceHashKey, "os"))
	curClientType, _ := redis.Int(redisConn.Do("HGET", curDeviceHashKey, "clientType"))
	curLogonAt, _ := redis.Uint64(redisConn.Do("HGET", curDeviceHashKey, "logonAt"))

	nc.logger.Debug("HandleSync",
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
		nc.logger.Error("Protobuf Unmarshal Error", zap.Error(err))
		goto COMPLETE

	} else {
		nc.logger.Debug("Sync  payload",
			zap.Uint64("MyInfoAt", req.MyInfoAt),
			zap.Uint64("FriendsAt", req.FriendsAt),
			zap.Uint64("FriendUsersAt", req.FriendUsersAt),
			zap.Uint64("TeamsAt", req.TeamsAt),
			zap.Uint64("TagsAt", req.TagsAt),
			zap.Uint64("SystemMsgAt", req.SystemMsgAt),
			zap.Uint64("WatchAt", req.WatchAt),
			zap.Uint64("ProductAt", req.ProductAt),
			zap.Uint64("GeneralProductAt", req.GeneralProductAt),
		)

		//所有同步的时间戳数量
		var wg sync.WaitGroup
		wg.Add(common.TotalSyncCount)

		//异步
		go func() {
			defer wg.Done()

			if err := nc.SyncMyInfoAt(username, token, deviceID, req); err != nil {
				nc.logger.Error("SyncMyInfoAt 失败，Error", zap.Error(err))
			} else {
				nc.logger.Debug("SyncMyInfoAt is done")
			}
		}()

		go func() {
			defer wg.Done()

			if err := nc.SyncFriendsAt(username, token, deviceID, req); err != nil {
				nc.logger.Error("SyncFriendsAt 失败，Error", zap.Error(err))
			} else {
				nc.logger.Debug("SyncFriendsAt is done")
			}
		}()

		go func() {
			defer wg.Done()

			if err := nc.SyncFriendUsersAt(username, token, deviceID, req); err != nil {
				nc.logger.Error("SyncFriendUsersAt 失败，Error", zap.Error(err))
			} else {
				nc.logger.Debug("SyncFriendUsersAt is done")
			}
		}()

		go func() {
			defer wg.Done()

			if err := nc.SyncTeamsAt(username, token, deviceID, req); err != nil {
				nc.logger.Error("SyncTeamsAt 失败，Error", zap.Error(err))
			} else {
				nc.logger.Debug("SyncTeamsAt is done")
			}
		}()

		go func() {
			defer wg.Done()

			if err := nc.SyncTagsAt(username, token, deviceID, req); err != nil {
				nc.logger.Error("SyncTagsAt 失败，Error", zap.Error(err))
			} else {
				nc.logger.Debug("SyncTagsAt is done")
			}
		}()

		go func() {
			defer wg.Done()

			if err := nc.SyncSystemMsgAt(username, token, deviceID, req); err != nil {
				nc.logger.Error("SyncSystemMsgAt 失败，Error", zap.Error(err))
			} else {
				nc.logger.Debug("SyncSystemMsgAt is done")
			}
		}()

		go func() {
			defer wg.Done()

			if err := nc.SyncWatchAt(username, token, deviceID, req); err != nil {
				nc.logger.Error("SyncWatchAt 失败，Error", zap.Error(err))
			} else {
				nc.logger.Debug("SyncWatchAt is done")
			}
		}()

		go func() {
			defer wg.Done()

			if err := nc.SyncProductAt(username, token, deviceID, req); err != nil {
				nc.logger.Error("SyncProductAt 失败，Error", zap.Error(err))
			} else {
				nc.logger.Debug("SyncProductAt is done")
			}
		}()

		go func() {
			defer wg.Done()

			if err := nc.SyncGeneralProductAt(username, token, deviceID, req); err != nil {
				nc.logger.Error("GeneralProductAt 失败，Error", zap.Error(err))
			} else {
				nc.logger.Debug("GeneralProductAt is done")
			}
		}()

		// 等待执行结束
		wg.Wait()

		//发送SyncDoneEvent
		nc.SendSyncDoneEventToUser(username, deviceID, token)
		nc.logger.Debug("All Sync done")

	}

COMPLETE:
	msg.SetCode(int32(errorCode)) //状态码
	if errorCode == 200 {
		//只需返回200
		msg.FillBody(nil)
	} else {
		msg.SetErrorMsg([]byte(errorMsg)) //错误提示
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
	targetMsg.BuildHeader("AuthService", time.Now().UnixNano()/1e6)
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