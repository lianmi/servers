package kafkaBackend

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/gomodule/redigo/redis"
	Friends "github.com/lianmi/servers/api/proto/friends"
	Msg "github.com/lianmi/servers/api/proto/msg"
	Team "github.com/lianmi/servers/api/proto/team"
	User "github.com/lianmi/servers/api/proto/user"
	"github.com/lianmi/servers/internal/common"
	"github.com/lianmi/servers/internal/pkg/models"
	uuid "github.com/satori/go.uuid"
	"go.uber.org/zap"
)

/*
4-1 创建群组

1. 此版本只支持创建群，不支持创建讨论组。
2. 网点APP的UI有“申请建群”入口按钮， 普通用户没有此按钮。
3. 网点一旦注册并在管理后台被审核通过后就拥有一个群， 另外可继续创建新群，普通用户无法创建群。
4. 群创建时自动设置网点用户账号为群主，群主可以增设管理员。
5. 群组不开放名称/id搜索
6. 凡是绑定了网点的新注册用户自动加入群，如果群成员数量已满，则等待群主创建新群后自动加入。
7. 自由退群
8. 群组成员上限600
9. 支持自由加入及由用户注册绑定网点后自动加入。
10. 群组创建后，不会马上生效，需要运营后台审核并开通群组，使用方法: GET /approveteam/:teamid

权限说明：
1. 用户被封禁后，不能创建群
2. 用户达到建群上限后，不能再创建新群

*/
func (kc *KafkaClient) HandleCreateTeam(msg *models.Message) error {
	var err error
	errorCode := 200
	var errorMsg string

	//响应参数
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
		errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
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
			errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
			errorMsg = fmt.Sprintf("Query user error[teamOwner=%s]", teamOwner)
			goto COMPLETE

		} else {
			if !isExists {
				errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
				errorMsg = fmt.Sprintf("Owner is not exists[teamOwner=%s]", teamOwner)
				goto COMPLETE
			}

			//判断群主是否已经注册为网点用户类型
			userType, _ := redis.Int(redisConn.Do("HGET", fmt.Sprintf("userData:%s", teamOwner), "UserType"))
			if User.UserType(userType) != User.UserType_Ut_Business {
				errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
				errorMsg = fmt.Sprintf("serType is not business type [userType=%d]", userType)
				goto COMPLETE
			}

			//用户拥有的群的总数量是否已经达到上限
			if count, err := redis.Int(redisConn.Do("ZCARD", fmt.Sprintf("Team:%s", teamOwner))); err != nil {
				kc.logger.Error("ZCARD Error", zap.Error(err))
				errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
				errorMsg = fmt.Sprintf("Can not query team count ")
				goto COMPLETE
			} else {
				if count >= common.MaxTeamLimit {
					errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
					errorMsg = fmt.Sprintf("Reach team max limit[count=%d]", count)
					goto COMPLETE
				}

			}
			//写入MySQL数据库
			var newTeamIndex uint64
			if newTeamIndex, err = redis.Uint64(redisConn.Do("INCR", "TeamIndex")); err != nil {
				kc.logger.Error("redisConn GET TeamIndex Error", zap.Error(err))
				errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
				errorMsg = fmt.Sprintf("serType is not business type [userType=%d]", userType)
				goto COMPLETE
			}

			pTeam := new(models.Team)
			pTeam.CreatedAt = time.Now().UnixNano() / 1e6
			pTeam.TeamID = fmt.Sprintf("team%d", newTeamIndex) //群id， 自动生成
			pTeam.Teamname = req.GetName()
			pTeam.Nick = req.GetName()
			pTeam.Owner = req.GetOwner()
			pTeam.Type = int(req.GetType())
			pTeam.VerifyType = int(req.GetVerifyType())
			pTeam.InviteMode = int(req.GetInviteMode())

			//默认的设置
			pTeam.Status = 1 //Init(1) - 初始状态,审核中 Normal(2) - 正常状态 Blocked(3) - 封禁状态
			pTeam.MemberLimit = common.PerTeamMembersLimit
			pTeam.MemberNum = 1  //刚刚建群是只有群主1人
			pTeam.MuteType = 1   //None(1) - 所有人可发言
			pTeam.InviteMode = 1 //邀请模式,初始为1

			if err = kc.SaveTeam(pTeam); err != nil {
				kc.logger.Error("Save CreateTeam Error", zap.Error(err))
				errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
				errorMsg = "无法保存到数据库"
				goto COMPLETE
			}

			//群信息
			rsp.TeamInfo = &Team.TeamInfo{
				TeamId:       pTeam.TeamID,
				Name:         pTeam.Teamname,
				Icon:         "", //TODO 需要改为默认
				Announcement: "", //群公告
				Introduce:    "", //群简介
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
		kc.logger.Info("Succeed succeed send message to ProduceChannel", zap.String("topic", topic))
	} else {
		kc.logger.Error("Failed to send  message to ProduceChannel", zap.Error(err))
	}
	_ = err
	return nil

}

/*
4-2 获取群组成员
该接口用于增量获取普通群和Vip群群成员信息
权限说明:
1. 根据timeAt增量返回群成员,首次timeAt请初始化为0，服务器返回全量群成员信息，后续采取增量方式更新
2. 如果removedUsers不为空，终端根据removedUsers移除本机群成员缓存数据
3. 终端开发获取群成员接口的流程是: 发起获取成员请求 → 更新本地数据库 → 返回数据给UI


*/
func (kc *KafkaClient) HandleGetTeamMembers(msg *models.Message) error {
	var err error
	errorCode := 200
	var errorMsg string

	//响应参数
	rsp := &Team.GetTeamMembersRsp{
		TimeAt:       uint64(time.Now().UnixNano() / 1e6),
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
		errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
		errorMsg = fmt.Sprintf("Protobuf Unmarshal Error: %s", err.Error())
		goto COMPLETE

	} else {
		kc.logger.Debug("GetTeamMembers body",
			zap.String("teamId", req.GetTeamId()),
			zap.Int("timeAt", int(req.GetTimeAt())),
		)

		teamID := req.GetTeamId()

		if isExists, err := redis.Bool(redisConn.Do("EXISTS", fmt.Sprintf("TeamInfo:%s", teamID))); err != nil {
			errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
			errorMsg = fmt.Sprintf("Query team info error[teamID=%s]", teamID)
			goto COMPLETE

		} else {
			if !isExists {
				errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
				errorMsg = fmt.Sprintf("Team is not exists[teamID=%s]", teamID)
				goto COMPLETE
			}

			//redis查出此群的成员, 从TimeAt开始到最大。
			teamMembers, _ := redis.Strings(redisConn.Do("ZRANGEBYSCORE", fmt.Sprintf("TeamUsers:%s", teamID), req.GetTimeAt(), "+inf"))
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
					TeamId:          teamID,
					Username:        teamUser.Username,
					Invitedusername: teamUser.InvitedUsername,
					Nick:            teamUser.Nick,
					Avatar:          teamUser.Avatar,
					Label:           teamUser.Label,
					Source:          teamUser.Source,
					Type:            Team.TeamMemberType(teamUser.TeamMemberType),
					NotifyType:      Team.NotifyType(teamUser.NotifyType),
					Mute:            teamUser.IsMute,
					Ex:              teamUser.Extend,
					JoinTime:        uint64(teamUser.JoinAt),
					UpdateTime:      uint64(teamUser.UpdatedAt),
				})
			}

			//群成员退群用户列表
			teamRemoveMembers, _ := redis.Strings(redisConn.Do("ZRANGEBYSCORE", fmt.Sprintf("RemoveTeamMembers:%s", teamID), req.GetTimeAt(), "+inf"))
			for _, teamRemoveMember := range teamRemoveMembers {
				rsp.RemovedUsers = append(rsp.RemovedUsers, teamRemoveMember)
			}

			//回包
			data, _ = proto.Marshal(rsp)
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
		kc.logger.Info("Succeed succeed send message to ProduceChannel", zap.String("topic", topic))
	} else {
		kc.logger.Error("Failed to send  message to ProduceChannel", zap.Error(err))
	}
	_ = err
	return nil

}

