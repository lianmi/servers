package repositories

import (
	"fmt"
	// "time"

	// "github.com/golang/protobuf/proto"
	Global "github.com/lianmi/servers/api/proto/global"
	Order "github.com/lianmi/servers/api/proto/order"
	LMCommon "github.com/lianmi/servers/internal/common"

	"math"

	"github.com/lianmi/servers/internal/pkg/models"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

//获取某个商户的所有商品列表
func (s *MysqlLianmiRepository) GetProductsList(req *Order.ProductsListReq) (*Order.ProductsListResp, error) {
	var err error
	total := new(int64) //总页数
	pageIndex := int(req.Page)
	pageSize := int(req.Limit)

	columns := []string{"*"}
	orderBy := "updated_at desc"

	redisConn := s.redisPool.Get()
	defer redisConn.Close()

	var products []*models.Product
	var mod Order.Product
	wheres := make([]interface{}, 0)
	if req.ProductType > 0 {
		wheres = append(wheres, []interface{}{"product_type", "=", int(req.ProductType)})
	}
	if req.BusinessUsername != "" {
		wheres = append(wheres, []interface{}{"business_username", "=", req.BusinessUsername})
	}

	db2 := s.db
	db2, err = s.base.BuildQueryList(db2, wheres, columns, orderBy, pageIndex, pageSize)
	if err != nil {
		return nil, err
	}
	err = db2.Find(&products).Error

	if err != nil {
		s.logger.Error("Find错误", zap.Error(err))
		return nil, err
	}

	db2, err = s.base.BuildWhere(db2, wheres)
	if err != nil {
		s.logger.Error("BuildWhere错误", zap.Error(err))
		return nil, err
	}

	db2 = s.db
	db2.Model(&mod).Count(total)

	pages := math.Ceil(float64(*total) / float64(pageSize))

	resp := &Order.ProductsListResp{
		TotalPage: uint64(pages),
	}

	for _, product := range products {
		var thumbnail string
		if product.ShortVideo != "" {
			thumbnail = LMCommon.OSSUploadPicPrefix + product.ShortVideo + "?x-oss-process=video/snapshot,t_500,f_jpg,w_800,h_600"
		}
		oProduct := &Order.Product{
			ProductId:         product.ProductID,                       //商品ID
			Expire:            uint64(product.Expire),                  //商品过期时间
			ProductName:       product.ProductName,                     //商品名称
			ProductType:       Global.ProductType(product.ProductType), //商品种类类型  枚举
			SubType:           Global.LotteryType(product.SubType),     // 商品子 类型
			ProductDesc:       product.ProductDesc,                     //商品详细介绍
			ShortVideo:        product.ShortVideo,                      //商品短视频
			Thumbnail:         thumbnail,                               //商品短视频缩略图
			Price:             product.Price,                           //价格
			LeftCount:         product.LeftCount,                       //库存数量
			Discount:          product.Discount,                        //折扣 实际数字，例如: 0.95, UI显示为九五折
			DiscountDesc:      product.DiscountDesc,                    //折扣说明
			DiscountStartTime: uint64(product.DiscountStartTime),       //折扣开始时间
			DiscountEndTime:   uint64(product.DiscountEndTime),         //折扣结束时间
			AllowCancel:       product.AllowCancel,                     //是否允许撤单， 默认是可以，彩票类的不可以
		}

		if product.ProductPic1Large != "" {
			// 动态拼接
			oProduct.ProductPics = append(oProduct.ProductPics, &Order.ProductPic{
				Small:  LMCommon.OSSUploadPicPrefix + product.ProductPic1Large + "?x-oss-process=image/resize,w_50/quality,q_50",
				Middle: LMCommon.OSSUploadPicPrefix + product.ProductPic1Large + "?x-oss-process=image/resize,w_100/quality,q_100",
				Large:  LMCommon.OSSUploadPicPrefix + product.ProductPic1Large,
			})
		}

		if product.ProductPic2Large != "" {
			// 动态拼接
			oProduct.ProductPics = append(oProduct.ProductPics, &Order.ProductPic{
				Small:  LMCommon.OSSUploadPicPrefix + product.ProductPic2Large + "?x-oss-process=image/resize,w_50/quality,q_50",
				Middle: LMCommon.OSSUploadPicPrefix + product.ProductPic2Large + "?x-oss-process=image/resize,w_100/quality,q_100",
				Large:  LMCommon.OSSUploadPicPrefix + product.ProductPic2Large,
			})
		}

		if product.ProductPic3Large != "" {
			// 动态拼接
			oProduct.ProductPics = append(oProduct.ProductPics, &Order.ProductPic{
				Small:  LMCommon.OSSUploadPicPrefix + product.ProductPic3Large + "?x-oss-process=image/resize,w_50/quality,q_50",
				Middle: LMCommon.OSSUploadPicPrefix + product.ProductPic3Large + "?x-oss-process=image/resize,w_100/quality,q_100",
				Large:  LMCommon.OSSUploadPicPrefix + product.ProductPic3Large,
			})
		}

		if product.DescPic1 != "" {
			oProduct.DescPics = append(oProduct.DescPics, LMCommon.OSSUploadPicPrefix+product.DescPic1)
		}

		if product.DescPic2 != "" {
			oProduct.DescPics = append(oProduct.DescPics, LMCommon.OSSUploadPicPrefix+product.DescPic2)
		}

		if product.DescPic3 != "" {
			oProduct.DescPics = append(oProduct.DescPics, LMCommon.OSSUploadPicPrefix+product.DescPic3)
		}

		if product.DescPic4 != "" {
			oProduct.DescPics = append(oProduct.DescPics, LMCommon.OSSUploadPicPrefix+product.DescPic4)
		}

		if product.DescPic5 != "" {
			oProduct.DescPics = append(oProduct.DescPics, LMCommon.OSSUploadPicPrefix+product.DescPic5)
		}

		if product.DescPic6 != "" {
			oProduct.DescPics = append(oProduct.DescPics, LMCommon.OSSUploadPicPrefix+product.DescPic6)

		}

		resp.Products = append(resp.Products, oProduct)
	}
	return resp, nil

}

//设置商品的子类型
func (s *MysqlLianmiRepository) SetProductSubType(req *Order.ProductSetSubTypeReq) error {
	product := new(models.Product)
	where := models.Product{
		ProductInfo: models.ProductInfo{
			ProductID: req.ProductId,
		},
	}

	var curSubType int
	if err := s.db.Model(&models.Product{}).Where(&where).First(product).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			//记录找不到也会触发错误 记录不存在
			return errors.Wrapf(err, "Record not exists[productID=%s]", req.ProductId)
		} else {
			return errors.Wrapf(err, "Get Product info error[productID=%s]", req.ProductId)
		}

	} else {
		curSubType = product.SubType
	}

	result := s.db.Model(&models.Product{}).Where(&where).Update("sub_type", req.SubType)
	//updated records count
	s.logger.Debug("SetProductSubType result: ", zap.Int64("RowsAffected", result.RowsAffected), zap.Error(result.Error))

	if result.Error != nil {
		s.logger.Error("设置商品的子类型失败", zap.Error(result.Error))
		return result.Error
	} else {
		mtxt := fmt.Sprintf("设置商品的子类型成功:  旧的子类型: %d, 新的子类型: %d", curSubType, req.SubType)
		s.logger.Debug(mtxt)
	}

	return nil
}

