package repositories

import (
	"fmt"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/gomodule/redigo/redis"
	Global "github.com/lianmi/servers/api/proto/global"
	Msg "github.com/lianmi/servers/api/proto/msg"
	Team "github.com/lianmi/servers/api/proto/team"
	"github.com/lianmi/servers/internal/common"
	"github.com/lianmi/servers/internal/pkg/models"
	"github.com/pkg/errors"
	uuid "github.com/satori/go.uuid"
	"go.uber.org/zap"
	"gorm.io/gorm/clause"
)

//授权新创建的群组: 将Status变为 Normal(1) - 正常状态
func (s *MysqlLianmiRepository) ApproveTeam(teamID string) error {
	var err error

	redisConn := s.redisPool.Get()
	defer redisConn.Close()

	p := new(models.Team)

	if err = s.db.Model(p).Where(&models.Team{
		TeamInfo: models.TeamInfo{
			TeamID: teamID,
		},
	}).First(p).Error; err != nil {
		return errors.Wrapf(err, "Get team info error[teamID=%s]", teamID)
	}

	//用户拥有的群的总数量
	if count, err := redis.Int(redisConn.Do("ZCARD", fmt.Sprintf("Team:%s", p.Owner))); err != nil {
		s.logger.Error("ZCARD Error", zap.Error(err))
	} else {
		if count >= common.MaxTeamLimit {
			return errors.Wrapf(err, "Reach team max limit[count=%d]", count)
		}

	}

	p.Status = int(Team.TeamStatus_Status_Normal) //状态 Init(0) - 初始状态,未审核 Normal(1) - 正常状态 Blocked(2) - 封禁状态

	//存储群成员信息 TeamUser
	memberNick, _ := redis.String(redisConn.Do("HGET", "userData:%s", p.Owner, "Nick"))
	memberAvatar, _ := redis.String(redisConn.Do("HGET", "userData:%s", p.Owner, "Avatar"))
	memberLabel, _ := redis.String(redisConn.Do("HGET", "userData:%s", p.Owner, "Label"))
	memberExtend, _ := redis.String(redisConn.Do("HGET", "userData:%s", p.Owner, "Extend"))

	teamUser := new(models.TeamUser)
	teamUser.TeamUserInfo.JoinAt = time.Now().UnixNano() / 1e6
	teamUser.TeamUserInfo.Teamname = p.Teamname
	teamUser.TeamUserInfo.Username = p.Owner
	teamUser.TeamUserInfo.Nick = memberNick                                   //群成员呢称
	teamUser.TeamUserInfo.Avatar = memberAvatar                               //群成员头像
	teamUser.TeamUserInfo.Label = memberLabel                                 //群成员标签
	teamUser.TeamUserInfo.Source = ""                                         //群成员来源  TODO
	teamUser.TeamUserInfo.Extend = memberExtend                               //群成员扩展字段
	teamUser.TeamUserInfo.TeamMemberType = int(Team.TeamMemberType_Tmt_Owner) //群成员类型 Owner(4) - 创建者
	teamUser.TeamUserInfo.IsMute = false                                      //是否被禁言
	teamUser.TeamUserInfo.NotifyType = 1                                      //群消息通知方式 All(1) - 群全部消息提醒

	//将Status变为 Normal(1) - 正常状态
	result := s.db.Model(&models.Team{}).Where(&models.Team{
		TeamInfo: models.TeamInfo{
			TeamID: teamID,
		},
	}).Update("status", int(Team.TeamStatus_Status_Normal))

	//updated records count
	s.logger.Debug("ApproveTeam result: ", zap.Int64("RowsAffected", result.RowsAffected), zap.Error(result.Error))

	//增加teamuser表记录
	if err := s.db.Clauses(clause.OnConflict{DoNothing: true}).Create(&teamUser).Error; err != nil {
		s.logger.Error("ApproveTeam, failed to upsert teamUser", zap.Error(err))
	} else {
		s.logger.Debug("ApproveTeam, upsert teamUser succeed")
	}

	/*
		1. 用户拥有的群，用有序集合存储，Key: Team:{Owner}, 成员元素是: TeamnID
		3. 每个群在用有序集合存储, key格式为： TeamUsers:{TeamnID}, 成员元素是: Username
		4. 每个群成员用哈希表存储，Key格式为： TeamUser:{TeamnID}:{Username} , 字段为: Teamname Username Nick JoinAt 等TeamUser表的字段
		5. 被移除的成员列表，Key格式为： TeamUsersRemoved:{TeamnID}
	*/

	//存储所有群组， 方便查询及定时任务解禁
	err = redisConn.Send("ZADD", "Teams", time.Now().UnixNano()/1e6, p.TeamID)
	err = redisConn.Send("ZADD", fmt.Sprintf("Team:%s", p.Owner), time.Now().UnixNano()/1e6, p.TeamID)

	//当前只有群主一个成员
	err = redisConn.Send("ZADD", fmt.Sprintf("TeamUsers:%s", p.TeamID), time.Now().UnixNano()/1e6, p.Owner)

	err = redisConn.Send("HMSET", redis.Args{}.Add(fmt.Sprintf("TeamUser:%s:%s", p.TeamID, p.Owner)).AddFlat(teamUser.TeamUserInfo)...)

	//存储群信息
	teamInfoKey := fmt.Sprintf("TeamInfo:%s", p.TeamID)
	teamInfo := &models.TeamInfo{
		TeamID:       p.TeamID,
		Teamname:     p.Teamname,
		Nick:         p.Nick,
		Icon:         p.Icon,
		Announcement: p.Announcement,
		Introductory: p.Introductory,
		Status:       p.Status,
		Extend:       p.Extend,
		Owner:        p.Owner,
		Type:         p.Type,
		VerifyType:   p.VerifyType,
		InviteMode:   p.InviteMode,
		MemberLimit:  p.MemberLimit,
		MemberNum:    1, //刚刚建群是只有群主1人
		MuteType:     p.MuteType,
		Ex:           p.Ex,
		ModifiedBy:   p.ModifiedBy,
		IsMute:       p.IsMute,
	}

	err = redisConn.Send("HMSET", redis.Args{}.Add(teamInfoKey).AddFlat(teamInfo)...)

	//更新redis的sync:{用户账号} teamsAt 时间戳
	err = redisConn.Send("HSET",
		fmt.Sprintf("sync:%s", p.Owner),
		"teamsAt",
		time.Now().UnixNano()/1e6)

	redisConn.Flush()

	//向群主推送通知，此群已经审核通过

	//群资料主要字段
	updateTeamInfo := &Team.TeamInfo{
		TeamName:     p.Teamname,
		Icon:         p.Icon,
		Announcement: p.Announcement,
		Introduce:    p.Introductory,
		VerifyType:   Team.VerifyType(p.VerifyType),
		InviteMode:   Team.InviteMode(p.InviteMode),
		Owner:        teamInfo.Owner,
		Type:         Team.TeamType(p.Type),
		MemberLimit:  int32(common.PerTeamMembersLimit),
		MemberNum:    int32(1),
		Status:       Team.TeamStatus(teamInfo.Status),
		MuteType:     Team.MuteMode(teamInfo.MuteType),
		Ex:           teamInfo.Extend,
		IsMute:       teamInfo.IsMute,
	}
	updateTeamInfoData, _ := proto.Marshal(updateTeamInfo)

	body := Msg.MessageNotificationBody{
		Type:           Msg.MessageNotificationType_MNT_Approveteam, //群审核通过，成为正常状态，可以加群及拉人
		HandledAccount: p.Owner,                                     //群主
		HandledMsg:     "approveteam passed",
		Status:         Msg.MessageStatus_MOS_Passed, //已通过验证
		Data:           updateTeamInfoData,           //群信息
		To:             p.Owner,                      //群主
	}
	bodyData, _ := proto.Marshal(&body)

	eRsp := &Msg.RecvMsgEventRsp{
		Scene:        Msg.MessageScene_MsgScene_S2C,        //系统消息
		Type:         Msg.MessageType_MsgType_Notification, //通知类型
		Body:         bodyData,                             //字节流
		From:         p.Teamname,
		FromDeviceId: "",
		ServerMsgId:  uuid.NewV4().String(), //服务器分配的消息ID
		Recv:         teamID,                //接收方, 根据场景判断to是个人还是群
		WorkflowID:   "",                    //工作流ID
		Seq:          0,                     //消息序号，单个会话内自然递增, 这里是对inviteUsername这个用户的通知序号
		Uuid:         "",
		Time:         uint64(time.Now().UnixNano() / 1e6),
	}

	data, _ := proto.Marshal(eRsp)
	/*
		//删除7天前的缓存系统消息
		nTime := time.Now()
		yesTime := nTime.AddDate(0, 0, -7).Unix()
		offLineMsgListKey := fmt.Sprintf("offLineMsgList:%s", p.Owner)

		_, err = redisConn.Do("ZREMRANGEBYSCORE", offLineMsgListKey, "-inf", yesTime)

		//Redis里缓存此系统消息,目的是6-1同步接口里的 systemmsgAt, 然后同步给用户
		systemMsgAt := time.Now().UnixNano() / 1e6
		if _, err := redisConn.Do("ZADD", offLineMsgListKey, systemMsgAt, eRsp.GetServerMsgId()); err != nil {
			s.logger.Error("ZADD Error", zap.Error(err))
		}

		//系统消息具体内容
		systemMsgKey := fmt.Sprintf("systemMsg:%s:%s", p.Owner, eRsp.GetServerMsgId())

		_, err = redisConn.Do("HMSET",
			systemMsgKey,
			"Username", p.Owner,
			"SystemMsgAt", systemMsgAt,
			"Seq", eRsp.Seq,
			"Data", data,
		)

		_, err = redisConn.Do("EXPIRE", systemMsgKey, 7*24*3600) //设置有效期为7天
	*/

	//向toUser所有端发送
	deviceListKey := fmt.Sprintf("devices:%s", p.Owner)
	eDeviceID, _ := redis.String(redisConn.Do("GET", deviceListKey))

	targetMsg := &models.Message{}
	curDeviceKey := fmt.Sprintf("DeviceJwtToken:%s", eDeviceID)
	curJwtToken, _ := redis.String(redisConn.Do("GET", curDeviceKey))
	s.logger.Debug("Redis GET ", zap.String("curDeviceKey", curDeviceKey), zap.String("curJwtToken", curJwtToken))

	targetMsg.UpdateID()
	//构建消息路由, 第一个参数是要处理的业务类型，后端服务器处理完成后，需要用此来拼接topic: {businessTypeName.Frontend}
	targetMsg.BuildRouter("Msg", "", "Msg.Frontend")

	targetMsg.SetJwtToken(curJwtToken)
	targetMsg.SetUserName(p.Owner)
	targetMsg.SetDeviceID(eDeviceID)
	targetMsg.SetBusinessTypeName("Msg")
	targetMsg.SetBusinessType(uint32(Global.BusinessType_Msg))           //消息模块
	targetMsg.SetBusinessSubType(uint32(Global.MsgSubType_RecvMsgEvent)) //接收消息事件

	targetMsg.BuildHeader("Dispatcher", time.Now().UnixNano()/1e6)

	targetMsg.FillBody(data) //网络包的body，承载真正的业务数据

	targetMsg.SetCode(200) //成功的状态码

	//构建数据完成，向NsqChan发送
	s.multiChan.NsqChan <- targetMsg

	s.logger.Info("Broadcast Msg To AllDevices Succeed",
		zap.String("Username:", p.Owner),
		zap.String("DeviceID:", curDeviceKey),
		zap.Int64("Now", time.Now().UnixNano()/1e6))

	_ = err

	return nil

}

