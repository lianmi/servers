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

	var list []*Order.Product
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
	err = db.Find(&list).Error

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

	resp := &Order.ProductsListResp{
		TotalPage: uint64(*total),
	}

	for _, product := range list {

		resp.Products = append(resp.Products, &Order.Product{
			ProductId:         product.ProductId,         //商品ID
			Expire:            product.Expire,            //商品过期时间
			ProductName:       product.ProductName,       //商品名称
			ProductType:       product.ProductType,       //商品种类类型  枚举
			ProductDesc:       product.ProductDesc,       //商品详细介绍
			ProductPic1Small:  product.ProductPic1Small,  //商品图片1-小图
			ProductPic1Middle: product.ProductPic1Middle, //商品图片1-中图
			ProductPic1Large:  product.ProductPic1Large,  //商品图片1-大图
			ProductPic2Small:  product.ProductPic2Small,  //商品图片2-小图
			ProductPic2Middle: product.ProductPic2Middle, //商品图片2-中图
			ProductPic2Large:  product.ProductPic2Large,  //商品图片2-大图
			ProductPic3Small:  product.ProductPic3Small,  //商品图片3-小图
			ProductPic3Middle: product.ProductPic3Middle, //商品图片3-中图
			ProductPic3Large:  product.ProductPic3Large,  //商品图片3-大图
			Thumbnail:         product.Thumbnail,         //商品短视频缩略图
			ShortVideo:        product.ShortVideo,        //商品短视频
			Price:             product.Price,             //价格
			LeftCount:         product.LeftCount,         //库存数量
			Discount:          product.Discount,          //折扣 实际数字，例如: 0.95, UI显示为九五折
			DiscountDesc:      product.DiscountDesc,      //折扣说明
			DiscountStartTime: product.DiscountStartTime, //折扣开始时间
			DiscountEndTime:   product.DiscountEndTime,   //折扣结束时间
			CreateAt:          product.CreateAt,          //创建时间
			ModifyAt:          product.ModifyAt,          //最后修改时间
			AllowCancel:       product.AllowCancel,       //是否允许撤单， 默认是可以，彩票类的不可以
		})
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

	return &Order.Product{
		ProductId:         productID,
		Expire:            uint64(product.Expire),
		ProductName:       product.ProductName,
		ProductType:       Global.ProductType(product.ProductType),
		ProductDesc:       product.ProductDesc,
		ProductPic1Small:  product.ProductPic1Small,
		ProductPic1Middle: product.ProductPic1Middle,
		ProductPic1Large:  product.ProductPic1Large,

		ProductPic2Small:  product.ProductPic2Small,
		ProductPic2Middle: product.ProductPic2Middle,
		ProductPic2Large:  product.ProductPic2Large,

		ProductPic3Small:  product.ProductPic3Small,
		ProductPic3Middle: product.ProductPic3Middle,
		ProductPic3Large:  product.ProductPic3Large,

		Thumbnail:         product.Thumbnail,
		ShortVideo:        product.ShortVideo,
		Price:             product.Price,
		LeftCount:         product.LeftCount,
		Discount:          product.Discount,
		DiscountDesc:      product.DiscountDesc,
		DiscountStartTime: uint64(product.DiscountStartTime),
		DiscountEndTime:   uint64(product.DiscountEndTime),
		CreateAt:          uint64(product.CreatedAt),
		ModifyAt:          uint64(product.ModifyAt),
		AllowCancel:       product.AllowCancel,
	}, nil

}