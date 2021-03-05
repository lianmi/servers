package repositories

import (
	"github.com/lianmi/servers/internal/pkg/models"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

//用户类型 1-普通
func IsNormalUser(db *gorm.DB) *gorm.DB {
	return db.Where("user_type = ? ", 1)
}

//用户类型 2-商户  处于预审核状态
func IsPreBusinessUser(db *gorm.DB) *gorm.DB {
	return db.Where("user_type = ?  and  state = ?", 2, 0)
}

//用户类型 2-商户  处于已审核状态
func IsBusinessUser(db *gorm.DB) *gorm.DB {
	return db.Where("user_type = ?  and  state = ?", 2, 1)
}

//按createAt的时间段
func BetweenCreatedAt(startAt, endAt uint64) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("created_at>= ? and created_at<= ? ", startAt, endAt)
	}
}

//按createAt的时间段
func BetweenUpdatedAt(startAt, endAt uint64) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("updated_at>= ? and updated_at<= ? ", startAt, endAt)
	}
}

//实名
func TrueName(trueNames []string) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("true_name IN (?)", trueNames)
	}
}

//法人
func LegalPerson(legalPersons []string) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("legal_person IN (?)", legalPersons)
	}
}

//店铺名称
func Branchesname(branchesnames []string) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("branchesname IN (?)", branchesnames)
	}
}

func Paginate(page, pageSize int) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if page == 0 {
			page = 1
		}

		switch {
		case pageSize > 100:
			pageSize = 20
		case pageSize <= 0:
			pageSize = 10
		}

		offset := (page - 1) * pageSize
		return db.Offset(offset).Limit(pageSize)
	}
}

//分页查询lnmc_collection_histories列表, pageSize default 20
func (s *MysqlWalletRepository) GetCollectionHistorys(toUsername, fromUsername string, startAt, endAt uint64, pageNum int, pageSize int, total *int64) ([]*models.LnmcCollectionHistory, error) {

	var count int64
	var lnmcCollectionHistories []*models.LnmcCollectionHistory
	model := new(models.LnmcCollectionHistory)

	s.db.Model(&model).Find(&lnmcCollectionHistories, "to_username = ? AND from_username=?", toUsername, fromUsername)

	resultError := s.db.Model(&model).Scopes(BetweenCreatedAt(startAt, endAt), Paginate(pageNum, pageSize)).Order("created_at DESC").Find(&lnmcCollectionHistories)

	count = int64(len(lnmcCollectionHistories))
	total = &count
	s.logger.Debug("分页查询lnmc_collection_histories列表 ", zap.Int64("count", count))

	for _, history := range lnmcCollectionHistories {
		s.logger.Debug("分页显示lnmc_collection_histories列表 ",
			zap.Int64("CreatedAt", history.CreatedAt),
			zap.String("ToUsername", history.ToUsername),
			zap.String("FromUsername", history.FromUsername),
			zap.String("OrderID", history.OrderID),
		)
	}
	if resultError.Error != nil {
		return nil, resultError.Error
	}

	return lnmcCollectionHistories, nil

}
