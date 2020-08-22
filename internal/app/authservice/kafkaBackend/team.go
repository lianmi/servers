package kafkaBackend

import (
	"fmt"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/gomodule/redigo/redis"
	Msg "github.com/lianmi/servers/api/proto/msg"
	Team "github.com/lianmi/servers/api/proto/team"
	"github.com/lianmi/servers/internal/common"
	"github.com/lianmi/servers/internal/pkg/models"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

/*
4-1 创建群组
*/
func (kc *KafkaClient) HandleCreateTeam(msg *models.Message) error {
	var err error
	var errorMsg string
	rsp := &Team.CreateTeamRsp{}
	var data []byte

	redisConn := kc.redisPool.Get()
	defer redisConn.Close()

	username := msg.GetUserName() //用户自己的账号
	// token := msg.GetJwtToken()
	deviceID := msg.GetDeviceID()

	kc.logger.Info("HandleCreateTeam start...",
		zap.String("username", username),
		zap.String("deviceId", deviceID))

	//取出当前设备的os， clientType， logonAt
	curDeviceHashKey := fmt.Sprintf("devices:%s:%s", username, deviceID)
	isMaster, _ := redis.Bool(redisConn.Do("HGET", curDeviceHashKey, "ismaster"))
	curOs, _ := redis.String(redisConn.Do("HGET", curDeviceHashKey, "os"))
	curClientType, _ := redis.Int(redisConn.Do("HGET", curDeviceHashKey, "clientType"))
	curLogonAt, _ := redis.Uint64(redisConn.Do("HGET", curDeviceHashKey, "logonAt"))

	kc.logger.Debug("CreateTeam",
		zap.Bool("isMaster", isMaster),
		zap.String("username", username),
		zap.String("deviceID", deviceID),
		zap.String("curOs", curOs),
		zap.Int("curClientType", curClientType),
		zap.Uint64("curLogonAt", curLogonAt))

	//打开msg里的负载， 获取请求参数
	body := msg.GetContent()

	//解包body
	req := &Team.CreateTeamReq{}
	if err := proto.Unmarshal(body, req); err != nil {
		kc.logger.Error("Protobuf Unmarshal Error", zap.Error(err))
		errorMsg = fmt.Sprintf("Protobuf Unmarshal Error: %s", err.Error())
		goto COMPLETE

	} else {
		kc.logger.Debug("CreateTeam body",
			zap.String("群主账号", req.GetOwner()),
			zap.Int("群类型", int(req.GetType())), // Normal(1) - 普通群 Advanced(2) - vip群
			zap.String("群组名称", req.GetName()),
			zap.Int("verifyType", int(req.GetVerifyType())), //如果普通群，只能选4，如果vip群，可以选其它
			zap.Int("inviteMode", int(req.GetInviteMode())), //邀请模式
		)

		teamOwner := req.GetOwner()

		if isExists, err := redis.Bool(redisConn.Do("EXISTS", fmt.Sprintf("userData:%s", teamOwner))); err != nil {
			errorMsg = fmt.Sprintf("Query user error[teamOwner=%s]", teamOwner)
			goto COMPLETE

		} else {
			if !isExists {
				err = errors.Wrapf(err, "Owner is not exists[teamOwner=%s]", teamOwner)
				errorMsg = fmt.Sprintf("Owner is not exists[teamOwner=%s]", teamOwner)
				goto COMPLETE
			}

			//判断群主是否已经注册为网点用户类型
			userType, _ := redis.Int(redisConn.Do("HGET", fmt.Sprintf("userData:%s", teamOwner), "UserType"))
			if userType != 2 {
				err = errors.Wrapf(err, "userType is not business type [userType=%d]", userType)
				errorMsg = fmt.Sprintf("serType is not business type [userType=%d]", userType)
				goto COMPLETE
			}

			//写入MySQL数据库

			var newTeamIndex uint64
			if newTeamIndex, err = redis.Uint64(redisConn.Do("INCR", "TeamIndex")); err != nil {
				kc.logger.Error("redisConn GET TeamIndex Error", zap.Error(err))
				errorMsg = fmt.Sprintf("serType is not business type [userType=%d]", userType)
				goto COMPLETE
			}

			pTeam := new(models.Team)
			pTeam.CreatedAt = time.Now().Unix()
			pTeam.TeamID = fmt.Sprintf("team%d", newTeamIndex) //群id， 自动生成
			pTeam.Teamname = req.GetName()
			pTeam.Nick = req.GetName()
			pTeam.Owner = req.GetOwner()
			pTeam.Type = int(req.GetType())
			pTeam.VerifyType = int(req.GetVerifyType())
			pTeam.InviteMode = int(req.GetInviteMode())

			//默认的设置
			pTeam.Status = 1 //Init(1) - 初始状态,审核中 Normal(2) - 正常状态 Blocked(3) - 封禁状态
			pTeam.MemberLimit = 600
			pTeam.MemberNum = 1  //刚刚建群是只有群主
			pTeam.MuteType = 1   //None(1) - 所有人可发言
			pTeam.InviteMode = 1 //邀请模式,初始为1

			if err = kc.SaveCreateTeam(pTeam); err != nil {
				kc.logger.Error("Save CreateTeam Error", zap.Error(err))
				errorMsg = "无法保存到数据库"
				goto COMPLETE
			}

			//这里不写入redis，等管理员审核通过后才写入redis里

			//回包
			rsp.TeamInfo = &Team.TeamInfo{
				TeamId:       pTeam.TeamID,
				Name:         pTeam.Teamname,
				Icon:         "",
				Announcement: "",
				Introduce:    "",
				Owner:        pTeam.Owner,
				Type:         Team.TeamType(pTeam.Type),
				VerifyType:   Team.VerifyType(pTeam.VerifyType),
				MemberLimit:  int32(pTeam.MemberLimit),
				MemberNum:    int32(pTeam.MemberNum),
				Status:       Team.Status(pTeam.Status),
				MuteType:     Team.MuteMode(pTeam.MuteType),
				InviteMode:   Team.InviteMode(pTeam.InviteMode),
				Ex:           pTeam.Extend,
			}

			data, _ = proto.Marshal(rsp)
		}
	}

COMPLETE:
	if err != nil {
		msg.SetCode(400)                  //状态码
		msg.SetErrorMsg([]byte(errorMsg)) //错误提示
		msg.FillBody(nil)

	} else {
		msg.SetCode(200) //状态码
		data, _ = proto.Marshal(rsp)
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
4-2 获取群组成员
*/
func (kc *KafkaClient) HandleGetTeamMembers(msg *models.Message) error {
	var err error
	var errorMsg string
	rsp := &Team.GetTeamMembersRsp{
		TimeAt:       uint64(time.Now().Unix()),
		Tmembers:     make([]*Team.Tmember, 0),
		RemovedUsers: make([]string, 0),
	}
	var data []byte

	redisConn := kc.redisPool.Get()
	defer redisConn.Close()

	username := msg.GetUserName() //用户自己的账号
	// token := msg.GetJwtToken()
	deviceID := msg.GetDeviceID()

	kc.logger.Info("HandleGetTeamMembers start...",
		zap.String("username", username),
		zap.String("deviceId", deviceID))

	//取出当前设备的os， clientType， logonAt
	curDeviceHashKey := fmt.Sprintf("devices:%s:%s", username, deviceID)
	isMaster, _ := redis.Bool(redisConn.Do("HGET", curDeviceHashKey, "ismaster"))
	curOs, _ := redis.String(redisConn.Do("HGET", curDeviceHashKey, "os"))
	curClientType, _ := redis.Int(redisConn.Do("HGET", curDeviceHashKey, "clientType"))
	curLogonAt, _ := redis.Uint64(redisConn.Do("HGET", curDeviceHashKey, "logonAt"))

	kc.logger.Debug("GetTeamMembers",
		zap.Bool("isMaster", isMaster),
		zap.String("username", username),
		zap.String("deviceID", deviceID),
		zap.String("curOs", curOs),
		zap.Int("curClientType", curClientType),
		zap.Uint64("curLogonAt", curLogonAt))

	//打开msg里的负载， 获取请求参数
	body := msg.GetContent()

	//解包body
	req := &Team.GetTeamMembersReq{}
	if err := proto.Unmarshal(body, req); err != nil {
		kc.logger.Error("Protobuf Unmarshal Error", zap.Error(err))
		errorMsg = fmt.Sprintf("Protobuf Unmarshal Error: %s", err.Error())
		goto COMPLETE

	} else {
		kc.logger.Debug("GetTeamMembers body",
			zap.String("teamId", req.GetTeamId()),
			zap.Int("timeAt", int(req.GetTimeAt())),
		)

		teamID := req.GetTeamId()

		if isExists, err := redis.Bool(redisConn.Do("EXISTS", fmt.Sprintf("TeamInfo:%s", teamID))); err != nil {
			errorMsg = fmt.Sprintf("Query team info error[teamID=%s]", teamID)
			goto COMPLETE

		} else {
			if !isExists {
				err = errors.Wrapf(err, "Team is not exists[teamID=%s]", teamID)
				errorMsg = fmt.Sprintf("Team is not exists[teamID=%s]", teamID)
				goto COMPLETE
			}

			//redis查出此群的成员
			teamMembers, _ := redis.Strings(redisConn.Do("ZRANGEBYSCORE", fmt.Sprintf("TeamUsers:%s", teamID), "-inf", "+inf"))
			for _, teamMember := range teamMembers {
				key := fmt.Sprintf("TeamUser:%s:%s", teamID, teamMember)
				teamUser := new(models.TeamUser)
				if result, err := redis.Values(redisConn.Do("HGETALL", key)); err == nil {
					if err := redis.ScanStruct(result, teamUser); err != nil {

						kc.logger.Error("错误：ScanStruct", zap.Error(err))
						continue
					}
				}

				rsp.Tmembers = append(rsp.Tmembers, &Team.Tmember{
					TeamId:     teamID,
					Username:   teamUser.Username,
					Nick:       teamUser.Nick,
					Avatar:     teamUser.Avatar,
					Source:     teamUser.Source,
					Type:       Team.TeamMemberType(teamUser.TeamMemberType),
					NotifyType: Team.NotifyType(teamUser.NotifyType),
					Mute:       teamUser.IsMute,
					Ex:         teamUser.Extend,
					JoinTime:   uint64(teamUser.JoinAt),
				})
			}

			//被移除的成员列表
			teamRemoveMembers, _ := redis.Strings(redisConn.Do("ZRANGEBYSCORE", fmt.Sprintf("TeamUsersRemoved:%s", teamID), "-inf", "+inf"))
			for _, teamRemoveMember := range teamRemoveMembers {
				rsp.RemovedUsers = append(rsp.RemovedUsers, teamRemoveMember)
			}

			//回包
			data, _ = proto.Marshal(rsp)
		}
	}

COMPLETE:
	if err != nil {
		msg.SetCode(400)                  //状态码
		msg.SetErrorMsg([]byte(errorMsg)) //错误提示
		msg.FillBody(nil)

	} else {
		msg.SetCode(200) //状态码
		data, _ = proto.Marshal(rsp)
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
4-3 查询群信息
*/
func (kc *KafkaClient) HandleGetTeam(msg *models.Message) error {
	var err error
	var errorMsg string
	var count int
	rsp := &Team.GetTeamRsp{}
	var data []byte

	redisConn := kc.redisPool.Get()
	defer redisConn.Close()

	username := msg.GetUserName() //用户自己的账号
	// token := msg.GetJwtToken()
	deviceID := msg.GetDeviceID()

	kc.logger.Info("HandleGetTeam start...",
		zap.String("username", username),
		zap.String("deviceId", deviceID))

	//取出当前设备的os， clientType， logonAt
	curDeviceHashKey := fmt.Sprintf("devices:%s:%s", username, deviceID)
	isMaster, _ := redis.Bool(redisConn.Do("HGET", curDeviceHashKey, "ismaster"))
	curOs, _ := redis.String(redisConn.Do("HGET", curDeviceHashKey, "os"))
	curClientType, _ := redis.Int(redisConn.Do("HGET", curDeviceHashKey, "clientType"))
	curLogonAt, _ := redis.Uint64(redisConn.Do("HGET", curDeviceHashKey, "logonAt"))

	kc.logger.Debug("GetTeam",
		zap.Bool("isMaster", isMaster),
		zap.String("username", username),
		zap.String("deviceID", deviceID),
		zap.String("curOs", curOs),
		zap.Int("curClientType", curClientType),
		zap.Uint64("curLogonAt", curLogonAt))

	//打开msg里的负载， 获取请求参数
	body := msg.GetContent()

	//解包body
	req := &Team.GetTeamReq{}
	if err := proto.Unmarshal(body, req); err != nil {
		kc.logger.Error("Protobuf Unmarshal Error", zap.Error(err))
		errorMsg = fmt.Sprintf("Protobuf Unmarshal Error: %s", err.Error())
		goto COMPLETE

	} else {
		kc.logger.Debug("GetTeam body",
			zap.String("teamId", req.GetTeamId()),
		)

		teamID := req.GetTeamId()

		if isExists, err := redis.Bool(redisConn.Do("EXISTS", fmt.Sprintf("TeamInfo:%s", teamID))); err != nil {
			errorMsg = fmt.Sprintf("Query team info error[teamID=%s]", teamID)
			goto COMPLETE

		} else {
			if !isExists {
				err = errors.Wrapf(err, "Team is not exists[teamID=%s]", teamID)
				errorMsg = fmt.Sprintf("Team is not exists[teamID=%s]", teamID)
				goto COMPLETE
			}

			//Zcount 用于计算有序集合中指定分数区间的成员数量。
			if count, err = redis.Int(redisConn.Do("ZCOUNT", fmt.Sprintf("TeamUsers:%s", teamID), "-inf", "+inf")); err != nil {
				kc.logger.Error("ZCOUNT Error", zap.Error(err))
				errorMsg = fmt.Sprintf("TeamUsers is not exists[teamID=%s]", teamID)
				goto COMPLETE
			}

			//redis查出此群的成员
			key := fmt.Sprintf("TeamInfo:%s", teamID)
			teamInfo := new(models.Team)
			if result, err := redis.Values(redisConn.Do("HGETALL", key)); err == nil {
				if err := redis.ScanStruct(result, teamInfo); err != nil {
					kc.logger.Error("错误：ScanStruct", zap.Error(err))
					errorMsg = fmt.Sprintf("Team is not exists[teamID=%s]", teamID)
					goto COMPLETE
				}

				rsp.TeamInfo = &Team.TeamInfo{
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
				}
			}

			//回包
			data, _ = proto.Marshal(rsp)
		}
	}

COMPLETE:
	if err != nil {
		msg.SetCode(400)                  //状态码
		msg.SetErrorMsg([]byte(errorMsg)) //错误提示
		msg.FillBody(nil)

	} else {
		msg.SetCode(200) //状态码
		data, _ = proto.Marshal(rsp)
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
4-4 邀请用户加群
*/
func (kc *KafkaClient) HandleInviteTeamMembers(msg *models.Message) error {
	var err error
	var errorMsg string
	var count int
	rsp := &Team.InviteTeamMembersRsp{
		AbortedUsers: make([]string, 0),
	}
	var data []byte

	redisConn := kc.redisPool.Get()
	defer redisConn.Close()

	username := msg.GetUserName() //用户自己的账号
	// token := msg.GetJwtToken()
	deviceID := msg.GetDeviceID()
	userNick, _ := redis.String(redisConn.Do("HGET", fmt.Sprintf("userData:%s", username), "Nick"))

	kc.logger.Info("HandleInviteTeamMembers start...",
		zap.String("username", username),
		zap.String("deviceId", deviceID))

	//取出当前设备的os， clientType， logonAt
	curDeviceHashKey := fmt.Sprintf("devices:%s:%s", username, deviceID)
	isMaster, _ := redis.Bool(redisConn.Do("HGET", curDeviceHashKey, "ismaster"))
	curOs, _ := redis.String(redisConn.Do("HGET", curDeviceHashKey, "os"))
	curClientType, _ := redis.Int(redisConn.Do("HGET", curDeviceHashKey, "clientType"))
	curLogonAt, _ := redis.Uint64(redisConn.Do("HGET", curDeviceHashKey, "logonAt"))

	kc.logger.Debug("InviteTeamMembers",
		zap.Bool("isMaster", isMaster),
		zap.String("username", username),
		zap.String("deviceID", deviceID),
		zap.String("curOs", curOs),
		zap.Int("curClientType", curClientType),
		zap.Uint64("curLogonAt", curLogonAt))

	//打开msg里的负载， 获取请求参数
	body := msg.GetContent()

	//解包body
	req := &Team.InviteTeamMembersReq{}
	if err := proto.Unmarshal(body, req); err != nil {
		kc.logger.Error("Protobuf Unmarshal Error", zap.Error(err))
		errorMsg = fmt.Sprintf("Protobuf Unmarshal Error: %s", err.Error())
		goto COMPLETE

	} else {
		kc.logger.Debug("InviteTeamMembers body",
			zap.String("teamId", req.GetTeamId()),
			zap.String("ps", req.GetPs()),
			zap.Strings("usernames", req.GetUsernames()),
		)

		teamID := req.GetTeamId()

		//判断 teamID 是否存在
		if isExists, err := redis.Bool(redisConn.Do("EXISTS", fmt.Sprintf("TeamInfo:%s", teamID))); err != nil {
			errorMsg = fmt.Sprintf("Query team info error[teamID=%s]", teamID)
			goto COMPLETE

		} else {
			if !isExists {
				err = errors.Wrapf(err, "Team is not exists[teamID=%s]", teamID)
				errorMsg = fmt.Sprintf("Team is not exists[teamID=%s]", teamID)
				goto COMPLETE
			}

			//获取到群信息
			key := fmt.Sprintf("TeamInfo:%s", teamID)
			teamInfo := new(models.Team)
			if result, err := redis.Values(redisConn.Do("HGETALL", key)); err == nil {
				if err := redis.ScanStruct(result, teamInfo); err != nil {
					kc.logger.Error("错误：ScanStruct", zap.Error(err))
					errorMsg = fmt.Sprintf("Team is not exists[teamID=%s]", teamID)
					goto COMPLETE
				}
			}

			//一天最多只能邀请50人入群，在服务端控制
			nTime := time.Now()
			yesTime := nTime.AddDate(0, 0, -1).Unix()

			if count, err = redis.Int(redisConn.Do("ZCOUNT", fmt.Sprintf("TeamUsers:%s", teamID), yesTime, "+inf")); err != nil {
				kc.logger.Error("ZCOUNT Error", zap.Error(err))
				errorMsg = fmt.Sprintf("TeamUsers is not exists[teamID=%s]", teamID)
				goto COMPLETE
			}

			if count > common.OnedayInvitedLimit {
				err = errors.Wrapf(err, "Reach one day invite limit[count=%d]", count)
				errorMsg = fmt.Sprintf("Reach one day invite limit[count=%d]", count)
				goto COMPLETE
			}
			for _, inviteUsername := range req.GetUsernames() {
				//首先判断一下是否已经是群成员了
				if reply, err := redisConn.Do("ZRANK", fmt.Sprintf("TeamUsers:%s", teamID), inviteUsername); err == nil {
					if reply != nil {
						//已经是群成员
					} else {
						//是否被封禁
						var state int
						if state, err = redis.Int(redisConn.Do("HGET", fmt.Sprintf("userData:%s", inviteUsername))); err != nil {
							kc.logger.Error("redisConn HGET Error", zap.Error(err))
							continue
						}
						if state == common.UserBlocked {
							kc.logger.Debug("User is blocked", zap.String("Username", inviteUsername))
							continue
						}

						var newSeq uint64

						if newSeq, err = redis.Uint64(redisConn.Do("INCR", fmt.Sprintf("userSeq:%s", inviteUsername))); err != nil {
							kc.logger.Error("redisConn INCR userSeq Error", zap.Error(err))
							continue
						}
						//向inviteUsername发出入群请求, 如果此人已经关注了网点，则不需要发出邀请
						inviteEventRsp := &Msg.RecvMsgEventRsp{
							Scene:        Msg.MessageScene_MsgScene_S2C,        //系统消息
							Type:         Msg.MessageType_MsgType_Notification, //通知类型
							Body:         []byte(""),                           //JSON
							From:         username,                             //邀请人
							FromNick:     userNick,
							FromDeviceId: deviceID,
							ServerMsgId:  msg.GetID(),                        //服务器分配的消息ID
							Seq:          newSeq,                             //消息序号，单个会话内自然递增, 这里是对inviteUsername这个用户的通知序号
							Uuid:         fmt.Sprintf("%d", msg.GetTaskID()), //客户端分配的消息ID，SDK生成的消息id，这里返回TaskID
							Time:         uint64(time.Now().Unix()),
						}
						data, _ = proto.Marshal(inviteEventRsp)
						go kc.BroadcastMsgToAllDevices(data, inviteUsername)
					}

				}

			}

			// 	rsp.TeamInfo = &Team.TeamInfo{
			// 		TeamId:       teamInfo.TeamID,
			// 		Name:         teamInfo.Teamname,
			// 		Icon:         teamInfo.Icon,
			// 		Announcement: teamInfo.Announcement,
			// 		Introduce:    teamInfo.Introductory,
			// 		Owner:        teamInfo.Owner,
			// 		Type:         Team.TeamType(teamInfo.Type),
			// 		VerifyType:   Team.VerifyType(teamInfo.VerifyType),
			// 		MemberLimit:  int32(common.PerTeamMembersLimit),
			// 		MemberNum:    int32(count),
			// 		Status:       Team.Status(teamInfo.Status),
			// 		MuteType:     Team.MuteMode(teamInfo.MuteType),
			// 		InviteMode:   Team.InviteMode(teamInfo.InviteMode),
			// 		Ex:           teamInfo.Extend,
			// 		CreateAt:     uint64(teamInfo.CreatedAt),
			// 		UpdateAt:     uint64(teamInfo.UpdatedAt),
			// 	}
			// }

			//回包
			data, _ = proto.Marshal(rsp)
		}
	}

COMPLETE:
	if err != nil {
		msg.SetCode(400)                  //状态码
		msg.SetErrorMsg([]byte(errorMsg)) //错误提示
		msg.FillBody(nil)

	} else {
		msg.SetCode(200) //状态码
		data, _ = proto.Marshal(rsp)
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
