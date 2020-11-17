package repositories

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
	// "github.com/lianmi/servers/internal/app/dispatcher/nsqMq"
	uuid "github.com/satori/go.uuid"
	"go.uber.org/zap"
)

//授权新创建的群组
func (s *MysqlLianmiRepository) ApproveTeam(teamID string) error {
	var err error

	redisConn := s.redisPool.Get()
	defer redisConn.Close()

	p := new(models.Team)

	if err = s.db.Model(p).Where(&models.Team{
		TeamID: teamID,
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

	p.Status = 2 //状态 Init(1) - 初始状态,未审核 Normal(2) - 正常状态 Blocked(3) - 封禁状态

	//存储群成员信息 TeamUser
	memberNick, _ := redis.String(redisConn.Do("HGET", "userData:%s", p.Owner, "Nick"))
	memberAvatar, _ := redis.String(redisConn.Do("HGET", "userData:%s", p.Owner, "Avatar"))
	memberLabel, _ := redis.String(redisConn.Do("HGET", "userData:%s", p.Owner, "Label"))
	memberExtend, _ := redis.String(redisConn.Do("HGET", "userData:%s", p.Owner, "Extend"))
	memberProvince, _ := redis.String(redisConn.Do("HGET", "userData:%s", p.Owner, "Province"))
	memberCity, _ := redis.String(redisConn.Do("HGET", "userData:%s", p.Owner, "City"))
	memberCounty, _ := redis.String(redisConn.Do("HGET", "userData:%s", p.Owner, "County"))
	memberStreet, _ := redis.String(redisConn.Do("HGET", "userData:%s", p.Owner, "Street"))
	memberAddress, _ := redis.String(redisConn.Do("HGET", "userData:%s", p.Owner, "Address"))

	teamUser := new(models.TeamUser)
	teamUser.JoinAt = time.Now().UnixNano() / 1e6
	teamUser.Teamname = p.Teamname
	teamUser.Username = p.Owner
	teamUser.Nick = memberNick                                   //群成员呢称
	teamUser.Avatar = memberAvatar                               //群成员头像
	teamUser.Label = memberLabel                                 //群成员标签
	teamUser.Source = ""                                         //群成员来源  TODO
	teamUser.Extend = memberExtend                               //群成员扩展字段
	teamUser.TeamMemberType = int(Team.TeamMemberType_Tmt_Owner) //群成员类型 Owner(4) - 创建者
	teamUser.IsMute = false                                      //是否被禁言
	teamUser.NotifyType = 1                                      //群消息通知方式 All(1) - 群全部消息提醒
	teamUser.Province = memberProvince                           //省份, 如广东省
	teamUser.City = memberCity                                   //城市，如广州市
	teamUser.County = memberCounty                               //区，如天河区
	teamUser.Street = memberStreet                               //街道
	teamUser.Address = memberAddress                             //地址

	tx := s.base.GetTransaction()

	//将Status变为2 正常状态
	p.Status = 2
	if err := tx.Save(p).Error; err != nil {
		s.logger.Error("授权新创建的群组失败", zap.Error(err))
		tx.Rollback()
		return err

	}
	if err := tx.Save(teamUser).Error; err != nil {
		s.logger.Error("更新teamUser失败", zap.Error(err))
		tx.Rollback()
		return err

	}
	//提交
	tx.Commit()

	/*
		1. 用户拥有的群，用有序集合存储，Key: Team:{Owner}, 成员元素是: TeamnID
		2. 群记录哈希表, key格式为: TeamInfo:{TeamnID}, 字段为: Teamname Nick Icon 等Team表的字段
		3. 每个群在用有序集合存储, key格式为： TeamUsers:{TeamnID}, 成员元素是: Username
		4. 每个群成员用哈希表存储，Key格式为： TeamUser:{TeamnID}:{Username} , 字段为: Teamname Username Nick JoinAt 等TeamUser表的字段
		5. 被移除的成员列表，Key格式为： TeamUsersRemoved:{TeamnID}
	*/

	//存储所有群组， 方便查询及定时任务解禁
	err = redisConn.Send("ZADD", "Teams", time.Now().UnixNano()/1e6, p.TeamID)
	err = redisConn.Send("ZADD", fmt.Sprintf("Team:%s", p.Owner), time.Now().UnixNano()/1e6, p.TeamID)
	err = redisConn.Send("HMSET", redis.Args{}.Add(fmt.Sprintf("TeamInfo:%s", p.TeamID)).AddFlat(p)...)

	//当前只有群主一个成员
	err = redisConn.Send("ZADD", fmt.Sprintf("TeamUsers:%s", p.TeamID), time.Now().UnixNano()/1e6, p.Owner)

	err = redisConn.Send("HMSET", redis.Args{}.Add(fmt.Sprintf("TeamUser:%s:%s", p.TeamID, p.Owner)).AddFlat(teamUser)...)

	//更新redis的sync:{用户账号} teamsAt 时间戳
	err = redisConn.Send("HSET",
		fmt.Sprintf("sync:%s", p.Owner),
		"teamsAt",
		time.Now().UnixNano()/1e6)

	redisConn.Flush()

	//向群主推送通知，此群已经审核通过

	body := Msg.MessageNotificationBody{
		Type:           Msg.MessageNotificationType_MNT_Approveteam, //群审核通过，成为正常状态，可以加群及拉人
		HandledAccount: "operator",
		HandledMsg:     "approveteam passed",
		Status:         Msg.MessageStatus_MOS_Passed, //已通过验证
		Data:           []byte(""),
		To:             p.Owner, //群主
	}
	bodyData, _ := proto.Marshal(&body)

	eRsp := &Msg.RecvMsgEventRsp{
		Scene:        Msg.MessageScene_MsgScene_S2C,        //系统消息
		Type:         Msg.MessageType_MsgType_Notification, //通知类型
		Body:         bodyData,                             //字节流
		From:         "",
		FromDeviceId: "",
		ServerMsgId:  uuid.NewV4().String(), //服务器分配的消息ID
		Recv:         teamID,                //接收方, 根据场景判断to是个人还是群
		WorkflowID:   "",                    //工作流ID
		Seq:          0,                     //消息序号，单个会话内自然递增, 这里是对inviteUsername这个用户的通知序号
		Uuid:         "",
		Time:         uint64(time.Now().UnixNano() / 1e6),
	}

	//TODO
	// s.nsqClient.BroadcastSystemMsgToAllDevices(eRsp, p.Owner, "")
	_ = eRsp

	return nil

}

//封禁群组
func (s *MysqlLianmiRepository) BlockTeam(teamID string) error {
	p := new(models.Team)
	if err := s.db.Model(p).Where(&models.Team{
		TeamID: teamID,
	}).First(p).Error; err != nil {
		return errors.Wrapf(err, "Get team info error[teamID=%s]", teamID)
	}

	p.Status = 3 //状态 Init(1) - 初始状态,未审核 Normal(2) - 正常状态 Blocked(3) - 封禁状态

	tx := s.base.GetTransaction()

	if err := tx.Save(p).Error; err != nil {
		s.logger.Error("封禁群组失败", zap.Error(err))
		tx.Rollback()
		return err
	}
	//提交
	tx.Commit()

	return nil

}

//解封群组
func (s *MysqlLianmiRepository) DisBlockTeam(teamID string) error {
	p := new(models.Team)
	if err := s.db.Model(p).Where(&models.Team{
		TeamID: teamID,
	}).First(p).Error; err != nil {
		return errors.Wrapf(err, "Get team info error[teamID=%s]", teamID)
	}

	p.Status = 2 //状态 Init(1) - 初始状态,未审核 Normal(2) - 正常状态 Blocked(3) - 封禁状态

	tx := s.base.GetTransaction()

	if err := tx.Save(p).Error; err != nil {
		s.logger.Error("解封群组失败", zap.Error(err))
		tx.Rollback()
		return err
	}
	//提交
	tx.Commit()

	return nil
}

//更新或增加群成员
func (s *MysqlLianmiRepository) SaveTeamUser(pTeamUser *models.TeamUser) error {
	//使用事务同时更新或增加群成员
	tx := s.base.GetTransaction()

	if err := tx.Save(pTeamUser).Error; err != nil {
		s.logger.Error("更新TeamUser表失败", zap.Error(err))
		tx.Rollback()
		return err
	}

	//提交
	tx.Commit()

	return nil
}

//删除群成员
func (s *MysqlLianmiRepository) DeleteTeamUser(teamID, username string) error {
	where := models.TeamUser{TeamID: teamID, Username: username}
	db := s.db.Where(&where).Delete(models.TeamUser{})
	err := db.Error
	if err != nil {
		s.logger.Error("DeleteTeamUser", zap.Error(err))
		return err
	}
	count := db.RowsAffected
	s.logger.Debug("DeleteTeamUser成功", zap.Int64("count", count))
	return nil
}

//设置群管理员
func (s *MysqlLianmiRepository) SetTeamManager(teamID, username string) error {
	where := models.TeamUser{TeamID: teamID, Username: username}
	db := s.db.Where(&where).Save(models.TeamUser{})
	err := db.Error
	if err != nil {
		s.logger.Error("DeleteTeamUser", zap.Error(err))
		return err
	}
	count := db.RowsAffected
	s.logger.Debug("DeleteTeamUser成功", zap.Int64("count", count))
	return nil
}

// GetPages 分页返回数据
func (s *MysqlLianmiRepository) GetPages(model interface{}, out interface{}, pageIndex, pageSize int, totalCount *uint64, where interface{}, orders ...string) error {
	db := s.db.Model(model).Where(model)
	db = db.Where(where)
	if len(orders) > 0 {
		for _, order := range orders {
			db = db.Order(order)
		}
	}
	err := db.Count(totalCount).Error
	if err != nil {
		s.logger.Error("查询总数出错", zap.Error(err))
		return err
	}
	if *totalCount == 0 {
		return nil
	}
	return db.Offset((pageIndex - 1) * pageSize).Limit(pageSize).Find(out).Error
}

//分页获取群成员
func (s *MysqlLianmiRepository) GetTeamUsers(teamID string, PageNum int, PageSize int, total *uint64, where interface{}) []*models.TeamUser {
	var teamUsers []*models.TeamUser
	if err := s.GetPages(&models.TeamUser{TeamID: teamID}, &teamUsers, PageNum, PageSize, total, where); err != nil {
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

//创建群
func (s *MysqlLianmiRepository) SaveTeam(pTeam *models.Team) error {
	//使用事务同时更新创建群数据
	tx := s.base.GetTransaction()

	if err := tx.Save(pTeam).Error; err != nil {
		s.logger.Error("更新群team表失败", zap.Error(err))
		tx.Rollback()
		return err
	}

	//提交
	tx.Commit()

	return nil
}
