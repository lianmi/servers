package repositories

import (
	"fmt"
	// "time"

	// "github.com/golang/protobuf/proto"
	"github.com/gomodule/redigo/redis"
	Auth "github.com/lianmi/servers/api/proto/auth"
	Global "github.com/lianmi/servers/api/proto/global"
	Order "github.com/lianmi/servers/api/proto/order"
	User "github.com/lianmi/servers/api/proto/user"
	LMCommon "github.com/lianmi/servers/internal/common"
	"github.com/lianmi/servers/internal/pkg/models"
	"github.com/pkg/errors"
	uuid "github.com/satori/go.uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

//修改或增加店铺资料
func (s *MysqlLianmiRepository) AddStore(req *User.Store) error {
	var err error
	store := new(models.Store)

	redisConn := s.redisPool.Get()
	defer redisConn.Close()

	//判断商户的注册id的合法性以及是否封禁等
	userData := new(models.User)

	userKey := fmt.Sprintf("userData:%s", req.BusinessUsername)
	if result, err := redis.Values(redisConn.Do("HGETALL", userKey)); err == nil {
		if err := redis.ScanStruct(result, userData); err != nil {

			s.logger.Error("错误：ScanStruct", zap.Error(err))
			return errors.Wrapf(err, "查询redis出错[Businessusername=%s]", req.BusinessUsername)

		}
	}
	// 判断是否是商户类型
	if userData.UserType != 2 {
		s.logger.Error("错误：此注册账号id不是商户类型")
		return errors.Wrapf(err, "此注册账号id不是商户类型[Businessusername=%s]", req.BusinessUsername)
	}

	//判断是否被封禁
	if userData.State == LMCommon.UserBlocked {
		s.logger.Debug("User is blocked", zap.String("Businessusername", req.BusinessUsername))
		return errors.Wrapf(err, "User is blocked[Businessusername=%s]", req.BusinessUsername)
	}

	//先查询对应的记录是否存在

	where := models.Store{
		BusinessUsername: req.BusinessUsername,
	}

	err = s.db.Model(&models.Store{}).Where(&where).First(&store).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			s.logger.Debug("记录不存在")
			store := models.Store{
				StoreUUID:         uuid.NewV4().String(),  //店铺的uuid
				StoreType:         int(req.StoreType),     //店铺类型,对应Global.proto里的StoreType枚举
				BusinessUsername:  req.BusinessUsername,   //商户注册号
				Introductory:      req.Introductory,       //商店简介 Text文本类型
				Province:          req.Province,           //省份, 如广东省
				City:              req.City,               //城市，如广州市
				County:            req.County,             //区，如天河区
				Street:            req.Street,             //街道
				Address:           req.Address,            //地址
				Branchesname:      req.Branchesname,       //网点名称
				LegalPerson:       req.LegalPerson,        //法人姓名
				LegalIdentityCard: req.LegalIdentityCard,  //法人身份证
				Longitude:         req.Longitude,          //商户地址的经度
				Latitude:          req.Latitude,           //商户地址的纬度
				WeChat:            req.Wechat,             //商户联系人微信号
				Keys:              req.Keys,               //商户经营范围搜索关键字
				LicenseURL:        req.BusinessLicenseUrl, //商户营业执照阿里云url
				AuditState:        0,                      //初始值
			}

			//如果没有记录，则增加
			if err := s.db.Clauses(clause.OnConflict{DoNothing: true}).Create(&store).Error; err != nil {
				s.logger.Error("AddStore, failed to upsert stores", zap.Error(err))
				return err
			} else {
				s.logger.Debug("AddStore, upsert stores succeed")
			}

		} else {

			s.logger.Error("修改或增加店铺资料失败 db err", zap.Error(err))
			return err

		}
	} else {

		s.logger.Debug("记录存在")

		if store.AuditState == 1 {
			return errors.Wrapf(err, "已经审核通过的不能修改资料[Businessusername=%s]", req.BusinessUsername)
		}

		where2 := models.Store{
			StoreUUID: store.StoreUUID,
		}
		// 同时更新多个字段
		result := s.db.Model(&models.Store{}).Where(&where2).Updates(models.Store{
			StoreType:         int(req.StoreType),     //店铺类型,对应Global.proto里的StoreType枚举
			BusinessUsername:  req.BusinessUsername,   //商户注册号
			Introductory:      req.Introductory,       //商店简介 Text文本类型
			Province:          req.Province,           //省份, 如广东省
			City:              req.City,               //城市，如广州市
			County:            req.County,             //区，如天河区
			Street:            req.Street,             //街道
			Address:           req.Address,            //地址
			Branchesname:      req.Branchesname,       //网点名称
			LegalPerson:       req.LegalPerson,        //法人姓名
			LegalIdentityCard: req.LegalIdentityCard,  //法人身份证
			Longitude:         req.Longitude,          //商户地址的经度
			Latitude:          req.Latitude,           //商户地址的纬度
			WeChat:            req.Wechat,             //商户联系人微信号
			Keys:              req.Keys,               //商户经营范围搜索关键字
			LicenseURL:        req.BusinessLicenseUrl, //商户营业执照阿里云url
		})

		//updated records count
		s.logger.Debug("修改店铺记录  result: ",
			zap.Int64("RowsAffected", result.RowsAffected),
			zap.Error(result.Error))

		if result.Error != nil {
			s.logger.Error("Update Store失败", zap.Error(result.Error))
			return result.Error
		} else {
			s.logger.Debug("Update Store成功")
		}

	}

	return nil

}