//封禁群组 状态 Init(0) - 初始状态,未审核 Normal(1) - 正常状态 Blocked(2) - 封禁状态
func (s *MysqlLianmiRepository) BlockTeam(teamID string) error {

	result := s.db.Model(&models.Team{}).Where(&models.Team{
		TeamInfo: models.TeamInfo{
			TeamID: teamID,
		},
	}).Update("status", int(Team.TeamStatus_Status_Blocked)) //更改Status

	//updated records count
	s.logger.Debug("BlockTeam result: ", zap.Int64("RowsAffected", result.RowsAffected), zap.Error(result.Error))

	if result.Error != nil {
		s.logger.Error("封禁群组失败", zap.Error(result.Error))
		return result.Error
	}

	return nil

}

//解封群组
func (s *MysqlLianmiRepository) DisBlockTeam(teamID string) error {
	result := s.db.Model(&models.Team{}).Where(&models.Team{
		TeamInfo: models.TeamInfo{
			TeamID: teamID,
		},
	}).Update("status", int(Team.TeamStatus_Status_Normal)) //更改Status

	//updated records count
	s.logger.Debug("DisBlockTeam result: ", zap.Int64("RowsAffected", result.RowsAffected), zap.Error(result.Error))

	if result.Error != nil {
		s.logger.Error("解封群组失败", zap.Error(result.Error))
		return result.Error
	}

	return nil
}