//根据商品ID查询商品详情
func (s *MysqlLianmiRepository) GetProductInfo(productID string) (*Order.Product, error) {
	product := new(models.Product)
	where := models.Product{
		ProductInfo: models.ProductInfo{
			ProductID: productID,
		},
	}

	if err := s.db.Model(&models.Product{}).Where(&where).First(product).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			//记录找不到也会触发错误 记录不存在
			return nil, errors.Wrapf(err, "Record not exists[productID=%s]", productID)
		} else {
			return nil, errors.Wrapf(err, "Get Product info error[productID=%s]", productID)
		}

	}
	var thumbnail string
	if product.ShortVideo != "" {

		thumbnail = LMCommon.OSSUploadPicPrefix + product.ShortVideo + "?x-oss-process=video/snapshot,t_500,f_jpg,w_800,h_600"
	}
	oProduct := &Order.Product{
		ProductId:         productID,                               //商品ID
		Expire:            uint64(product.Expire),                  //商品过期时间
		ProductName:       product.ProductName,                     //商品名称
		ProductType:       Global.ProductType(product.ProductType), //商品种类类型  枚举
		SubType:           Global.LotteryType(product.SubType),     //商品子类型
		ProductDesc:       product.ProductDesc,                     //商品详细介绍
		ShortVideo:        product.ShortVideo,                      //商品短视频
		Thumbnail:         thumbnail,                               //商品短视频缩略图
		Price:             product.Price,                           //价格
		LeftCount:         product.LeftCount,                       //库存数量
		Discount:          product.Discount,                        //折扣 实际数字，例如: 0.95, UI显示为九五折
		DiscountDesc:      product.DiscountDesc,                    //折扣说明
		DiscountStartTime: uint64(product.DiscountStartTime),       //折扣开始时间
		DiscountEndTime:   uint64(product.DiscountEndTime),         //折扣结束时间
		AllowCancel:       product.AllowCancel,                     //是否允许撤单， 默认是可以，彩票类的不可以
	}

	if product.ProductPic1Large != "" {
		// 动态拼接
		oProduct.ProductPics = append(oProduct.ProductPics, &Order.ProductPic{
			Small:  LMCommon.OSSUploadPicPrefix + product.ProductPic1Large + "?x-oss-process=image/resize,w_50/quality,q_50",
			Middle: LMCommon.OSSUploadPicPrefix + product.ProductPic1Large + "?x-oss-process=image/resize,w_100/quality,q_100",
			Large:  LMCommon.OSSUploadPicPrefix + product.ProductPic1Large,
		})
	}

	if product.ProductPic2Large != "" {
		// 动态拼接
		oProduct.ProductPics = append(oProduct.ProductPics, &Order.ProductPic{
			Small:  LMCommon.OSSUploadPicPrefix + product.ProductPic2Large + "?x-oss-process=image/resize,w_50/quality,q_50",
			Middle: LMCommon.OSSUploadPicPrefix + product.ProductPic2Large + "?x-oss-process=image/resize,w_100/quality,q_100",
			Large:  LMCommon.OSSUploadPicPrefix + product.ProductPic2Large,
		})
	}

	if product.ProductPic3Large != "" {
		// 动态拼接
		oProduct.ProductPics = append(oProduct.ProductPics, &Order.ProductPic{
			Small:  LMCommon.OSSUploadPicPrefix + product.ProductPic3Large + "?x-oss-process=image/resize,w_50/quality,q_50",
			Middle: LMCommon.OSSUploadPicPrefix + product.ProductPic3Large + "?x-oss-process=image/resize,w_100/quality,q_100",
			Large:  LMCommon.OSSUploadPicPrefix + product.ProductPic3Large,
		})
	}

	if product.DescPic1 != "" {
		oProduct.DescPics = append(oProduct.DescPics, LMCommon.OSSUploadPicPrefix+product.DescPic1)
	}

	if product.DescPic2 != "" {
		oProduct.DescPics = append(oProduct.DescPics, LMCommon.OSSUploadPicPrefix+product.DescPic2)
	}

	if product.DescPic3 != "" {
		oProduct.DescPics = append(oProduct.DescPics, LMCommon.OSSUploadPicPrefix+product.DescPic3)
	}

	if product.DescPic4 != "" {
		oProduct.DescPics = append(oProduct.DescPics, LMCommon.OSSUploadPicPrefix+product.DescPic4)
	}

	if product.DescPic5 != "" {
		oProduct.DescPics = append(oProduct.DescPics, LMCommon.OSSUploadPicPrefix+product.DescPic5)
	}

	if product.DescPic6 != "" {
		oProduct.DescPics = append(oProduct.DescPics, LMCommon.OSSUploadPicPrefix+product.DescPic6)

	}

	return oProduct, nil

}
func (s *MysqlLianmiRepository) GetStoreProductLists(req *Order.ProductsListReq) (p *[]models.StoreProductItems, err error) {
	//panic("implement me")
	// 通过商户id 获取商品列表
	p = new([]models.StoreProductItems)
	err = s.db.Model(&models.StoreProductItems{}).Where(&models.StoreProductItems{StoreUUID: req.BusinessUsername}).Find(p).Error
	return
}
