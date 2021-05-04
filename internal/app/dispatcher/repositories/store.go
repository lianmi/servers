package repositories

import (
	"fmt"
	"math/rand"
	"time"

	"strings"

	// "github.com/golang/protobuf/proto"
	"math"

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
	userData := new(models.UserBase)

	userKey := fmt.Sprintf("userData:%s", req.BusinessUsername)
	if result, err := redis.Values(redisConn.Do("HGETALL", userKey)); err == nil {
		if err := redis.ScanStruct(result, userData); err != nil {

			s.logger.Error("错误：ScanStruct", zap.Error(err))
			return errors.Wrapf(err, "查询redis出错[Businessusername=%s]", req.BusinessUsername)

		}
	}
	// 判断是否是商户类型
	if userData.UserType != int(User.UserType_Ut_Business) {
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
				StoreUUID:             uuid.NewV4().String(), //店铺的uuid
				StoreType:             int(req.StoreType),    //店铺类型,对应Global.proto里的StoreType枚举
				ImageURL:              req.ImageUrl,
				BusinessUsername:      req.BusinessUsername,      //商户注册号
				Introductory:          req.Introductory,          //商店简介 Text文本类型
				Province:              req.Province,              //省份, 如广东省
				City:                  req.City,                  //城市，如广州市
				Area:                  req.Area,                  //区，如天河区
				Street:                req.Street,                //街道
				Address:               req.Address,               //地址
				Branchesname:          req.Branchesname,          //网点名称
				LegalPerson:           req.LegalPerson,           //法人姓名
				LegalIdentityCard:     req.LegalIdentityCard,     //法人身份证
				Longitude:             req.Longitude,             //商户地址的经度
				Latitude:              req.Latitude,              //商户地址的纬度
				ContactMobile:         req.ContactMobile,         //联系手机
				WeChat:                req.Wechat,                //商户联系人微信号
				Keys:                  req.Keys,                  //商户经营范围搜索关键字
				LicenseURL:            req.BusinessLicenseUrl,    //商户营业执照阿里云url
				BusinessCode:          req.BusinessCode,          //网点编码
				NotaryServiceUsername: req.NotaryServiceUsername, //网点的公证注册id
				AuditState:            0,                         //初始值
				OpeningHours:          req.OpeningHours,          //营业时间
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

		s.logger.Debug("记录存在", zap.Int("AuditState", store.AuditState))

		// if store.AuditState == 1 {
		// 	s.logger.Debug("已经审核通过的不能修改资料", zap.Int("AuditState", store.AuditState))
		// 	return errors.New("已经审核通过的不能修改资料")
		// }

		where2 := models.Store{
			StoreUUID: store.StoreUUID,
		}
		// 同时更新多个字段
		result := s.db.Model(&models.Store{}).Where(&where2).Updates(models.Store{
			StoreType:             int(req.StoreType),        //店铺类型,对应Global.proto里的StoreType枚举
			ImageURL:              req.ImageUrl,              //店铺外景照片或形象图片
			BusinessUsername:      req.BusinessUsername,      //商户注册号
			Introductory:          req.Introductory,          //商店简介 Text文本类型
			Province:              req.Province,              //省份, 如广东省
			City:                  req.City,                  //城市，如广州市
			Area:                  req.Area,                  //区，如天河区
			Street:                req.Street,                //街道
			Address:               req.Address,               //地址
			Branchesname:          req.Branchesname,          //网点名称
			LegalPerson:           req.LegalPerson,           //法人姓名
			LegalIdentityCard:     req.LegalIdentityCard,     //法人身份证
			Longitude:             req.Longitude,             //商户地址的经度
			Latitude:              req.Latitude,              //商户地址的纬度
			WeChat:                req.Wechat,                //商户联系人微信号
			Keys:                  req.Keys,                  //商户经营范围搜索关键字
			LicenseURL:            req.BusinessLicenseUrl,    //商户营业执照阿里云url
			OpeningHours:          req.OpeningHours,          //营业时间
			ContactMobile:         req.ContactMobile,         //联系电话
			BusinessCode:          req.BusinessCode,          //商户的网点编码，适合彩票店或连锁网点
			NotaryServiceUsername: req.NotaryServiceUsername, //商户对应的公证处注册id
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

//更新商店
func (s *MysqlLianmiRepository) UpdateStore(username string, store *models.Store) error {
	where := models.Store{
		Branchesname: username,
	}
	// 同时更新多个字段
	result := s.db.Model(&models.Store{}).Where(&where).Updates(store)

	//updated records count
	s.logger.Debug("UpdateStore result: ",
		zap.Int64("RowsAffected", result.RowsAffected),
		zap.Error(result.Error))

	if result.Error != nil {
		s.logger.Error("UpdateStore, 修改店铺资料数据失败", zap.Error(result.Error))
		return result.Error
	} else {
		s.logger.Error("UpdateStore, 修改店铺资料数据成功")
	}
	return nil
}

func (s *MysqlLianmiRepository) GetStore(businessUsername string) (*User.Store, error) {
	var err error

	redisConn := s.redisPool.Get()
	defer redisConn.Close()
	// 判断businessUsername是否是商户

	//用户类型 1-普通，2-商户
	userType, _ := redis.Int(redisConn.Do("HGET", fmt.Sprintf("userData:%s", businessUsername), "UserType"))
	if userType != int(User.UserType_Ut_Business) {
		return nil, errors.Wrap(err, "此用户非商户类型")
	}
	p := new(models.Store)
	if err = s.db.Model(p).Where(&models.Store{
		BusinessUsername: businessUsername,
	}).First(p).Error; err != nil {
		s.logger.Error("MySQL里读取错误或记录不存在", zap.Error(err))
		return nil, errors.Wrapf(err, "Query error[BusinessUsername=%s]", businessUsername)
	}

	//获取店铺头像
	avatar, err := redis.String(redisConn.Do("HGET", fmt.Sprintf("userData:%s", businessUsername), "Avatar"))
	if err != nil {
		s.logger.Error("HGET Avatar error", zap.Error(err))
		return nil, errors.Wrapf(err, "Query error[BusinessUsername=%s]", businessUsername)
	}
	//智能判断一下，是否是带 http(s) 前缀
	if !strings.HasPrefix(avatar, "http") {

		avatar = LMCommon.OSSUploadPicPrefix + avatar //拼接URL

	}

	var imageURL, licenseURL string
	if p.ImageURL != "" {
		imageURL = LMCommon.OSSUploadPicPrefix + p.ImageURL
	}

	if p.LicenseURL != "" {
		licenseURL = LMCommon.OSSUploadPicPrefix + p.LicenseURL
	}
	return &User.Store{
		StoreUUID:             p.StoreUUID,                   //店铺的uuid
		StoreType:             Global.StoreType(p.StoreType), //店铺类型,对应Global.proto里的StoreType枚举
		BusinessUsername:      p.BusinessUsername,            //商户注册号
		Avatar:                avatar,                        //头像
		ImageUrl:              imageURL,                      //店铺形象图片
		Introductory:          p.Introductory,                //商店简介 Text文本类型
		Province:              p.Province,                    //省份, 如广东省
		City:                  p.City,                        //城市，如广州市
		Area:                  p.Area,                        //区，如天河区
		Street:                p.Street,                      //街道
		Address:               p.Address,                     //地址
		Branchesname:          p.Branchesname,                //网点名称
		LegalPerson:           p.LegalPerson,                 //法人姓名
		LegalIdentityCard:     p.LegalIdentityCard,           //法人身份证
		Longitude:             p.Longitude,                   //商户地址的经度
		Latitude:              p.Latitude,                    //商户地址的纬度
		Wechat:                p.WeChat,                      //商户联系人微信号
		Keys:                  p.Keys,                        //商户经营范围搜索关键字
		BusinessLicenseUrl:    licenseURL,                    //商户营业执照阿里云url
		AuditState:            int32(p.AuditState),           //审核状态，0-预审核，1-审核通过, 2-占位
		OpeningHours:          p.OpeningHours,                //营业时间
		BusinessCode:          p.BusinessCode,
		NotaryServiceUsername: p.NotaryServiceUsername, //第三方公证
	}, nil

}

//根据gps位置获取一定范围内的店铺列表
func (s *MysqlLianmiRepository) GetStores(req *Order.QueryStoresNearbyReq) (*Order.QueryStoresNearbyResp, error) {

	var err error

	redisConn := s.redisPool.Get()
	defer redisConn.Close()

	total := new(int64) //总页数
	pageIndex := int(req.Page)
	pageSize := int(req.Limit)
	if pageSize == 0 {
		pageSize = 20
	}

	columns := []string{"*"}
	orderBy := "updated_at desc"

	var list []*User.Store
	var mod User.Store
	wheres := make([]interface{}, 0)
	if req.StoreType > 0 {
		wheres = append(wheres, []interface{}{"store_type", "=", int(req.StoreType)})
	}

	//审核状态
	if req.State > 0 {
		wheres = append(wheres, []interface{}{"audit_state", "=", int(req.State)})
	}

	if req.Province != "" {
		wheres = append(wheres, []interface{}{"province", "=", req.Province})
	}
	if req.City != "" {
		wheres = append(wheres, []interface{}{"city", "=", req.City})
	}
	if req.Area != "" {
		wheres = append(wheres, []interface{}{"area", "=", req.Area})
	}

	if req.Keys != "" && req.Address != "" {
		// wheres = append(wheres, []interface{}{"keys", "like", "%" + req.Keys + "%"})
		wheres = append(wheres, []interface{}{"address like ? or keys like ?", "%" + req.Address + "%", "%" + req.Keys + "%"})
	} else if req.Keys == "" && req.Address != "" {
		wheres = append(wheres, []interface{}{"address like ? ", "%" + req.Address + "%"})
	} else if req.Keys != "" && req.Address == "" {
		wheres = append(wheres, []interface{}{"keys like ? ", "%" + req.Keys + "%"})
	}

	db2 := s.db
	db2, err = s.base.BuildQueryList(db2, wheres, columns, orderBy, pageIndex, pageSize)
	if err != nil {
		return nil, err
	}
	err = db2.Find(&list).Error

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

	s.logger.Debug("math.Ceil",
		zap.Float64("float64(*total)", float64(*total)),
		zap.Float64("float64(pageSize)", float64(pageSize)),
		zap.Float64("pages", pages),
	)

	resp := &Order.QueryStoresNearbyResp{
		TotalPage: uint64(pages),
	}

	for _, store := range list {
		var imageUrl, businessLicenseUrl string
		//获取店铺头像
		avatar, _ := redis.String(redisConn.Do("HGET", fmt.Sprintf("userData:%s", store.BusinessUsername), "Avatar"))
		if avatar == "" {
			//没默认头像
			avatar = LMCommon.PubAvatar
		}

		//智能判断一下，是否是带 http(s) 前缀
		if !strings.HasPrefix(avatar, "http") {

			avatar = LMCommon.OSSUploadPicPrefix + avatar //拼接URL

		}

		if store.ImageUrl != "" {
			imageUrl = LMCommon.OSSUploadPicPrefix + store.ImageUrl
		}

		if store.BusinessLicenseUrl != "" {
			businessLicenseUrl = LMCommon.OSSUploadPicPrefix + store.BusinessLicenseUrl
		}

		//设置随机种子
		rand.Seed(time.Now().UnixNano())
		likes := rand.Intn(999)
		commentcount := rand.Intn(400)

		resp.Stores = append(resp.Stores, &User.Store{
			StoreUUID:          store.StoreUUID,                   //店铺的uuid
			StoreType:          Global.StoreType(store.StoreType), //店铺类型,对应Global.proto里的StoreType枚举
			BusinessUsername:   store.BusinessUsername,            //商户注册号
			Avatar:             avatar,                            //头像
			ImageUrl:           imageUrl,                          //头像
			Introductory:       store.Introductory,                //商店简介 Text文本类型
			Province:           store.Province,                    //省份, 如广东省
			City:               store.City,                        //城市，如广州市
			Area:               store.Area,                        //区，如天河区
			Street:             store.Street,                      //街道
			Address:            store.Address,                     //地址
			Branchesname:       store.Branchesname,                //网点名称
			LegalPerson:        store.LegalPerson,                 //法人姓名
			LegalIdentityCard:  store.LegalIdentityCard,           //法人身份证
			Longitude:          store.Longitude,                   //商户地址的经度
			Latitude:           store.Latitude,                    //商户地址的纬度
			Wechat:             store.Wechat,                      //商户联系人微信号
			Keys:               store.Keys,                        //商户经营范围搜索关键字
			BusinessLicenseUrl: businessLicenseUrl,                //商户营业执照阿里云url
			BusinessCode:       store.BusinessCode,
			AuditState:         store.AuditState, //审核状态，0-预审核，1-审核通过, 2-占位
			CreatedAt:          uint64(store.CreatedAt),
			UpdatedAt:          uint64(store.UpdatedAt),
			Commentcount:       uint64(commentcount), //TODO 暂时是虚拟的
			Likes:              uint64(likes),        //TODO 暂时是虚拟的
			OpeningHours:       store.OpeningHours,   //营业时间
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

	// 判断businessUsername是否是商户

	//用户类型 1-普通，2-商户
	userType, _ := redis.Int(redisConn.Do("HGET", fmt.Sprintf("userData:%s", req.BusinessUsername), "UserType"))
	if userType != int(User.UserType_Ut_Business) {
		return errors.Wrap(err, "此用户非商户类型")
	}

	//判断商户的注册id的合法性以及是否封禁等
	userData := new(models.UserBase)

	userKey := fmt.Sprintf("userData:%s", req.BusinessUsername)
	if result, err := redis.Values(redisConn.Do("HGETALL", userKey)); err == nil {
		if err := redis.ScanStruct(result, userData); err != nil {

			s.logger.Error("错误：ScanStruct", zap.Error(err))
			return errors.Wrapf(err, "查询redis出错[Businessusername=%s]", req.BusinessUsername)

		}
	}
	// 判断是否是商户类型
	if userData.UserType != int(User.UserType_Ut_Business) {
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

//保存excel某一行的网点
func (s *MysqlLianmiRepository) SaveExcelToDb(lotteryStore *models.LotteryStore) error {
	var err error
	if err = s.db.Clauses(clause.OnConflict{DoNothing: true}).Create(lotteryStore).Error; err != nil {
		s.logger.Error("SaveExcelToDb error ", zap.Error(err))
		return err
	} else {
		s.logger.Debug("SaveExcelToDb succeed")
	}

	return nil

}

//查询并分页获取采集的网点
func (s *MysqlLianmiRepository) GetLotteryStores(req *models.LotteryStoreReq) ([]*models.LotteryStore, error) {
	var err error
	pageIndex := int(req.Offset)
	pageSize := int(req.Limit)

	columns := []string{"*"}
	orderBy := "id"

	redisConn := s.redisPool.Get()
	defer redisConn.Close()

	var lotteryStores []*models.LotteryStore
	wheres := make([]interface{}, 0)

	if req.Keyword != "" {
		wheres = append(wheres, []interface{}{"keyword", "=", req.Keyword})
	}
	if req.Province != "" {
		wheres = append(wheres, []interface{}{"province", "=", req.Province})
	}
	if req.City != "" {
		wheres = append(wheres, []interface{}{"city", "=", req.City})
	}
	if req.Area != "" {
		wheres = append(wheres, []interface{}{"area", "=", req.Area})
	}
	if req.Address != "" {
		wheres = append(wheres, []interface{}{"address", "like", "%" + req.Address + "%"})
	}

	db2 := s.db
	db2, err = s.base.BuildQueryList(db2, wheres, columns, orderBy, pageIndex, pageSize)
	if err != nil {
		return nil, err
	}
	err = db2.Find(&lotteryStores).Error

	if err != nil {
		s.logger.Error("Find错误", zap.Error(err))
		return nil, err
	}

	return lotteryStores, nil
}

// 批量增加网点
func (s *MysqlLianmiRepository) BatchAddStores(req *models.LotteryStoreReq) error {
	var err error
	var newIndex uint64
	pageIndex := int(req.Offset)
	pageSize := int(req.Limit)

	columns := []string{"*"}
	orderBy := "id"

	redisConn := s.redisPool.Get()
	defer redisConn.Close()

	if newIndex, err = redis.Uint64(redisConn.Do("INCR", "usernameindex")); err != nil {
		s.logger.Error("redisConn GET usernameindex Error", zap.Error(err))
		return err
	}

	var lotteryStores []*models.LotteryStore
	wheres := make([]interface{}, 0)

	wheres = append(wheres, []interface{}{"status", "=", 0})
	if req.Keyword != "" {
		wheres = append(wheres, []interface{}{"keyword", "=", req.Keyword})
	}
	if req.Province != "" {
		wheres = append(wheres, []interface{}{"province", "=", req.Province})
	}
	if req.City != "" {
		wheres = append(wheres, []interface{}{"city", "=", req.City})
	}
	if req.Area != "" {
		wheres = append(wheres, []interface{}{"area", "=", req.Area})
	}
	if req.Address != "" {
		wheres = append(wheres, []interface{}{"address", "like", "%" + req.Address + "%"})
	}

	db2 := s.db
	db2, err = s.base.BuildQueryList(db2, wheres, columns, orderBy, pageIndex, pageSize)
	if err != nil {
		return err
	}
	err = db2.Find(&lotteryStores).Error

	if err != nil {
		s.logger.Error("Find错误", zap.Error(err))
		return err
	}

	storeType := 1 //福彩

	avatar := ""
	label := ""
	mobileNum := uint64(18977012300)

	imageUrl := ""
	for _, lotteryStore := range lotteryStores {

		if strings.Contains(lotteryStore.Keyword, "福彩") {
			avatar = "http://git.geejoan.cn/wujehy/lianmi_images/-/raw/lianmi/fuli_avatar.jpg"
			label = "扶老 助残 救孤 济困"

			rand.Seed(time.Now().UnixNano())
			number := rand.Intn(10)
			if number == 10 {
				imageUrl = "http://git.geejoan.cn/wujehy/lianmi_images/-/raw/lianmi/fuli010.jpg"
			} else {
				imageUrl = fmt.Sprintf("http://git.geejoan.cn/wujehy/lianmi_images/-/raw/lianmi/fuli00%d.jpg", number)
			}

		} else {
			avatar = "http://git.geejoan.cn/wujehy/lianmi_images/-/raw/lianmi/tiyu_aavatar.jpg"
			label = "公益 快乐 健康 希望"
			storeType = 2

			rand.Seed(time.Now().UnixNano())
			number := rand.Intn(10)
			if number == 10 {
				imageUrl = "http://git.geejoan.cn/wujehy/lianmi_images/-/raw/lianmi/tiyu010.jpg"
			} else {
				imageUrl = fmt.Sprintf("http://git.geejoan.cn/wujehy/lianmi_images/-/raw/lianmi/tiyu00%d.jpg", number)
			}
		}

		s.logger.Debug("变量输出",
			zap.String("avatar", avatar),
			zap.String("label", label),
			zap.String("imageUrl", imageUrl),
			zap.Int("storeType", storeType),
			zap.Any("lotteryStore", lotteryStore),
		)
		/*
			//创建一个商户
			user := &models.User{
				UserBase: models.UserBase{
					Username:         fmt.Sprintf("id%d", newIndex),                 //用户注册号，自动生成，字母 + 数字
					Password:         "C33367701511B4F6020EC61DED352059",            //用户密码，md5加密
					Nick:             lotteryStore.StoreName,                        //用户呢称，必填
					Gender:           1,                                             //性别
					Avatar:           avatar,                                        //头像url
					Label:            label,                                         //签名标签
					Mobile:           fmt.Sprintf("%d", mobileNum+newIndex),         //注册手机
					Email:            fmt.Sprintf("%d@139.com", mobileNum+newIndex), //密保邮件，需要发送校验邮件确认
					AllowType:        3,                                             //用户加好友枚举，默认是3
					UserType:         2,                                             //用户类型 1-普通，2-商户
					State:            0,                                             //状态 0-普通用户，非VIP 1-付费用户(购买会员) 2-封号
					TrueName:         lotteryStore.StoreName,                        //实名
					ReferrerUsername: "id98",                                        //推荐人，上线；介绍人, 账号的数字部分，app的推荐码就是用户id的数字
				},
			}
			if err := s.base.Create(user); err != nil {
				s.logger.Error("db写入错误，注册用户失败")
				return err
			}

			//创建店铺

			store := models.Store{
				StoreUUID:             uuid.NewV4().String(), //店铺的uuid
				StoreType:             storeType,             //店铺类型,对应Global.proto里的StoreType枚举
				ImageURL:              avatar,
				BusinessUsername:      user.Username,          //商户注册号
				Introductory:          label,                  //商店简介 Text文本类型
				Province:              lotteryStore.Province,  //省份, 如广东省
				City:                  lotteryStore.City,      //城市，如广州市
				Area:                  lotteryStore.Area,      //区，如天河区
				Address:               lotteryStore.Address,   //地址
				Branchesname:          lotteryStore.StoreName, //网点名称
				LegalPerson:           "xxx",                  //法人姓名
				LegalIdentityCard:     "xxxxxxxx",             //法人身份证
				Longitude:             lotteryStore.Longitude, //商户地址的经度
				Latitude:              lotteryStore.Latitude,  //商户地址的纬度
				ContactMobile:         user.Mobile,            //联系手机
				WeChat:                user.Mobile,            //商户联系人微信号
				Keys:                  "",                     //商户经营范围搜索关键字
				LicenseURL:            "xxxx",                 //商户营业执照阿里云url
				BusinessCode:          "21313223",             //网点编码
				NotaryServiceUsername: "id119",                //网点的公证注册id
				AuditState:            1,                      //初始值
				OpeningHours:          "10:00-22:00",          //营业时间
			}

			//如果没有记录，则增加
			if err := s.db.Clauses(clause.OnConflict{DoNothing: true}).Create(&store).Error; err != nil {
				s.logger.Error("AddStore, failed to upsert stores", zap.Error(err))
				return err
			} else {
				s.logger.Debug("AddStore, upsert stores succeed")
			}

			//网点商户自动建群
			var newTeamIndex uint64
			if newTeamIndex, err = redis.Uint64(redisConn.Do("INCR", "TeamIndex")); err != nil {
				s.logger.Error("redisConn GET TeamIndex Error", zap.Error(err))
				return err
			}
			pTeam := new(models.Team)
			pTeam.TeamID = fmt.Sprintf("team%d", newTeamIndex) //群id， 自动生成
			pTeam.Teamname = fmt.Sprintf("team%d", newTeamIndex)
			pTeam.Nick = fmt.Sprintf("%s的群", user.Nick)
			pTeam.Owner = user.Username
			pTeam.Type = 1
			pTeam.VerifyType = 1
			pTeam.InviteMode = 1

			//默认的设置
			pTeam.Status = 1 //Init(1) - 初始状态,审核中 Normal(2) - 正常状态 Blocked(3) - 封禁状态
			pTeam.MemberLimit = LMCommon.PerTeamMembersLimit
			pTeam.MemberNum = 1  //刚刚建群是只有群主1人
			pTeam.MuteType = 1   //None(1) - 所有人可发言
			pTeam.InviteMode = 1 //邀请模式,初始为1

			//创建群数据 增加记录
			if err := s.db.Clauses(clause.OnConflict{DoNothing: true}).Create(&pTeam).Error; err != nil {
				s.logger.Error("Register, failed to upsert team", zap.Error(err))
				return err
			} else {
				s.logger.Debug("CreateTeam, upsert team succeed")
			}

			//将用户信息缓存到redis里
			userKey := fmt.Sprintf("userData:%s", user.Username)
			if _, err := redisConn.Do("HMSET", redis.Args{}.Add(userKey).AddFlat(user.UserBase)...); err != nil {
				s.logger.Error("错误：HMSET", zap.Error(err))
			}

			//创建redis的sync:{用户账号} myInfoAt 时间戳
			//myInfoAt, friendsAt, friendUsersAt, teamsAt, tagsAt, watchAt, productAt

			syncKey := fmt.Sprintf("sync:%s", user.Username)
			redisConn.Do("HSET", syncKey, "myInfoAt", time.Now().UnixNano()/1e6)
			redisConn.Do("HSET", syncKey, "friendsAt", time.Now().UnixNano()/1e6)
			redisConn.Do("HSET", syncKey, "friendUsersAt", time.Now().UnixNano()/1e6)
			redisConn.Do("HSET", syncKey, "teamsAt", time.Now().UnixNano()/1e6)
			redisConn.Do("HSET", syncKey, "tagsAt", time.Now().UnixNano()/1e6)
			redisConn.Do("HSET", syncKey, "watchAt", time.Now().UnixNano()/1e6)

			s.ApproveTeam(pTeam.TeamID)

			s.logger.Debug("注册商户成功", zap.String("Username", user.Username))

			s.db.Model(&models.LotteryStore{}).Where(&models.LotteryStore{
				MapID: lotteryStore.MapID,
			}).Update("status", 1)

		*/
		break //只录入一条记录

	}
	_ = newIndex
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

	redisConn := s.redisPool.Get()
	defer redisConn.Close()

	// 判断businessUsername是否是商户

	//用户类型 1-普通，2-商户
	userType, _ := redis.Int(redisConn.Do("HGET", fmt.Sprintf("userData:%s", username), "UserType"))
	if userType != int(User.UserType_Ut_Business) {
		return nil, errors.Wrap(err, "此用户非商户类型")
	}

	var businessUsers []string
	rsp := &User.UserLikesResp{
		Username: username,
	}

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
	s.logger.Debug("StoreLikes start ...")
	var err error

	redisConn := s.redisPool.Get()
	defer redisConn.Close()

	// 判断businessUsername是否是商户

	//用户类型 1-普通，2-商户
	userType, _ := redis.Int(redisConn.Do("HGET", fmt.Sprintf("userData:%s", businessUsername), "UserType"))
	if userType != int(User.UserType_Ut_Business) {
		return nil, errors.Wrap(err, "此用户非商户类型")
	}

	var users []string
	rsp := &User.StoreLikesResp{
		BusinessUsername: businessUsername,
	}
	storelikeKey := fmt.Sprintf("StoreLike:%s", businessUsername)
	s.logger.Debug("StoreLikes", zap.String("storelikeKey", storelikeKey))

	if users, err = redis.Strings(redisConn.Do("SMEMBERS", storelikeKey)); err != nil {
		s.logger.Error("SMEMBERS Error", zap.Error(err))
		return nil, err
	}

	for _, user := range users {
		if strings.HasPrefix(user, "id") {
			rsp.Usernames = append(rsp.Usernames, user)

		}
	}

	return rsp, nil
}

//获取店铺的所有点赞总数
func (s *MysqlLianmiRepository) StoreLikesCount(businessUsername string) (int, error) {
	s.logger.Debug("StoreLikesCount start ...")
	var err error

	redisConn := s.redisPool.Get()
	defer redisConn.Close()

	// 判断businessUsername是否是商户

	//用户类型 1-普通，2-商户
	userType, _ := redis.Int(redisConn.Do("HGET", fmt.Sprintf("userData:%s", businessUsername), "UserType"))
	if userType != int(User.UserType_Ut_Business) {
		return 0, errors.Wrap(err, "此用户非商户类型")
	}

	var count int

	storelikeKey := fmt.Sprintf("StoreLike:%s", businessUsername)
	s.logger.Debug("StoreLikes", zap.String("storelikeKey", storelikeKey))

	if count, err = redis.Int(redisConn.Do("SCARD", storelikeKey)); err != nil {
		s.logger.Error("SCARD Error", zap.Error(err))
		return 0, err
	}

	return count, nil
}

//对某个店铺点赞，返回当前所有的点赞总数
func (s *MysqlLianmiRepository) ClickLike(username, businessUsername string) (int64, error) {

	var err error
	var totalLikeCount int64

	s.logger.Debug("ClickLike start ...")

	redisConn := s.redisPool.Get()
	defer redisConn.Close()

	// 判断businessUsername是否是商户

	//用户类型 1-普通，2-商户
	userType, _ := redis.Int(redisConn.Do("HGET", fmt.Sprintf("userData:%s", username), "UserType"))
	if userType != int(User.UserType_Ut_Business) {
		return 0, errors.Wrap(err, "此用户非商户类型")
	}

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
	// Scard 命令返回集合中元素的数量。
	if totalLikeCount, err = redis.Int64(redisConn.Do("SCARD", storelikeKey)); err != nil {
		s.logger.Error("SCARD TotalLike Error", zap.Error(err))
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

	// 判断businessUsername是否是商户

	//用户类型 1-普通，2-商户
	userType, _ := redis.Int(redisConn.Do("HGET", fmt.Sprintf("userData:%s", username), "UserType"))
	if userType != int(User.UserType_Ut_Business) {
		return 0, errors.Wrap(err, "此用户非商户类型")
	}

	userlikeKey := fmt.Sprintf("UserLike:%s", username)
	if _, err = redisConn.Do("SREM", userlikeKey, businessUsername); err != nil {
		s.logger.Error("SREM UserLike Error", zap.Error(err))
		return 0, err
	}

	//删除店铺点赞用户列表
	storelikeKey := fmt.Sprintf("StoreLike:%s", businessUsername)
	if _, err = redisConn.Do("SREM", storelikeKey, username); err != nil {
		s.logger.Error("SREM StoreLike Error", zap.Error(err))
		return 0, err
	}

	//此店铺的总点赞数
	// Scard 命令返回集合中元素的数量。
	if totalLikeCount, err = redis.Int64(redisConn.Do("SCARD", storelikeKey)); err != nil {
		s.logger.Error("SCARD TotalLike Error", zap.Error(err))
		return 0, err
	}
	return totalLikeCount, nil
}

//取消对某个店铺点赞
func (s *MysqlLianmiRepository) GetIsLike(username, businessUsername string) (bool, error) {
	redisConn := s.redisPool.Get()
	defer redisConn.Close()

	//SISMEMBER
	storelikeKey := fmt.Sprintf("StoreLike:%s", businessUsername)

	return redis.Bool(redisConn.Do("SCARD", storelikeKey))

}

//将点赞记录插入到UserLike表
func (s *MysqlLianmiRepository) AddUserLike(username, businessUser string) error {
	var err error
	redisConn := s.redisPool.Get()
	defer redisConn.Close()

	// 判断businessUsername是否是商户

	//用户类型 1-普通，2-商户
	userType, _ := redis.Int(redisConn.Do("HGET", fmt.Sprintf("userData:%s", username), "UserType"))
	if userType != int(User.UserType_Ut_Business) {
		return errors.Wrap(err, "此用户非商户类型")
	}

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
	var err error
	redisConn := s.redisPool.Get()
	defer redisConn.Close()

	// 判断businessUsername是否是商户

	//用户类型 1-普通，2-商户
	userType, _ := redis.Int(redisConn.Do("HGET", fmt.Sprintf("userData:%s", businessUsername), "UserType"))
	if userType != int(User.UserType_Ut_Business) {
		return errors.Wrap(err, "此用户非商户类型")
	}

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

//获取各种彩票的开售及停售时刻
func (s *MysqlLianmiRepository) QueryLotterySaleTimes() (*Order.QueryLotterySaleTimesRsp, error) {
	var err error

	lotterySaleTimesRsp := &Order.QueryLotterySaleTimesRsp{}

	var lotterySaleTimes []*models.LotterySaleTime

	where := models.LotterySaleTime{
		IsActive: true,
	}

	db2 := s.db
	db2, err = s.base.BuildWhere(db2, where)
	if err != nil {
		s.logger.Error("BuildWhere错误", zap.Error(err))
		return nil, err
	}

	db2.Find(&lotterySaleTimes)

	for _, lotterySaleTime := range lotterySaleTimes {

		orderLotterySaleTime := &Order.LotterySaleTime{
			LotteryType:   int32(lotterySaleTime.LotteryType),
			LotteryName:   lotterySaleTime.LotteryName,
			SaleEndHour:   int32(lotterySaleTime.SaleEndHour),
			SaleEndMinute: int32(lotterySaleTime.SaleEndMinute),
		}
		orderLotterySaleTime.SaleEndWeeks = strings.Split(lotterySaleTime.SaleEndWeek, ",")
		orderLotterySaleTime.Holidays = strings.Split(lotterySaleTime.Holidays, ",")

		lotterySaleTimesRsp.LotterySaleTimes = append(lotterySaleTimesRsp.LotterySaleTimes, orderLotterySaleTime)

	}

	return lotterySaleTimesRsp, nil

}

//设置当前商户默认OPK
func (s *MysqlLianmiRepository) SetDefaultOPK(username, opk string) error {
	var err error
	s.logger.Debug("SetDefaultOPK start ...")

	redisConn := s.redisPool.Get()
	defer redisConn.Close()

	//用户类型 1-普通，2-商户
	userType, _ := redis.Int(redisConn.Do("HGET", fmt.Sprintf("userData:%s", username), "UserType"))
	if userType != int(User.UserType_Ut_Business) {
		return errors.Wrap(err, "此用户非商户类型")
	}
	_, err = redisConn.Do("SET", fmt.Sprintf("DefaultOPK:%s", username), opk)
	if err != nil {
		return err
	}

	//更新MySQL stores表
	result := s.db.Model(&models.Store{}).Where(&models.Store{
		BusinessUsername: username,
	}).Update("default_opk", opk)

	//updated records count
	s.logger.Debug("修改 stores表 result: ",
		zap.Int64("RowsAffected", result.RowsAffected),
		zap.Error(result.Error))

	if result.Error != nil {
		s.logger.Error("Update Store default_opk 失败", zap.Error(result.Error))
		return result.Error
	} else {
		s.logger.Debug("Update Store default_opk  成功")
	}

	s.logger.Debug("SetDefaultOPK end")

	return nil

}
