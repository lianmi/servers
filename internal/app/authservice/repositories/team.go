package repositories

import (
	"fmt"
	"time"

	"github.com/gomodule/redigo/redis"
	Team "github.com/lianmi/servers/api/proto/team"
	"github.com/lianmi/servers/internal/common"
	"github.com/lianmi/servers/internal/pkg/models"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

//授权新创建的群组
func (s *MysqlUsersRepository) ApproveTeam(teamID string) error {
	var err error

	redisConn := s.redisPool.Get()
	defer redisConn.Close()

	p := new(models.Team)
	if err := s.db.Model(p).Where("team_id = ?", teamID).First(p).Error; err != nil {
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
	teamUser.JoinAt = time.Now().UnixNano()/1e6
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
	if _, err = redisConn.Do("ZADD", fmt.Sprintf("Team:%s", p.Owner), time.Now().UnixNano()/1e6, p.TeamID); err != nil {
		s.logger.Error("ZADD Error", zap.Error(err))
	}
	if _, err = redisConn.Do("HMSET", redis.Args{}.Add(fmt.Sprintf("TeamInfo:%s", p.TeamID)).AddFlat(p)...); err != nil {
		s.logger.Error("错误：HMSET TeamInfo", zap.Error(err))
	}

	//当前只有群主一个成员
	if _, err = redisConn.Do("ZADD", fmt.Sprintf("TeamUsers:%s", p.TeamID), time.Now().UnixNano()/1e6, p.Owner); err != nil {
		s.logger.Error("ZADD Error", zap.Error(err))
	}

	if _, err = redisConn.Do("HMSET", redis.Args{}.Add(fmt.Sprintf("TeamUser:%s:%s", p.TeamID, p.Owner)).AddFlat(teamUser)...); err != nil {
		s.logger.Error("错误：HMSET TeamUser", zap.Error(err))
	}
	//更新redis的sync:{用户账号} teamsAt 时间戳
	redisConn.Do("HSET",
		fmt.Sprintf("sync:%s", p.Owner),
		"teamsAt",
		time.Now().UnixNano()/1e6)
	return nil

}

//封禁群组
func (s *MysqlUsersRepository) BlockTeam(teamID string) error {
	p := new(models.Team)
	if err := s.db.Model(p).Where("team_id = ?", teamID).First(p).Error; err != nil {
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
func (s *MysqlUsersRepository) DisBlockTeam(teamID string) error {
	p := new(models.Team)
	if err := s.db.Model(p).Where("team_id = ?", teamID).First(p).Error; err != nil {
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
