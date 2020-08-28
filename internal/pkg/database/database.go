package database

import (
	"github.com/google/wire"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"github.com/lianmi/servers/internal/pkg/models"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
	"go.uber.org/zap"
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
	db, err := gorm.Open("mysql", o.URL) //打开mysql这个系统表
	if err != nil {
		return nil, errors.Wrap(err, "gorm open database connection error")
	}

	if o.Debug {
		db = db.Debug()
	}

	//自动迁移仅仅会创建表，缺少列和索引，并且不会改变现有列的类型或删除未使用的列以保护数据
	db.AutoMigrate(&models.User{})     // 用户表
	db.AutoMigrate(&models.Token{})    // 令牌表
	db.AutoMigrate(&models.Role{})     // 权限表
	db.AutoMigrate(&models.Tag{})      // 标签表
	db.AutoMigrate(&models.Friend{})   // 好友表
	db.AutoMigrate(&models.Team{})     // 群组表
	db.AutoMigrate(&models.TeamUser{}) // 群成员表
	db.AutoMigrate(&models.Prekey{})   // OPK表, 商户上传

	return db, nil
}

var ProviderSet = wire.NewSet(New, NewOptions)
