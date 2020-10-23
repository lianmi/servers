package order

import (
	"fmt"
	"github.com/gomodule/redigo/redis"
	"github.com/lianmi/servers/internal/pkg/models"
	LMCommon "github.com/lianmi/servers/lmSdkClient/common"
	uuid "github.com/satori/go.uuid"
	"log"
	"time"
)

func MockGeneralProduct() error {

	redisConn, err := redis.Dial("tcp", LMCommon.RedisAddr)
	if err != nil {
		log.Fatalln(err)
		return err
	}

	defer redisConn.Close()

	//增加一个通用商品
	//uuid
	productID := uuid.NewV4().String()

	key := fmt.Sprintf("GeneralProduct:%s", productID)
	//fb18ef07-0e23-4141-aff3-dbff653599d9

	productInfo := &models.GeneralProduct{
		ProductID:         productID,                       //商品ID
		ProductName:       "猪肚",                            //商品名称
		ProductType:       1,                               // 商品种类枚举 : 肉类
		ProductDesc:       "新鲜,补脾之要品,配伍党参、白术、薏苡仁、莲子、陈皮煮熟食", //商品详细介绍
		ProductPic1Small:  "",                              //商品图片1-小图
		ProductPic1Middle: "",                              //商品图片1-中图
		ProductPic1Large:  "",                              //商品图片1-大图
		ProductPic2Small:  "",                              //商品图片2-小图
		ProductPic2Middle: "",                              //商品图片2-中图
		ProductPic2Large:  "",                              //商品图片2-大图
		ProductPic3Small:  "",                              //商品图片3-小图
		ProductPic3Middle: "",                              //商品图片3-中图
		ProductPic3Large:  "",                              //商品图片3-大图
		Thumbnail:         "",                              //商品短视频缩略图
		ShortVideo:        "",                              //商品短视频
		AllowCancel:       false,
	}
	if _, err := redisConn.Do("HMSET", redis.Args{}.Add(key).AddFlat(productInfo)...); err != nil {
		log.Println("错误：HMSET", err.Error())
		// continue
	}

	//增加到 GeneralProducts
	if _, err = redisConn.Do("ZADD", "GeneralProducts", time.Now().UnixNano()/1e6, productID); err != nil {
		log.Println("ZADD", err.Error())
	}

	return nil
}