//保存禁言的值，用于设置群禁言或解禁
func (s *MysqlLianmiRepository) UpdateTeamMute(teamID string, muteType int) error {
	result := s.db.Model(&models.Team{}).Where(&models.Team{
		TeamInfo: models.TeamInfo{
			TeamID: teamID,
		},
	}).Update("mute_type", muteType) //修改MuteType

	//updated records count
	s.logger.Debug("UpdateTeamMute result: ",
		zap.Int64("RowsAffected", result.RowsAffected),
		zap.Error(result.Error))

	if result.Error != nil {
		s.logger.Error("设置群禁言或解禁失败", zap.Error(result.Error))
		return result.Error
	}

	return nil
}

//增加群成员资料
func (s *MysqlLianmiRepository) AddTeamUser(pTeamUser *models.TeamUser) error {
	if pTeamUser == nil {
		return errors.New("pTeamUser is nil")
	}

	//增加记录
	if err := s.db.Clauses(clause.OnConflict{DoNothing: true}).Create(&pTeamUser).Error; err != nil {
		s.logger.Error("增加TeamUser失败", zap.Error(err))
		return err
	} else {
		s.logger.Debug("增加TeamUser成功")
	}

	return nil
}

//增加或移除群管理员, isAdd = true为增加 isAdd=false为移除
func (s *MysqlLianmiRepository) UpdateTeamUserManager(teamID, managerUsername string, isAdd bool) error {

	where := models.TeamUser{
		TeamUserInfo: models.TeamUserInfo{
			TeamID:   teamID,
			Username: managerUsername,
		},
	}
	var teamMemberType int
	if isAdd {
		teamMemberType = 2
	} else {
		teamMemberType = 3
	}

	// 同时更新多个字段
	result := s.db.Model(&models.TeamUser{}).Where(where).Updates(&models.TeamUser{
		TeamUserInfo: models.TeamUserInfo{
			TeamMemberType: teamMemberType, //群管理员 or 普通群成员s
		},
	})

	//updated records count
	s.logger.Debug("UpdateTeamUserManager result: ",
		zap.Int64("RowsAffected", result.RowsAffected),
		zap.Error(result.Error))

	if result.Error != nil {
		s.logger.Error("UpdateTeamUserManager失败", zap.Error(result.Error))
		return result.Error
	} else {
		s.logger.Debug("UpdateTeamUserManager成功")
	}

	return nil
}

