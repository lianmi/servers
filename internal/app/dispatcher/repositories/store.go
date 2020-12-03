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
	var imageUrl string
	var businessLicenseUrl string

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

	imageUrl = LMCommon.OSSUploadPicPrefix + req.ImageUrl
	businessLicenseUrl = LMCommon.OSSUploadPicPrefix + req.BusinessLicenseUrl

	err = s.db.Model(&models.Store{}).Where(&where).First(&store).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			s.logger.Debug("记录不存在")
			store := models.Store{
				StoreUUID:         uuid.NewV4().String(), //店铺的uuid
				StoreType:         int(req.StoreType),    //店铺类型,对应Global.proto里的StoreType枚举
				ImageURL:          imageUrl,
				BusinessUsername:  req.BusinessUsername,  //商户注册号
				Introductory:      req.Introductory,      //商店简介 Text文本类型
				Province:          req.Province,          //省份, 如广东省
				City:              req.City,              //城市，如广州市
				County:            req.County,            //区，如天河区
				Street:            req.Street,            //街道
				Address:           req.Address,           //地址
				Branchesname:      req.Branchesname,      //网点名称
				LegalPerson:       req.LegalPerson,       //法人姓名
				LegalIdentityCard: req.LegalIdentityCard, //法人身份证
				Longitude:         req.Longitude,         //商户地址的经度
				Latitude:          req.Latitude,          //商户地址的纬度
				WeChat:            req.Wechat,            //商户联系人微信号
				Keys:              req.Keys,              //商户经营范围搜索关键字
				LicenseURL:        businessLicenseUrl,    //商户营业执照阿里云url
				AuditState:        0,                     //初始值
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
			StoreType:         int(req.StoreType), //店铺类型,对应Global.proto里的StoreType枚举
			ImageURL:          imageUrl,
			BusinessUsername:  req.BusinessUsername,  //商户注册号
			Introductory:      req.Introductory,      //商店简介 Text文本类型
			Province:          req.Province,          //省份, 如广东省
			City:              req.City,              //城市，如广州市
			County:            req.County,            //区，如天河区
			Street:            req.Street,            //街道
			Address:           req.Address,           //地址
			Branchesname:      req.Branchesname,      //网点名称
			LegalPerson:       req.LegalPerson,       //法人姓名
			LegalIdentityCard: req.LegalIdentityCard, //法人身份证
			Longitude:         req.Longitude,         //商户地址的经度
			Latitude:          req.Latitude,          //商户地址的纬度
			WeChat:            req.Wechat,            //商户联系人微信号
			Keys:              req.Keys,              //商户经营范围搜索关键字
			LicenseURL:        businessLicenseUrl,    //商户营业执照阿里云url
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
		ImageUrl:           p.ImageURL,
		Introductory:       p.Introductory,      //商店简介 Text文本类型
		Province:           p.Province,          //省份, 如广东省
		City:               p.City,              //城市，如广州市
		County:             p.County,            //区，如天河区
		Street:             p.Street,            //街道
		Address:            p.Address,           //地址
		Branchesname:       p.Branchesname,      //网点名称
		LegalPerson:        p.LegalPerson,       //法人姓名
		LegalIdentityCard:  p.LegalIdentityCard, //法人身份证
		Longitude:          p.Longitude,         //商户地址的经度
		Latitude:           p.Latitude,          //商户地址的纬度
		Wechat:             p.WeChat,            //商户联系人微信号
		Keys:               p.Keys,              //商户经营范围搜索关键字
		BusinessLicenseUrl: p.LicenseURL,        //商户营业执照阿里云url
		AuditState:         int32(p.AuditState), //审核状态，0-预审核，1-审核通过, 2-占位
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
	if pageSize == 0 {
		pageSize = 20
	}

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
			ImageUrl:           store.ImageUrl,                    //头像
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

/*

用户对所有店铺的点赞数列表
使用HashMap数据结构，HashMap中的key为BusinessUsername，value为Set，Set中的值为用户Username，即HashMap<String, Set<String>>

用户点赞的店铺列表
使用HashMap数据结构，HashMap中的key为Username，value为Set，Set中的值为BusinessUsername，即HashMap<String, Set<String>>

*/
//获取某个用户对所有店铺点赞情况, UI会保存在本地表里,  UI主动发起同步
func (s *MysqlLianmiRepository) UserLikes(username string) (*User.UserLikesResp, error) {
	var err error
	var businessUsers []string
	rsp := &User.UserLikesResp{
		Username: username,
	}

	redisConn := s.redisPool.Get()
	defer redisConn.Close()

	userlikeKey := fmt.Sprintf("UserLike:%s", username)

	if businessUsers, err = redis.Strings(redisConn.Do("SMEMBERS", userlikeKey)); err != nil {
		s.logger.Error("SMEMBERS Error", zap.Error(err))
		return nil, err
	}

	for _, user := range businessUsers {
		rsp.Businessusernames = append(rsp.Businessusernames, user)
	}

	return rsp, nil
}

//获取店铺的所有点赞的用户列表
func (s *MysqlLianmiRepository) StoreLikes(businessUsername string) (*User.StoreLikesResp, error) {
	var err error
	var users []string
	rsp := &User.StoreLikesResp{
		BusinessUsername: businessUsername,
	}

	redisConn := s.redisPool.Get()
	defer redisConn.Close()

	storelikeKey := fmt.Sprintf("StoreLike:%s", businessUsername)

	if users, err = redis.Strings(redisConn.Do("SMEMBERS", storelikeKey)); err != nil {
		s.logger.Error("SMEMBERS Error", zap.Error(err))
		return nil, err
	}

	for _, user := range users {
		rsp.Usernames = append(rsp.Usernames, user)
	}

	return rsp, nil
}

//对某个店铺点赞，返回当前所有的点赞总数
func (s *MysqlLianmiRepository) ClickLike(username, businessUsername string) (int64, error) {

	var err error
	var totalLikeCount int64

	redisConn := s.redisPool.Get()
	defer redisConn.Close()

	//增加到此用户的店铺点赞列表
	userlikeKey := fmt.Sprintf("UserLike:%s", username)
	if _, err = redisConn.Do("SADD", userlikeKey, businessUsername); err != nil {
		s.logger.Error("SADD userlikeKey Error", zap.Error(err))
		return 0, err
	}

	//增加到店铺点赞用户列表
	storelikeKey := fmt.Sprintf("StoreLike:%s", businessUsername)
	if _, err = redisConn.Do("SADD", storelikeKey, username); err != nil {
		s.logger.Error("SADD storelikeKey Error", zap.Error(err))
		return 0, err
	}

	//此店铺的总点赞数， 包括其他用户的点赞
	totalLikeKey := fmt.Sprintf("TotalLikeCount:%s", businessUsername)
	if totalLikeCount, err = redis.Int64(redisConn.Do("INCR", totalLikeKey)); err != nil {
		s.logger.Error("INCR TotalLike Error", zap.Error(err))
		return 0, err
	}

	return totalLikeCount, nil
}

//取消对某个店铺点赞
func (s *MysqlLianmiRepository) DeleteClickLike(username, businessUsername string) (int64, error) {
	var err error
	var totalLikeCount int64

	redisConn := s.redisPool.Get()
	defer redisConn.Close()

	userlikeKey := fmt.Sprintf("UserLike:%s", username)
	if _, err = redisConn.Do("SREM", userlikeKey, businessUsername); err != nil {
		s.logger.Error("SREM Error", zap.Error(err))
		return 0, err
	}

	//删除店铺点赞用户列表
	storelikeKey := fmt.Sprintf("StoreLike:%s", username)
	if _, err = redisConn.Do("SREM", storelikeKey); err != nil {
		s.logger.Error("SADD storelikeKey Error", zap.Error(err))
		return 0, err
	}

	//此店铺的总点赞数
	totalLikeKey := fmt.Sprintf("TotalLikeCount:%s", businessUsername)
	if totalLikeCount, err = redis.Int64(redisConn.Do("DECR", totalLikeKey)); err != nil {
		s.logger.Error("DECR TotalLike Error", zap.Error(err))
		return 0, err
	}
	return totalLikeCount, nil
}

//将点赞记录插入到UserLike表
func (s *MysqlLianmiRepository) AddUserLike(username, businessUser string) error {
	userLike := &models.UserLike{
		Username:         username,
		BusinessUsername: businessUser,
	}
	//如果没有记录，则增加
	if err := s.db.Clauses(clause.OnConflict{DoNothing: true}).Create(&userLike).Error; err != nil {
		s.logger.Error("AddUserLike, failed to upsert UserLike", zap.Error(err))
		return err
	} else {
		s.logger.Debug("AddUserLike, upsert UserLike succeed")
	}
	return nil
}

//将用户对店铺的点赞记录插入到StoreLike表
func (s *MysqlLianmiRepository) AddStoreLike(businessUsername, user string) error {
	storeLike := &models.StoreLike{
		BusinessUsername: businessUsername,
		Username:         user,
	}
	//如果没有记录，则增加
	if err := s.db.Clauses(clause.OnConflict{DoNothing: true}).Create(&storeLike).Error; err != nil {
		s.logger.Error("AddStoreLike, failed to upsert UserLike", zap.Error(err))
		return err
	} else {
		s.logger.Debug("AddStoreLike, upsert UserLike succeed")
	}
	return nil
}