func (s *MysqlLianmiRepository) GetStore(businessUsername string) (*User.Store, error) {
	var err error
	p := new(models.Store)
	if err = s.db.Model(p).Where(&models.Store{
		BusinessUsername: businessUsername,
	}).First(p).Error; err != nil {
		s.logger.Error("MySQL里读取错误或记录不存在", zap.Error(err))
		return nil, errors.Wrapf(err, "Query error[BusinessUsername=%s]", businessUsername)
	}

	redisConn := s.redisPool.Get()
	defer redisConn.Close()

	//获取店铺头像
	avatar, err := redis.String(redisConn.Do("HGET", fmt.Sprintf("userData:%s", businessUsername), "Avatar"))
	if err != nil {
		s.logger.Error("HGET Avatar error", zap.Error(err))
		return nil, errors.Wrapf(err, "Query error[BusinessUsername=%s]", businessUsername)
	}

	return &User.Store{
		StoreUUID:          p.StoreUUID,                   //店铺的uuid
		StoreType:          Global.StoreType(p.StoreType), //店铺类型,对应Global.proto里的StoreType枚举
		BusinessUsername:   p.BusinessUsername,            //商户注册号
		Avatar:             avatar,                        //头像
		Introductory:       p.Introductory,                //商店简介 Text文本类型
		Province:           p.Province,                    //省份, 如广东省
		City:               p.City,                        //城市，如广州市
		County:             p.County,                      //区，如天河区
		Street:             p.Street,                      //街道
		Address:            p.Address,                     //地址
		Branchesname:       p.Branchesname,                //网点名称
		LegalPerson:        p.LegalPerson,                 //法人姓名
		LegalIdentityCard:  p.LegalIdentityCard,           //法人身份证
		Longitude:          p.Longitude,                   //商户地址的经度
		Latitude:           p.Latitude,                    //商户地址的纬度
		Wechat:             p.WeChat,                      //商户联系人微信号
		Keys:               p.Keys,                        //商户经营范围搜索关键字
		BusinessLicenseUrl: p.LicenseURL,                  //商户营业执照阿里云url
		AuditState:         int32(p.AuditState),           //审核状态，0-预审核，1-审核通过, 2-占位
		CreatedAt:          uint64(p.CreatedAt),
		UpdatedAt:          uint64(p.UpdatedAt),
	}, nil

}

