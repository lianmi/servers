/*
GORM 英文文档 https://gorm.io/docs
     中文文档 https://gorm.io/zh_CN/docs/index.html
*/
package database

import (
	"github.com/google/wire"
	LMCommon "github.com/lianmi/servers/internal/common"
	"github.com/lianmi/servers/internal/pkg/models"
	"github.com/pkg/errors"
	uuid "github.com/satori/go.uuid"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// Options is  configuration of database
type Options struct {
	URL   string `yaml:"url"`
	Debug bool
}

func NewOptions(v *viper.Viper, logger *zap.Logger) (*Options, error) {
	var err error
	o := new(Options)
	if err = v.UnmarshalKey("db", o); err != nil {
		return nil, errors.Wrap(err, "unmarshal db option error")
	}

	logger.Info("load database options success", zap.String("url", o.URL))

	return o, err
}

// Init 初始化数据库
func New(o *Options) (*gorm.DB, error) {
	var err error
	db, err := gorm.Open(mysql.Open(o.URL), &gorm.Config{})
	if err != nil {
		return nil, errors.Wrap(err, "gorm open database connection error")
	}

	if o.Debug {
		db = db.Debug()
	}

	//自动迁移仅仅会创建表，缺少列和索引，并且不会改变现有列的类型或删除未使用的列以保护数据
	db.AutoMigrate(&models.User{})                     // 用户表
	db.AutoMigrate(&models.Token{})                    // 令牌表
	db.AutoMigrate(&models.Role{})                     // 权限表
	db.AutoMigrate(&models.Tag{})                      // 标签表
	db.AutoMigrate(&models.Friend{})                   // 好友表
	db.AutoMigrate(&models.Team{})                     // 群组表
	db.AutoMigrate(&models.TeamUser{})                 // 群成员表
	db.AutoMigrate(&models.Prekey{})                   // OPK表, 商户上传
	db.AutoMigrate(&models.Product{})                  // 商品表
	db.AutoMigrate(&models.GeneralProduct{})           // 通用商品表
	db.AutoMigrate(&models.UserWallet{})               // 用户钱包表
	db.AutoMigrate(&models.LnmcDepositHistory{})       // 用户充值记录表
	db.AutoMigrate(&models.LnmcTransferHistory{})      // 用户转账及支付记录表
	db.AutoMigrate(&models.LnmcWithdrawHistory{})      // 用户提现记录表
	db.AutoMigrate(&models.LnmcCollectionHistory{})    // 用户收款记录
	db.AutoMigrate(&models.LnmcOrderTransferHistory{}) // 订单完成后的商户到账或撤单退款记录表
	db.AutoMigrate(&models.CustomerServiceInfo{})      // 在线客服技术表
	db.AutoMigrate(&models.Grade{})                    // 客服满意度评分
	db.AutoMigrate(&models.Distribution{})             // 用户层级表
	db.AutoMigrate(&models.Commission{})               // 用户佣金表
	db.AutoMigrate(&models.CommissionWithdraw{})       // 佣金提现申请表表
	db.AutoMigrate(&models.CommissionStatistics{})     // 商户/用户佣金月统计表
	db.AutoMigrate(&models.BusinessUnderling{})        // 商户下属的会员表
	db.AutoMigrate(&models.BusinessUserStatistics{})   // 商户下属的月会员统计表
	db.AutoMigrate(&models.Store{})                    // 商户店铺表
	db.AutoMigrate(&models.StoreLike{})                // 商店的点赞明细表
	db.AutoMigrate(&models.UserLike{})                 // 用户点赞的店铺记录表
	db.AutoMigrate(&models.OrderImagesHistory{})       // 服务端的订单图片上链历史表
	db.AutoMigrate(&models.AliPayHistory{})            //  支付宝充值历史表
	db.AutoMigrate(&models.VipPrice{})                 //  VIP会员价格表
	db.AutoMigrate(&models.ECoupon{})                  //  系统电子优惠券表

	vipPrice := &models.VipPrice{
		ID:               1,
		BusinessUsername: LMCommon.VipBusinessUsername,
		ProductID:        uuid.NewV4().String(),
		PayType:          1,
		Title:            "包年",
		Price:            99.0,
		Days:             365,
		IsActive:         true,
	}
	//如果没有记录，则增加，如果有记录，则更新全部字段
	if err = db.Clauses(clause.OnConflict{DoNothing: true}).Create(&vipPrice).Error; err != nil {
		// return nil, err
	}
	vipPrice2 := &models.VipPrice{
		ID:               2,
		BusinessUsername: LMCommon.VipBusinessUsername,
		ProductID:        uuid.NewV4().String(),
		PayType:          2,
		Title:            "包季",
		Price:            38.0,
		Days:             90,
		IsActive:         true,
	}

	if err = db.Clauses(clause.OnConflict{DoNothing: true}).Create(&vipPrice2).Error; err != nil {
		// return nil, err
	}
	vipPrice3 := &models.VipPrice{
		ID:               3,
		BusinessUsername: LMCommon.VipBusinessUsername,
		ProductID:        uuid.NewV4().String(),
		PayType:          3,
		Title:            "包月",
		Price:            9.9,
		Days:             30,
		IsActive:         true,
	}

	if err = db.Clauses(clause.OnConflict{DoNothing: true}).Create(&vipPrice3).Error; err != nil {
		// return nil, err
	}

	currYearMonth := "2020-12"
	username := "id1"
	total := 120.22

	newnucs := &models.CommissionStatistics{
		Username:        username,
		YearMonth:       currYearMonth,
		TotalCommission: total, //本月返佣总金额
		IsRebate:        false, //默认返现的值是false
	}

	//create
	resultErr := db.Model(newnucs).Save(&newnucs).Error
	if resultErr != nil {
		if resultErr != gorm.ErrRecordNotFound {
			// log.Println("Record Not Found")
		} else {
			// log.Println(resultErr)
			// return
		}
	}

	return db, nil
}

var ProviderSet = wire.NewSet(New, NewOptions)
