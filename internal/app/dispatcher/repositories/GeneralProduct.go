package repositories

import (
	Global "github.com/lianmi/servers/api/proto/global"
	Order "github.com/lianmi/servers/api/proto/order"
	"github.com/lianmi/servers/internal/pkg/models"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

//======通用商品======/

// 增加通用商品- Create
func (s *MysqlLianmiRepository) AddGeneralProduct(generalProduct *models.GeneralProduct) error {
	var err error
	if generalProduct == nil {
		return errors.New("generalProduct is nil")
	}
	//判断ProductName是否存在， 如果存在，则无法增加
	where := models.GeneralProduct{
		ProductName: generalProduct.ProductName,
	}

	p := new(models.GeneralProduct)

	if err = s.db.Model(p).Where(&where).First(p).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			//记录找不到也会触发错误 记录不存在
		} else {
			return errors.Wrapf(err, "Get GeneralProduct info error[ProductName=%s]", generalProduct.ProductName)
		}

	}

	//如果没有记录，则增加，如果有记录，则更新全部字段
	if err := s.db.Clauses(clause.OnConflict{DoNothing: true}).Create(&generalProduct).Error; err != nil {
		s.logger.Error("AddGeneralProduct, failed to upsert generalProduct", zap.Error(err))
		return err
	} else {
		s.logger.Debug("AddGeneralProduct, upsert generalProduct succeed")
	}

	return nil

}

//查询通用商品(productID) - Read
func (s *MysqlLianmiRepository) GetGeneralProductByID(productID string) (p *models.GeneralProduct, err error) {
	p = new(models.GeneralProduct)

	if err = s.db.Model(p).Where(&models.GeneralProduct{
		ProductID: productID,
	}).First(p).Error; err != nil {
		//记录找不到也会触发错误
		return nil, errors.Wrapf(err, "Get GeneralProduct error[productID=%s]", productID)
	}

	s.logger.Debug("GetUser run...")
	return
}

//查询通用商品分页 - Page
func (s *MysqlLianmiRepository) GetGeneralProductPage(req *Order.GetGeneralProductPageReq) (*Order.GetGeneralProductPageResp, error) {
	var generalProducts []*models.GeneralProduct
	pageIndex := int(req.Page)
	if pageIndex == 0 {
		pageIndex = 1
	}
	pageSize := int(req.Limit)
	if pageSize == 0 {
		pageSize = 20
	}
	total := new(int64)
	where := Order.GeneralProduct{}
	if req.ProductType > 0 {
		where = Order.GeneralProduct{
			ProductType: Global.ProductType(req.ProductType),
		}
	}
	if err := s.base.GetPages(&models.GeneralProduct{}, &generalProducts, pageIndex, pageSize, total, &where); err != nil {
		s.logger.Error("获取通用商品分页失败", zap.Error(err))
		return nil, err
	}
	resp := &Order.GetGeneralProductPageResp{
		TotalPage: uint64(*total),
	}
	for _, generalProduct := range generalProducts {
		resp.Generalproducts = append(resp.Generalproducts, &Order.GeneralProduct{
			ProductId:         generalProduct.ProductID,                       //通用商品ID
			ProductName:       generalProduct.ProductName,                     //商品名称
			ProductType:       Global.ProductType(generalProduct.ProductType), //商品种类类型  枚举
			ProductDesc:       generalProduct.ProductDesc,                     //商品详细介绍
			ProductPic1Small:  generalProduct.ProductPic1Small,                //商品图片1-小图
			ProductPic1Middle: generalProduct.ProductPic1Middle,               //商品图片1-中图
			ProductPic1Large:  generalProduct.ProductPic1Large,                //商品图片1-大图
			ProductPic2Small:  generalProduct.ProductPic2Small,                //商品图片2-小图
			ProductPic2Middle: generalProduct.ProductPic2Middle,               //商品图片2-中图
			ProductPic2Large:  generalProduct.ProductPic2Large,                //商品图片2-大图
			ProductPic3Small:  generalProduct.ProductPic3Small,                //商品图片3-小图
			ProductPic3Middle: generalProduct.ProductPic3Small,                //商品图片3-中图
			ProductPic3Large:  generalProduct.ProductPic3Large,                //商品图片3-大图
			Thumbnail:         generalProduct.Thumbnail,                       //商品短视频缩略图
			ShortVideo:        generalProduct.ShortVideo,                      //商品短视频
			CreateAt:          uint64(generalProduct.CreatedAt),               //创建时间
			ModifyAt:          uint64(generalProduct.ModifyAt),                //最后修改时间
			AllowCancel:       generalProduct.AllowCancel,                     //是否允许撤单， 默认是可以，彩票类的不可以
		})
	}

	return resp, nil

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