// 修改群成员呢称、扩展
func (s *MysqlLianmiRepository) UpdateTeamUserMyInfo(teamID, username, nick, ex string) error {
	where := models.TeamUser{
		TeamUserInfo: models.TeamUserInfo{
			TeamID:   teamID,
			Username: username,
		},
	}
	// 同时更新多个字段
	result := s.db.Model(&models.TeamUser{}).Where(where).Updates(models.TeamUser{
		TeamUserInfo: models.TeamUserInfo{
			Nick:   nick,
			Extend: ex,
		},
	})

	//updated records count
	s.logger.Debug("UpdateTeamUserMyInfo result: ",
		zap.Int64("RowsAffected", result.RowsAffected),
		zap.Error(result.Error))

	if result.Error != nil {
		s.logger.Error("UpdateTeamUserMyInfo失败", zap.Error(result.Error))
		return result.Error
	} else {
		s.logger.Debug("UpdateTeamUserMyInfo成功")
	}

	return nil
}

//修改群通知方式
func (s *MysqlLianmiRepository) UpdateTeamUserNotifyType(teamID string, notifyType int) error {
	where := models.TeamUser{
		TeamUserInfo: models.TeamUserInfo{
			TeamID: teamID,
		},
	}

	//更新notify_type字段
	result := s.db.Model(&models.TeamUser{}).Where(where).Update("notify_type", notifyType)

	//updated records count
	s.logger.Debug("UpdateTeamUserManager result: ",
		zap.Int64("RowsAffected", result.RowsAffected),
		zap.Error(result.Error))

	if result.Error != nil {
		s.logger.Error("UpdateTeamUserNotifyType失败", zap.Error(result.Error))
		return result.Error
	} else {
		s.logger.Debug("UpdateTeamUserNotifyType成功")
	}

	return nil
}

