package repositories

import (
	Global "github.com/lianmi/servers/api/proto/global"
	Order "github.com/lianmi/servers/api/proto/order"
	LMCommon "github.com/lianmi/servers/internal/common"
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
	where := models.GeneralProduct{}
	if req.ProductType > 0 {
		where = models.GeneralProduct{
			ProductType: int(req.ProductType),
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
		var thumbnail string
		if generalProduct.ShortVideo != "" {
			thumbnail = LMCommon.OSSUploadPicPrefix + generalProduct.ShortVideo + "?x-oss-process=video/snapshot,t_500,f_jpg,w_800,h_600"
		}

		gProduct := &Order.GeneralProduct{
			ProductId:   generalProduct.ProductID,                       //通用商品ID
			ProductName: generalProduct.ProductName,                     //商品名称
			ProductType: Global.ProductType(generalProduct.ProductType), //商品种类类型  枚举
			ProductDesc: generalProduct.ProductDesc,                     //商品详细介绍
			Thumbnail:   thumbnail,                                      //商品短视频缩略图
			ShortVideo:  generalProduct.ShortVideo,                      //商品短视频
			CreateAt:    uint64(generalProduct.CreatedAt),               //创建时间
			ModifyAt:    uint64(generalProduct.ModifyAt),                //最后修改时间
			AllowCancel: generalProduct.AllowCancel,                     //是否允许撤单， 默认是可以，彩票类的不可以
		}

		if generalProduct.ProductPic1Large != "" {
			// 动态拼接
			gProduct.ProductPics = append(gProduct.ProductPics, &Order.ProductPic{
				Small:  LMCommon.OSSUploadPicPrefix + generalProduct.ProductPic1Large + "?x-oss-process=image/resize,w_50/quality,q_50",
				Middle: LMCommon.OSSUploadPicPrefix + generalProduct.ProductPic1Large + "?x-oss-process=image/resize,w_100/quality,q_100",
				Large:  LMCommon.OSSUploadPicPrefix + generalProduct.ProductPic1Large,
			})

		}

		if generalProduct.ProductPic2Large != "" {
			// 动态拼接
			gProduct.ProductPics = append(gProduct.ProductPics, &Order.ProductPic{
				Small:  LMCommon.OSSUploadPicPrefix + generalProduct.ProductPic2Large + "?x-oss-process=image/resize,w_50/quality,q_50",
				Middle: LMCommon.OSSUploadPicPrefix + generalProduct.ProductPic2Large + "?x-oss-process=image/resize,w_100/quality,q_100",
				Large:  LMCommon.OSSUploadPicPrefix + generalProduct.ProductPic2Large,
			})
		}

		if generalProduct.ProductPic3Large != "" {
			// 动态拼接
			gProduct.ProductPics = append(gProduct.ProductPics, &Order.ProductPic{
				Small:  LMCommon.OSSUploadPicPrefix + generalProduct.ProductPic3Large + "?x-oss-process=image/resize,w_50/quality,q_50",
				Middle: LMCommon.OSSUploadPicPrefix + generalProduct.ProductPic3Large + "?x-oss-process=image/resize,w_100/quality,q_100",
				Large:  LMCommon.OSSUploadPicPrefix + generalProduct.ProductPic3Large,
			})
		}

		//商品介绍图片，6张
		if generalProduct.DescPic1 != "" {
			gProduct.DescPics = append(gProduct.DescPics, LMCommon.OSSUploadPicPrefix+generalProduct.DescPic1)
		}
		if generalProduct.DescPic2 != "" {
			gProduct.DescPics = append(gProduct.DescPics, LMCommon.OSSUploadPicPrefix+generalProduct.DescPic2)
		}
		if generalProduct.DescPic3 != "" {
			gProduct.DescPics = append(gProduct.DescPics, LMCommon.OSSUploadPicPrefix+generalProduct.DescPic3)
		}
		if generalProduct.DescPic4 != "" {
			gProduct.DescPics = append(gProduct.DescPics, LMCommon.OSSUploadPicPrefix+generalProduct.DescPic4)
		}
		if generalProduct.DescPic5 != "" {
			gProduct.DescPics = append(gProduct.DescPics, LMCommon.OSSUploadPicPrefix+generalProduct.DescPic5)
		}
		if generalProduct.DescPic6 != "" {
			gProduct.DescPics = append(gProduct.DescPics, LMCommon.OSSUploadPicPrefix+generalProduct.DescPic6)
		}

		resp.Generalproducts = append(resp.Generalproducts)
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
