package repositories

import (
	// "encoding/json"
	// "fmt"
	// "time"

	// "github.com/golang/protobuf/proto"
	// "github.com/gomodule/redigo/redis"
	// // "github.com/jinzhu/gorm"
	// Auth "github.com/lianmi/servers/api/proto/auth"
	// pb "github.com/lianmi/servers/api/proto/user"
	// "github.com/lianmi/servers/internal/app/authservice/nsqMq"
	// "github.com/lianmi/servers/internal/common"
	"github.com/lianmi/servers/internal/pkg/models"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

//======后台相关======/

// 增加通用商品- Create
func (s *MysqlLianmiRepository) AddGeneralProduct(generalProduct *models.GeneralProduct) error {
	if err := s.base.Create(generalProduct); err != nil {
		s.logger.Error("db写入错误，增加GeneralProduct表失败")
		return err
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
		// fmt.Println("GetUser first error:", err.Error())
		return nil, errors.Wrapf(err, "Get GeneralProduct error[productID=%d]", productID)
	}
	s.logger.Debug("GetUser run...")
	return
}

//查询通用商品分页 - Page
func (s *MysqlLianmiRepository) GetGeneralProductPage(pageIndex, pageSize int, total *uint64, where interface{}) ([]*models.GeneralProduct, error) {
	var generalProducts []*models.GeneralProduct
	if err := s.base.GetPages(&models.GeneralProduct{}, &generalProducts, pageIndex, pageSize, total, where); err != nil {
		s.logger.Error("获取通用商品分页失败", zap.Error(err))
		return nil, err
	}
	return generalProducts, nil

}

// 修改通用商品 - Update
func (s *MysqlLianmiRepository) UpdateGeneralProduct(generalProduct *models.GeneralProduct) error {
	//使用事务批量保存
	tx := s.base.GetTransaction()

	if err := tx.Save(generalProduct).Error; err != nil {
		s.logger.Error("保存GeneralProduct表失败", zap.Error(err))
		tx.Rollback()
		return err
	}

	//提交
	tx.Commit()
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
