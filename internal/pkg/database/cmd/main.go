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
	dsn = "lianmidba:12345678@tcp(127.0.0.1:3306)/lianmicloud?charset=utf8&parseTime=True&loc=Local"
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
	pageSize := 5
	// var count int64
	var users []models.User
	userModel := new(models.User)
	
	db.Model(&userModel).Association("users").Find(&users, "user_type=?", 0)

	log.Println(" *********  查询 users *******  ", len(users))
	PrintPretty(users)

	_ = page
	_ = pageSize

}
