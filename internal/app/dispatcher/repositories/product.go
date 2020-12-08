package repositories

import (
	// "time"

	// "github.com/golang/protobuf/proto"
	Global "github.com/lianmi/servers/api/proto/global"
	Order "github.com/lianmi/servers/api/proto/order"
	"github.com/lianmi/servers/internal/pkg/models"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"math"
)

//获取某个商户的所有商品列表
func (s *MysqlLianmiRepository) GetProductsList(req *Order.ProductsListReq) (*Order.ProductsListResp, error) {
	var err error
	total := new(int64) //总页数
	pageIndex := int(req.Page)
	pageSize := int(req.Limit)

	columns := []string{"*"}
	orderBy := "modify_at desc"

	redisConn := s.redisPool.Get()
	defer redisConn.Close()

	var products []*models.Product
	var mod Order.Product
	wheres := make([]interface{}, 0)
	if req.ProductType > 0 {
		wheres = append(wheres, []interface{}{"product_type", "=", int(req.ProductType)})
	}

	db := s.db
	db, err = s.base.BuildQueryList(db, wheres, columns, orderBy, pageIndex, pageSize)
	if err != nil {
		return nil, err
	}
	err = db.Find(&products).Error

	if err != nil {
		s.logger.Error("Find错误", zap.Error(err))
		return nil, err
	}

	db, err = s.base.BuildWhere(db, wheres)
	if err != nil {
		s.logger.Error("BuildWhere错误", zap.Error(err))
		return nil, err
	}

	db = s.db
	db.Model(&mod).Count(total)

	pages := math.Ceil(float64(*total) / float64(pageSize))

	resp := &Order.ProductsListResp{
		TotalPage: uint64(pages),
	}

	for _, product := range products {
		oProduct := &Order.Product{
			ProductId:         product.ProductID,                       //商品ID
			Expire:            uint64(product.Expire),                  //商品过期时间
			ProductName:       product.ProductName,                     //商品名称
			ProductType:       Global.ProductType(product.ProductType), //商品种类类型  枚举
			ProductDesc:       product.ProductDesc,                     //商品详细介绍
			ShortVideo:        product.ShortVideo,                      //商品短视频
			Thumbnail:         product.Thumbnail,                       //商品短视频缩略图
			Price:             product.Price,                           //价格
			LeftCount:         product.LeftCount,                       //库存数量
			Discount:          product.Discount,                        //折扣 实际数字，例如: 0.95, UI显示为九五折
			DiscountDesc:      product.DiscountDesc,                    //折扣说明
			DiscountStartTime: uint64(product.DiscountStartTime),       //折扣开始时间
			DiscountEndTime:   uint64(product.DiscountEndTime),         //折扣结束时间
			CreateAt:          uint64(product.CreatedAt),               //创建时间
			ModifyAt:          uint64(product.ModifyAt),                //最后修改时间
			AllowCancel:       product.AllowCancel,                     //是否允许撤单， 默认是可以，彩票类的不可以
		}
		oProduct.ProductPics = append(oProduct.ProductPics, &Order.ProductPic{
			Small:  product.ProductPic1Small,
			Middle: product.ProductPic1Middle,
			Large:  product.ProductPic1Large,
		})

		oProduct.ProductPics = append(oProduct.ProductPics, &Order.ProductPic{
			Small:  product.ProductPic2Small,
			Middle: product.ProductPic2Middle,
			Large:  product.ProductPic2Large,
		})

		oProduct.ProductPics = append(oProduct.ProductPics, &Order.ProductPic{
			Small:  product.ProductPic3Small,
			Middle: product.ProductPic3Middle,
			Large:  product.ProductPic3Large,
		})

		oProduct.DescPics = append(oProduct.DescPics, product.DescPic1)
		oProduct.DescPics = append(oProduct.DescPics, product.DescPic2)
		oProduct.DescPics = append(oProduct.DescPics, product.DescPic3)
		oProduct.DescPics = append(oProduct.DescPics, product.DescPic4)
		oProduct.DescPics = append(oProduct.DescPics, product.DescPic5)
		oProduct.DescPics = append(oProduct.DescPics, product.DescPic6)

		resp.Products = append(resp.Products, oProduct)
	}
	return resp, nil

}

func (s *MysqlLianmiRepository) GetProductInfo(productID string) (*Order.Product, error) {
	product := new(models.Product)
	where := models.Product{
		ProductID: productID,
	}

	if err := s.db.Model(&models.Product{}).Where(&where).First(product).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			//记录找不到也会触发错误 记录不存在
			return nil, errors.Wrapf(err, "Record not exists[productID=%s]", productID)
		} else {
			return nil, errors.Wrapf(err, "Get Product info error[productID=%s]", productID)
		}

	}
	oProduct := &Order.Product{
		ProductId:         productID,                               //商品ID
		Expire:            uint64(product.Expire),                  //商品过期时间
		ProductName:       product.ProductName,                     //商品名称
		ProductType:       Global.ProductType(product.ProductType), //商品种类类型  枚举
		ProductDesc:       product.ProductDesc,                     //商品详细介绍
		ShortVideo:        product.ShortVideo,                      //商品短视频
		Thumbnail:         product.Thumbnail,                       //商品短视频缩略图
		Price:             product.Price,                           //价格
		LeftCount:         product.LeftCount,                       //库存数量
		Discount:          product.Discount,                        //折扣 实际数字，例如: 0.95, UI显示为九五折
		DiscountDesc:      product.DiscountDesc,                    //折扣说明
		DiscountStartTime: uint64(product.DiscountStartTime),       //折扣开始时间
		DiscountEndTime:   uint64(product.DiscountEndTime),         //折扣结束时间
		CreateAt:          uint64(product.CreatedAt),               //创建时间
		ModifyAt:          uint64(product.ModifyAt),                //最后修改时间
		AllowCancel:       product.AllowCancel,                     //是否允许撤单， 默认是可以，彩票类的不可以
	}
	oProduct.ProductPics = append(oProduct.ProductPics, &Order.ProductPic{
		Small:  product.ProductPic1Small,
		Middle: product.ProductPic1Middle,
		Large:  product.ProductPic1Large,
	})

	oProduct.ProductPics = append(oProduct.ProductPics, &Order.ProductPic{
		Small:  product.ProductPic2Small,
		Middle: product.ProductPic2Middle,
		Large:  product.ProductPic2Large,
	})

	oProduct.ProductPics = append(oProduct.ProductPics, &Order.ProductPic{
		Small:  product.ProductPic3Small,
		Middle: product.ProductPic3Middle,
		Large:  product.ProductPic3Large,
	})
	oProduct.DescPics = append(oProduct.DescPics, product.DescPic1)
	oProduct.DescPics = append(oProduct.DescPics, product.DescPic2)
	oProduct.DescPics = append(oProduct.DescPics, product.DescPic3)
	oProduct.DescPics = append(oProduct.DescPics, product.DescPic4)
	oProduct.DescPics = append(oProduct.DescPics, product.DescPic5)
	oProduct.DescPics = append(oProduct.DescPics, product.DescPic6)

	return oProduct, nil

}