//解除群成员的禁言
func (s *MysqlLianmiRepository) SetMuteTeamUser(teamID, dissMuteUser string, isMute bool, mutedays int) error {

	where := models.TeamUser{
		TeamUserInfo: models.TeamUserInfo{
			TeamID:   teamID,
			Username: dissMuteUser,
		},
	}
	result := s.db.Model(&models.TeamUser{}).Where(&where).Updates(models.TeamUser{
		TeamUserInfo: models.TeamUserInfo{
			IsMute:   isMute,
			Mutedays: mutedays,
		},
	})

	//updated records count
	s.logger.Debug("SetMuteTeamUser result: ",
		zap.Int64("RowsAffected", result.RowsAffected),
		zap.Error(result.Error))

	if result.Error != nil {
		s.logger.Error("设置群成员的禁言失败", zap.Error(result.Error))
		return result.Error
	} else {
		s.logger.Debug("设置群成员的禁言成功")
	}
	return nil

}

//删除群成员
func (s *MysqlLianmiRepository) DeleteTeamUser(teamID, username string) error {
	where := models.TeamUser{
		TeamUserInfo: models.TeamUserInfo{
			TeamID:   teamID,
			Username: username,
		},
	}
	db2 := s.db.Where(&where).Delete(models.TeamUser{})
	err := db2.Error
	if err != nil {
		s.logger.Error("DeleteTeamUser", zap.Error(err))
		return err
	}
	count := db2.RowsAffected
	s.logger.Debug("DeleteTeamUser成功", zap.Int64("count", count))
	return nil
}

// GetPages 分页返回数据
func (s *MysqlLianmiRepository) GetPages(model interface{}, out interface{}, pageIndex, pageSize int, totalCount *int64, where interface{}, orders ...string) error {
	db2 := s.db.Model(model).Where(model).Where(where)
	if len(orders) > 0 {
		for _, order := range orders {
			db2 = db2.Order(order)
		}
	}
	err := db2.Count(totalCount).Error
	if err != nil {
		s.logger.Error("查询总数出错", zap.Error(err))
		return err
	}
	if *totalCount == 0 {
		return nil
	}
	return db2.Offset((pageIndex - 1) * pageSize).Limit(pageSize).Find(out).Error
}

//分页获取群成员
func (s *MysqlLianmiRepository) GetTeamUsers(teamID string, PageNum int, PageSize int, total *int64, where interface{}) []*models.TeamUser {
	var teamUsers []*models.TeamUser
	if err := s.GetPages(&models.TeamUser{
		TeamUserInfo: models.TeamUserInfo{
			TeamID: teamID,
		},
	}, &teamUsers, PageNum, PageSize, total, where); err != nil {
		s.logger.Error("获取群成员信息失败", zap.Error(err))
	}
	return teamUsers
}

//获取所有群组id， 返回一个切片
func (s *MysqlLianmiRepository) GetTeams() []string {
	var teamIDs []string
	s.db.Model(&models.Team{}).Pluck("team_id", &teamIDs)
	return teamIDs
}

//创建群,  增加一条群信息数据
func (s *MysqlLianmiRepository) CreateTeam(pTeam *models.Team) error {
	//增加记录
	if err := s.db.Clauses(clause.OnConflict{DoNothing: true}).Create(&pTeam).Error; err != nil {
		s.logger.Error("CreateTeam, failed to upsert team", zap.Error(err))
		return err
	} else {
		s.logger.Debug("CreateTeam, upsert team succeed")
	}

	return nil
}

//更新群数据
func (s *MysqlLianmiRepository) UpdateTeam(teamID string, pTeam *models.Team) error {

	where := models.Team{
		TeamInfo: models.TeamInfo{
			TeamID: teamID,
		},
	}
	// 同时更新多个字段
	result := s.db.Model(&models.Team{}).Where(&where).Select("nick", "icon", "announcement", "introductory", "verify_type", "invite_mode").Updates(pTeam)

	//updated records count
	s.logger.Debug("UpdateTeam result: ",
		zap.Int64("RowsAffected", result.RowsAffected),
		zap.Error(result.Error))

	if result.Error != nil {
		s.logger.Error("更新群数据失败", zap.Error(result.Error))
		return result.Error
	}

	return nil
}
