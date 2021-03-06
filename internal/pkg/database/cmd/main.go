package main

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/lianmi/servers/internal/pkg/database/cmd/core"
	"github.com/lianmi/servers/internal/pkg/database/cmd/global"
	"github.com/lianmi/servers/internal/pkg/database/cmd/internal"
	"github.com/lianmi/servers/internal/pkg/models"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var (
	//使用 lianmicloud 数据库
	// dsn = "lianmidba:12345678@tcp(127.0.0.1:3306)/lianmicloud?charset=utf8&parseTime=True&loc=Local"
	dsn = "root:password@tcp(127.0.0.1:3306)/lianmicloud?charset=utf8&parseTime=True&loc=Local"
	db  *gorm.DB
	// GVA_LOG *zap.Logger
)

const LogZap = "zap"

func gormConfig(mod bool) *gorm.Config {
	var config = &gorm.Config{DisableForeignKeyConstraintWhenMigrating: true}
	switch LogZap {
	case "silent", "Silent":
		config.Logger = internal.Default.LogMode(logger.Silent)
	case "error", "Error":
		config.Logger = internal.Default.LogMode(logger.Error)
	case "warn", "Warn":
		config.Logger = internal.Default.LogMode(logger.Warn)
	case "info", "Info":
		config.Logger = internal.Default.LogMode(logger.Info)
	case "zap", "Zap":
		config.Logger = internal.Default.LogMode(logger.Info)
	default:
		if mod {
			config.Logger = internal.Default.LogMode(logger.Info)
			break
		}
		config.Logger = internal.Default.LogMode(logger.Silent)
	}
	return config
}

func init() {
	var err error
	global.GVA_LOG = core.Zap() // 初始化zap日志库

	db, err = gorm.Open(mysql.Open(dsn), gormConfig(true))
	if err != nil {
		log.Fatalln(err)
	}
	db = db.Debug()

}

func PrintPretty(i interface{}) {
	data, err := json.MarshalIndent(i, "", "    ")
	if err != nil {
		log.Fatalf("JSON marshaling failed: %s", err)
	}
	fmt.Printf("%s\n", data)
}

func main() {
	page := 0
	pageSize := 20
	// var count int64
	var users []models.User
	userModel := new(models.User)

	db.Model(&userModel).Find(&users, "user_type=?", 0)

	log.Println(" *********  查询 users *******  ", len(users))
	// PrintPretty(users)

	// db.Model(&userModel).Scopes(IsNormalUser, Paginate(page, pageSize)).Find(&users).Order("updated_at DESC")
	//注意！Order必须在Find之前
	db.Model(&userModel).Scopes(IsNormalUser, Paginate(page, pageSize), BetweenCreateAt(1606408683991, 1606514952437)).Order("created_at DESC").Find(&users)
	// db = db
	// db.Model(&userModel).Scopes(IsBusinessUser, Paginate(page, pageSize)).Find(&users)
	// db.Model(&userModel).Scopes(IsPreBusinessUser, LegalPerson([]string{"杜老板"}), Paginate(page, pageSize)).Find(&users)

	log.Println("分页显示users列表, count: ", len(users))

	for idx, user := range users {
		log.Printf("idx=%d, create_at %d, username=%s, mobile=%s\n", idx, user.CreatedAt, user.Username, user.Mobile)
	}
	_ = page
	_ = pageSize

}

//用户类型 1-普通
func IsNormalUser(db *gorm.DB) *gorm.DB {
	return db.Where("user_type = ? ", 1)
}

//用户类型 2-商户  处于预审核状态
func IsPreBusinessUser(db *gorm.DB) *gorm.DB {
	return db.Where("user_type = ?  and  state = ?", 2, 0)
}

//用户类型 2-商户  处于已审核状态
func IsBusinessUser(db *gorm.DB) *gorm.DB {
	return db.Where("user_type = ?  and  state = ?", 2, 1)
}

//实名
func TrueName(trueNames []string) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("true_name IN (?)", trueNames)
	}
}

//法人
func LegalPerson(legalPersons []string) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("legal_person IN (?)", legalPersons)
	}
}

//店铺名称
func Branchesname(branchesnames []string) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("branchesname IN (?)", branchesnames)
	}
}

//按createAt的时间段
func BetweenCreateAt(startAt, endAt uint64) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("created_at>= ? and created_at<= ? ", startAt, endAt)
	}
}
func Paginate(page, pageSize int) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if page == 0 {
			page = 1
		}

		switch {
		case pageSize > 100:
			pageSize = 20
		case pageSize <= 0:
			pageSize = 10
		}

		offset := (page - 1) * pageSize
		return db.Offset(offset).Limit(pageSize)
	}
}

// SELECT * FROM `users` WHERE user_type = 1  AND (create_at>= 1605328128169 and create_at<= 1603789653918 ) AND `users`.`deleted_at` IS NULL LIMIT 20
