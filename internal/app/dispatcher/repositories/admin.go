package repositories

import (
	"github.com/lianmi/servers/internal/pkg/models"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

//管理员修改app版本号
func (s *MysqlLianmiRepository) ManagerSetVersionLast(req *models.VersionInfo) error {
	var err error
	redisConn := s.redisPool.Get()
	defer redisConn.Close()

	_, err = redisConn.Do("SET", "AppVersionLast", req.Version)
	if err != nil {
		s.logger.Error("ManagerSetVersionLast", zap.Error(err))
		return err
	}

	// 写入App版本升级历史表
	//先查询数据是否存在，如果存在，则返回，如果不存在，则新增
	where := models.AppVersionHistory{
		VersionInfo: models.VersionInfo{
			Version: req.Version,
		},
	}
	avh := &models.AppVersionHistory{
		VersionInfo: models.VersionInfo{
			Version: req.Version,
			Details: req.Details,
		},
	}

	appVersionHistory := new(models.AppVersionHistory)
	if err := s.db.Model(&models.AppVersionHistory{}).Where(&where).First(appVersionHistory).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {

			if err := s.db.Clauses(clause.OnConflict{DoNothing: true}).Create(&avh).Error; err != nil {
				s.logger.Error("ManagerSetVersionLast, failed to upsert AppVersionHistory", zap.Error(err))
				return err
			} else {
				s.logger.Debug("ManagerSetVersionLast, upsert AppVersionHistory  succeed")
			}
			return nil

		} else {
			return err
		}

	} else {
		// 修改

		// 同时更新多个字段
		result := s.db.Model(&models.AppVersionHistory{}).Where(&where).Updates(&avh)

		//updated records count
		s.logger.Debug("Update AppVersionHistory result: ",
			zap.Int64("RowsAffected", result.RowsAffected),
			zap.Error(result.Error))

		if result.Error != nil {
			s.logger.Error("Update AppVersionHistory, 修改app版本信息数据失败", zap.Error(result.Error))
			return result.Error
		} else {
			s.logger.Error("Update AppVersionHistory, 修改app版本信息数据成功")
		}
		return nil

	}
}

func (s *MysqlLianmiRepository) ManagerAddSystemMsg(level int, title, content string) error {
	systemMsg := new(models.SystemMsg)
	where := models.SystemMsg{
		Title: title,
	}

	if err := s.db.Model(&models.SystemMsg{}).Where(&where).First(systemMsg).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {

			if err := s.db.Clauses(clause.OnConflict{DoNothing: true}).Create(&models.SystemMsg{
				Level:   level,
				Title:   title,
				Content: content,
				Active:  true,
			}).Error; err != nil {
				s.logger.Error("ManagerAddSystemMsg, failed to upsert SystemMsg", zap.Error(err))
				return err
			} else {
				s.logger.Debug("ManagerAddSystemMsg, upsert SystemMsg  succeed")
			}
			return nil

		} else {
			return err
		}

	}
	return nil
}

func (s *MysqlLianmiRepository) ManagerDeleteSystemMsg(id uint) error {
	//采用事务同时删除
	var (
		gpWhere   = models.SystemMsg{}
		systemMsg models.SystemMsg
	)
	gpWhere.ID = id
	tx := s.base.GetTransaction()
	if err := tx.Where(&gpWhere).Delete(&systemMsg).Error; err != nil {
		s.logger.Error("删除系失败", zap.Error(err))
		tx.Rollback()
		return err
	}
	tx.Commit()
	return nil
}
