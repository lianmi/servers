package kafkaBackend

import (
	"github.com/jinzhu/gorm"
	"github.com/lianmi/servers/internal/pkg/models"
	"go.uber.org/zap"
)

//GetTransaction 获取事务
func (kc *KafkaClient) GetTransaction() *gorm.DB {
	return kc.db.Begin()
}

//修改用户资料
func (kc *KafkaClient) SaveUser(user *models.User) error {
	//使用事务同时更新用户数据
	tx := kc.GetTransaction()

	if err := tx.Save(user).Error; err != nil {
		kc.logger.Error("更新用户表失败", zap.Error(err))
		tx.Rollback()

	}
	//提交
	tx.Commit()

	return nil
}

//更新好友
func (kc *KafkaClient) SaveFriend(pFriend *models.Friend) error {
	//使用事务同时更新好友数据
	tx := kc.GetTransaction()

	if err := tx.Save(pFriend).Error; err != nil {
		kc.logger.Error("更新好友表失败", zap.Error(err))
		tx.Rollback()

	}
	//提交
	tx.Commit()

	return nil
}

//删除好友
func (kc *KafkaClient) DeleteFriend(userID, friendUserID uint64) error {
	where := models.Friend{UserID: userID, FriendUserID: friendUserID}
	db := kc.db.Where(&where).Delete(models.Friend{})
	err := db.Error
	if err != nil {
		kc.logger.Error("DeleteFriend出错", zap.Error(err))
		return err
	}
	count := db.RowsAffected
	kc.logger.Debug("DeleteFriend成功", zap.Int64("count", count))
	return nil
}

//创建群
func (kc *KafkaClient) SaveTeam(pTeam *models.Team) error {
	//使用事务同时更新创建群数据
	tx := kc.GetTransaction()

	if err := tx.Save(pTeam).Error; err != nil {
		kc.logger.Error("更新群team表失败", zap.Error(err))
		tx.Rollback()
		return err
	}

	//提交
	tx.Commit()

	return nil
}

//增加群成员
func (kc *KafkaClient) SaveTeamUser(pTeamUser *models.TeamUser) error {
	//使用事务同时更新增加群成员
	tx := kc.GetTransaction()

	if err := tx.Save(pTeamUser).Error; err != nil {
		kc.logger.Error("更新TeamUser表失败", zap.Error(err))
		tx.Rollback()
		return err
	}

	//提交
	tx.Commit()

	return nil
}

//删除群成员
func (kc *KafkaClient) DeleteTeamUser(teamID, username string) error {
	where := models.TeamUser{TeamID: teamID, Username: username}
	db := kc.db.Where(&where).Delete(models.TeamUser{})
	err := db.Error
	if err != nil {
		kc.logger.Error("DeleteTeamUser", zap.Error(err))
		return err
	}
	count := db.RowsAffected
	kc.logger.Debug("DeleteTeamUser成功", zap.Int64("count", count))
	return nil
}

//设置群管理员
func (kc *KafkaClient) SetTeamManager(teamID, username string) error {
	where := models.TeamUser{TeamID: teamID, Username: username}
	db := kc.db.Where(&where).Save(models.TeamUser{})
	err := db.Error
	if err != nil {
		kc.logger.Error("DeleteTeamUser", zap.Error(err))
		return err
	}
	count := db.RowsAffected
	kc.logger.Debug("DeleteTeamUser成功", zap.Int64("count", count))
	return nil
}

// GetPages 分页返回数据
func (kc *KafkaClient) GetPages(model interface{}, out interface{}, pageIndex, pageSize int, totalCount *uint64, where interface{}, orders ...string) error {
	db := kc.db.Model(model).Where(model)
	db = db.Where(where)
	if len(orders) > 0 {
		for _, order := range orders {
			db = db.Order(order)
		}
	}
	err := db.Count(totalCount).Error
	if err != nil {
		kc.logger.Error("查询总数出错", zap.Error(err))
		return err
	}
	if *totalCount == 0 {
		return nil
	}
	return db.Offset((pageIndex - 1) * pageSize).Limit(pageSize).Find(out).Error
}

//分页获取群成员
func (kc *KafkaClient) GetTeamUsers(PageNum int, PageSize int, total *uint64, where interface{}) []*models.TeamUser {
	var teamUsers []*models.TeamUser
	if err := kc.GetPages(&models.User{}, &teamUsers, PageNum, PageSize, total, where); err != nil {
		kc.logger.Error("获取用户信息失败", zap.Error(err))
	}
	return teamUsers
}