/*
4-3 查询群信息
该接口用于根据群id查询群的信息
*/
func (kc *KafkaClient) HandleGetTeam(msg *models.Message) error {
	var err error
	errorCode := 200
	var errorMsg string
	var count int

	//响应参数
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
		errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
		errorMsg = fmt.Sprintf("Protobuf Unmarshal Error: %s", err.Error())
		goto COMPLETE

	} else {
		kc.logger.Debug("GetTeam body",
			zap.String("teamId", req.GetTeamId()),
		)

		teamID := req.GetTeamId()

		if isExists, err := redis.Bool(redisConn.Do("EXISTS", fmt.Sprintf("TeamInfo:%s", teamID))); err != nil {
			kc.logger.Error("EXISTS TeamInfo Error", zap.Error(err))
			errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
			errorMsg = fmt.Sprintf("Query team info error[teamID=%s]", teamID)
			goto COMPLETE

		} else {
			if !isExists {
				errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
				errorMsg = fmt.Sprintf("Team is not exists[teamID=%s]", teamID)
				goto COMPLETE
			}

			//计算群成员数量。
			if count, err = redis.Int(redisConn.Do("ZCARD", fmt.Sprintf("TeamUsers:%s", teamID))); err != nil {
				kc.logger.Error("ZCARD Error", zap.Error(err))
				errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
				errorMsg = fmt.Sprintf("TeamUsers is not exists[teamID=%s]", teamID)
				goto COMPLETE
			}

			//查出此群信息
			teamInfo := new(models.Team)
			if result, err := redis.Values(redisConn.Do("HGETALL", fmt.Sprintf("TeamInfo:%s", teamID))); err == nil {
				if err := redis.ScanStruct(result, teamInfo); err != nil {
					kc.logger.Error("错误：ScanStruct", zap.Error(err))
					errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
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
		kc.logger.Info("Succeed succeed send message to ProduceChannel", zap.String("topic", topic))
	} else {
		kc.logger.Error("Failed to send  message to ProduceChannel", zap.Error(err))
	}
	_ = err
	return nil

}

/*
4-4 邀请用户加群
工作流ID说明:
1. 生成工作流ID，下发给SDK
2. SDK响应的时候，需要携带此工作流ID

其它说明:
1. 普通群: 用户注册时输入推荐码（网点用户账号的数字部分）或 用户关注网点，就会自动加群,
2. Vip群: 群成员是否可以拉取用户入群由管理员设置，邀请用户需要用户同意， 可以不是好友也可以邀请入群，类似微信的弱管理。
3. 一天最多只能邀请50人入群，在服务端控制

权限要求：
1. 群没有被封禁
2. 拉人入群设定
3. 不是群成员
*/
func (kc *KafkaClient) HandleInviteTeamMembers(msg *models.Message) error {
	var err error
	errorCode := 200
	var errorMsg string
	var newSeq uint64
	var count int

	//响应参数
	rsp := &Team.InviteTeamMembersRsp{
		AbortedUsers: make([]string, 0),
	}
	var data []byte

	redisConn := kc.redisPool.Get()
	defer redisConn.Close()

	username := msg.GetUserName() //用户自己的账号
	// token := msg.GetJwtToken()
	deviceID := msg.GetDeviceID()

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
		errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
		errorMsg = fmt.Sprintf("Protobuf Unmarshal Error: %s", err.Error())
		goto COMPLETE

	} else {
		kc.logger.Debug("InviteTeamMembers payload",
			zap.String("teamId", req.GetTeamId()),
			zap.String("ps", req.GetPs()),
			zap.Strings("usernames", req.GetUsernames()),
		)

		teamID := req.GetTeamId()

		//判断 teamID 是否存在
		if isExists, err := redis.Bool(redisConn.Do("EXISTS", fmt.Sprintf("TeamInfo:%s", teamID))); err != nil {
			kc.logger.Error("EXISTS Error", zap.Error(err))
			errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
			errorMsg = fmt.Sprintf("Query team info error[teamID=%s]", teamID)
			goto COMPLETE

		} else {
			if !isExists {
				errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
				errorMsg = fmt.Sprintf("Team is not exists[teamID=%s]", teamID)
				goto COMPLETE
			}

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

			//一天最多只能邀请50人入群，在服务端控制
			nTime := time.Now()
			yesTime := nTime.AddDate(0, 0, -1).Unix()

			if count, err = redis.Int(redisConn.Do("ZCOUNT", fmt.Sprintf("TeamUsers:%s", teamID), yesTime, "+inf")); err != nil {
				kc.logger.Error("ZCOUNT Error", zap.Error(err))
				errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
				errorMsg = fmt.Sprintf("TeamUsers is not exists[teamID=%s]", teamID)
				goto COMPLETE
			}

			if count > common.OnedayInvitedLimit {
				errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
				errorMsg = fmt.Sprintf("Reach one day invite limit[count=%d]", count)
				goto COMPLETE
			}

			//此群的拉人进群的模式设定
			switch Team.InviteMode(teamInfo.InviteMode) {
			case Team.InviteMode_Invite_All: //所有人都可以邀请其他人入群
				//处理待入群用户列表
				abortUsers := kc.processInviteMembers(redisConn, teamID, username, deviceID, req.GetPs(), req.GetUsernames())
				for _, abortUser := range abortUsers {
					rsp.AbortedUsers = append(rsp.AbortedUsers, abortUser)
				}

			case Team.InviteMode_Invite_Manager: //只有管理员可以邀请其他人入群
				//判断当前用户的类型是否是管理员
				//判断操作者是不是群主或管理员
				opUser := new(models.TeamUser)
				if result, err := redis.Values(redisConn.Do("HGETALL", fmt.Sprintf("TeamUser:%s:%s", teamID, username))); err == nil {
					if err := redis.ScanStruct(result, opUser); err != nil {
						kc.logger.Error("TeamUser is not exist", zap.Error(err))
						errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
						errorMsg = fmt.Sprintf("TeamUser is not exists[teamID=%s, teamUser=%s]", teamID, username)
						goto COMPLETE
					}
				}
				teamMemberType := Team.TeamMemberType(opUser.TeamMemberType)
				if teamMemberType == Team.TeamMemberType_Tmt_Owner || teamMemberType == Team.TeamMemberType_Tmt_Manager {
					//pass
				} else {
					kc.logger.Warn("User is not team owner or manager", zap.String("Username", username))
					errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
					errorMsg = fmt.Sprintf("User is not team owner or manager[Username=%s]", username)
					goto COMPLETE
				}

				//处理待入群用户列表
				abortUsers := kc.processInviteMembers(redisConn, teamID, username, deviceID, req.GetPs(), req.GetUsernames())
				for _, abortUser := range abortUsers {
					rsp.AbortedUsers = append(rsp.AbortedUsers, abortUser)
				}

			case Team.InviteMode_Invite_Check: //邀请用户入群时需要管理员审核，需要向所有群管理员发送系统通知，管理员利用4-26 回复
				//向群主或管理员推送此用户的主动加群通知
				managers, _ := kc.GetOwnerAndManagers(teamInfo.TeamID)
				for _, manager := range managers {
					//遍历整个被邀请加群用户列表, 注意：每个用户都必须有独立的工作流ID
					for _, inviteUsername := range req.GetUsernames() {
						if reply, err := redisConn.Do("ZRANK", fmt.Sprintf("TeamUsers:%s", teamID), inviteUsername); err == nil {
							if reply != nil {
								//已经是群成员
								rsp.AbortedUsers = append(rsp.AbortedUsers, inviteUsername)
							} else {
								//是否被封禁
								if state, err := redis.Int(redisConn.Do("HGET", fmt.Sprintf("userData:%s", inviteUsername), "State")); err != nil {
									kc.logger.Error("redisConn HGET Error", zap.Error(err))
									continue
								} else {
									if state == common.UserBlocked {
										kc.logger.Debug("User is blocked", zap.String("inviteUsername", inviteUsername))
										rsp.AbortedUsers = append(rsp.AbortedUsers, inviteUsername)
										continue
									}
								}
							}
						}
						workflowID := uuid.NewV4().String()

						//将被邀请方存入InviteTeamMembers:{teamID}里，以便被邀请方同意或拒绝的时候校验，其它人没被邀请，则直接退出
						if _, err = redisConn.Do("ZADD", fmt.Sprintf("InviteTeamMembers:%s", teamID), time.Now().UnixNano()/1e6, inviteUsername); err != nil {
							kc.logger.Error("ZADD Error", zap.Error(err))
						}

						//将此工作流ID作为key保存此加群事件的哈希表, InviteWorkflow:{member}:{workflowID}
						workflowKey := fmt.Sprintf("InviteWorkflow:%s:%s", inviteUsername, workflowID)
						_, err = redisConn.Do("HMSET",
							workflowKey,
							"Inviter", username, //邀请人
							"Invitee", inviteUsername, //受邀请人
							"TeamID", teamID, //群ID
							"Ps", req.GetPs(), //附言
						)

						if newSeq, err = redis.Uint64(redisConn.Do("INCR", fmt.Sprintf("userSeq:%s", manager))); err != nil {
							kc.logger.Error("redisConn INCR userSeq Error", zap.Error(err))
							errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
							errorMsg = fmt.Sprintf("INCR error[Owner=%s]", manager)
							goto COMPLETE
						}
						nick, err := redis.String(redisConn.Do("HGET", fmt.Sprintf("userData:%s", username), "Nick"))
						if err != nil {
							kc.logger.Error("获取用户呢称错误", zap.Error(err))
							continue
						}
						inviteeNick, err := redis.String(redisConn.Do("HGET", fmt.Sprintf("userData:%s", inviteUsername), "Nick"))
						if err != nil {
							kc.logger.Error("获取受邀请用户呢称错误", zap.Error(err))
							continue
						}

						handledMsg := fmt.Sprintf("用户: %s 邀请 %s 进群", nick, inviteeNick)
						serverMsgId := uuid.NewV4().String()

						body := Msg.MessageNotificationBody{
							Type:           Msg.MessageNotificationType_MNT_CheckTeamInvite, //向群主推送审核入群通知
							HandledAccount: username,
							HandledMsg:     handledMsg,
							Status:         Msg.MessageStatus_MOS_Processing, //处理中
							Data:           []byte(""),
							To:             teamInfo.TeamID, //群id
						}
						bodyData, _ := proto.Marshal(&body)
						inviteEventRsp := &Msg.RecvMsgEventRsp{
							Scene:        Msg.MessageScene_MsgScene_S2C,        //系统消息
							Type:         Msg.MessageType_MsgType_Notification, //通知类型
							Body:         bodyData,
							From:         username, //发起人
							FromDeviceId: deviceID,
							ServerMsgId:  serverMsgId, //服务器分配的消息ID
							WorkflowID:   workflowID,
							Seq:          newSeq,                             //消息序号，单个会话内自然递增, 这里是对inviteUsername这个用户的通知序号
							Uuid:         fmt.Sprintf("%d", msg.GetTaskID()), //客户端分配的消息ID，SDK生成的消息id，这里返回TaskID
							Time:         uint64(time.Now().UnixNano() / 1e6),
						}
						go kc.BroadcastMsgToAllDevices(inviteEventRsp, manager) //群主或管理员
					}

				}
				goto COMPLETE
			}

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
		kc.logger.Info("Succeed succeed send message to ProduceChannel", zap.String("topic", topic))
	} else {
		kc.logger.Error("Failed to send  message to ProduceChannel", zap.Error(err))
	}
	_ = err
	return nil

}

func (kc *KafkaClient) processInviteMembers(redisConn redis.Conn, teamID, inviter, fromDeviceId, ps string, inviteUsername []string) []string {
	var newSeq uint64
	abortedUsers := make([]string, 0)

	//遍历整个被邀请加群用户列表, 注意：每个用户都必须有独立的工作流ID
	for _, inviteUsername := range inviteUsername {
		//首先判断一下是否已经是群成员了
		if reply, err := redisConn.Do("ZRANK", fmt.Sprintf("TeamUsers:%s", teamID), inviteUsername); err == nil {
			if reply != nil {
				//已经是群成员
				abortedUsers = append(abortedUsers, inviteUsername)
			} else {
				//是否被封禁
				if state, err := redis.Int(redisConn.Do("HGET", fmt.Sprintf("userData:%s", inviteUsername), "State")); err != nil {
					kc.logger.Error("redisConn HGET Error", zap.Error(err))
					continue
				} else {
					if state == common.UserBlocked {
						kc.logger.Debug("User is blocked", zap.String("inviteUsername", inviteUsername))
						abortedUsers = append(abortedUsers, inviteUsername)
						continue
					}
				}

				//TODO 生成工作流ID
				workflowID := uuid.NewV4().String()
				serverMsgId := uuid.NewV4().String()

				if newSeq, err = redis.Uint64(redisConn.Do("INCR", fmt.Sprintf("userSeq:%s", inviteUsername))); err != nil {
					kc.logger.Error("redisConn INCR userSeq Error", zap.Error(err))
					continue
				}

				//将被邀请方存入InviteTeamMembers:{teamID}里，以便被邀请方同意或拒绝的时候校验，其它人没被邀请，则直接退出
				if _, err = redisConn.Do("ZADD", fmt.Sprintf("InviteTeamMembers:%s", teamID), time.Now().UnixNano()/1e6, inviteUsername); err != nil {
					kc.logger.Error("ZADD Error", zap.Error(err))
				}

				//将此工作流ID作为key保存此加群事件的哈希表, InviteWorkflow:{member}:{workflowID}
				workflowKey := fmt.Sprintf("InviteWorkflow:%s:%s", inviteUsername, workflowID)
				_, err = redisConn.Do("HMSET",
					workflowKey,
					"Inviter", inviter, //邀请人
					"Invitee", inviteUsername, //受邀请人
					"TeamID", teamID, //群ID
					"Ps", ps, //附言
				)
				nick, err := redis.String(redisConn.Do("HGET", fmt.Sprintf("userData:%s", inviter), "Nick"))
				if err != nil {
					kc.logger.Error("获取用户呢称错误", zap.Error(err))
					continue
				}
				inviteeNick, err := redis.String(redisConn.Do("HGET", fmt.Sprintf("userData:%s", inviteUsername), "Nick"))
				if err != nil {
					kc.logger.Error("获取受邀请用户呢称错误", zap.Error(err))
					continue
				}

				handledMsg := fmt.Sprintf("用户: %s 邀请 %s 进群", nick, inviteeNick)

				body := Msg.MessageNotificationBody{
					Type:           Msg.MessageNotificationType_MNT_TeamInvite, //邀请加群
					HandledAccount: inviter,
					HandledMsg:     handledMsg,
					Status:         Msg.MessageStatus_MOS_Init, //未处理
					Data:           []byte(""),
					To:             teamID, //群id
				}
				bodyData, _ := proto.Marshal(&body)

				inviteEventRsp := &Msg.RecvMsgEventRsp{
					Scene:        Msg.MessageScene_MsgScene_S2C,        //系统消息
					Type:         Msg.MessageType_MsgType_Notification, //通知类型
					Body:         bodyData,                             //字节流
					From:         inviter,                              //邀请人
					FromDeviceId: fromDeviceId,
					ServerMsgId:  serverMsgId, //服务器分配的消息ID
					WorkflowID:   workflowID,  //工作流ID
					Seq:          newSeq,      //消息序号，单个会话内自然递增, 这里是对inviteUsername这个用户的通知序号
					Uuid:         "",
					Time:         uint64(time.Now().UnixNano() / 1e6),
				}

				//向被邀请加群的用户推送系统通知
				go kc.BroadcastMsgToAllDevices(inviteEventRsp, inviteUsername)
			}
		}
	}
	return abortedUsers
}

/*
4-5 删除群组成员
管理员移除群成员
*/

func (kc *KafkaClient) HandleRemoveTeamMembers(msg *models.Message) error {
	var err error
	errorCode := 200
	var errorMsg string
	var newSeq uint64

	//响应参数
	rsp := &Team.RemoveTeamMembersRsp{
		AbortedUsers: make([]string, 0),
	}
	var data []byte

	redisConn := kc.redisPool.Get()
	defer redisConn.Close()

	username := msg.GetUserName() //用户自己的账号
	// token := msg.GetJwtToken()
	deviceID := msg.GetDeviceID()

	kc.logger.Info("HandleRemoveTeamMembers start...",
		zap.String("username", username),
		zap.String("deviceId", deviceID))

	//取出当前设备的os， clientType， logonAt
	curDeviceHashKey := fmt.Sprintf("devices:%s:%s", username, deviceID)
	isMaster, _ := redis.Bool(redisConn.Do("HGET", curDeviceHashKey, "ismaster"))
	curOs, _ := redis.String(redisConn.Do("HGET", curDeviceHashKey, "os"))
	curClientType, _ := redis.Int(redisConn.Do("HGET", curDeviceHashKey, "clientType"))
	curLogonAt, _ := redis.Uint64(redisConn.Do("HGET", curDeviceHashKey, "logonAt"))

	kc.logger.Debug("RemoveTeamMembers",
		zap.Bool("isMaster", isMaster),
		zap.String("username", username),
		zap.String("deviceID", deviceID),
		zap.String("curOs", curOs),
		zap.Int("curClientType", curClientType),
		zap.Uint64("curLogonAt", curLogonAt))

	//打开msg里的负载， 获取请求参数
	body := msg.GetContent()

	//解包body
	req := &Team.RemoveTeamMembersReq{}
	if err := proto.Unmarshal(body, req); err != nil {
		kc.logger.Error("Protobuf Unmarshal Error", zap.Error(err))
		errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
		errorMsg = fmt.Sprintf("Protobuf Unmarshal Error: %s", err.Error())
		goto COMPLETE

	} else {
		kc.logger.Debug("RemoveTeamMembersReq payload",
			zap.String("teamId", req.GetTeamId()),
			zap.Strings("usernames", req.GetUsernames()),
		)

		teamID := req.GetTeamId()

		//判断 teamID 是否存在
		if isExists, err := redis.Bool(redisConn.Do("EXISTS", fmt.Sprintf("TeamInfo:%s", teamID))); err != nil {
			kc.logger.Error("EXISTS Error", zap.Error(err))
			errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
			errorMsg = fmt.Sprintf("Query team info error[teamID=%s]", teamID)
			goto COMPLETE

		} else {
			if !isExists {
				errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
				errorMsg = fmt.Sprintf("Team is not exists[teamID=%s]", teamID)
				goto COMPLETE
			}

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

			//判断usename是不是管理员身份或群主，如果不是，则返回
			teamUser := new(models.TeamUser)
			if result, err := redis.Values(redisConn.Do("HGETALL", fmt.Sprintf("TeamUser:%s:%s", teamID, username))); err == nil {
				if err := redis.ScanStruct(result, teamUser); err != nil {
					errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
					errorMsg = fmt.Sprintf("TeamUser is not exists[teamID=%s, teamUser=%s]", teamID, username)
					goto COMPLETE
				}
			}
			teamMemberType := Team.TeamMemberType(teamUser.TeamMemberType)

			if teamMemberType == Team.TeamMemberType_Tmt_Owner || teamMemberType == Team.TeamMemberType_Tmt_Manager {
				//管理员或群主
			} else {
				kc.logger.Error("无权删除群成员", zap.Error(err))
				errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
				errorMsg = fmt.Sprintf("User is not owner or manager[username=%s]", username)
				goto COMPLETE
			}

			for _, removedUsername := range req.GetUsernames() {
				//首先判断一下是否是群成员
				if reply, err := redisConn.Do("ZRANK", fmt.Sprintf("TeamUsers:%s", teamID), removedUsername); err == nil {
					if reply != nil { //是群成员
						//判断是否有权移除， 例如，管理员不能在这里移除， 群主不能被移除

						removeUser := new(models.TeamUser)
						if result, err := redis.Values(redisConn.Do("HGETALL", fmt.Sprintf("TeamUser:%s:%s", teamID, removedUsername))); err == nil {
							if err := redis.ScanStruct(result, removeUser); err != nil {
								errorMsg = fmt.Sprintf("TeamUser is not exists[teamID=%s, teamUser=%s]", teamID, removedUsername)
								kc.logger.Error("TeamUser is not exist", zap.Error(err))

								//增加到无法移除列表
								rsp.AbortedUsers = append(rsp.AbortedUsers, removedUsername)
								continue
							}
						}
						teamMemberType := Team.TeamMemberType(removeUser.TeamMemberType)

						if teamMemberType == Team.TeamMemberType_Tmt_Owner || teamMemberType == Team.TeamMemberType_Tmt_Manager {
							//管理员或群主
							kc.logger.Error("无权移除管理员或群主", zap.Error(err))

							//增加到无法移除列表
							rsp.AbortedUsers = append(rsp.AbortedUsers, removedUsername)

							continue
						} else {
							//删除此用户在群里的数据
							if err := kc.DeleteTeamUser(teamID, removedUsername); err != nil {
								kc.logger.Error("移除群成员失败", zap.Error(err))

								//增加到无法移除列表
								rsp.AbortedUsers = append(rsp.AbortedUsers, removedUsername)
								continue
							}

							//删除redis里的TeamUser哈希表
							err = redisConn.Send("DEL", fmt.Sprintf("TeamUser:%s:%s", teamInfo.TeamID, removedUsername))
							//删除群成员的有序集合
							err = redisConn.Send("ZREM", fmt.Sprintf("TeamUsers:%s", teamID), removedUsername)
							//将群成员自己加入的群里删除teamID
							err = redisConn.Send("ZREM", fmt.Sprintf("Team:%s", removedUsername), teamID)
							//增加到此用户自己的退群列表
							err = redisConn.Send("ZADD", fmt.Sprintf("RemoveTeam:%s", removedUsername), time.Now().UnixNano()/1e6, teamID)

							//更新redis的sync:{用户账号} teamsAt 时间戳
							redisConn.Send("HSET",
								fmt.Sprintf("sync:%s", removedUsername),
								"teamsAt",
								time.Now().UnixNano()/1e6)

							//一次性写入到Redis
							redisConn.Flush()

							teamMembers, _ := redis.Strings(redisConn.Do("ZRANGEBYSCORE", fmt.Sprintf("TeamUsers:%s", teamID), "-inf", "+inf"))
							//向其它群成员推送MNT_KickOffTeam
							for _, teamMember := range teamMembers {

								//更新redis的sync:{用户账号} teamsAt 时间戳
								redisConn.Do("HSET",
									fmt.Sprintf("sync:%s", teamMember),
									"teamsAt",
									time.Now().UnixNano()/1e6)

								if newSeq, err = redis.Uint64(redisConn.Do("INCR", fmt.Sprintf("userSeq:%s", teamMember))); err != nil {
									kc.logger.Error("redisConn INCR userSeq Error", zap.Error(err))
									continue
								}

								//向所有群成员发出移除removedUsername出群通知
								psSource := &Friends.PsSource{
									Ps:     "",
									Source: removedUsername, //被移除出群的用户
								}
								psSourceData, _ := proto.Marshal(psSource)

								handledMsg := fmt.Sprintf("用户 %s被管理员移除出群", removeUser.Nick)
								body := Msg.MessageNotificationBody{
									Type:           Msg.MessageNotificationType_MNT_KickOffTeam, //被管理员踢出群
									HandledAccount: username,
									HandledMsg:     handledMsg,
									Status:         Msg.MessageStatus_MOS_Done,
									Data:           psSourceData, //包含信息
									To:             teamID,       //群id
								}
								bodyData, _ := proto.Marshal(&body)
								mrsp := &Msg.RecvMsgEventRsp{
									Scene:        Msg.MessageScene_MsgScene_S2C,        //系统消息
									Type:         Msg.MessageType_MsgType_Notification, //通知类型
									Body:         bodyData,
									From:         username,
									FromDeviceId: deviceID,
									ServerMsgId:  msg.GetID(),                        //服务器分配的消息ID
									Seq:          newSeq,                             //消息序号，单个会话内自然递增
									Uuid:         fmt.Sprintf("%d", msg.GetTaskID()), //客户端分配的消息ID，SDK生成的消息id，这里返回TaskID
									Time:         uint64(time.Now().UnixNano() / 1e6),
								}
								go kc.BroadcastMsgToAllDevices(mrsp, teamMember)

							}

						}

					} else {
						//增加到无法移除列表
						rsp.AbortedUsers = append(rsp.AbortedUsers, removedUsername)
					}
				}
			}

			//回包
			data, _ = proto.Marshal(rsp)
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
		kc.logger.Info("Succeed succeed send message to ProduceChannel", zap.String("topic", topic))
	} else {
		kc.logger.Error("Failed to send  message to ProduceChannel", zap.Error(err))
	}
	_ = err
	return nil

}

/*
4-6 接受群邀请
说明：
1. 被拉的人系统通知有显示入群的通知,点接收,注意拒绝后,再接受会出现群成员状态不对,通知只能操作一次
2. 向所有群成员推送用户入群通知

权限:

*/

func (kc *KafkaClient) HandleAcceptTeamInvite(msg *models.Message) error {
	var err error
	errorCode := 200
	var errorMsg string
	var newSeq uint64
	var count int

	//响应参数
	rsp := &Team.AcceptTeamInviteRsp{}
	var data []byte

	redisConn := kc.redisPool.Get()
	defer redisConn.Close()

	username := msg.GetUserName() //用户自己的账号
	// token := msg.GetJwtToken()
	deviceID := msg.GetDeviceID()

	kc.logger.Info("HandleAcceptTeamInvite start...",
		zap.String("username", username),
		zap.String("deviceId", deviceID))

	//取出当前设备的os， clientType， logonAt
	curDeviceHashKey := fmt.Sprintf("devices:%s:%s", username, deviceID)
	isMaster, _ := redis.Bool(redisConn.Do("HGET", curDeviceHashKey, "ismaster"))
	curOs, _ := redis.String(redisConn.Do("HGET", curDeviceHashKey, "os"))
	curClientType, _ := redis.Int(redisConn.Do("HGET", curDeviceHashKey, "clientType"))
	curLogonAt, _ := redis.Uint64(redisConn.Do("HGET", curDeviceHashKey, "logonAt"))

	kc.logger.Debug("AcceptTeamInvite",
		zap.Bool("isMaster", isMaster),
		zap.String("username", username),
		zap.String("deviceID", deviceID),
		zap.String("curOs", curOs),
		zap.Int("curClientType", curClientType),
		zap.Uint64("curLogonAt", curLogonAt))

	//打开msg里的负载， 获取请求参数
	body := msg.GetContent()

	//解包body
	req := &Team.AcceptTeamInviteReq{}
	if err := proto.Unmarshal(body, req); err != nil {
		kc.logger.Error("Protobuf Unmarshal Error", zap.Error(err))
		errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
		errorMsg = fmt.Sprintf("Protobuf Unmarshal Error: %s", err.Error())
		goto COMPLETE

	} else {
		kc.logger.Debug("AcceptTeamInvite payload",
			zap.String("teamId", req.GetTeamId()),
			zap.String("from", req.GetFrom()),             //邀请方
			zap.String("workflowID", req.GetWorkflowID()), //工作流ID
		)

		teamID := req.GetTeamId()

		//获取邀请方的呢称
		fromNick, err := redis.String(redisConn.Do("HGET", fmt.Sprintf("userData:%s", req.GetFrom()), "Nick"))
		if err != nil {
			kc.logger.Error("获取邀请方的呢称错误", zap.Error(err))
			errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
			errorMsg = fmt.Sprintf("HGET error[from=%s]", req.GetFrom())
			goto COMPLETE
		}
		//校验用户是否曾经被人拉入群
		if reply, err := redisConn.Do("ZRANK", fmt.Sprintf("InviteTeamMembers:%s", teamID), username); err == nil {
			if reply != nil {
				//曾经被人拉入群, 删除有序集合
				_, err = redisConn.Do("ZREM", fmt.Sprintf("InviteTeamMembers:%s", teamID), username)

			} else {
				kc.logger.Warn("校验用户是否曾经被人拉入群: 否", zap.String("username", username))
				errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
				errorMsg = fmt.Sprintf("Not in invited list")
				goto COMPLETE
			}
		}

		//判断 teamID 是否存在
		if isExists, err := redis.Bool(redisConn.Do("EXISTS", fmt.Sprintf("TeamInfo:%s", teamID))); err != nil {
			kc.logger.Error("EXISTS Error", zap.Error(err))
			errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
			errorMsg = fmt.Sprintf("Query team info error[teamID=%s]", teamID)
			goto COMPLETE

		} else {
			if !isExists {
				errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
				errorMsg = fmt.Sprintf("Team is not exists[teamID=%s]", teamID)
				goto COMPLETE
			}

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

			//判断username是不是被封禁了，如果是则返回
			if state, err := redis.Int(redisConn.Do("HGET", fmt.Sprintf("userData:%s", username), "State")); err != nil {
				kc.logger.Error("redisConn HGET Error", zap.Error(err))
				errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
				errorMsg = fmt.Sprintf("ser is not exists[Username=%s]", username)
				goto COMPLETE
			} else {
				if state == common.UserBlocked {
					kc.logger.Debug("User is blocked", zap.String("Username", username))
					errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
					errorMsg = fmt.Sprintf("ser is blocked[Username=%s]", username)
					goto COMPLETE
				}
			}

			//判断username是不是已经是群成员了，如果是，则返回
			//首先判断一下是否是群成员
			if reply, err := redisConn.Do("ZRANK", fmt.Sprintf("TeamUsers:%s", teamID), username); err == nil {
				if reply != nil { //是群成员
					err = nil
					kc.logger.Debug("User is already member", zap.String("Username", username))
					goto COMPLETE
				}
			}

			userData := new(models.User)
			userKey := fmt.Sprintf("userData:%s", username)
			if result, err := redis.Values(redisConn.Do("HGETALL", userKey)); err == nil {
				if err := redis.ScanStruct(result, userData); err != nil {

					kc.logger.Error("错误：ScanStruct", zap.Error(err))
					errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
					errorMsg = fmt.Sprintf("ScanStruct Error[Username=%s]", username)
					goto COMPLETE

				}
			}

			//存储群成员信息 TeamUser
			teamUser := new(models.TeamUser)
			teamUser.JoinAt = time.Now().UnixNano() / 1e6
			teamUser.Teamname = teamInfo.Teamname
			teamUser.Username = userData.Username
			teamUser.InvitedUsername = req.GetFrom()                      //邀请者
			teamUser.Nick = userData.Nick                                 //群成员呢称
			teamUser.Avatar = userData.Avatar                             //群成员头像
			teamUser.Label = userData.Label                               //群成员标签
			teamUser.Source = ""                                          //群成员来源  TODO
			teamUser.Extend = userData.Extend                             //群成员扩展字段
			teamUser.TeamMemberType = int(Team.TeamMemberType_Tmt_Normal) //群成员类型 3-普通
			teamUser.IsMute = false                                       //是否被禁言
			teamUser.NotifyType = 1                                       //群消息通知方式 All(1) - 群全部消息提醒
			teamUser.Province = userData.Province                         //省份, 如广东省
			teamUser.City = userData.City                                 //城市，如广州市
			teamUser.County = userData.County                             //区，如天河区
			teamUser.Street = userData.Street                             //街道
			teamUser.Address = userData.Address                           //地址

			if err := kc.SaveTeamUser(teamUser); err != nil {
				kc.logger.Error("更新teamUser失败", zap.Error(err))
				errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
				errorMsg = fmt.Sprintf("更新teamUser失败[Username=%s]", username)
				goto COMPLETE

			}

			/*
				1. 用户拥有的群，用有序集合存储，Key: Team:{Owner}, 成员元素是: TeamnID
				2. 群信息哈希表, key格式为: TeamInfo:{TeamnID}, 字段为: Teamname Nick Icon 等Team表的字段
				3. 用户有拥有的群用有序集合存储, key格式为： TeamUsers:{TeamnID}, 成员元素是: Username
				4. 每个群成员用哈希表存储，Key格式为： TeamUser:{TeamnID}:{Username} , 字段为: Teamname Username Nick JoinAt 等TeamUser表的字段
				5. 被移除的成员列表，Key格式为： RemoveTeamMembers:{TeamnID}
			*/
			err = redisConn.Send("ZADD", fmt.Sprintf("Team:%s", username), time.Now().UnixNano()/1e6, teamInfo.TeamID)

			err = redisConn.Send("HMSET", redis.Args{}.Add(fmt.Sprintf("TeamInfo:%s", teamInfo.TeamID)).AddFlat(teamInfo)...)

			//删除退群名单列表里的此用户
			err = redisConn.Send("ZREM", fmt.Sprintf("RemoveTeamMembers:%s", teamInfo.TeamID), time.Now().UnixNano()/1e6, username)
			//add群成员
			err = redisConn.Send("ZADD", fmt.Sprintf("TeamUsers:%s", teamInfo.TeamID), time.Now().UnixNano()/1e6, username)

			err = redisConn.Send("HMSET", redis.Args{}.Add(fmt.Sprintf("TeamUser:%s:%s", teamInfo.TeamID, username)).AddFlat(teamUser)...)

			//更新redis的sync:{用户账号} teamsAt 时间戳
			redisConn.Send("HSET",
				fmt.Sprintf("sync:%s", username),
				"teamsAt",
				time.Now().UnixNano()/1e6)

			//一次性写入到Redis
			redisConn.Flush()

			//向群推送此用户的入群通知
			teamMembers, _ := redis.Strings(redisConn.Do("ZRANGEBYSCORE", fmt.Sprintf("TeamUsers:%s", teamID), "-inf", "+inf"))
			psSource := &Friends.PsSource{
				Ps:     "",
				Source: req.GetFrom(), //发起邀请方
			}
			psSourceData, _ := proto.Marshal(psSource)

			handledMsg := fmt.Sprintf("\"%s\"邀请\"%s\"入群", fromNick, userData.Nick)

			for _, teamMember := range teamMembers {
				if newSeq, err = redis.Uint64(redisConn.Do("INCR", fmt.Sprintf("userSeq:%s", teamMember))); err != nil {
					kc.logger.Error("redisConn INCR userSeq Error", zap.Error(err))
					continue
				}
				body := Msg.MessageNotificationBody{
					Type:           Msg.MessageNotificationType_MNT_PassTeamInvite, //用户同意群邀请
					HandledAccount: username,
					HandledMsg:     handledMsg,
					Status:         Msg.MessageStatus_MOS_Passed,
					Data:           psSourceData,
					To:             teamID, //群组id
				}
				bodyData, _ := proto.Marshal(&body)
				inviteEventRsp := &Msg.RecvMsgEventRsp{
					Scene:        Msg.MessageScene_MsgScene_S2C,        //系统消息
					Type:         Msg.MessageType_MsgType_Notification, //通知类型
					Body:         bodyData,                             //字节流
					From:         username,
					FromDeviceId: deviceID,
					ServerMsgId:  msg.GetID(),                        //服务器分配的消息ID
					WorkflowID:   req.GetWorkflowID(),                //工作流ID
					Seq:          newSeq,                             //消息序号，单个会话内自然递增, 这里是对inviteUsername这个用户的通知序号
					Uuid:         fmt.Sprintf("%d", msg.GetTaskID()), //客户端分配的消息ID，SDK生成的消息id，这里返回TaskID
					Time:         uint64(time.Now().UnixNano() / 1e6),
				}

				go kc.BroadcastMsgToAllDevices(inviteEventRsp, teamMember) //向群成员广播
			}

			//TODO 向自己的其它端广播

			//计算群成员数量。
			if count, err = redis.Int(redisConn.Do("ZCARD", fmt.Sprintf("TeamUsers:%s", teamID))); err != nil {
				kc.logger.Error("ZCARD Error", zap.Error(err))
				errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
				errorMsg = fmt.Sprintf("TeamUsers is not exists[teamID=%s]", teamID)
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
				UpdateAt:     uint64(time.Now().UnixNano() / 1e6), //更新时间
			}

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
		kc.logger.Info("Succeed succeed send message to ProduceChannel", zap.String("topic", topic))
	} else {
		kc.logger.Error("Failed to send  message to ProduceChannel", zap.Error(err))
	}
	_ = err
	return nil

}

/*
4-7 拒绝群邀请
说明：
1. 被拉的人系统通知有显示入群的通知, 点拒绝, 注意不能重复点拒绝

*/
func (kc *KafkaClient) HandleRejectTeamInvitee(msg *models.Message) error {
	var err error
	errorCode := 200
	var errorMsg string
	var newSeq uint64

	redisConn := kc.redisPool.Get()
	defer redisConn.Close()

	username := msg.GetUserName() //用户自己的账号
	// token := msg.GetJwtToken()
	deviceID := msg.GetDeviceID()

	kc.logger.Info("HandleRejectTeamInvitee start...",
		zap.String("username", username),
		zap.String("deviceId", deviceID))

	//取出当前设备的os， clientType， logonAt
	curDeviceHashKey := fmt.Sprintf("devices:%s:%s", username, deviceID)
	isMaster, _ := redis.Bool(redisConn.Do("HGET", curDeviceHashKey, "ismaster"))
	curOs, _ := redis.String(redisConn.Do("HGET", curDeviceHashKey, "os"))
	curClientType, _ := redis.Int(redisConn.Do("HGET", curDeviceHashKey, "clientType"))
	curLogonAt, _ := redis.Uint64(redisConn.Do("HGET", curDeviceHashKey, "logonAt"))

	kc.logger.Debug("RejectTeamInvitee",
		zap.Bool("isMaster", isMaster),
		zap.String("username", username),
		zap.String("deviceID", deviceID),
		zap.String("curOs", curOs),
		zap.Int("curClientType", curClientType),
		zap.Uint64("curLogonAt", curLogonAt))

	//打开msg里的负载， 获取请求参数
	body := msg.GetContent()

	//解包body
	req := &Team.RejectTeamInviteReq{}
	if err := proto.Unmarshal(body, req); err != nil {
		kc.logger.Error("Protobuf Unmarshal Error", zap.Error(err))
		errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
		errorMsg = fmt.Sprintf("Protobuf Unmarshal Error: %s", err.Error())
		goto COMPLETE

	} else {
		kc.logger.Debug("RejectTeamInvitee payload",
			zap.String("teamId", req.GetTeamId()),
			zap.String("from", req.GetFrom()),             //邀请方
			zap.String("workflowID", req.GetWorkflowID()), //工作流ID
			zap.String("ps", req.GetPs()),                 //拒绝的附言
		)

		teamID := req.GetTeamId()

		//校验用户是否曾经被人拉入群
		if reply, err := redisConn.Do("ZRANK", fmt.Sprintf("InviteTeamMembers:%s", teamID), username); err == nil {
			if reply != nil {
				//曾经被人拉入群 , 删除有序集合
				_, err = redisConn.Do("ZREM", fmt.Sprintf("InviteTeamMembers:%s", teamID), username)
			} else {
				kc.logger.Warn("校验用户是否曾经被人拉入群: 否", zap.String("username", username))
				errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
				errorMsg = fmt.Sprintf("Not Invite for this team")
				goto COMPLETE
			}
		}

		//判断 teamID 是否存在
		if isExists, err := redis.Bool(redisConn.Do("EXISTS", fmt.Sprintf("TeamInfo:%s", teamID))); err != nil {
			kc.logger.Error("EXISTS Error", zap.Error(err))
			errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
			errorMsg = fmt.Sprintf("Query team info error[teamID=%s]", teamID)
			goto COMPLETE

		} else {
			if !isExists {
				errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
				errorMsg = fmt.Sprintf("Team is not exists[teamID=%s]", teamID)
				goto COMPLETE
			}

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

			if newSeq, err = redis.Uint64(redisConn.Do("INCR", fmt.Sprintf("userSeq:%s", teamInfo.Owner))); err != nil {
				kc.logger.Error("redisConn INCR userSeq Error", zap.Error(err))
				errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
				errorMsg = fmt.Sprintf("INCR error[Owner=%s]", teamInfo.Owner)
				goto COMPLETE
			}

			psSource := &Friends.PsSource{
				Ps:     req.GetPs(),
				Source: req.GetFrom(), //发起邀请方
			}
			psSourceData, _ := proto.Marshal(psSource)

			//获取当前用户的呢称
			nick, err := redis.String(redisConn.Do("HGET", fmt.Sprintf("userData:%s", username), "Nick"))
			if err != nil {
				kc.logger.Error("获取邀请方的呢称错误", zap.Error(err))
				errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
				errorMsg = fmt.Sprintf("HGET error[from=%s]", req.GetFrom())
				goto COMPLETE
			}

			body := Msg.MessageNotificationBody{
				Type:           Msg.MessageNotificationType_MNT_RejectTeamInvite, //用户拒绝群邀请
				HandledAccount: username,
				HandledMsg:     fmt.Sprintf("用户 %s  拒绝群邀请", nick),
				Status:         Msg.MessageStatus_MOS_Declined,
				Data:           psSourceData,
				To:             teamInfo.Owner,
			}
			bodyData, _ := proto.Marshal(&body)
			inviteEventRsp := &Msg.RecvMsgEventRsp{
				Scene:        Msg.MessageScene_MsgScene_S2C,        //系统消息
				Type:         Msg.MessageType_MsgType_Notification, //通知类型
				Body:         bodyData,
				From:         username,
				FromDeviceId: deviceID,
				ServerMsgId:  msg.GetID(),                        //服务器分配的消息ID
				WorkflowID:   req.GetWorkflowID(),                //工作流ID
				Seq:          newSeq,                             //消息序号，单个会话内自然递增, 这里是对inviteUsername这个用户的通知序号
				Uuid:         fmt.Sprintf("%d", msg.GetTaskID()), //客户端分配的消息ID，SDK生成的消息id，这里返回TaskID
				Time:         uint64(time.Now().UnixNano() / 1e6),
			}
			go kc.BroadcastMsgToAllDevices(inviteEventRsp, req.GetFrom()) //向邀请者发送此用户拒绝入群的通知

			//TODO 向自己的其它端推送

		}
	}

COMPLETE:
	msg.SetCode(int32(errorCode)) //状态码
	if errorCode == 200 {
		//只需200
	} else {
		msg.SetErrorMsg([]byte(errorMsg)) //错误提示
		msg.FillBody(nil)
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
4-8 主动申请加群
必须根据群的VerifyType设定来进行逻辑判断

说明：
1. 用户主动申请进入群组
   如果群组设置为需要审核，申请后管理员和群主会受到申请入群系统通知，需要等待管理员或者群主审核，如果群组设置为任何人可加入，则直接入群成功。

2. 向所有群成员推送用户入群通知

*/
func (kc *KafkaClient) HandleApplyTeam(msg *models.Message) error {
	var err error
	errorCode := 200
	var errorMsg string
	var newSeq uint64
	// var count int

	redisConn := kc.redisPool.Get()
	defer redisConn.Close()

	username := msg.GetUserName() //用户自己的账号
	// token := msg.GetJwtToken()
	deviceID := msg.GetDeviceID()

	kc.logger.Info("HandleApplyTeam start...",
		zap.String("username", username),
		zap.String("deviceId", deviceID))

	//取出当前设备的os， clientType， logonAt
	curDeviceHashKey := fmt.Sprintf("devices:%s:%s", username, deviceID)
	isMaster, _ := redis.Bool(redisConn.Do("HGET", curDeviceHashKey, "ismaster"))
	curOs, _ := redis.String(redisConn.Do("HGET", curDeviceHashKey, "os"))
	curClientType, _ := redis.Int(redisConn.Do("HGET", curDeviceHashKey, "clientType"))
	curLogonAt, _ := redis.Uint64(redisConn.Do("HGET", curDeviceHashKey, "logonAt"))

	kc.logger.Debug("ApplyTeam",
		zap.Bool("isMaster", isMaster),
		zap.String("username", username),
		zap.String("deviceID", deviceID),
		zap.String("curOs", curOs),
		zap.Int("curClientType", curClientType),
		zap.Uint64("curLogonAt", curLogonAt))

	//打开msg里的负载， 获取请求参数
	body := msg.GetContent()

	//解包body
	req := &Team.ApplyTeamReq{}
	if err := proto.Unmarshal(body, req); err != nil {
		kc.logger.Error("Protobuf Unmarshal Error", zap.Error(err))
		errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
		errorMsg = fmt.Sprintf("Protobuf Unmarshal Error: %s", err.Error())
		goto COMPLETE

	} else {
		kc.logger.Debug("ApplyTeam payload",
			zap.String("teamId", req.GetTeamId()),
			zap.String("ps", req.GetPs()),
		)

		teamID := req.GetTeamId()

		userData := new(models.User)
		userKey := fmt.Sprintf("userData:%s", username)
		if result, err := redis.Values(redisConn.Do("HGETALL", userKey)); err == nil {
			if err := redis.ScanStruct(result, userData); err != nil {

				kc.logger.Error("错误：ScanStruct", zap.Error(err))
				errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
				errorMsg = fmt.Sprintf("ScanStruct Error[Username=%s]", username)
				goto COMPLETE

			}
		}

		psSource := &Friends.PsSource{
			Ps:     req.GetPs(),
			Source: username, //发起邀请方， 主动的
		}
		psSourceData, _ := proto.Marshal(psSource)

		//判断 teamID 是否存在
		if isExists, err := redis.Bool(redisConn.Do("EXISTS", fmt.Sprintf("TeamInfo:%s", teamID))); err != nil {
			kc.logger.Error("EXISTS Error", zap.Error(err))
			errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
			errorMsg = fmt.Sprintf("Query team info error[teamID=%s]", teamID)
			goto COMPLETE

		} else {
			if !isExists {
				errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
				errorMsg = fmt.Sprintf("Team is not exists[teamID=%s]", teamID)
				goto COMPLETE
			}

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

			//判断username是不是被封禁了，如果是则返回
			if state, err := redis.Int(redisConn.Do("HGET", fmt.Sprintf("userData:%s", username), "State")); err != nil {
				kc.logger.Error("redisConn HGET Error", zap.Error(err))
				errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
				errorMsg = fmt.Sprintf("ser is not exists[Username=%s]", username)
				goto COMPLETE
			} else {
				if state == common.UserBlocked {
					kc.logger.Debug("User is blocked", zap.String("Username", username))
					errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
					errorMsg = fmt.Sprintf("ser is blocked[Username=%s]", username)
					goto COMPLETE
				}
			}

			//判断targetUsername是不是已经是群成员了，如果是，则返回

			//首先判断一下是否是群成员
			if reply, err := redisConn.Do("ZRANK", fmt.Sprintf("TeamUsers:%s", teamID), username); err == nil {
				if reply != nil { //是群成员
					err = nil
					kc.logger.Debug("User is already member", zap.String("Username", username))
					goto COMPLETE
				}
			}

			//TODO 生成工作流ID
			workflowID := uuid.NewV4().String()

			//判断入群校验模式
			switch Team.VerifyType(teamInfo.VerifyType) {
			case Team.VerifyType_Vt_Free: //所有人可加入

				//存储群成员信息 TeamUser
				teamUser := new(models.TeamUser)
				teamUser.JoinAt = time.Now().UnixNano() / 1e6
				teamUser.Teamname = teamInfo.Teamname
				teamUser.Username = userData.Username
				teamUser.InvitedUsername = ""
				teamUser.Nick = userData.Nick                                 //群成员呢称
				teamUser.Avatar = userData.Avatar                             //群成员头像
				teamUser.Label = userData.Label                               //群成员标签
				teamUser.Source = ""                                          //群成员来源  TODO
				teamUser.Extend = userData.Extend                             //群成员扩展字段
				teamUser.TeamMemberType = int(Team.TeamMemberType_Tmt_Normal) //群成员类型
				teamUser.IsMute = false                                       //是否被禁言
				teamUser.NotifyType = 1                                       //群消息通知方式 All(1) - 群全部消息提醒
				teamUser.Province = userData.Province                         //省份, 如广东省
				teamUser.City = userData.City                                 //城市，如广州市
				teamUser.County = userData.County                             //区，如天河区
				teamUser.Street = userData.Street                             //街道
				teamUser.Address = userData.Address                           //地址

				kc.SaveTeamUser(teamUser)

				//向群推送此用户的入群通知
				teamMembers, _ := redis.Strings(redisConn.Do("ZRANGEBYSCORE", fmt.Sprintf("TeamUsers:%s", teamID), "-inf", "+inf"))
				for _, teamMember := range teamMembers {

					if newSeq, err = redis.Uint64(redisConn.Do("INCR", fmt.Sprintf("userSeq:%s", teamMember))); err != nil {
						kc.logger.Error("redisConn INCR userSeq Error", zap.Error(err))
						errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
						errorMsg = fmt.Sprintf("INCR error[Username=%s]", teamMember)
						goto COMPLETE
					}
					body := Msg.MessageNotificationBody{
						Type:           Msg.MessageNotificationType_MNT_ApplyJoinTeam, //申请加群请求
						HandledAccount: username,
						HandledMsg:     fmt.Sprintf("用户: %s 申请加群请求获得通过", userData.Nick),
						Status:         Msg.MessageStatus_MOS_Passed,
						Data:           psSourceData, // 附带的文本 该系统消息的文本
						To:             teamID,       //群组id
					}
					bodyData, _ := proto.Marshal(&body)
					inviteEventRsp := &Msg.RecvMsgEventRsp{
						Scene:        Msg.MessageScene_MsgScene_S2C,        //系统消息
						Type:         Msg.MessageType_MsgType_Notification, //通知类型
						Body:         bodyData,                             //字节流
						From:         username,                             //发起人
						FromDeviceId: deviceID,
						ServerMsgId:  msg.GetID(),                        //服务器分配的消息ID
						WorkflowID:   workflowID,                         //工作流ID
						Seq:          newSeq,                             //消息序号，单个会话内自然递增, 这里是对inviteUsername这个用户的通知序号
						Uuid:         fmt.Sprintf("%d", msg.GetTaskID()), //客户端分配的消息ID，SDK生成的消息id，这里返回TaskID
						Time:         uint64(time.Now().UnixNano() / 1e6),
					}
					go kc.BroadcastMsgToAllDevices(inviteEventRsp, teamMember) //向群成员广播
				}

				/*
					1. 用户拥有的群，用有序集合存储，Key: Team:{Owner}, 成员元素是: TeamnID
					2. 群信息哈希表, key格式为: TeamInfo:{TeamnID}, 字段为: Teamname Nick Icon 等Team表的字段
					3. 用户有拥有的群用有序集合存储, key格式为： TeamUsers:{TeamnID}, 成员元素是: Username
					4. 每个群成员用哈希表存储，Key格式为： TeamUser:{TeamnID}:{Username} , 字段为: Teamname Username Nick JoinAt 等TeamUser表的字段
					5. 被移除的成员列表，Key格式为： RemoveTeamMembers:{TeamnID}
				*/
				err = redisConn.Send("ZADD", fmt.Sprintf("Team:%s", username), time.Now().UnixNano()/1e6, teamInfo.TeamID)
				err = redisConn.Send("HMSET", redis.Args{}.Add(fmt.Sprintf("TeamInfo:%s", teamInfo.TeamID)).AddFlat(teamInfo)...)

				//删除退群名单列表里的此用户
				err = redisConn.Send("ZREM", fmt.Sprintf("RemoveTeamMembers:%s", teamInfo.TeamID), time.Now().UnixNano()/1e6, username)

				err = redisConn.Send("ZADD", fmt.Sprintf("TeamUsers:%s", teamInfo.TeamID), time.Now().UnixNano()/1e6, username)
				err = redisConn.Send("HMSET", redis.Args{}.Add(fmt.Sprintf("TeamUser:%s:%s", teamInfo.TeamID, username)).AddFlat(teamUser)...)
				//更新redis的sync:{用户账号} teamsAt 时间戳
				err = redisConn.Send("HSET",
					fmt.Sprintf("sync:%s", username),
					"teamsAt",
					time.Now().UnixNano()/1e6)

				redisConn.Flush()

			case Team.VerifyType_Vt_Apply: //需要审核加入

				//向群主或管理员推送此用户的主动加群通知
				managers, _ := kc.GetOwnerAndManagers(teamInfo.TeamID)
				for _, manager := range managers {
					if newSeq, err = redis.Uint64(redisConn.Do("INCR", fmt.Sprintf("userSeq:%s", manager))); err != nil {
						kc.logger.Error("redisConn INCR userSeq Error", zap.Error(err))
						errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
						errorMsg = fmt.Sprintf("INCR error[Owner=%s]", manager)
						goto COMPLETE
					}
					body := Msg.MessageNotificationBody{
						Type:           Msg.MessageNotificationType_MNT_CheckTeamInvite, //向群主推送此用户的主动入群通知
						HandledAccount: username,
						HandledMsg:     fmt.Sprintf("用户: %s 发出申请加群请求", userData.Nick),
						Status:         Msg.MessageStatus_MOS_Processing,
						Data:           psSourceData,
						To:             teamInfo.TeamID, //群id
					}
					bodyData, _ := proto.Marshal(&body)
					inviteEventRsp := &Msg.RecvMsgEventRsp{
						Scene:        Msg.MessageScene_MsgScene_S2C,        //系统消息
						Type:         Msg.MessageType_MsgType_Notification, //通知类型
						Body:         bodyData,
						From:         username, //发起人
						FromDeviceId: deviceID,
						ServerMsgId:  msg.GetID(),                        //服务器分配的消息ID
						WorkflowID:   workflowID,                         //工作流ID
						Seq:          newSeq,                             //消息序号，单个会话内自然递增, 这里是对inviteUsername这个用户的通知序号
						Uuid:         fmt.Sprintf("%d", msg.GetTaskID()), //客户端分配的消息ID，SDK生成的消息id，这里返回TaskID
						Time:         uint64(time.Now().UnixNano() / 1e6),
					}
					go kc.BroadcastMsgToAllDevices(inviteEventRsp, manager) //群主或管理员
				}

			case Team.VerifyType_Vt_Private: //仅限邀请加入
				kc.logger.Warn("此群仅限邀请加入")
				errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
				errorMsg = fmt.Sprintf("Team is private, not support apply join")
				goto COMPLETE
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
		kc.logger.Info("Succeed succeed send message to ProduceChannel", zap.String("topic", topic))
	} else {
		kc.logger.Error("Failed to send  message to ProduceChannel", zap.Error(err))
	}
	_ = err
	return nil

}

/*
4-9 批准加群申请

权限:
只有群主及管理员才能批准通过群申请
*/

func (kc *KafkaClient) HandlePassTeamApply(msg *models.Message) error {
	var err error
	errorCode := 200
	var errorMsg string
	var newSeq uint64
	// var count int

	redisConn := kc.redisPool.Get()
	defer redisConn.Close()

	username := msg.GetUserName() //用户自己的账号
	// token := msg.GetJwtToken()
	deviceID := msg.GetDeviceID()

	kc.logger.Info("HandlePassTeamApply start...",
		zap.String("username", username),
		zap.String("deviceId", deviceID))

	//取出当前设备的os， clientType， logonAt
	curDeviceHashKey := fmt.Sprintf("devices:%s:%s", username, deviceID)
	isMaster, _ := redis.Bool(redisConn.Do("HGET", curDeviceHashKey, "ismaster"))
	curOs, _ := redis.String(redisConn.Do("HGET", curDeviceHashKey, "os"))
	curClientType, _ := redis.Int(redisConn.Do("HGET", curDeviceHashKey, "clientType"))
	curLogonAt, _ := redis.Uint64(redisConn.Do("HGET", curDeviceHashKey, "logonAt"))

	kc.logger.Debug("PassTeamApply ",
		zap.Bool("isMaster", isMaster),
		zap.String("username", username),
		zap.String("deviceID", deviceID),
		zap.String("curOs", curOs),
		zap.Int("curClientType", curClientType),
		zap.Uint64("curLogonAt", curLogonAt))

	//打开msg里的负载， 获取请求参数
	body := msg.GetContent()

	//解包body
	req := &Team.PassTeamApplyReq{}
	if err := proto.Unmarshal(body, req); err != nil {
		kc.logger.Error("Protobuf Unmarshal Error", zap.Error(err))
		errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
		errorMsg = fmt.Sprintf("Protobuf Unmarshal Error: %s", err.Error())
		goto COMPLETE

	} else {
		kc.logger.Debug("PassTeamApply  payload",
			zap.String("teamId", req.GetTeamId()),         // 群组ID
			zap.String("from", req.GetFrom()),             //申请方账号
			zap.String("workflowID", req.GetWorkflowID()), //工作流ID
		)

		teamID := req.GetTeamId()
		targetUsername := req.GetFrom() //申请方
		psSource := &Friends.PsSource{
			Ps:     "",
			Source: req.GetFrom(), //主动加群
		}
		psSourceData, _ := proto.Marshal(psSource)

		//判断 teamID 是否存在
		if isExists, err := redis.Bool(redisConn.Do("EXISTS", fmt.Sprintf("TeamInfo:%s", teamID))); err != nil {
			kc.logger.Error("EXISTS Error", zap.Error(err))
			errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
			errorMsg = fmt.Sprintf("Query team info error[teamID=%s]", teamID)
			goto COMPLETE

		} else {
			if !isExists {
				errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
				errorMsg = fmt.Sprintf("Team is not exists[teamID=%s]", teamID)
				goto COMPLETE
			}

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

			//判断targetUsername是不是被封禁了，如果是则返回
			if state, err := redis.Int(redisConn.Do("HGET", fmt.Sprintf("userData:%s", targetUsername), "State")); err != nil {
				kc.logger.Error("redisConn HGET Error", zap.Error(err))
				errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
				errorMsg = fmt.Sprintf("ser is not exists[Username=%s]", targetUsername)
				goto COMPLETE
			} else {
				if state == common.UserBlocked {
					kc.logger.Debug("User is blocked", zap.String("Username", targetUsername))
					errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
					errorMsg = fmt.Sprintf("ser is blocked[Username=%s]", targetUsername)
					goto COMPLETE
				}
			}

			//判断targetUsername是不是已经是群成员了，如果是，则返回

			//首先判断一下是否是群成员
			if reply, err := redisConn.Do("ZRANK", fmt.Sprintf("TeamUsers:%s", teamID), targetUsername); err == nil {
				if reply != nil { //是群成员
					err = nil
					kc.logger.Debug("User is already member", zap.String("Username", targetUsername))
					goto COMPLETE
				}
			}

			//判断操作者是不是群主或管理员
			opUser := new(models.TeamUser)
			if result, err := redis.Values(redisConn.Do("HGETALL", fmt.Sprintf("TeamUser:%s:%s", teamID, username))); err == nil {
				if err := redis.ScanStruct(result, opUser); err != nil {
					kc.logger.Error("TeamUser is not exist", zap.Error(err))
					errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
					errorMsg = fmt.Sprintf("TeamUser is not exists[teamID=%s, teamUser=%s]", teamID, username)
					goto COMPLETE
				}
			}
			teamMemberType := Team.TeamMemberType(opUser.TeamMemberType)
			if teamMemberType == Team.TeamMemberType_Tmt_Owner || teamMemberType == Team.TeamMemberType_Tmt_Manager {
				//pass
			} else {
				kc.logger.Warn("User is not team owner or manager", zap.String("Username", username))
				errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
				errorMsg = fmt.Sprintf("User is not team owner[Username=%s]", username)
				goto COMPLETE
			}

			userData := new(models.User)
			userKey := fmt.Sprintf("userData:%s", targetUsername)
			if result, err := redis.Values(redisConn.Do("HGETALL", userKey)); err == nil {
				if err := redis.ScanStruct(result, userData); err != nil {

					kc.logger.Error("错误：ScanStruct", zap.Error(err))
					errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
					errorMsg = fmt.Sprintf("ScanStruct Error[Username=%s]", targetUsername)
					goto COMPLETE

				}
			}
			//存储群成员信息 TeamUser
			teamUser := new(models.TeamUser)
			teamUser.JoinAt = time.Now().UnixNano() / 1e6
			teamUser.Teamname = teamInfo.Teamname
			teamUser.Username = userData.Username
			teamUser.Nick = userData.Nick                                 //群成员呢称
			teamUser.Avatar = userData.Avatar                             //群成员头像
			teamUser.Label = userData.Label                               //群成员标签
			teamUser.Source = ""                                          //群成员来源  TODO
			teamUser.Extend = userData.Extend                             //群成员扩展字段
			teamUser.TeamMemberType = int(Team.TeamMemberType_Tmt_Normal) //群成员类型
			teamUser.IsMute = false                                       //是否被禁言
			teamUser.NotifyType = 1                                       //群消息通知方式 All(1) - 群全部消息提醒
			teamUser.Province = userData.Province                         //省份, 如广东省
			teamUser.City = userData.City                                 //城市，如广州市
			teamUser.County = userData.County                             //区，如天河区
			teamUser.Street = userData.Street                             //街道
			teamUser.Address = userData.Address                           //地址

			kc.SaveTeamUser(teamUser)

			handledMsg := fmt.Sprintf("管理员: %s 同意 %s 入群申请", opUser.Nick, userData.Nick)

			//向所有群成员推送此用户的入群通知
			teamMembers, _ := redis.Strings(redisConn.Do("ZRANGEBYSCORE", fmt.Sprintf("TeamUsers:%s", teamID), "-inf", "+inf"))
			for _, teamMember := range teamMembers {

				if newSeq, err = redis.Uint64(redisConn.Do("INCR", fmt.Sprintf("userSeq:%s", teamMember))); err != nil {
					kc.logger.Error("redisConn INCR userSeq Error", zap.Error(err))
					errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
					errorMsg = fmt.Sprintf("INCR error[Username=%s]", teamMember)
					goto COMPLETE
				}
				body := Msg.MessageNotificationBody{
					Type:           Msg.MessageNotificationType_MNT_PassTeamApply, //同意群邀请
					HandledAccount: username,
					HandledMsg:     handledMsg,
					Status:         Msg.MessageStatus_MOS_Passed,
					Data:           psSourceData, // 附带的文本 该系统消息的文本
					To:             teamID,       //群组id
				}
				bodyData, _ := proto.Marshal(&body)
				inviteEventRsp := &Msg.RecvMsgEventRsp{
					Scene:        Msg.MessageScene_MsgScene_S2C,        //系统消息
					Type:         Msg.MessageType_MsgType_Notification, //通知类型
					Body:         bodyData,
					From:         username, //当前用户
					FromDeviceId: deviceID,
					ServerMsgId:  msg.GetID(),                        //服务器分配的消息ID
					WorkflowID:   req.GetWorkflowID(),                //工作流ID
					Seq:          newSeq,                             //消息序号，单个会话内自然递增, 这里是对inviteUsername这个用户的通知序号
					Uuid:         fmt.Sprintf("%d", msg.GetTaskID()), //客户端分配的消息ID，SDK生成的消息id，这里返回TaskID
					Time:         uint64(time.Now().UnixNano() / 1e6),
				}
				go kc.BroadcastMsgToAllDevices(inviteEventRsp, teamMember) //向群成员广播
			}
			/*
				1. 用户拥有的群，用有序集合存储，Key: Team:{Owner}, 成员元素是: TeamnID
				2. 群信息哈希表, key格式为: TeamInfo:{TeamnID}, 字段为: Teamname Nick Icon 等Team表的字段
				3. 用户有拥有的群用有序集合存储, key格式为： TeamUsers:{TeamnID}, 成员元素是: Username
				4. 每个群成员用哈希表存储，Key格式为： TeamUser:{TeamnID}:{Username} , 字段为: Teamname Username Nick JoinAt 等TeamUser表的字段
				5. 被移除的成员列表，Key格式为： RemoveTeamMembers:{TeamnID}
			*/
			err = redisConn.Send("ZADD", fmt.Sprintf("Team:%s", targetUsername), time.Now().UnixNano()/1e6, teamInfo.TeamID)
			err = redisConn.Send("HMSET", redis.Args{}.Add(fmt.Sprintf("TeamInfo:%s", teamInfo.TeamID)).AddFlat(teamInfo)...)
			//删除退群名单列表里的此用户
			err = redisConn.Send("ZREM", fmt.Sprintf("RemoveTeamMembers:%s", teamInfo.TeamID), time.Now().UnixNano()/1e6, targetUsername)

			err = redisConn.Send("ZADD", fmt.Sprintf("TeamUsers:%s", teamInfo.TeamID), time.Now().UnixNano()/1e6, targetUsername)
			err = redisConn.Send("HMSET", redis.Args{}.Add(fmt.Sprintf("TeamUser:%s:%s", teamInfo.TeamID, targetUsername)).AddFlat(teamUser)...)
			//更新redis的sync:{用户账号} teamsAt 时间戳
			redisConn.Send("HSET",
				fmt.Sprintf("sync:%s", targetUsername),
				"teamsAt",
				time.Now().UnixNano()/1e6)

			redisConn.Flush()

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
		kc.logger.Info("Succeed succeed send message to ProduceChannel", zap.String("topic", topic))
	} else {
		kc.logger.Error("Failed to send  message to ProduceChannel", zap.Error(err))
	}
	_ = err
	return nil

}

/*
4-10 否决加群申请

权限:
只有群主及管理员才能否决加群申请
*/

func (kc *KafkaClient) HandleRejectTeamApply(msg *models.Message) error {
	var err error
	errorCode := 200
	var errorMsg string
	var newSeq uint64

	redisConn := kc.redisPool.Get()
	defer redisConn.Close()

	username := msg.GetUserName() //用户自己的账号
	// token := msg.GetJwtToken()
	deviceID := msg.GetDeviceID()

	kc.logger.Info("HandleRejectTeamApply start...",
		zap.String("username", username),
		zap.String("deviceId", deviceID))

	//取出当前设备的os， clientType， logonAt
	curDeviceHashKey := fmt.Sprintf("devices:%s:%s", username, deviceID)
	isMaster, _ := redis.Bool(redisConn.Do("HGET", curDeviceHashKey, "ismaster"))
	curOs, _ := redis.String(redisConn.Do("HGET", curDeviceHashKey, "os"))
	curClientType, _ := redis.Int(redisConn.Do("HGET", curDeviceHashKey, "clientType"))
	curLogonAt, _ := redis.Uint64(redisConn.Do("HGET", curDeviceHashKey, "logonAt"))

	kc.logger.Debug("RejectTeamApply",
		zap.Bool("isMaster", isMaster),
		zap.String("username", username),
		zap.String("deviceID", deviceID),
		zap.String("curOs", curOs),
		zap.Int("curClientType", curClientType),
		zap.Uint64("curLogonAt", curLogonAt))

	//打开msg里的负载， 获取请求参数
	body := msg.GetContent()

	//解包body
	req := &Team.RejectTeamApplyReq{}
	if err := proto.Unmarshal(body, req); err != nil {
		kc.logger.Error("Protobuf Unmarshal Error", zap.Error(err))
		errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
		errorMsg = fmt.Sprintf("Protobuf Unmarshal Error: %s", err.Error())
		goto COMPLETE

	} else {
		kc.logger.Debug("RejectTeamApply  payload",
			zap.String("teamId", req.GetTeamId()),
			zap.String("from", req.GetFrom()),
			zap.String("workflowID", req.GetWorkflowID()),
			zap.String("ps", req.GetPs()),
		)

		teamID := req.GetTeamId()
		targetUsername := req.GetFrom()

		psSource := &Friends.PsSource{
			Ps:     "",
			Source: req.GetFrom(), //主动加群
		}
		psSourceData, _ := proto.Marshal(psSource)

		//判断 teamID 是否存在
		if isExists, err := redis.Bool(redisConn.Do("EXISTS", fmt.Sprintf("TeamInfo:%s", teamID))); err != nil {
			kc.logger.Error("EXISTS Error", zap.Error(err))
			errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
			errorMsg = fmt.Sprintf("Query team info error[teamID=%s]", teamID)
			goto COMPLETE

		} else {
			if !isExists {
				errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
				errorMsg = fmt.Sprintf("Team is not exists[teamID=%s]", teamID)
				goto COMPLETE
			}

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

			//判断targetUsername是不是被封禁了，如果是则返回
			if state, err := redis.Int(redisConn.Do("HGET", fmt.Sprintf("userData:%s", targetUsername), "State")); err != nil {
				kc.logger.Error("redisConn HGET Error", zap.Error(err))
				errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
				errorMsg = fmt.Sprintf("ser is not exists[Username=%s]", targetUsername)
				goto COMPLETE
			} else {
				if state == common.UserBlocked {
					kc.logger.Debug("User is blocked", zap.String("Username", targetUsername))
					errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
					errorMsg = fmt.Sprintf("ser is blocked[Username=%s]", targetUsername)
					goto COMPLETE
				}
			}

			//判断操作者是不是群主或管理员
			opUser := new(models.TeamUser)
			if result, err := redis.Values(redisConn.Do("HGETALL", fmt.Sprintf("TeamUser:%s:%s", teamID, username))); err == nil {
				if err := redis.ScanStruct(result, opUser); err != nil {
					kc.logger.Error("Operate User is not exist", zap.Error(err))
					errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
					errorMsg = fmt.Sprintf("Operate is not exists[teamID=%s, teamUser=%s]", teamID, username)
					goto COMPLETE
				}
			}
			teamMemberType := Team.TeamMemberType(opUser.TeamMemberType)
			if teamMemberType == Team.TeamMemberType_Tmt_Owner || teamMemberType == Team.TeamMemberType_Tmt_Manager {
				//pass
			} else {
				kc.logger.Warn("Operate User is not team owner or manager", zap.String("Username", username))
				errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
				errorMsg = fmt.Sprintf("Operate User is not team owner[Username=%s]", username)
				goto COMPLETE
			}

			userNick, _ := redis.String(redisConn.Do("HGET", fmt.Sprintf("userData:%s", targetUsername), "Nick"))
			handledMsg := fmt.Sprintf("管理员: %s 拒绝 %s 入群申请", opUser.Nick, userNick)

			//向此用户推送拒绝入群的通知
			if newSeq, err = redis.Uint64(redisConn.Do("INCR", fmt.Sprintf("userSeq:%s", targetUsername))); err != nil {
				kc.logger.Error("redisConn INCR userSeq Error", zap.Error(err))
				errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
				errorMsg = fmt.Sprintf("INCR error[Username=%s]", targetUsername)
				goto COMPLETE
			}
			body := Msg.MessageNotificationBody{
				Type:           Msg.MessageNotificationType_MNT_RejectTeamApply, //管理员拒绝加群申请
				HandledAccount: username,                                        //当前是管理员
				HandledMsg:     handledMsg,
				Status:         Msg.MessageStatus_MOS_Declined,
				Data:           psSourceData,
				To:             teamID, //群组id
			}
			bodyData, _ := proto.Marshal(&body)
			inviteEventRsp := &Msg.RecvMsgEventRsp{
				Scene:        Msg.MessageScene_MsgScene_S2C,        //系统消息
				Type:         Msg.MessageType_MsgType_Notification, //通知类型
				Body:         bodyData,
				From:         username, //当前用户
				FromDeviceId: deviceID,
				ServerMsgId:  msg.GetID(),                        //服务器分配的消息ID
				WorkflowID:   req.GetWorkflowID(),                //工作流ID
				Seq:          newSeq,                             //消息序号，单个会话内自然递增, 这里是对inviteUsername这个用户的通知序号
				Uuid:         fmt.Sprintf("%d", msg.GetTaskID()), //客户端分配的消息ID，SDK生成的消息id，这里返回TaskID
				Time:         uint64(time.Now().UnixNano() / 1e6),
			}
			go kc.BroadcastMsgToAllDevices(inviteEventRsp, targetUsername)

			//向群的群主及管理员发送拒绝入群消息
			managers, _ := kc.GetOwnerAndManagers(teamID)
			for _, manager := range managers {
				if manager == opUser.Username { //不发给当前管理员
					continue
				}
				if newSeq, err = redis.Uint64(redisConn.Do("INCR", fmt.Sprintf("userSeq:%s", manager))); err != nil {
					kc.logger.Error("redisConn INCR userSeq Error", zap.Error(err))
					continue
				}

				body := Msg.MessageNotificationBody{
					Type:           Msg.MessageNotificationType_MNT_RejectTeamApply, //管理员拒绝加群申请
					HandledAccount: opUser.Username,
					HandledMsg:     handledMsg,
					Status:         Msg.MessageStatus_MOS_Declined,
					Data:           psSourceData, // 附带的文本 该系统消息的文本
					To:             teamID,       //群组id
				}
				bodyData, _ := proto.Marshal(&body)
				inviteEventRsp := &Msg.RecvMsgEventRsp{
					Scene:        Msg.MessageScene_MsgScene_S2C,        //系统消息
					Type:         Msg.MessageType_MsgType_Notification, //通知类型
					Body:         bodyData,
					From:         username, //当前用户
					FromDeviceId: deviceID,
					ServerMsgId:  msg.GetID(),                        //服务器分配的消息ID
					Seq:          newSeq,                             //消息序号，单个会话内自然递增, 这里是对inviteUsername这个用户的通知序号
					Uuid:         fmt.Sprintf("%d", msg.GetTaskID()), //客户端分配的消息ID，SDK生成的消息id，这里返回TaskID
					Time:         uint64(time.Now().UnixNano() / 1e6),
				}
				go kc.BroadcastMsgToAllDevices(inviteEventRsp, manager) //向群的群主及管理员广播
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
		kc.logger.Info("Succeed succeed send message to ProduceChannel", zap.String("topic", topic))
	} else {
		kc.logger.Error("Failed to send  message to ProduceChannel", zap.Error(err))
	}
	_ = err
	return nil

}

/*
4-11 更新群组信息
只有群主才能更新群组信息
需要 修改 sync:{用户} teamsAt
*/

func (kc *KafkaClient) HandleUpdateTeam(msg *models.Message) error {
	var err error
	errorCode := 200
	var errorMsg string
	var data []byte
	rsp := &Team.UpdateTeamRsp{}

	redisConn := kc.redisPool.Get()
	defer redisConn.Close()

	username := msg.GetUserName() //用户自己的账号
	// token := msg.GetJwtToken()
	deviceID := msg.GetDeviceID()

	kc.logger.Info("HandleUpdateTeam start...",
		zap.String("username", username),
		zap.String("deviceId", deviceID))

	//取出当前设备的os， clientType， logonAt
	curDeviceHashKey := fmt.Sprintf("devices:%s:%s", username, deviceID)
	isMaster, _ := redis.Bool(redisConn.Do("HGET", curDeviceHashKey, "ismaster"))
	curOs, _ := redis.String(redisConn.Do("HGET", curDeviceHashKey, "os"))
	curClientType, _ := redis.Int(redisConn.Do("HGET", curDeviceHashKey, "clientType"))
	curLogonAt, _ := redis.Uint64(redisConn.Do("HGET", curDeviceHashKey, "logonAt"))

	kc.logger.Debug("UpdateTeam",
		zap.Bool("isMaster", isMaster),
		zap.String("username", username),
		zap.String("deviceID", deviceID),
		zap.String("curOs", curOs),
		zap.Int("curClientType", curClientType),
		zap.Uint64("curLogonAt", curLogonAt))

	//打开msg里的负载， 获取请求参数
	body := msg.GetContent()

	//解包body
	req := &Team.UpdateTeamReq{}
	if err := proto.Unmarshal(body, req); err != nil {
		kc.logger.Error("Protobuf Unmarshal Error", zap.Error(err))
		errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
		errorMsg = fmt.Sprintf("Protobuf Unmarshal Error: %s", err.Error())
		goto COMPLETE

	} else {
		kc.logger.Debug("UpdateTeam payload",
			zap.String("teamId", req.GetTeamId()),
		)

		teamID := req.GetTeamId()

		//判断 teamID 是否存在
		if isExists, err := redis.Bool(redisConn.Do("EXISTS", fmt.Sprintf("TeamInfo:%s", teamID))); err != nil {
			kc.logger.Error("EXISTS Error", zap.Error(err))
			errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
			errorMsg = fmt.Sprintf("Query team info error[teamID=%s]", teamID)
			goto COMPLETE

		} else {
			if !isExists {
				errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
				errorMsg = fmt.Sprintf("Team is not exists[teamID=%s]", teamID)
				goto COMPLETE
			}

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

			//判断操作者是不是群主或管理员
			teamMemberType, _ := redis.Int(redisConn.Do("HGET", fmt.Sprintf("TeamUser:%s:%s", teamID, username), "TeamMemberType"))

			if Team.TeamMemberType(teamMemberType) == Team.TeamMemberType_Tmt_Owner || Team.TeamMemberType(teamMemberType) == Team.TeamMemberType_Tmt_Manager {
				//管理员或群主

			} else {
				kc.logger.Warn("User is not team owner or manager", zap.String("Username", username))
				errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
				errorMsg = fmt.Sprintf("User is not team owner or manager [Username=%s]", username)
				goto COMPLETE
			}

			if nick, ok := req.Fields[1]; ok {
				//修改群组呢称
				teamInfo.Nick = nick

			}
			if icon, ok := req.Fields[2]; ok {
				//修改群组Icon
				teamInfo.Icon = icon

			}
			if announcement, ok := req.Fields[3]; ok {
				//修改群组Announcement
				teamInfo.Announcement = announcement

			}
			if introduce, ok := req.Fields[4]; ok {
				//修改群组Introductory
				teamInfo.Introductory = introduce
			}

			if verifyTypeStr, ok := req.Fields[5]; ok {
				verifyType := 1 //默认
				if verifyTypeStr != "" {
					if n, err := strconv.ParseUint(verifyTypeStr, 10, 64); err == nil {
						verifyType = int(n)
					}
				}
				//修改群组VerifyType
				teamInfo.VerifyType = verifyType
			}

			if inviteModeStr, ok := req.Fields[6]; ok {
				inviteMode := 1 //默认
				if inviteModeStr != "" {
					if n, err := strconv.ParseUint(inviteModeStr, 10, 64); err == nil {
						inviteMode = int(n)
					}
				}
				//修改群组InviteMode
				teamInfo.InviteMode = inviteMode
			}

			kc.SaveTeam(teamInfo)

			//对所有群成员
			teamMembers, _ := redis.Strings(redisConn.Do("ZRANGEBYSCORE", fmt.Sprintf("TeamUsers:%s", teamID), "-inf", "+inf"))
			curAt := time.Now().UnixNano() / 1e6
			var newSeq uint64
			for _, teamMember := range teamMembers {

				if newSeq, err = redis.Uint64(redisConn.Do("INCR", fmt.Sprintf("userSeq:%s", teamMember))); err != nil {
					kc.logger.Error("redisConn INCR userSeq Error", zap.Error(err))
					continue
				}
				userNick, _ := redis.String(redisConn.Do("HGET", fmt.Sprintf("userData:%s", username), "Nick"))
				handledMsg := fmt.Sprintf("管理员 %s 更新群资料", userNick)
				//向群成员发出更新群资料通知
				body := Msg.MessageNotificationBody{
					Type:           Msg.MessageNotificationType_MNT_UpdateTeam, //更新群资料
					HandledAccount: username,
					HandledMsg:     handledMsg,
					Status:         Msg.MessageStatus_MOS_Processing,
					Data:           []byte(""),
					To:             teamID, //群id
				}
				bodyData, _ := proto.Marshal(&body)
				mrsp := &Msg.RecvMsgEventRsp{
					Scene:        Msg.MessageScene_MsgScene_S2C,        //系统消息
					Type:         Msg.MessageType_MsgType_Notification, //通知类型
					Body:         bodyData,
					From:         username,
					FromDeviceId: deviceID,
					ServerMsgId:  msg.GetID(),                        //服务器分配的消息ID
					Seq:          newSeq,                             //消息序号，单个会话内自然递增
					Uuid:         fmt.Sprintf("%d", msg.GetTaskID()), //客户端分配的消息ID，SDK生成的消息id，这里返回TaskID
					Time:         uint64(curAt),
				}
				go kc.BroadcastMsgToAllDevices(mrsp, teamMember)

				//更新redis的sync:{用户账号} teamsAt 时间戳
				redisConn.Send("HSET",
					fmt.Sprintf("sync:%s", teamMember),
					"teamsAt",
					curAt)
			}

			//一次性写入到Redis
			redisConn.Flush()

			rsp.TeamId = teamID
			rsp.TimeAt = uint64(time.Now().UnixNano() / 1e6)
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
		kc.logger.Info("Succeed succeed send message to ProduceChannel", zap.String("topic", topic))
	} else {
		kc.logger.Error("Failed to send  message to ProduceChannel", zap.Error(err))
	}
	_ = err
	return nil

}

/*
4-13 退群

*/

func (kc *KafkaClient) HandleLeaveTeam(msg *models.Message) error {
	var err error
	errorCode := 200
	var errorMsg string
	var newSeq uint64
	var teamID string

	redisConn := kc.redisPool.Get()
	defer redisConn.Close()

	username := msg.GetUserName() //用户自己的账号
	// token := msg.GetJwtToken()
	deviceID := msg.GetDeviceID()

	kc.logger.Info("HandleLeaveTeam start...",
		zap.String("username", username),
		zap.String("deviceId", deviceID))

	//取出当前设备的os， clientType， logonAt
	curDeviceHashKey := fmt.Sprintf("devices:%s:%s", username, deviceID)
	isMaster, _ := redis.Bool(redisConn.Do("HGET", curDeviceHashKey, "ismaster"))
	curOs, _ := redis.String(redisConn.Do("HGET", curDeviceHashKey, "os"))
	curClientType, _ := redis.Int(redisConn.Do("HGET", curDeviceHashKey, "clientType"))
	curLogonAt, _ := redis.Uint64(redisConn.Do("HGET", curDeviceHashKey, "logonAt"))

	kc.logger.Debug("LeaveTeam ",
		zap.Bool("isMaster", isMaster),
		zap.String("username", username),
		zap.String("deviceID", deviceID),
		zap.String("curOs", curOs),
		zap.Int("curClientType", curClientType),
		zap.Uint64("curLogonAt", curLogonAt))

	//打开msg里的负载， 获取请求参数
	body := msg.GetContent()

	//解包body
	req := &Team.LeaveTeamReq{}
	if err := proto.Unmarshal(body, req); err != nil {
		kc.logger.Error("Protobuf Unmarshal Error", zap.Error(err))
		errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
		errorMsg = fmt.Sprintf("Protobuf Unmarshal Error: %s", err.Error())
		goto COMPLETE

	} else {
		kc.logger.Debug("LeaveTeam  payload",
			zap.String("teamId", req.GetTeamId()),
		)

		teamID = req.GetTeamId()

		psSource := &Friends.PsSource{
			Ps:     "",
			Source: username, //主动退群的用户
		}
		psSourceData, _ := proto.Marshal(psSource)

		//判断 teamID 是否存在
		if isExists, err := redis.Bool(redisConn.Do("EXISTS", fmt.Sprintf("TeamInfo:%s", teamID))); err != nil {
			kc.logger.Error("EXISTS Error", zap.Error(err))
			errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
			errorMsg = fmt.Sprintf("Query team info error[teamID=%s]", teamID)
			goto COMPLETE

		} else {
			if !isExists {
				errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
				errorMsg = fmt.Sprintf("Team is not exists[teamID=%s]", teamID)
				goto COMPLETE
			}

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

			//首先判断一下是否是群成员
			if reply, err := redisConn.Do("ZRANK", fmt.Sprintf("TeamUsers:%s", teamID), username); err == nil {
				if reply != nil { //是群成员
					//判断是否有权移除， 例如，管理员不能在这里移除， 群主不能被移除
					removeUser := new(models.TeamUser)
					if result, err := redis.Values(redisConn.Do("HGETALL", fmt.Sprintf("TeamUser:%s:%s", teamID, username))); err == nil {
						if err := redis.ScanStruct(result, removeUser); err != nil {
							kc.logger.Error("TeamUser is not exist", zap.Error(err))
							errorCode = http.StatusBadRequest //错误码，400
							errorMsg = fmt.Sprintf("TeamUser is not exists[teamID=%s, teamUser=%s]", teamID, username)
							goto COMPLETE
						}
					}
					teamMemberType := Team.TeamMemberType(removeUser.TeamMemberType)

					if teamMemberType == Team.TeamMemberType_Tmt_Owner || teamMemberType == Team.TeamMemberType_Tmt_Manager {
						//管理员或群主
						kc.logger.Error("管理员或群主不能退群，必须由群主删除")
						errorCode = http.StatusBadRequest //错误码，400
						errorMsg = fmt.Sprintf("管理员或群主不能退群，必须由群主删除")
						goto COMPLETE

					} else {

						//删除此用户在群里的数据
						if err := kc.DeleteTeamUser(teamID, username); err != nil {
							kc.logger.Error("移除群成员失败", zap.Error(err))
							errorCode = http.StatusBadRequest //错误码，400
							errorMsg = fmt.Sprintf("移除群成员失败")
							goto COMPLETE
						}

						//删除redis里的TeamUser哈希表
						err = redisConn.Send("DEL", fmt.Sprintf("TeamUser:%s:%s", teamInfo.TeamID, username))
						//删除群成员的有序集合
						err = redisConn.Send("ZREM", fmt.Sprintf("TeamUsers:%s", teamID), username)
						//增加此群的退群名单
						err = redisConn.Send("ZADD", fmt.Sprintf("RemoveTeamMembers:%s", teamID), username)

						//删除Team:{username}里teamID
						err = redisConn.Send("ZREM", fmt.Sprintf("Team:%s", username), teamInfo.TeamID)
						//增加到用户自己的退群列表
						err = redisConn.Send("ZADD", fmt.Sprintf("RemoveTeam:%s", username), time.Now().UnixNano()/1e6, teamID)

						//一次性写入到Redis
						redisConn.Flush()

						//对所有群成员
						teamMembers, _ := redis.Strings(redisConn.Do("ZRANGEBYSCORE", fmt.Sprintf("TeamUsers:%s", teamID), "-inf", "+inf"))
						curAt := time.Now().UnixNano() / 1e6

						for _, teamMember := range teamMembers {

							if newSeq, err = redis.Uint64(redisConn.Do("INCR", fmt.Sprintf("userSeq:%s", teamMember))); err != nil {
								kc.logger.Error("redisConn INCR userSeq Error", zap.Error(err))
								continue
							}

							//向群成员发出用户退群通知
							body := Msg.MessageNotificationBody{
								Type:           Msg.MessageNotificationType_MNT_QuitTeam, //主动退群
								HandledAccount: username,
								HandledMsg:     fmt.Sprintf("用户 %s 退出本群", removeUser.Nick),
								Status:         Msg.MessageStatus_MOS_Done,
								Data:           psSourceData,
								To:             teamID, //群id
							}
							bodyData, _ := proto.Marshal(&body)
							mrsp := &Msg.RecvMsgEventRsp{
								Scene:        Msg.MessageScene_MsgScene_S2C,        //系统消息
								Type:         Msg.MessageType_MsgType_Notification, //通知类型
								Body:         bodyData,
								From:         username,
								FromDeviceId: deviceID,
								ServerMsgId:  msg.GetID(),                        //服务器分配的消息ID
								Seq:          newSeq,                             //消息序号，单个会话内自然递增
								Uuid:         fmt.Sprintf("%d", msg.GetTaskID()), //客户端分配的消息ID，SDK生成的消息id，这里返回TaskID
								Time:         uint64(curAt),
							}
							go kc.BroadcastMsgToAllDevices(mrsp, teamMember)

						}

					}

				}
			}

		}
	}

COMPLETE:
	msg.SetCode(int32(errorCode)) //状态码
	if errorCode == 200 {
		//200
		kc.logger.Info("退群成功", zap.String("teamID", teamID))

	} else {
		msg.SetErrorMsg([]byte(errorMsg)) //错误提示
		msg.FillBody(nil)
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
4-14 设置群管理员

群主设置群管理员

权限:
只有群主才能设置或删除

*/

func (kc *KafkaClient) HandleAddTeamManagers(msg *models.Message) error {
	var err error
	errorCode := 200
	var errorMsg string
	rsp := &Team.AddTeamManagersRsp{
		AbortedUsernames: make([]string, 0),
	}
	var data []byte

	redisConn := kc.redisPool.Get()
	defer redisConn.Close()

	username := msg.GetUserName() //用户自己的账号
	// token := msg.GetJwtToken()
	deviceID := msg.GetDeviceID()

	kc.logger.Info("HandleAddTeamManagers start...",
		zap.String("username", username),
		zap.String("deviceId", deviceID))

	//取出当前设备的os， clientType， logonAt
	curDeviceHashKey := fmt.Sprintf("devices:%s:%s", username, deviceID)
	isMaster, _ := redis.Bool(redisConn.Do("HGET", curDeviceHashKey, "ismaster"))
	curOs, _ := redis.String(redisConn.Do("HGET", curDeviceHashKey, "os"))
	curClientType, _ := redis.Int(redisConn.Do("HGET", curDeviceHashKey, "clientType"))
	curLogonAt, _ := redis.Uint64(redisConn.Do("HGET", curDeviceHashKey, "logonAt"))

	kc.logger.Debug("AddTeamManagers ",
		zap.Bool("isMaster", isMaster),
		zap.String("username", username),
		zap.String("deviceID", deviceID),
		zap.String("curOs", curOs),
		zap.Int("curClientType", curClientType),
		zap.Uint64("curLogonAt", curLogonAt))

	//打开msg里的负载， 获取请求参数
	body := msg.GetContent()

	//解包body
	req := &Team.AddTeamManagersReq{}
	if err := proto.Unmarshal(body, req); err != nil {
		kc.logger.Error("Protobuf Unmarshal Error", zap.Error(err))
		errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
		errorMsg = fmt.Sprintf("Protobuf Unmarshal Error: %s", err.Error())
		goto COMPLETE

	} else {
		kc.logger.Debug("AddTeamManagers  payload",
			zap.String("teamId", req.GetTeamId()),
			zap.Strings("usernames", req.GetUsernames()),
		)

		teamID := req.GetTeamId()

		//判断 teamID 是否存在
		if isExists, err := redis.Bool(redisConn.Do("EXISTS", fmt.Sprintf("TeamInfo:%s", teamID))); err != nil {
			kc.logger.Error("EXISTS Error", zap.Error(err))
			errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
			errorMsg = fmt.Sprintf("Query team info error[teamID=%s]", teamID)
			goto COMPLETE

		} else {
			if !isExists {
				errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
				errorMsg = fmt.Sprintf("Team is not exists[teamID=%s]", teamID)
				goto COMPLETE
			}

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

			//判断操作者是不是群主
			if username != teamInfo.Owner {
				kc.logger.Warn("User is not team owner", zap.String("Username", username))
				errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
				errorMsg = fmt.Sprintf("User is not team owner[Username=%s]", username)
				goto COMPLETE
			}

			for _, manager := range req.GetUsernames() {
				//首先判断一下是否是群成员
				if reply, err := redisConn.Do("ZRANK", fmt.Sprintf("TeamUsers:%s", teamID), manager); err == nil {
					if reply != nil { //是群成员
						//判断是否封号，是否存在
						if state, err := redis.Int(redisConn.Do("HGET", fmt.Sprintf("userData:%s", manager), "State")); err != nil {
							kc.logger.Error("redisConn HGET Error", zap.Error(err))
							//增加到放弃列表
							rsp.AbortedUsernames = append(rsp.AbortedUsernames, manager)
							continue
						} else {
							if state == common.UserBlocked {
								kc.logger.Debug("User is blocked", zap.String("Username", manager))
								//增加到放弃列表
								rsp.AbortedUsernames = append(rsp.AbortedUsernames, manager)
								continue
							}
						}

						managerUser := new(models.TeamUser)
						if result, err := redis.Values(redisConn.Do("HGETALL", fmt.Sprintf("TeamUser:%s:%s", teamID, manager))); err == nil {
							if err := redis.ScanStruct(result, manager); err != nil {
								errorMsg = fmt.Sprintf("TeamUser is not exists[teamID=%s, teamUser=%s]", teamID, manager)
								kc.logger.Error("TeamUser is not exist", zap.Error(err))

								//增加到放弃列表
								rsp.AbortedUsernames = append(rsp.AbortedUsernames, manager)
								continue
							}
						}
						teamMemberType := Team.TeamMemberType(managerUser.TeamMemberType)

						if teamMemberType == Team.TeamMemberType_Tmt_Owner || teamMemberType == Team.TeamMemberType_Tmt_Manager {
							//管理员或群主
							kc.logger.Error("已经是管理员或群主", zap.Error(err))

							//增加到放弃列表
							rsp.AbortedUsernames = append(rsp.AbortedUsernames, manager)
							continue
						} else {
							//将用户设置为管理员
							managerUser.TeamMemberType = 2 //管理员

							if err := kc.SaveTeamUser(managerUser); err != nil {
								kc.logger.Error("SaveTeamUser error", zap.Error(err))

								//增加到放弃列表
								rsp.AbortedUsernames = append(rsp.AbortedUsernames, manager)
								continue
							}

							//刷新redis
							if _, err = redisConn.Do("HMSET", redis.Args{}.Add(fmt.Sprintf("TeamInfo:%s", teamInfo.TeamID)).AddFlat(teamInfo)...); err != nil {
								kc.logger.Error("错误：HMSET TeamInfo", zap.Error(err))
							}

							//向所有群成员推送
							teamMembers, _ := redis.Strings(redisConn.Do("ZRANGEBYSCORE", fmt.Sprintf("TeamUsers:%s", teamID), "-inf", "+inf"))
							curAt := time.Now().UnixNano() / 1e6
							var newSeq uint64

							for _, teamMember := range teamMembers {

								if newSeq, err = redis.Uint64(redisConn.Do("INCR", fmt.Sprintf("userSeq:%s", teamMember))); err != nil {
									kc.logger.Error("redisConn INCR userSeq Error", zap.Error(err))
									continue
								}

								//向群成员发出通知
								body := Msg.MessageNotificationBody{
									Type:           Msg.MessageNotificationType_MNT_GrantManager, //设置管理员
									HandledAccount: username,
									HandledMsg:     fmt.Sprintf("用户 %s 被群主设为管理员", managerUser.Nick),
									Status:         Msg.MessageStatus_MOS_Done,
									Data:           []byte(""),
									To:             teamID, //群id
								}
								bodyData, _ := proto.Marshal(&body)
								mrsp := &Msg.RecvMsgEventRsp{
									Scene:        Msg.MessageScene_MsgScene_S2C,        //系统消息
									Type:         Msg.MessageType_MsgType_Notification, //通知类型
									Body:         bodyData,
									From:         username,
									FromDeviceId: deviceID,
									ServerMsgId:  msg.GetID(),                        //服务器分配的消息ID
									Seq:          newSeq,                             //消息序号，单个会话内自然递增
									Uuid:         fmt.Sprintf("%d", msg.GetTaskID()), //客户端分配的消息ID，SDK生成的消息id，这里返回TaskID
									Time:         uint64(curAt),
								}
								go kc.BroadcastMsgToAllDevices(mrsp, teamMember)

							}

						}

					} else {
						//增加到放弃列表
						rsp.AbortedUsernames = append(rsp.AbortedUsernames, manager)
					}
				}
			}

			//回包
			data, _ = proto.Marshal(rsp)

		}
	}

COMPLETE:
	msg.SetCode(int32(errorCode)) //状态码
	if errorCode == 200 {
		data, _ = proto.Marshal(rsp)
		msg.FillBody(data) //网络包的body，承载真正的业务数据
	} else {
		msg.SetErrorMsg([]byte(errorMsg)) //错误提示
		msg.FillBody(nil)
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
4-15 撤销群管理员

群主设置群管理员

权限:
只有群主才能设置或删除

*/
func (kc *KafkaClient) HandleRemoveTeamManagers(msg *models.Message) error {
	var err error
	errorCode := 200
	var errorMsg string
	rsp := &Team.RemoveTeamManagersRsp{
		AbortedUsernames: make([]string, 0),
	}
	var data []byte
	// var newSeq uint64
	// var count int

	redisConn := kc.redisPool.Get()
	defer redisConn.Close()

	username := msg.GetUserName() //用户自己的账号
	// token := msg.GetJwtToken()
	deviceID := msg.GetDeviceID()

	kc.logger.Info("HandleRemoveTeamManagers start...",
		zap.String("username", username),
		zap.String("deviceId", deviceID))

	//取出当前设备的os， clientType， logonAt
	curDeviceHashKey := fmt.Sprintf("devices:%s:%s", username, deviceID)
	isMaster, _ := redis.Bool(redisConn.Do("HGET", curDeviceHashKey, "ismaster"))
	curOs, _ := redis.String(redisConn.Do("HGET", curDeviceHashKey, "os"))
	curClientType, _ := redis.Int(redisConn.Do("HGET", curDeviceHashKey, "clientType"))
	curLogonAt, _ := redis.Uint64(redisConn.Do("HGET", curDeviceHashKey, "logonAt"))

	kc.logger.Debug("RemoveTeamManagers",
		zap.Bool("isMaster", isMaster),
		zap.String("username", username),
		zap.String("deviceID", deviceID),
		zap.String("curOs", curOs),
		zap.Int("curClientType", curClientType),
		zap.Uint64("curLogonAt", curLogonAt))

	//打开msg里的负载， 获取请求参数
	body := msg.GetContent()

	//解包body
	req := &Team.RemoveTeamManagersReq{}
	if err := proto.Unmarshal(body, req); err != nil {
		kc.logger.Error("Protobuf Unmarshal Error", zap.Error(err))
		errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
		errorMsg = fmt.Sprintf("Protobuf Unmarshal Error: %s", err.Error())
		goto COMPLETE

	} else {
		kc.logger.Debug("RemoveTeamManagers payload",
			zap.String("teamId", req.GetTeamId()),
			zap.Strings("usernames", req.GetUsernames()),
		)

		teamID := req.GetTeamId()

		//判断 teamID 是否存在
		if isExists, err := redis.Bool(redisConn.Do("EXISTS", fmt.Sprintf("TeamInfo:%s", teamID))); err != nil {
			kc.logger.Error("EXISTS Error", zap.Error(err))
			errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
			errorMsg = fmt.Sprintf("Query team info error[teamID=%s]", teamID)
			goto COMPLETE

		} else {
			if !isExists {
				errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
				errorMsg = fmt.Sprintf("Team is not exists[teamID=%s]", teamID)
				goto COMPLETE
			}

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

			//判断操作者是不是群主
			if username != teamInfo.Owner {
				kc.logger.Warn("User is not team owner", zap.String("Username", username))
				errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
				errorMsg = fmt.Sprintf("User is not team owner[Username=%s]", username)
				goto COMPLETE
			}
			for _, manager := range req.GetUsernames() {
				//首先判断一下是否是群成员
				if reply, err := redisConn.Do("ZRANK", fmt.Sprintf("TeamUsers:%s", teamID), manager); err == nil {
					if reply != nil { //是群成员
						//判断是否封号，是否存在
						if state, err := redis.Int(redisConn.Do("HGET", fmt.Sprintf("userData:%s", manager), "State")); err != nil {
							kc.logger.Error("redisConn HGET Error", zap.Error(err))
							//增加到放弃列表
							rsp.AbortedUsernames = append(rsp.AbortedUsernames, manager)
							continue
						} else {
							if state == common.UserBlocked {
								kc.logger.Debug("User is blocked", zap.String("Username", manager))
								//增加到放弃列表
								rsp.AbortedUsernames = append(rsp.AbortedUsernames, manager)
								continue
							}
						}

						managerUser := new(models.TeamUser)
						if result, err := redis.Values(redisConn.Do("HGETALL", fmt.Sprintf("TeamUser:%s:%s", teamID, manager))); err == nil {
							if err := redis.ScanStruct(result, manager); err != nil {
								errorMsg = fmt.Sprintf("TeamUser is not exists[teamID=%s, teamUser=%s]", teamID, manager)
								kc.logger.Error("TeamUser is not exist", zap.Error(err))

								//增加到放弃列表
								rsp.AbortedUsernames = append(rsp.AbortedUsernames, manager)
								continue
							}
						}
						teamMemberType := Team.TeamMemberType(managerUser.TeamMemberType)

						if teamMemberType == Team.TeamMemberType_Tmt_Owner || teamMemberType == Team.TeamMemberType_Tmt_Manager {
							//管理员或群主
							kc.logger.Error("已经是管理员或群主", zap.Error(err))

							//增加到放弃列表
							rsp.AbortedUsernames = append(rsp.AbortedUsernames, manager)
							continue

						} else {

							//撤销管理员
							managerUser.TeamMemberType = 3 //普通成员

							if err := kc.SaveTeamUser(managerUser); err != nil {
								kc.logger.Error("SaveTeamUser error", zap.Error(err))

								//增加到放弃列表
								rsp.AbortedUsernames = append(rsp.AbortedUsernames, manager)
								continue
							}

							//刷新redis
							if _, err = redisConn.Do("HMSET", redis.Args{}.Add(fmt.Sprintf("TeamInfo:%s", teamInfo.TeamID)).AddFlat(teamInfo)...); err != nil {
								kc.logger.Error("错误：HMSET TeamInfo", zap.Error(err))
							}

							//向所有群成员推送
							teamMembers, _ := redis.Strings(redisConn.Do("ZRANGEBYSCORE", fmt.Sprintf("TeamUsers:%s", teamID), "-inf", "+inf"))
							curAt := time.Now().UnixNano() / 1e6
							var newSeq uint64

							for _, teamMember := range teamMembers {

								if newSeq, err = redis.Uint64(redisConn.Do("INCR", fmt.Sprintf("userSeq:%s", teamMember))); err != nil {
									kc.logger.Error("redisConn INCR userSeq Error", zap.Error(err))
									continue
								}

								//向群成员发出通知
								body := Msg.MessageNotificationBody{
									Type:           Msg.MessageNotificationType_MNT_CancelManager, //取消管理员
									HandledAccount: username,
									HandledMsg:     fmt.Sprintf("用户 %s 被群主撤销管理员", managerUser.Nick),
									Status:         Msg.MessageStatus_MOS_Done,
									Data:           []byte(""),
									To:             teamID, //群id
								}
								bodyData, _ := proto.Marshal(&body)
								mrsp := &Msg.RecvMsgEventRsp{
									Scene:        Msg.MessageScene_MsgScene_S2C,        //系统消息
									Type:         Msg.MessageType_MsgType_Notification, //通知类型
									Body:         bodyData,
									From:         username,
									FromDeviceId: deviceID,
									ServerMsgId:  msg.GetID(),                        //服务器分配的消息ID
									Seq:          newSeq,                             //消息序号，单个会话内自然递增
									Uuid:         fmt.Sprintf("%d", msg.GetTaskID()), //客户端分配的消息ID，SDK生成的消息id，这里返回TaskID
									Time:         uint64(curAt),
								}
								go kc.BroadcastMsgToAllDevices(mrsp, teamMember)

							}

						}

					} else {
						//增加到放弃列表
						rsp.AbortedUsernames = append(rsp.AbortedUsernames, manager)
					}
				}
			}

			//回包
			data, _ = proto.Marshal(rsp)

		}
	}

COMPLETE:
	msg.SetCode(int32(errorCode)) //状态码
	if errorCode == 200 {
		data, _ = proto.Marshal(rsp)
		msg.FillBody(data) //网络包的body，承载真正的业务数据
	} else {
		msg.SetErrorMsg([]byte(errorMsg)) //错误提示
		msg.FillBody(nil)
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
4-18 设置群禁言模式

群主/管理修改群组发言模式, 全员禁言只能由群主设置

*/
func (kc *KafkaClient) HandleMuteTeam(msg *models.Message) error {
	var err error
	errorCode := 200
	var errorMsg string

	redisConn := kc.redisPool.Get()
	defer redisConn.Close()

	username := msg.GetUserName() //用户自己的账号
	// token := msg.GetJwtToken()
	deviceID := msg.GetDeviceID()

	kc.logger.Info("HandleMuteTeam start...",
		zap.String("username", username),
		zap.String("deviceId", deviceID))

	//取出当前设备的os， clientType， logonAt
	curDeviceHashKey := fmt.Sprintf("devices:%s:%s", username, deviceID)
	isMaster, _ := redis.Bool(redisConn.Do("HGET", curDeviceHashKey, "ismaster"))
	curOs, _ := redis.String(redisConn.Do("HGET", curDeviceHashKey, "os"))
	curClientType, _ := redis.Int(redisConn.Do("HGET", curDeviceHashKey, "clientType"))
	curLogonAt, _ := redis.Uint64(redisConn.Do("HGET", curDeviceHashKey, "logonAt"))

	kc.logger.Debug("MuteTeam ",
		zap.Bool("isMaster", isMaster),
		zap.String("username", username),
		zap.String("deviceID", deviceID),
		zap.String("curOs", curOs),
		zap.Int("curClientType", curClientType),
		zap.Uint64("curLogonAt", curLogonAt))

	//打开msg里的负载， 获取请求参数
	body := msg.GetContent()

	//解包body
	req := &Team.MuteTeamReq{}
	if err := proto.Unmarshal(body, req); err != nil {
		kc.logger.Error("Protobuf Unmarshal Error", zap.Error(err))
		errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
		errorMsg = fmt.Sprintf("Protobuf Unmarshal Error: %s", err.Error())
		goto COMPLETE

	} else {
		kc.logger.Debug("MuteTeam  payload",
			zap.String("teamId", req.GetTeamId()),
		)

		teamID := req.GetTeamId()
		mute := req.GetMute()

		//判断 teamID 是否存在
		if isExists, err := redis.Bool(redisConn.Do("EXISTS", fmt.Sprintf("TeamInfo:%s", teamID))); err != nil {
			kc.logger.Error("EXISTS Error", zap.Error(err))
			errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
			errorMsg = fmt.Sprintf("Query team info error[teamID=%s]", teamID)
			goto COMPLETE

		} else {
			if !isExists {
				errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
				errorMsg = fmt.Sprintf("Team is not exists[teamID=%s]", teamID)
				goto COMPLETE
			}

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

			key = fmt.Sprintf("TeamUser:%s:%s", teamID, username)
			teamUser := new(models.TeamUser)
			if result, err := redis.Values(redisConn.Do("HGETALL", key)); err == nil {
				if err := redis.ScanStruct(result, teamUser); err != nil {
					kc.logger.Error("错误：ScanStruct", zap.Error(err))
					errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
					errorMsg = fmt.Sprintf("Team user is not exists[username=%s]", username)
					goto COMPLETE
				}
			}

			//判断操作者是群主还是管理员
			teamMemberType := Team.TeamMemberType(teamUser.TeamMemberType)
			if teamMemberType == Team.TeamMemberType_Tmt_Owner {
				//群主可以自由设置
				teamInfo.MuteType = int(mute)
			} else if teamMemberType == Team.TeamMemberType_Tmt_Manager {
				if mute != 2 { // MuteALL
					teamInfo.MuteType = int(mute)
				} else {
					kc.logger.Warn("管理员无权设置全体禁言")
					errorCode = http.StatusBadRequest //错误码， 200是正常，其它是错误
					errorMsg = fmt.Sprintf("管理员无权设置全体禁言[username=%s]", username)
					goto COMPLETE
				}
			} else {
				//其它成员无权设置
				kc.logger.Warn("其它成员无权设置禁言")
				errorCode = http.StatusBadRequest //错误码， 200是正常，其它是错误
				errorMsg = fmt.Sprintf("其它成员无权设置禁言[username=%s]", username)
				goto COMPLETE
			}

			//写入MySQL
			if err = kc.SaveTeam(teamInfo); err != nil {
				kc.logger.Error("Save Team Error", zap.Error(err))
				errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
				errorMsg = "无法保存到Team"
				goto COMPLETE
			}

			//刷新redis
			if _, err = redisConn.Do("HMSET", redis.Args{}.Add(fmt.Sprintf("TeamInfo:%s", teamInfo.TeamID)).AddFlat(teamInfo)...); err != nil {
				kc.logger.Error("错误：HMSET TeamInfo", zap.Error(err))
			}

			//向所有群成员推送
			teamMembers, _ := redis.Strings(redisConn.Do("ZRANGEBYSCORE", fmt.Sprintf("TeamUsers:%s", teamID), "-inf", "+inf"))
			curAt := time.Now().UnixNano() / 1e6
			var newSeq uint64

			for _, teamMember := range teamMembers {

				if newSeq, err = redis.Uint64(redisConn.Do("INCR", fmt.Sprintf("userSeq:%s", teamMember))); err != nil {
					kc.logger.Error("redisConn INCR userSeq Error", zap.Error(err))
					continue
				}

				//向群成员发出通知
				body := Msg.MessageNotificationBody{
					Type:           Msg.MessageNotificationType_MNT_MuteTeam, //设置群组禁言模式
					HandledAccount: username,
					HandledMsg:     "设置群组禁言模式",
					Status:         Msg.MessageStatus_MOS_Done,
					Data:           []byte(""),
					To:             teamID, //群id
				}
				bodyData, _ := proto.Marshal(&body)
				mrsp := &Msg.RecvMsgEventRsp{
					Scene:        Msg.MessageScene_MsgScene_S2C,        //系统消息
					Type:         Msg.MessageType_MsgType_Notification, //通知类型
					Body:         bodyData,
					From:         username,
					FromDeviceId: deviceID,
					ServerMsgId:  msg.GetID(),                        //服务器分配的消息ID
					Seq:          newSeq,                             //消息序号，单个会话内自然递增
					Uuid:         fmt.Sprintf("%d", msg.GetTaskID()), //客户端分配的消息ID，SDK生成的消息id，这里返回TaskID
					Time:         uint64(curAt),
				}
				go kc.BroadcastMsgToAllDevices(mrsp, teamMember)

			}
		}
	}

COMPLETE:
	msg.SetCode(int32(errorCode)) //状态码
	if errorCode == 200 {
		//200
	} else {
		msg.SetErrorMsg([]byte(errorMsg)) //错误提示
		msg.FillBody(nil)
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
4-19 设置群成员禁言

群主/管理修改某个群成员发言模式
可以设置禁言时间，如果不设置mutedays，则永久禁言
*/

func (kc *KafkaClient) HandleMuteTeamMember(msg *models.Message) error {
	var err error
	errorCode := 200
	var errorMsg string

	var newSeq uint64

	redisConn := kc.redisPool.Get()
	defer redisConn.Close()

	username := msg.GetUserName() //用户自己的账号
	// token := msg.GetJwtToken()
	deviceID := msg.GetDeviceID()

	kc.logger.Info("HandleMuteTeamMember start...",
		zap.String("username", username),
		zap.String("deviceId", deviceID))

	//取出当前设备的os， clientType， logonAt
	curDeviceHashKey := fmt.Sprintf("devices:%s:%s", username, deviceID)
	isMaster, _ := redis.Bool(redisConn.Do("HGET", curDeviceHashKey, "ismaster"))
	curOs, _ := redis.String(redisConn.Do("HGET", curDeviceHashKey, "os"))
	curClientType, _ := redis.Int(redisConn.Do("HGET", curDeviceHashKey, "clientType"))
	curLogonAt, _ := redis.Uint64(redisConn.Do("HGET", curDeviceHashKey, "logonAt"))

	kc.logger.Debug("MuteTeamMember",
		zap.Bool("isMaster", isMaster),
		zap.String("username", username),
		zap.String("deviceID", deviceID),
		zap.String("curOs", curOs),
		zap.Int("curClientType", curClientType),
		zap.Uint64("curLogonAt", curLogonAt))

	//打开msg里的负载， 获取请求参数
	body := msg.GetContent()

	//解包body
	req := &Team.MuteTeamMemberReq{}
	if err := proto.Unmarshal(body, req); err != nil {
		kc.logger.Error("Protobuf Unmarshal Error", zap.Error(err))
		errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
		errorMsg = fmt.Sprintf("Protobuf Unmarshal Error: %s", err.Error())
		goto COMPLETE

	} else {
		kc.logger.Debug("MuteTeamMember payload",
			zap.String("teamId", req.GetTeamId()),       //所在群组
			zap.String("username", req.GetUsername()),   //被禁言的群成员
			zap.Bool("Mute", req.GetMute()),             //是否禁言,false/true
			zap.Int("Mutedays", int(req.GetMutedays())), //禁言天数，如：禁言3天
		)

		teamID := req.GetTeamId()
		// isMute := req.GetMute()
		// mutedays := req.GetMutedays()

		psSource := &Friends.PsSource{
			Ps:     "",
			Source: req.GetUsername(), //被禁言的群成员
		}
		psSourceData, _ := proto.Marshal(psSource)

		//判断 teamID 是否存在
		if isExists, err := redis.Bool(redisConn.Do("EXISTS", fmt.Sprintf("TeamInfo:%s", teamID))); err != nil {
			kc.logger.Error("EXISTS Error", zap.Error(err))
			errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
			errorMsg = fmt.Sprintf("Query team info error[teamID=%s]", teamID)
			goto COMPLETE

		} else {
			if !isExists {
				errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
				errorMsg = fmt.Sprintf("Team is not exists[teamID=%s]", teamID)
				goto COMPLETE
			}

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

			key = fmt.Sprintf("TeamUser:%s:%s", teamID, req.GetUsername())
			teamUser := new(models.TeamUser)
			if result, err := redis.Values(redisConn.Do("HGETALL", key)); err == nil {
				if err := redis.ScanStruct(result, teamUser); err != nil {
					kc.logger.Error("错误：ScanStruct", zap.Error(err))
					errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
					errorMsg = fmt.Sprintf("Team user is not exists[username=%s]", req.GetUsername())
					goto COMPLETE
				}
			}

			//判断操作者是群主还是管理员
			teamMemberType := Team.TeamMemberType(teamUser.TeamMemberType)
			if teamMemberType == Team.TeamMemberType_Tmt_Owner || teamMemberType == Team.TeamMemberType_Tmt_Manager {
				teamUser.IsMute = req.GetMute()
				teamUser.Mutedays = int(req.GetMutedays())
			} else {
				//其它成员无权设置
				kc.logger.Warn("其它成员无权设置禁言时长")
				errorCode = http.StatusBadRequest //错误码， 200是正常，其它是错误
				errorMsg = fmt.Sprintf("其它成员无权设置禁言时长[username=%s]", req.GetUsername())
				goto COMPLETE
			}

			//写入MySQL
			if err = kc.SaveTeamUser(teamUser); err != nil {
				kc.logger.Error("Save teamUser Error", zap.Error(err))
				errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
				errorMsg = "无法保存到teamUser"
				goto COMPLETE
			}

			//刷新redis
			if _, err = redisConn.Do("HMSET", redis.Args{}.Add(fmt.Sprintf("TeamUser:%s:%s", teamInfo.TeamID, req.GetUsername())).AddFlat(teamUser)...); err != nil {
				kc.logger.Error("错误：HMSET teamUser", zap.Error(err))
			}
			//向redis里的定时解禁任务列表DissMuteUsers:{群id}增加此用户， 由系统定时器cron将此用户到期解禁
			if req.GetMute() {
				if req.GetMutedays() > 0 {
					// handledMsg = fmt.Sprintf("群成员:%s被禁言%d天", teamUser.Nick, req.GetMutedays())
					now := time.Now()

					dd, _ := time.ParseDuration(fmt.Sprintf("%dh", 24*req.GetMutedays())) //什么时间解禁
					dd1 := now.Add(dd)
					//定时任务取出到时解禁的用户
					redisConn.Do("ZADD", fmt.Sprintf("DissMuteUsers:%s", teamInfo.TeamID), dd1.UnixNano()/1e6, req.GetUsername())
				}
			}

			//向群成员推送此用户被禁言
			teamMembers, _ := redis.Strings(redisConn.Do("ZRANGEBYSCORE", fmt.Sprintf("TeamUsers:%s", teamInfo.TeamID), "-inf", "+inf"))
			for _, teamMember := range teamMembers {

				if newSeq, err = redis.Uint64(redisConn.Do("INCR", fmt.Sprintf("userSeq:%s", teamMember))); err != nil {
					kc.logger.Error("redisConn INCR userSeq Error", zap.Error(err))
					errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
					errorMsg = fmt.Sprintf("INCR error[Username=%s]", teamMember)
					goto COMPLETE
				}

				var handledMsg string
				if req.GetMute() {
					if req.GetMutedays() > 0 {
						handledMsg = fmt.Sprintf("群成员:%s被禁言%d天", teamUser.Nick, req.GetMutedays())
					} else {
						handledMsg = fmt.Sprintf("群成员:%s被永久禁言", teamUser.Nick)
					}
				} else { //解禁
					handledMsg = fmt.Sprintf("群成员:%s解除禁言", teamUser.Nick)
				}

				body := Msg.MessageNotificationBody{
					Type:           Msg.MessageNotificationType_MNT_MuteTeamMember, //群成员禁言/解禁
					HandledAccount: username,                                       //当前用户
					HandledMsg:     handledMsg,
					Status:         Msg.MessageStatus_MOS_Done,
					Data:           psSourceData,
					To:             teamID, //群组id
				}
				bodyData, _ := proto.Marshal(&body)
				eRsp := &Msg.RecvMsgEventRsp{
					Scene:        Msg.MessageScene_MsgScene_S2C,        //系统消息
					Type:         Msg.MessageType_MsgType_Notification, //通知类型
					Body:         bodyData,
					From:         username,
					FromDeviceId: deviceID,
					ServerMsgId:  msg.GetID(),                        //服务器分配的消息ID
					Seq:          newSeq,                             //消息序号，单个会话内自然递增, 这里是对teamMembere这个用户的通知序号
					Uuid:         fmt.Sprintf("%d", msg.GetTaskID()), //客户端分配的消息ID，SDK生成的消息id，这里返回TaskID
					Time:         uint64(time.Now().UnixNano() / 1e6),
				}
				go kc.BroadcastMsgToAllDevices(eRsp, teamMember) //向群成员广播
			}
		}
	}

COMPLETE:
	msg.SetCode(int32(errorCode)) //状态码
	if errorCode == 200 {
		//200
	} else {
		msg.SetErrorMsg([]byte(errorMsg)) //错误提示
		msg.FillBody(nil)
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
4-20 用户设置群消息通知方式
群成员设置接收群消息的通知方式
*/

func (kc *KafkaClient) HandleSetNotifyType(msg *models.Message) error {
	var err error
	errorCode := 200
	var errorMsg string

	redisConn := kc.redisPool.Get()
	defer redisConn.Close()

	username := msg.GetUserName() //用户自己的账号
	// token := msg.GetJwtToken()
	deviceID := msg.GetDeviceID()

	kc.logger.Info("HandleSetNotifyType start...",
		zap.String("username", username),
		zap.String("deviceId", deviceID))

	//取出当前设备的os， clientType， logonAt
	curDeviceHashKey := fmt.Sprintf("devices:%s:%s", username, deviceID)
	isMaster, _ := redis.Bool(redisConn.Do("HGET", curDeviceHashKey, "ismaster"))
	curOs, _ := redis.String(redisConn.Do("HGET", curDeviceHashKey, "os"))
	curClientType, _ := redis.Int(redisConn.Do("HGET", curDeviceHashKey, "clientType"))
	curLogonAt, _ := redis.Uint64(redisConn.Do("HGET", curDeviceHashKey, "logonAt"))

	kc.logger.Debug("SetNotifyType ",
		zap.Bool("isMaster", isMaster),
		zap.String("username", username),
		zap.String("deviceID", deviceID),
		zap.String("curOs", curOs),
		zap.Int("curClientType", curClientType),
		zap.Uint64("curLogonAt", curLogonAt))

	//打开msg里的负载， 获取请求参数
	body := msg.GetContent()

	//解包body
	req := &Team.SetNotifyTypeReq{}
	if err := proto.Unmarshal(body, req); err != nil {
		kc.logger.Error("Protobuf Unmarshal Error", zap.Error(err))
		errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
		errorMsg = fmt.Sprintf("Protobuf Unmarshal Error: %s", err.Error())
		goto COMPLETE

	} else {
		kc.logger.Debug("SetNotifyType payload",
			zap.String("teamId", req.GetTeamId()),
			zap.Int("notifyType", int(req.GetNotifyType())),
		)

		teamID := req.GetTeamId()
		// notifyType := req.GetNotifyType()

		//判断 teamID 是否存在
		if isExists, err := redis.Bool(redisConn.Do("EXISTS", fmt.Sprintf("TeamInfo:%s", teamID))); err != nil {
			kc.logger.Error("EXISTS Error", zap.Error(err))
			errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
			errorMsg = fmt.Sprintf("Query team info error[teamID=%s]", teamID)
			goto COMPLETE

		} else {
			if !isExists {
				errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
				errorMsg = fmt.Sprintf("Team is not exists[teamID=%s]", teamID)
				goto COMPLETE
			}

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

			key = fmt.Sprintf("TeamUser:%s:%s", teamID, username)
			teamUser := new(models.TeamUser)
			if result, err := redis.Values(redisConn.Do("HGETALL", key)); err == nil {
				if err := redis.ScanStruct(result, teamUser); err != nil {
					kc.logger.Error("错误：ScanStruct", zap.Error(err))
					errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
					errorMsg = fmt.Sprintf("Team user is not exists[username=%s]", username)
					goto COMPLETE
				}
			}

			teamUser.NotifyType = int(req.GetNotifyType())
			//写入MySQL
			if err = kc.SaveTeamUser(teamUser); err != nil {
				kc.logger.Error("Save teamUser Error", zap.Error(err))
				errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
				errorMsg = "无法保存到teamUser"
				goto COMPLETE
			}

			//刷新redis
			if _, err = redisConn.Do("HMSET", redis.Args{}.Add(fmt.Sprintf("TeamUser:%s:%s", teamInfo.TeamID, username)).AddFlat(teamUser)...); err != nil {
				kc.logger.Error("错误：HMSET teamUser", zap.Error(err))
			}

			//TODO 向用户的其它设备
		}
	}

COMPLETE:
	msg.SetCode(int32(errorCode)) //状态码
	if errorCode == 200 {
		//200
	} else {
		msg.SetErrorMsg([]byte(errorMsg)) //错误提示
		msg.FillBody(nil)
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
4-21 用户设置其在群里的资料
群成员设置自己群里的个人资料
*/

func (kc *KafkaClient) HandleUpdateMyInfo(msg *models.Message) error {
	var err error
	errorCode := 200
	var errorMsg string

	redisConn := kc.redisPool.Get()
	defer redisConn.Close()

	username := msg.GetUserName() //用户自己的账号
	// token := msg.GetJwtToken()
	deviceID := msg.GetDeviceID()

	kc.logger.Info("HandleUpdateMyInfo start...",
		zap.String("username", username),
		zap.String("deviceId", deviceID))

	//取出当前设备的os， clientType， logonAt
	curDeviceHashKey := fmt.Sprintf("devices:%s:%s", username, deviceID)
	isMaster, _ := redis.Bool(redisConn.Do("HGET", curDeviceHashKey, "ismaster"))
	curOs, _ := redis.String(redisConn.Do("HGET", curDeviceHashKey, "os"))
	curClientType, _ := redis.Int(redisConn.Do("HGET", curDeviceHashKey, "clientType"))
	curLogonAt, _ := redis.Uint64(redisConn.Do("HGET", curDeviceHashKey, "logonAt"))

	kc.logger.Debug("UpdateMyInfo ",
		zap.Bool("isMaster", isMaster),
		zap.String("username", username),
		zap.String("deviceID", deviceID),
		zap.String("curOs", curOs),
		zap.Int("curClientType", curClientType),
		zap.Uint64("curLogonAt", curLogonAt))

	//打开msg里的负载， 获取请求参数
	body := msg.GetContent()

	//解包body
	req := &Team.UpdateMyInfoReq{}
	if err := proto.Unmarshal(body, req); err != nil {
		kc.logger.Error("Protobuf Unmarshal Error", zap.Error(err))
		errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
		errorMsg = fmt.Sprintf("Protobuf Unmarshal Error: %s", err.Error())
		goto COMPLETE

	} else {
		kc.logger.Debug("UpdateMyInfo  payload",
			zap.String("teamId", req.GetTeamId()),
		)

		teamID := req.GetTeamId()

		//判断 teamID 是否存在
		if isExists, err := redis.Bool(redisConn.Do("EXISTS", fmt.Sprintf("TeamInfo:%s", teamID))); err != nil {
			kc.logger.Error("EXISTS Error", zap.Error(err))
			errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
			errorMsg = fmt.Sprintf("Query team info error[teamID=%s]", teamID)
			goto COMPLETE

		} else {
			if !isExists {
				errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
				errorMsg = fmt.Sprintf("Team is not exists[teamID=%s]", teamID)
				goto COMPLETE
			}

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

			key = fmt.Sprintf("TeamUser:%s:%s", teamID, username)
			teamUser := new(models.TeamUser)
			if result, err := redis.Values(redisConn.Do("HGETALL", key)); err == nil {
				if err := redis.ScanStruct(result, teamUser); err != nil {
					kc.logger.Error("错误：ScanStruct", zap.Error(err))
					errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
					errorMsg = fmt.Sprintf("Team user is not exists[username=%s]", username)
					goto COMPLETE
				}
			}

			if nick, ok := req.Fields[1]; ok {
				//修改群组呢称
				teamUser.Nick = nick

			}
			if ex, ok := req.Fields[2]; ok {
				//修改群组呢称
				teamUser.Extend = ex

			}

			//写入MySQL
			if err = kc.SaveTeamUser(teamUser); err != nil {
				kc.logger.Error("Save teamUser Error", zap.Error(err))
				errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
				errorMsg = "无法保存到teamUser"
				goto COMPLETE
			}

			//刷新redis
			if _, err = redisConn.Do("HMSET", redis.Args{}.Add(fmt.Sprintf("TeamUser:%s:%s", teamInfo.TeamID, username)).AddFlat(teamUser)...); err != nil {
				kc.logger.Error("错误：HMSET teamUser", zap.Error(err))
			}

			//TODO 向用户的其它设备
		}
	}

COMPLETE:
	msg.SetCode(int32(errorCode)) //状态码
	if errorCode == 200 {
		//200
	} else {
		msg.SetErrorMsg([]byte(errorMsg)) //错误提示
		msg.FillBody(nil)
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
4-22 管理员设置群成员资料

管理员设置群成员资料
*/

func (kc *KafkaClient) HandleUpdateMemberInfo(msg *models.Message) error {
	var err error
	errorCode := 200
	var errorMsg string

	redisConn := kc.redisPool.Get()
	defer redisConn.Close()

	username := msg.GetUserName() //用户自己的账号
	// token := msg.GetJwtToken()
	deviceID := msg.GetDeviceID()

	kc.logger.Info("HandleUpdateMemberInfo start...",
		zap.String("username", username),
		zap.String("deviceId", deviceID))

	//取出当前设备的os， clientType， logonAt
	curDeviceHashKey := fmt.Sprintf("devices:%s:%s", username, deviceID)
	isMaster, _ := redis.Bool(redisConn.Do("HGET", curDeviceHashKey, "ismaster"))
	curOs, _ := redis.String(redisConn.Do("HGET", curDeviceHashKey, "os"))
	curClientType, _ := redis.Int(redisConn.Do("HGET", curDeviceHashKey, "clientType"))
	curLogonAt, _ := redis.Uint64(redisConn.Do("HGET", curDeviceHashKey, "logonAt"))

	kc.logger.Debug("UpdateMemberInfo ",
		zap.Bool("isMaster", isMaster),
		zap.String("username", username),
		zap.String("deviceID", deviceID),
		zap.String("curOs", curOs),
		zap.Int("curClientType", curClientType),
		zap.Uint64("curLogonAt", curLogonAt))

	//打开msg里的负载， 获取请求参数
	body := msg.GetContent()

	//解包body
	req := &Team.UpdateMemberInfoReq{}
	if err := proto.Unmarshal(body, req); err != nil {
		kc.logger.Error("Protobuf Unmarshal Error", zap.Error(err))
		errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
		errorMsg = fmt.Sprintf("Protobuf Unmarshal Error: %s", err.Error())
		goto COMPLETE

	} else {
		kc.logger.Debug("UpdateMemberInfo  payload",
			zap.String("teamId", req.GetTeamId()),
			zap.String("username", req.GetUsername()),
		)

		teamID := req.GetTeamId()

		//判断 teamID 是否存在
		if isExists, err := redis.Bool(redisConn.Do("EXISTS", fmt.Sprintf("TeamInfo:%s", teamID))); err != nil {
			kc.logger.Error("EXISTS Error", zap.Error(err))
			errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
			errorMsg = fmt.Sprintf("Query team info error[teamID=%s]", teamID)
			goto COMPLETE

		} else {
			if !isExists {
				errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
				errorMsg = fmt.Sprintf("Team is not exists[teamID=%s]", teamID)
				goto COMPLETE
			}

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

			key = fmt.Sprintf("TeamUser:%s:%s", teamID, req.GetUsername())
			teamUser := new(models.TeamUser)
			if result, err := redis.Values(redisConn.Do("HGETALL", key)); err == nil {
				if err := redis.ScanStruct(result, teamUser); err != nil {
					kc.logger.Error("错误：ScanStruct", zap.Error(err))
					errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
					errorMsg = fmt.Sprintf("Team user is not exists[username=%s]", req.GetUsername())
					goto COMPLETE
				}
			}

			//判断操作者是群主还是管理员
			teamMemberType := Team.TeamMemberType(teamUser.TeamMemberType)
			if teamMemberType == Team.TeamMemberType_Tmt_Owner || teamMemberType == Team.TeamMemberType_Tmt_Manager {
				// teamUser.IsMute = req.GetMute()
				// teamUser.Mutedays = int(req.GetMutedays())
				if nick, ok := req.Fields[1]; ok {
					//修改群组呢称
					teamUser.Nick = nick

				}
				if ex, ok := req.Fields[2]; ok {
					//修改群组呢称
					teamUser.Extend = ex

				}

			} else {
				//其它成员无权设置
				kc.logger.Warn("其它成员无权设置群成员资料")
				errorCode = http.StatusBadRequest //错误码， 200是正常，其它是错误
				errorMsg = fmt.Sprintf("其它成员无权设置群成员资料[username=%s]", req.GetUsername())
				goto COMPLETE
			}

			//写入MySQL
			if err = kc.SaveTeamUser(teamUser); err != nil {
				kc.logger.Error("Save teamUser Error", zap.Error(err))
				errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
				errorMsg = "无法保存到teamUser"
				goto COMPLETE
			}

			//刷新redis
			if _, err = redisConn.Do("HMSET", redis.Args{}.Add(fmt.Sprintf("TeamUser:%s:%s", teamInfo.TeamID, req.GetUsername())).AddFlat(teamUser)...); err != nil {
				kc.logger.Error("错误：HMSET teamUser", zap.Error(err))
			}

			// 向所有群成员推送
			var newSeq uint64
			teamMembers, _ := redis.Strings(redisConn.Do("ZRANGEBYSCORE", fmt.Sprintf("TeamUsers:%s", teamInfo.TeamID), "-inf", "+inf"))
			for _, teamMember := range teamMembers {

				if newSeq, err = redis.Uint64(redisConn.Do("INCR", fmt.Sprintf("userSeq:%s", teamMember))); err != nil {
					kc.logger.Error("redisConn INCR userSeq Error", zap.Error(err))
					errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
					errorMsg = fmt.Sprintf("INCR error[Username=%s]", teamMember)
					goto COMPLETE
				}

				body := Msg.MessageNotificationBody{
					Type:           Msg.MessageNotificationType_MNT_UpdateTeamMember, //管理员/群主修改群成员信息
					HandledAccount: username,                                         //当前用户
					HandledMsg:     "管理员/群主修改群成员信息",
					Status:         Msg.MessageStatus_MOS_Done,
					Data:           []byte(""),
					To:             teamID, //群组id
				}
				bodyData, _ := proto.Marshal(&body)
				eRsp := &Msg.RecvMsgEventRsp{
					Scene:        Msg.MessageScene_MsgScene_S2C,        //系统消息
					Type:         Msg.MessageType_MsgType_Notification, //通知类型
					Body:         bodyData,
					From:         username,
					FromDeviceId: deviceID,
					ServerMsgId:  msg.GetID(),                        //服务器分配的消息ID
					Seq:          newSeq,                             //消息序号，单个会话内自然递增, 这里是对teamMembere这个用户的通知序号
					Uuid:         fmt.Sprintf("%d", msg.GetTaskID()), //客户端分配的消息ID，SDK生成的消息id，这里返回TaskID
					Time:         uint64(time.Now().UnixNano() / 1e6),
				}
				go kc.BroadcastMsgToAllDevices(eRsp, teamMember) //向群成员广播
			}

		}
	}

COMPLETE:
	msg.SetCode(int32(errorCode)) //状态码
	if errorCode == 200 {
		//200
	} else {
		msg.SetErrorMsg([]byte(errorMsg)) //错误提示
		msg.FillBody(nil)
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
4-24 获取指定群组成员

根据群组用户ID获取最新群成员信息
*/

func (kc *KafkaClient) HandlePullTeamMembers(msg *models.Message) error {
	var err error
	errorCode := 200
	var errorMsg string
	rsp := &Team.PullTeamMembersRsp{
		Tmembers: make([]*Team.Tmember, 0),
	}
	var data []byte

	redisConn := kc.redisPool.Get()
	defer redisConn.Close()

	username := msg.GetUserName() //用户自己的账号
	// token := msg.GetJwtToken()
	deviceID := msg.GetDeviceID()

	kc.logger.Info("HandlePullTeamMembers start...",
		zap.String("username", username),
		zap.String("deviceId", deviceID))

	//取出当前设备的os， clientType， logonAt
	curDeviceHashKey := fmt.Sprintf("devices:%s:%s", username, deviceID)
	isMaster, _ := redis.Bool(redisConn.Do("HGET", curDeviceHashKey, "ismaster"))
	curOs, _ := redis.String(redisConn.Do("HGET", curDeviceHashKey, "os"))
	curClientType, _ := redis.Int(redisConn.Do("HGET", curDeviceHashKey, "clientType"))
	curLogonAt, _ := redis.Uint64(redisConn.Do("HGET", curDeviceHashKey, "logonAt"))

	kc.logger.Debug("PullTeamMembers ",
		zap.Bool("isMaster", isMaster),
		zap.String("username", username),
		zap.String("deviceID", deviceID),
		zap.String("curOs", curOs),
		zap.Int("curClientType", curClientType),
		zap.Uint64("curLogonAt", curLogonAt))

	//打开msg里的负载， 获取请求参数
	body := msg.GetContent()

	//解包body
	req := &Team.PullTeamMembersReq{}
	if err := proto.Unmarshal(body, req); err != nil {
		kc.logger.Error("Protobuf Unmarshal Error", zap.Error(err))
		errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
		errorMsg = fmt.Sprintf("Protobuf Unmarshal Error: %s", err.Error())
		goto COMPLETE

	} else {
		kc.logger.Debug("PullTeamMembers  payload",
			zap.String("teamId", req.GetTeamId()),
			zap.Strings("usernames", req.GetAccounts()),
		)

		teamID := req.GetTeamId()

		//判断 teamID 是否存在
		if isExists, err := redis.Bool(redisConn.Do("EXISTS", fmt.Sprintf("TeamInfo:%s", teamID))); err != nil {
			kc.logger.Error("EXISTS Error", zap.Error(err))
			errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
			errorMsg = fmt.Sprintf("Query team info error[teamID=%s]", teamID)
			goto COMPLETE

		} else {
			if !isExists {
				errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
				errorMsg = fmt.Sprintf("Team is not exists[teamID=%s]", teamID)
				goto COMPLETE
			}

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

			for _, account := range req.GetAccounts() {

				key = fmt.Sprintf("TeamUser:%s:%s", teamID, account)
				teamUser := new(models.TeamUser)
				if result, err := redis.Values(redisConn.Do("HGETALL", key)); err == nil {
					if err := redis.ScanStruct(result, teamUser); err != nil {
						kc.logger.Error("错误：ScanStruct", zap.Error(err))
						errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
						errorMsg = fmt.Sprintf("Team user is not exists[username=%s]", account)
						goto COMPLETE
					}
				}

				rsp.Tmembers = append(rsp.Tmembers, &Team.Tmember{
					TeamId:          teamUser.TeamID,
					Username:        teamUser.Username,
					Invitedusername: teamUser.InvitedUsername,
					Nick:            teamUser.Nick,
					Avatar:          teamUser.Avatar,
					Label:           teamUser.Label,
					Source:          teamUser.Source,
					Type:            Team.TeamMemberType(teamUser.TeamMemberType),
					NotifyType:      Team.NotifyType(teamUser.NotifyType),
					Mute:            teamUser.IsMute,
					Ex:              teamUser.Extend,
					JoinTime:        uint64(teamUser.JoinAt),
					UpdateTime:      uint64(teamUser.UpdatedAt),
				})
			}
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
		kc.logger.Info("Succeed succeed send message to ProduceChannel", zap.String("topic", topic))
	} else {
		kc.logger.Error("Failed to send  message to ProduceChannel", zap.Error(err))
	}
	_ = err
	return nil

}

/*
4-25 增量同步群组信息

增量同步群组信息
*/

func (kc *KafkaClient) HandleGetMyTeams(msg *models.Message) error {
	var err error
	errorCode := 200
	var errorMsg string
	rsp := &Team.GetMyTeamsRsp{
		Teams:        make([]*Team.TeamInfo, 0),
		RemovedTeams: make([]string, 0),
	}
	var data []byte

	redisConn := kc.redisPool.Get()
	defer redisConn.Close()

	username := msg.GetUserName() //用户自己的账号
	// token := msg.GetJwtToken()
	deviceID := msg.GetDeviceID()

	kc.logger.Info("HandleGetMyTeams start...",
		zap.String("username", username),
		zap.String("deviceId", deviceID))

	//取出当前设备的os， clientType， logonAt
	curDeviceHashKey := fmt.Sprintf("devices:%s:%s", username, deviceID)
	isMaster, _ := redis.Bool(redisConn.Do("HGET", curDeviceHashKey, "ismaster"))
	curOs, _ := redis.String(redisConn.Do("HGET", curDeviceHashKey, "os"))
	curClientType, _ := redis.Int(redisConn.Do("HGET", curDeviceHashKey, "clientType"))
	curLogonAt, _ := redis.Uint64(redisConn.Do("HGET", curDeviceHashKey, "logonAt"))

	kc.logger.Debug("GetMyTeams ",
		zap.Bool("isMaster", isMaster),
		zap.String("username", username),
		zap.String("deviceID", deviceID),
		zap.String("curOs", curOs),
		zap.Int("curClientType", curClientType),
		zap.Uint64("curLogonAt", curLogonAt))

	//打开msg里的负载， 获取请求参数
	body := msg.GetContent()

	//解包body
	req := &Team.GetMyTeamsReq{}
	if err := proto.Unmarshal(body, req); err != nil {
		kc.logger.Error("Protobuf Unmarshal Error", zap.Error(err))
		errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
		errorMsg = fmt.Sprintf("Protobuf Unmarshal Error: %s", err.Error())
		goto COMPLETE

	} else {
		kc.logger.Debug("GetMyTeams  payload",
			zap.Uint64("timeAt", req.GetTimeAt()),
		)

		//查出此用户的所有群组
		teamIDs, _ := redis.Strings(redisConn.Do("ZRANGEBYSCORE", fmt.Sprintf("Team:%s", username), "-inf", "+inf"))
		for _, teamID := range teamIDs {
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

			//计算群成员数量。
			var count int
			if count, err = redis.Int(redisConn.Do("ZCARD", fmt.Sprintf("TeamUsers:%s", teamID))); err != nil {
				kc.logger.Error("ZCARD Error", zap.Error(err))
				errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
				errorMsg = fmt.Sprintf("TeamUsers is not exists[teamID=%s]", teamID)
				goto COMPLETE
			}

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
				UpdateAt:     uint64(time.Now().UnixNano() / 1e6), //更新时间
			})
		}
		//用户自己的退群列表
		removeTeamIDs, _ := redis.Strings(redisConn.Do("ZRANGEBYSCORE", fmt.Sprintf("RemoveTeam:%s", username), "-inf", "+inf"))
		for _, removeTeamID := range removeTeamIDs {
			rsp.RemovedTeams = append(rsp.RemovedTeams, removeTeamID)
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
		kc.logger.Info("Succeed succeed send message to ProduceChannel", zap.String("topic", topic))
	} else {
		kc.logger.Error("Failed to send  message to ProduceChannel", zap.Error(err))
	}
	_ = err
	return nil

}

/*
TODO
4-26 管理员审核用户入群申请

管理员收到询问是否同意邀请用户入群的系统通知事件， 处理：同意或拒绝

说明:
1.  管理员根据工作流ID拉人入群进行甄别处理
2.  处理结果需要同时向其它管理员同步处理结果
3.  当同意后，需要将这个拉人入群事件向被邀请的用户发送
4. 当拒绝后，需要向发起拉人入群的用户发送拒绝
*/
func (kc *KafkaClient) HandleCheckTeamInvite(msg *models.Message) error {
	var err error
	errorCode := 200
	var errorMsg string
	var newSeq uint64

	redisConn := kc.redisPool.Get()
	defer redisConn.Close()

	username := msg.GetUserName() //用户自己的账号
	// token := msg.GetJwtToken()
	deviceID := msg.GetDeviceID()

	kc.logger.Info("HandleCheckTeamInvite start...",
		zap.String("username", username),
		zap.String("deviceId", deviceID))

	//取出当前设备的os， clientType， logonAt
	curDeviceHashKey := fmt.Sprintf("devices:%s:%s", username, deviceID)
	isMaster, _ := redis.Bool(redisConn.Do("HGET", curDeviceHashKey, "ismaster"))
	curOs, _ := redis.String(redisConn.Do("HGET", curDeviceHashKey, "os"))
	curClientType, _ := redis.Int(redisConn.Do("HGET", curDeviceHashKey, "clientType"))
	curLogonAt, _ := redis.Uint64(redisConn.Do("HGET", curDeviceHashKey, "logonAt"))

	kc.logger.Debug("CheckTeamInvite",
		zap.Bool("isMaster", isMaster),
		zap.String("username", username),
		zap.String("deviceID", deviceID),
		zap.String("curOs", curOs),
		zap.Int("curClientType", curClientType),
		zap.Uint64("curLogonAt", curLogonAt))

	//打开msg里的负载， 获取请求参数
	body := msg.GetContent()

	//解包body
	req := &Team.CheckTeamInviteReq{}
	if err := proto.Unmarshal(body, req); err != nil {
		kc.logger.Error("Protobuf Unmarshal Error", zap.Error(err))
		errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
		errorMsg = fmt.Sprintf("Protobuf Unmarshal Error: %s", err.Error())
		goto COMPLETE

	} else {
		kc.logger.Debug("CheckTeamInvite payload",
			zap.String("TeamId", req.GetTeamId()),
			zap.String("WorkflowID", req.GetWorkflowID()), //工作流ID
			zap.String("Inviter", req.GetInviter()),       //邀请人
			zap.String("Invitee", req.GetInvitee()),       //被邀请人
			zap.Bool("IsAgree", req.GetIsAgree()),         //是否同意邀请用户入群操作，true-同意，false-不同意
			zap.String("Ps", req.GetPs()),
		)

		teamID := req.GetTeamId()

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

		key = fmt.Sprintf("TeamUser:%s:%s", teamID, username)
		teamUser := new(models.TeamUser)
		if result, err := redis.Values(redisConn.Do("HGETALL", key)); err == nil {
			if err := redis.ScanStruct(result, teamUser); err != nil {
				kc.logger.Error("错误：ScanStruct", zap.Error(err))
				errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
				errorMsg = fmt.Sprintf("Team user is not exists[username=%s]", username)
				goto COMPLETE
			}
		}

		//判断操作者是群主还是管理员
		teamMemberType := Team.TeamMemberType(teamUser.TeamMemberType)
		if teamMemberType == Team.TeamMemberType_Tmt_Owner || teamMemberType == Team.TeamMemberType_Tmt_Manager {

			if req.GetIsAgree() {
				userData := new(models.User)
				userKey := fmt.Sprintf("userData:%s", req.GetInvitee()) //受邀请人
				if result, err := redis.Values(redisConn.Do("HGETALL", userKey)); err == nil {
					if err := redis.ScanStruct(result, userData); err != nil {

						kc.logger.Error("错误：ScanStruct", zap.Error(err))
						errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
						errorMsg = fmt.Sprintf("ScanStruct Error[Username=%s]", req.GetInvitee())
						goto COMPLETE

					}
				}

				//向受邀请人发送邀请加群通知
				if newSeq, err = redis.Uint64(redisConn.Do("INCR", fmt.Sprintf("userSeq:%s", req.GetInvitee()))); err != nil {
					kc.logger.Error("redisConn INCR userSeq Error", zap.Error(err))
					errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
					errorMsg = fmt.Sprintf("INCR error[Username=%s]", req.GetInviter())
					goto COMPLETE
				}
				inviterNick, _ := redis.String(redisConn.Do("HGET", fmt.Sprintf("userData:%s", req.GetInviter()), "Nick"))
				handledMsg := fmt.Sprintf("用户 %s 邀请你入群", inviterNick)
				body := Msg.MessageNotificationBody{
					Type:           Msg.MessageNotificationType_MNT_TeamInvite, //邀请加群
					HandledAccount: username,                                   //当前用户
					HandledMsg:     handledMsg,
					Status:         Msg.MessageStatus_MOS_Processing,
					Data:           []byte(""),
					To:             teamID, //群组id
				}
				bodyData, _ := proto.Marshal(&body)
				inviteEventRsp := &Msg.RecvMsgEventRsp{
					Scene:        Msg.MessageScene_MsgScene_S2C,        //系统消息
					Type:         Msg.MessageType_MsgType_Notification, //通知类型
					Body:         bodyData,
					From:         username,
					FromDeviceId: deviceID,
					ServerMsgId:  msg.GetID(),                        //服务器分配的消息ID
					WorkflowID:   req.GetWorkflowID(),                //工作流ID
					Seq:          newSeq,                             //消息序号，单个会话内自然递增, 这里是对inviteUsername这个用户的通知序号
					Uuid:         fmt.Sprintf("%d", msg.GetTaskID()), //客户端分配的消息ID，SDK生成的消息id，这里返回TaskID
					Time:         uint64(time.Now().UnixNano() / 1e6),
				}
				go kc.BroadcastMsgToAllDevices(inviteEventRsp, req.GetInvitee()) //向受邀请人推送

				//向其它管理员推送
				managers, _ := kc.GetOwnerAndManagers(teamID)
				for _, manager := range managers {
					if newSeq, err = redis.Uint64(redisConn.Do("INCR", fmt.Sprintf("userSeq:%s", manager))); err != nil {
						kc.logger.Error("redisConn INCR userSeq Error", zap.Error(err))
						errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
						errorMsg = fmt.Sprintf("INCR error[Username=%s]", manager)
						goto COMPLETE
					}
					handledMsg := fmt.Sprintf("管理员同意用户 %s 拉人入群申请", inviterNick)
					body := Msg.MessageNotificationBody{
						Type:           Msg.MessageNotificationType_MNT_RejectTeamInvite, //管理员拒绝加群申请
						HandledAccount: username,                                         //当前用户
						HandledMsg:     handledMsg,
						Status:         Msg.MessageStatus_MOS_Declined,
						Data:           []byte(""),
						To:             teamID, //群组id
					}
					bodyData, _ := proto.Marshal(&body)
					inviteEventRsp := &Msg.RecvMsgEventRsp{
						Scene:        Msg.MessageScene_MsgScene_S2C,        //系统消息
						Type:         Msg.MessageType_MsgType_Notification, //通知类型
						Body:         bodyData,
						From:         username,
						FromDeviceId: deviceID,
						ServerMsgId:  msg.GetID(),                        //服务器分配的消息ID
						WorkflowID:   req.GetWorkflowID(),                //工作流ID
						Seq:          newSeq,                             //消息序号，单个会话内自然递增, 这里是对inviteUsername这个用户的通知序号
						Uuid:         fmt.Sprintf("%d", msg.GetTaskID()), //客户端分配的消息ID，SDK生成的消息id，这里返回TaskID
						Time:         uint64(time.Now().UnixNano() / 1e6),
					}
					go kc.BroadcastMsgToAllDevices(inviteEventRsp, manager) //向其它管理员推送
				}

			} else {

				//向邀请人 及 其它管理员发送通知
				if newSeq, err = redis.Uint64(redisConn.Do("INCR", fmt.Sprintf("userSeq:%s", req.GetInviter()))); err != nil {
					kc.logger.Error("redisConn INCR userSeq Error", zap.Error(err))
					errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
					errorMsg = fmt.Sprintf("INCR error[Username=%s]", req.GetInviter())
					goto COMPLETE
				}
				inviterNick, _ := redis.String(redisConn.Do("HGET", fmt.Sprintf("userData:%s", req.GetInviter()), "Nick"))
				handledMsg := fmt.Sprintf("管理员拒绝用户 %s 拉人入群申请", inviterNick)
				body := Msg.MessageNotificationBody{
					Type:           Msg.MessageNotificationType_MNT_RejectTeamApply, //管理员拒绝加群申请
					HandledAccount: username,                                        //当前用户
					HandledMsg:     handledMsg,
					Status:         Msg.MessageStatus_MOS_Declined,
					Data:           []byte(""),
					To:             teamID, //群组id
				}
				bodyData, _ := proto.Marshal(&body)
				inviteEventRsp := &Msg.RecvMsgEventRsp{
					Scene:        Msg.MessageScene_MsgScene_S2C,        //系统消息
					Type:         Msg.MessageType_MsgType_Notification, //通知类型
					Body:         bodyData,
					From:         username,
					FromDeviceId: deviceID,
					ServerMsgId:  msg.GetID(),                        //服务器分配的消息ID
					WorkflowID:   req.GetWorkflowID(),                //工作流ID
					Seq:          newSeq,                             //消息序号，单个会话内自然递增, 这里是对inviteUsername这个用户的通知序号
					Uuid:         fmt.Sprintf("%d", msg.GetTaskID()), //客户端分配的消息ID，SDK生成的消息id，这里返回TaskID
					Time:         uint64(time.Now().UnixNano() / 1e6),
				}
				go kc.BroadcastMsgToAllDevices(inviteEventRsp, req.GetInviter()) //向邀请人推送

				//向其它管理员推送
				managers, _ := kc.GetOwnerAndManagers(teamID)
				for _, manager := range managers {
					if newSeq, err = redis.Uint64(redisConn.Do("INCR", fmt.Sprintf("userSeq:%s", manager))); err != nil {
						kc.logger.Error("redisConn INCR userSeq Error", zap.Error(err))
						errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
						errorMsg = fmt.Sprintf("INCR error[Username=%s]", manager)
						goto COMPLETE
					}
					body := Msg.MessageNotificationBody{
						Type:           Msg.MessageNotificationType_MNT_RejectTeamInvite, //管理员拒绝加群申请
						HandledAccount: username,                                         //当前用户
						HandledMsg:     handledMsg,
						Status:         Msg.MessageStatus_MOS_Declined,
						Data:           []byte(""),
						To:             teamID, //群组id
					}
					bodyData, _ := proto.Marshal(&body)
					inviteEventRsp := &Msg.RecvMsgEventRsp{
						Scene:        Msg.MessageScene_MsgScene_S2C,        //系统消息
						Type:         Msg.MessageType_MsgType_Notification, //通知类型
						Body:         bodyData,
						From:         username,
						FromDeviceId: deviceID,
						ServerMsgId:  msg.GetID(),                        //服务器分配的消息ID
						WorkflowID:   req.GetWorkflowID(),                //工作流ID
						Seq:          newSeq,                             //消息序号，单个会话内自然递增, 这里是对inviteUsername这个用户的通知序号
						Uuid:         fmt.Sprintf("%d", msg.GetTaskID()), //客户端分配的消息ID，SDK生成的消息id，这里返回TaskID
						Time:         uint64(time.Now().UnixNano() / 1e6),
					}
					go kc.BroadcastMsgToAllDevices(inviteEventRsp, manager) //向其它管理员推送
				}
			}

		} else {
			//其它成员无权设置
			kc.logger.Warn("其它成员无权审核用户入群申请")
			errorCode = http.StatusBadRequest //错误码， 200是正常，其它是错误
			errorMsg = fmt.Sprintf("其它成员无权审核用户入群申请[username=%s]", username)
			goto COMPLETE
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
		kc.logger.Info("Succeed succeed send message to ProduceChannel", zap.String("topic", topic))
	} else {
		kc.logger.Error("Failed to send  message to ProduceChannel", zap.Error(err))
	}
	_ = err
	return nil

}

/*
4-27 分页获取群成员信息
分页方式获取群成员信息，该接口仅支持在线获取，SDK不进行缓存
*/

func (kc *KafkaClient) HandleGetTeamMembersPage(msg *models.Message) error {
	var err error
	errorCode := 200
	var errorMsg string
	var data []byte
	// var newSeq uint64

	rsp := &Team.GetTeamMembersPageRsp{
		Members: make([]*Team.Tmember, 0),
	}

	redisConn := kc.redisPool.Get()
	defer redisConn.Close()

	username := msg.GetUserName() //用户自己的账号
	// token := msg.GetJwtToken()
	deviceID := msg.GetDeviceID()

	kc.logger.Info("HandleGetTeamMembersPage start...",
		zap.String("username", username),
		zap.String("deviceId", deviceID))

	//取出当前设备的os， clientType， logonAt
	curDeviceHashKey := fmt.Sprintf("devices:%s:%s", username, deviceID)
	isMaster, _ := redis.Bool(redisConn.Do("HGET", curDeviceHashKey, "ismaster"))
	curOs, _ := redis.String(redisConn.Do("HGET", curDeviceHashKey, "os"))
	curClientType, _ := redis.Int(redisConn.Do("HGET", curDeviceHashKey, "clientType"))
	curLogonAt, _ := redis.Uint64(redisConn.Do("HGET", curDeviceHashKey, "logonAt"))

	kc.logger.Debug("GetTeamMembersPage ",
		zap.Bool("isMaster", isMaster),
		zap.String("username", username),
		zap.String("deviceID", deviceID),
		zap.String("curOs", curOs),
		zap.Int("curClientType", curClientType),
		zap.Uint64("curLogonAt", curLogonAt))

	//打开msg里的负载， 获取请求参数
	body := msg.GetContent()

	//解包body
	req := &Team.GetTeamMembersPageReq{}
	if err := proto.Unmarshal(body, req); err != nil {
		kc.logger.Error("Protobuf Unmarshal Error", zap.Error(err))
		errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
		errorMsg = fmt.Sprintf("Protobuf Unmarshal Error: %s", err.Error())
		goto COMPLETE

	} else {
		kc.logger.Debug("GetTeamMembersPage  payload",
			zap.String("TeamId", req.GetTeamId()),
			// zap.Int32("QueryType", req.GetQueryType()),
			zap.Int32("Page", req.GetPage()),
			zap.Int32("PageSize", req.GetPageSize()),
		)
		teamID := req.GetTeamId()

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

		key = fmt.Sprintf("TeamUser:%s:%s", teamID, username)
		teamUser := new(models.TeamUser)
		if result, err := redis.Values(redisConn.Do("HGETALL", key)); err == nil {
			if err := redis.ScanStruct(result, teamUser); err != nil {
				kc.logger.Error("错误：ScanStruct", zap.Error(err))
				errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
				errorMsg = fmt.Sprintf("Team user is not exists[username=%s]", username)
				goto COMPLETE
			}
		}

		//TODO  GetPages 分页返回数据
		var maps string
		switch req.GetQueryType() {
		case Team.QueryType_Tmqt_Undefined, Team.QueryType_Tmqt_All:
			maps = "team_member_type != 0 "
		case Team.QueryType_Tmqt_Manager: //管理员
			maps = "team_member_type = 2 "
		case Team.QueryType_Tmqt_Muted:
			maps = "is_mute = true " //禁言
		}
		var total uint64
		teamUsers := kc.GetTeamUsers(int(req.GetPage()), int(req.GetPageSize()), &total, maps)
		rsp.Total = int32(total) //总页数
		for _, teamUser := range teamUsers {
			rsp.Members = append(rsp.Members, &Team.Tmember{
				TeamId:          teamUser.TeamID,
				Username:        teamUser.Username,
				Invitedusername: teamUser.InvitedUsername,
				Nick:            teamUser.Nick,
				Avatar:          teamUser.Avatar,
				Label:           teamUser.Label,
				Source:          teamUser.Source,
				Type:            Team.TeamMemberType(teamUser.TeamMemberType),
				NotifyType:      Team.NotifyType(teamUser.NotifyType),
				Mute:            teamUser.IsMute,
				Ex:              teamUser.Extend,
				JoinTime:        uint64(teamUser.JoinAt),
				UpdateTime:      uint64(teamUser.UpdatedAt),
			})
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
		kc.logger.Info("Succeed succeed send message to ProduceChannel", zap.String("topic", topic))
	} else {
		kc.logger.Error("Failed to send  message to ProduceChannel", zap.Error(err))
	}
	_ = err
	return nil

}

// 获取某个群的群主或管理员
func (kc *KafkaClient) GetOwnerAndManagers(teamID string) ([]string, error) {
	// var err error
	var teamMemberType int

	redisConn := kc.redisPool.Get()
	defer redisConn.Close()

	userNames := make([]string, 0)

	teamMembers, _ := redis.Strings(redisConn.Do("ZRANGEBYSCORE", fmt.Sprintf("TeamUsers:%s", teamID), "-inf", "+inf"))
	for _, teamMember := range teamMembers {
		key := fmt.Sprintf("TeamUser:%s:%s", teamID, teamMember)
		teamMemberType, _ = redis.Int(redisConn.Do("HGET", key, "TeamMemberType"))
		if Team.TeamMemberType(teamMemberType) == Team.TeamMemberType_Tmt_Owner || Team.TeamMemberType(teamMemberType) == Team.TeamMemberType_Tmt_Manager {
			//管理员或群主
			userNames = append(userNames, teamMember)
		}

	}

	return userNames, nil
}