//根据gps位置获取一定范围内的店铺列表
func (s *MysqlLianmiRepository) GetStores(req *Order.QueryStoresNearbyReq) (*Order.QueryStoresNearbyResp, error) {

	var err error
	total := new(int64) //总页数
	pageIndex := int(req.Page)
	pageSize := int(req.Limit)

	columns := []string{"*"}
	orderBy := "updated_at desc"

	redisConn := s.redisPool.Get()
	defer redisConn.Close()

	var list []*User.Store
	var mod User.Store
	wheres := make([]interface{}, 0)
	if req.StoreType > 0 {
		wheres = append(wheres, []interface{}{"store_type", "=", int(req.StoreType)})
	}
	if req.State > 0 {
		wheres = append(wheres, []interface{}{"state", "=", int(req.State)})
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

	resp := &Order.QueryStoresNearbyResp{
		TotalPage: uint64(*total),
	}

	for _, store := range list {
		//获取店铺头像
		avatar, _ := redis.String(redisConn.Do("HGET", fmt.Sprintf("userData:%s", store.BusinessUsername), "Avatar"))

		resp.Stores = append(resp.Stores, &User.Store{
			StoreUUID:          store.StoreUUID,                   //店铺的uuid
			StoreType:          Global.StoreType(store.StoreType), //店铺类型,对应Global.proto里的StoreType枚举
			BusinessUsername:   store.BusinessUsername,            //商户注册号
			Avatar:             avatar,                            //头像
			Introductory:       store.Introductory,                //商店简介 Text文本类型
			Province:           store.Province,                    //省份, 如广东省
			City:               store.City,                        //城市，如广州市
			County:             store.County,                      //区，如天河区
			Street:             store.Street,                      //街道
			Address:            store.Address,                     //地址
			Branchesname:       store.Branchesname,                //网点名称
			LegalPerson:        store.LegalPerson,                 //法人姓名
			LegalIdentityCard:  store.LegalIdentityCard,           //法人身份证
			Longitude:          store.Longitude,                   //商户地址的经度
			Latitude:           store.Latitude,                    //商户地址的纬度
			Wechat:             store.Wechat,                      //商户联系人微信号
			Keys:               store.Keys,                        //商户经营范围搜索关键字
			BusinessLicenseUrl: store.BusinessLicenseUrl,          //商户营业执照阿里云url
			AuditState:         store.AuditState,                  //审核状态，0-预审核，1-审核通过, 2-占位
			CreatedAt:          uint64(store.CreatedAt),
			UpdatedAt:          uint64(store.UpdatedAt),
		})
	}
	return resp, nil
}

//后台管理员将店铺审核通过, 将stores表里的对应的记录state设置为1
func (s *MysqlLianmiRepository) AuditStore(req *Auth.AuditStoreReq) error {
	var err error
	p := new(models.Store)

	redisConn := s.redisPool.Get()
	defer redisConn.Close()

	//判断商户的注册id的合法性以及是否封禁等
	userData := new(models.User)

	userKey := fmt.Sprintf("userData:%s", req.BusinessUsername)
	if result, err := redis.Values(redisConn.Do("HGETALL", userKey)); err == nil {
		if err := redis.ScanStruct(result, userData); err != nil {

			s.logger.Error("错误：ScanStruct", zap.Error(err))
			return errors.Wrapf(err, "查询redis出错[Businessusername=%s]", req.BusinessUsername)

		}
	}
	// 判断是否是商户类型
	if userData.UserType != 2 {
		s.logger.Error("错误：此注册账号id不是商户类型")
		return errors.Wrapf(err, "此注册账号id不是商户类型[Businessusername=%s]", req.BusinessUsername)
	}

	//判断是否被封禁
	if userData.State == LMCommon.UserBlocked {
		s.logger.Debug("User is blocked", zap.String("Businessusername", req.BusinessUsername))
		return errors.Wrapf(err, "User is blocked[Businessusername=%s]", req.BusinessUsername)
	}

	//先查询对应的记录是否存在
	where := &models.Store{
		BusinessUsername: req.BusinessUsername,
	}

	err = s.db.Model(&models.Store{}).Where(&where).First(p).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			s.logger.Error("错误：此商户没有提交资料")
			return errors.Wrapf(err, "此商户没有提交资料[Businessusername=%s]", req.BusinessUsername)
		} else {

			s.logger.Error("db err", zap.Error(err))
			return err

		}
	}

	//修改 audit_state 字段的值
	result := s.db.Model(&models.Store{}).Where(&models.Store{
		BusinessUsername: req.BusinessUsername,
	}).Update("audit_state", 1)

	//updated records count
	s.logger.Debug("AuditStore result: ", zap.Int64("RowsAffected", result.RowsAffected), zap.Error(result.Error))

	if result.Error != nil {
		return result.Error
	}
	return nil
}

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
