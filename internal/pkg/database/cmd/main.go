package main

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/lianmi/servers/internal/pkg/models"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var (
	//使用 lianmicloud 数据库
	// dsn = "lianmidba:12345678@tcp(127.0.0.1:3306)/newlianmidb?charset=utf8&parseTime=True&loc=Local"
	dsn = "root:password@tcp(127.0.0.1:3306)/lianmicloud?charset=utf8&parseTime=True&loc=Local"
	db  *gorm.DB
)

func init() {
	var err error
	db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
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
	page := 1
	pageSize := 20
	// var count int64
	var users []models.User
	userModel := new(models.User)

	db.Model(&userModel).Find(&users, "user_type=?", 0)

	log.Println(" *********  查询 users *******  ", len(users))
	// PrintPretty(users)

	// db.Model(&userModel).Scopes(IsNormalUser, Paginate(page, pageSize)).Find(&users)
	// db.Model(&userModel).Scopes(IsBusinessUser, Paginate(page, pageSize)).Find(&users)
	db.Model(&userModel).Scopes(IsPreBusinessUser, LegalPerson([]string{"杜老板"}), Paginate(page, pageSize)).Find(&users)

	log.Println("分页显示users列表, count: ", len(users))

	for idx, user := range users {
		log.Printf("idx=%d, username=%s, mobile=%d\n", idx, user.Username, user.Mobile)
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

func Paginate(page, pageSize int) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		// page, _ := strconv.Atoi(r.Query("page"))
		if page == 0 {
			page = 1
		}

		// pageSize, _ := strconv.Atoi(r.Query("page_size"))
		switch {
		case pageSize > 100:
			pageSize = 100
		case pageSize <= 0:
			pageSize = 10
		}

		offset := (page - 1) * pageSize
		return db.Offset(offset).Limit(pageSize)
	}
}
