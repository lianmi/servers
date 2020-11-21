package repositories

import (
	"github.com/lianmi/servers/internal/pkg/models"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"gorm.io/gorm/clause"
)

//======通用商品======/

// 增加通用商品- Create
func (s *MysqlLianmiRepository) AddGeneralProduct(generalProduct *models.GeneralProduct) error {
	if generalProduct == nil {
		return errors.New("generalProduct is nil")
	}
	//如果没有记录，则增加，如果有记录，则更新全部字段
	if err := s.db.Clauses(clause.OnConflict{UpdateAll: true}).Create(&generalProduct).Error; err != nil {
		s.logger.Error("AddGeneralProduct, failed to upsert generalProduct", zap.Error(err))
		return err
	} else {
		s.logger.Debug("AddGeneralProduct, upsert generalProduct succeed")
	}

	return nil

}

//查询通用商品(id) - Read
func (s *MysqlLianmiRepository) GetGeneralProductByID(productID string) (p *models.GeneralProduct, err error) {
	p = new(models.GeneralProduct)

	if err = s.db.Model(p).Where(&models.GeneralProduct{
		ProductID: productID,
	}).First(p).Error; err != nil {
		//记录找不到也会触发错误
		return nil, errors.Wrapf(err, "Get GeneralProduct error[productID=%d]", productID)
	}
	s.logger.Debug("GetUser run...")
	return
}

//查询通用商品分页 - Page
func (s *MysqlLianmiRepository) GetGeneralProductPage(pageIndex, pageSize int, total *int64, where interface{}) ([]*models.GeneralProduct, error) {
	var generalProducts []*models.GeneralProduct
	if err := s.base.GetPages(&models.GeneralProduct{}, &generalProducts, pageIndex, pageSize, total, where); err != nil {
		s.logger.Error("获取通用商品分页失败", zap.Error(err))
		return nil, err
	}
	return generalProducts, nil

}

// 修改通用商品 - Update
func (s *MysqlLianmiRepository) UpdateGeneralProduct(generalProduct *models.GeneralProduct) error {

	if generalProduct == nil {
		return errors.New("generalProduct is nil")
	}

	where := models.GeneralProduct{ProductID: generalProduct.ProductID}
	// 同时更新多个字段
	result := s.db.Model(&models.GeneralProduct{}).Where(&where).Updates(generalProduct)

	//updated records count
	s.logger.Debug("UpdateGeneralProduct result: ",
		zap.Int64("RowsAffected", result.RowsAffected),
		zap.Error(result.Error))

	if result.Error != nil {
		s.logger.Error("修改通用商品失败", zap.Error(result.Error))
		return result.Error
	} else {
		s.logger.Debug("修改通用商品成功")
	}
	return nil

}

// 删除通用商品 - Delete
func (s *MysqlLianmiRepository) DeleteGeneralProduct(productID string) bool {

	//采用事务同时删除
	var (
		gpWhere        = models.GeneralProduct{ProductID: productID}
		generalProduct models.GeneralProduct
	)
	tx := s.base.GetTransaction()
	if err := tx.Where(&gpWhere).Delete(&generalProduct).Error; err != nil {
		s.logger.Error("删除通用商品失败", zap.Error(err))
		tx.Rollback()
		return false
	}
	tx.Commit()
	return true

}
