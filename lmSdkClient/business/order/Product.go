package order

import (
	"context"
	"errors"
	"fmt"

	"github.com/eclipse/paho.golang/paho" //支持v5.0
	"github.com/golang/protobuf/proto"
	"github.com/gomodule/redigo/redis"
	"github.com/lianmi/servers/lmSdkClient/business"
	"github.com/lianmi/servers/util/array"

	"log"

	Global "github.com/lianmi/servers/api/proto/global"
	Order "github.com/lianmi/servers/api/proto/order"
	LMCommon "github.com/lianmi/servers/lmSdkClient/common"
)

// 7-1  查询某个商户的所有商品信息
func QueryProducts() error {

	redisConn, err := redis.Dial("tcp", LMCommon.RedisAddr)
	if err != nil {
		log.Fatalln(err)
		return err
	}

	defer redisConn.Close()

	topic := "lianmi/cloud/dispatcher"
	var localUserName string
	var localDeviceID string
	localUserName, _ = redis.String(redisConn.Do("GET", "LocalUserName"))

	localDeviceID, _ = redis.String(redisConn.Do("GET", "LocalDeviceID"))
	if localUserName == "" {
		log.Println("localUserName is  empty")
		return errors.New("localUserName is empty error")
	}
	if localDeviceID == "" {
		log.Println("localDeviceID is  empty")
		return errors.New("localDeviceID is empty error")
	}

	responseTopic := fmt.Sprintf("lianmi/cloud/%s", localDeviceID)

	//从本地redis里获取jwtToken，注意： 在auth模块的登录，登录成功后，需要写入，这里则读取
	key := fmt.Sprintf("jwtToken:%s", localUserName)
	jwtToken, err := redis.String(redisConn.Do("GET", key))
	if err != nil {
		log.Println("Redis GET jwtToken:{LocalUserName}", err)
		return err
	}
	if jwtToken == "" {
		return errors.New("jwtToken is empty error")
	}
	log.Println("localUserName: ", localUserName)
	log.Println("localDeviceID: ", localDeviceID)
	log.Println("jwtToken: ", jwtToken)
	taskId, _ := redis.Int(redisConn.Do("INCR", fmt.Sprintf("taksID:%s", localUserName)))
	taskIdStr := fmt.Sprintf("%d", taskId)

	req := &Order.QueryProductsReq{
		UserName: "id3", //测试指定为Id3
		TimeAt:   0,
	}

	content, _ := proto.Marshal(req)

	pb := &paho.Publish{
		Topic:   topic,
		QoS:     byte(1),
		Payload: content,
		Properties: &paho.PublishProperties{
			ResponseTopic: responseTopic,
			// User: map[string]string{
			// 	"jwtToken":        jwtToken,      // jwt令牌
			// 	"deviceId":        localDeviceID, // 设备号
			// 	"businessType":    "7",           // 业务号
			// 	"businessSubType": "1",           // 业务子号
			// 	"taskId":          taskIdStr,
			// 	"code":            "0",
			// 	"errormsg":        "",
			// },
		},
	}

	pb.Properties.User.Add("jwtToken", jwtToken)
	pb.Properties.User.Add("deviceId", localDeviceID)
	pb.Properties.User.Add("businessType", "7")
	pb.Properties.User.Add("businessSubType", "1")
	pb.Properties.User.Add("taskId", taskIdStr)
	pb.Properties.User.Add("code", "0")
	pb.Properties.User.Add("errormsg", "")

	var client *paho.Client
	var payloadCh chan []byte
	payloadCh = make(chan []byte, 0)

	client = business.CreateClient(payloadCh)

	cp := &paho.Connect{
		KeepAlive:  30,
		ClientID:   localDeviceID,
		CleanStart: true,
		Username:   "",
		Password:   []byte(""),
	}
	ca, err := client.Connect(context.Background(), cp)
	if err == nil {
		if ca.ReasonCode == 0 {
			subTopic := fmt.Sprintf("lianmi/cloud/device/%s", localDeviceID)
			if _, err := client.Subscribe(context.Background(), &paho.Subscribe{
				Subscriptions: map[string]paho.SubscribeOptions{
					subTopic: paho.SubscribeOptions{QoS: byte(1), NoLocal: true},
				},
			}); err != nil {
				log.Println("Failed to subscribe:", err)
			}
			log.Println("Subscribed succed: ", subTopic)
		}
	} else {
		log.Println("Failed to Connect mqtt server", err)
	}

	if _, err := client.Publish(context.Background(), pb); err != nil {
		log.Println("Failed to Publish:", err)
	} else {
		log.Println("Succeed Publish to mqtt broker:", topic)
	}

	//堵塞
	payload := <-payloadCh

	//解包负载 payload
	var rsq Order.QueryProductsRsp
	if err := proto.Unmarshal(payload, &rsq); err != nil {
		log.Println("Protobuf Unmarshal Error", err)

	} else {

		log.Println("回包内容 QueryProductsRsp ---------------------")
		log.Println("")
		log.Println(" 商户商品明细 rsq.Products---------------------")

		// for _, product := range rsq.Products {
		// 	log.Println("商品ID ProductId: ", product.ProductId)
		// 	log.Println("商品名称 ProductName: ", product.ProductName)
		// 	log.Println("商品描述 ProductDesc: ", product.ProductDesc)
		// }

		array.PrintPretty(rsq.Products)

		log.Println("")
		log.Println(" 下架商品明细 ---------------------")

		for _, ProductId := range rsq.SoldoutProducts {
			log.Println("商品ID ProductId: ", ProductId)

		}

	}
	log.Println("QueryProducts is Done.")

	return nil
}

//  模拟  7-2 商品上架
func AddProduct() error {

	redisConn, err := redis.Dial("tcp", LMCommon.RedisAddr)
	if err != nil {
		log.Fatalln(err)
		return err
	}

	defer redisConn.Close()

	topic := "lianmi/cloud/dispatcher"
	var localUserName string
	var localDeviceID string
	localUserName, _ = redis.String(redisConn.Do("GET", "LocalUserName"))
	localDeviceID, _ = redis.String(redisConn.Do("GET", "LocalDeviceID"))
	if localUserName == "" {
		log.Println("localUserName is  empty")
		return errors.New("localUserName is empty error")
	}
	if localDeviceID == "" {
		log.Println("localDeviceID is  empty")
		return errors.New("localDeviceID is empty error")
	}
	responseTopic := fmt.Sprintf("lianmi/cloud/%s", localDeviceID)

	//从本地redis里获取jwtToken，注意： 在auth模块的登录，登录成功后，需要写入，这里则读取
	key := fmt.Sprintf("jwtToken:%s", localUserName)
	jwtToken, err := redis.String(redisConn.Do("GET", key))
	if err != nil {
		log.Println("Redis GET jwtToken:{LocalUserName}", err)
		return err
	}
	if jwtToken == "" {
		return errors.New("jwtToken is empty error")
	}

	taskId, _ := redis.Int(redisConn.Do("INCR", fmt.Sprintf("taksID:%s", localUserName)))
	taskIdStr := fmt.Sprintf("%d", taskId)

	oProduct := &Order.Product{
		Expire:      uint64(0),
		ProductName: "双色球",                 //双色球
		ProductType: Global.ProductType(9), //9-彩票
		SubType:     1,                     //双色球枚举
		ProductDesc: "最新派彩，大奖最高一千万",        //大奖最高一千万
		ShortVideo:  "",
		Price:       float32(2.0),
		AllowCancel: true,
	}
	oProduct.ProductPics = append(oProduct.ProductPics, &Order.ProductPic{
		Large: "products/2020/12/09/2ff867f849548257d20c003e9de44f98.jpeg", // 原图1
	})
	oProduct.ProductPics = append(oProduct.ProductPics, &Order.ProductPic{
		Large: "products/2020/12/09/2ff867f849548257d20c003e9de44f98.jpeg", // 原图2
	})
	oProduct.ProductPics = append(oProduct.ProductPics, &Order.ProductPic{
		Large: "products/2020/12/09/2ff867f849548257d20c003e9de44f98.jpeg", // 原图3
	})

	//商品内容图片数组
	oProduct.DescPics = append(oProduct.DescPics, "products/07cb349819583706fee9c08d03434a30.jpeg")
	oProduct.DescPics = append(oProduct.DescPics, "products/048c9822ca1c424080fcbc195abf9624.jpeg")
	oProduct.DescPics = append(oProduct.DescPics, "products/fb5e2fa4e971b0aa4d3d8937e60997c.jpg")
	oProduct.DescPics = append(oProduct.DescPics, "products/4a3f0fe2d58092e2a7f48ce8f58a3df7.jpeg")
	oProduct.DescPics = append(oProduct.DescPics, "products/cf14e9281e6f3819a2001c4b1bdc1301.jpeg")
	oProduct.DescPics = append(oProduct.DescPics, "products/0ffb4f3bc3d419affa6d8fe3efa7eb31.jpeg")

	req := &Order.AddProductReq{
		Product: oProduct,
		// OrderType:       Global.OrderType(1), //1- 正常 2-任务抢单类型 3-竞猜类
	}

	content, _ := proto.Marshal(req)

	pb := &paho.Publish{
		Topic:   topic,
		QoS:     byte(1),
		Payload: content,
		Properties: &paho.PublishProperties{
			ResponseTopic: responseTopic,
			// User: map[string]string{
			// 	"jwtToken":        jwtToken,      // jwt令牌
			// 	"deviceId":        localDeviceID, // 设备号
			// 	"businessType":    "7",           // 业务号
			// 	"businessSubType": "2",           //  业务子号
			// 	"taskId":          taskIdStr,
			// 	"code":            "0",
			// 	"errormsg":        "",
			// },
		},
	}
	pb.Properties.User.Add("jwtToken", jwtToken)
	pb.Properties.User.Add("deviceId", localDeviceID)
	pb.Properties.User.Add("businessType", "7")
	pb.Properties.User.Add("businessSubType", "2")
	pb.Properties.User.Add("taskId", taskIdStr)
	pb.Properties.User.Add("code", "0")
	pb.Properties.User.Add("errormsg", "")

	var client *paho.Client
	var payloadCh chan []byte
	payloadCh = make(chan []byte, 0)

	client = business.CreateClient(payloadCh)

	cp := &paho.Connect{
		KeepAlive:  30,
		ClientID:   localDeviceID,
		CleanStart: true,
		Username:   "",
		Password:   []byte(""),
	}
	ca, err := client.Connect(context.Background(), cp)
	if err == nil {
		if ca.ReasonCode == 0 {
			subTopic := fmt.Sprintf("lianmi/cloud/device/%s", localDeviceID)
			if _, err := client.Subscribe(context.Background(), &paho.Subscribe{
				Subscriptions: map[string]paho.SubscribeOptions{
					subTopic: paho.SubscribeOptions{QoS: byte(1), NoLocal: true},
				},
			}); err != nil {
				log.Println("Failed to subscribe:", err)
			}
			log.Println("Subscribed succed: ", subTopic)
		}
	} else {
		log.Println("Failed to Connect mqtt server", err)
	}

	if _, err := client.Publish(context.Background(), pb); err != nil {
		log.Println("Failed to Publish:", err)
	} else {
		log.Println("Succeed Publish to mqtt broker:", topic)
	}

	//堵塞
	payload := <-payloadCh

	//解包负载 payload
	var rsq Order.AddProductRsp
	if err := proto.Unmarshal(payload, &rsq); err != nil {
		log.Println("Protobuf Unmarshal Error", err)

	} else {

		log.Println("回包内容 AddProductRsp ---------------------")
		log.Println("商品ID productID: ", rsq.Product.ProductId)
		pkey := fmt.Sprintf("ProductID:%s", localUserName)
		redisConn.Do("SET", pkey, rsq.Product.ProductId)
	}

	log.Println("AddProduct is Done.")

	return nil

}

//  模拟 7-3 商品编辑更新
func UpdateProduct() error {

	redisConn, err := redis.Dial("tcp", LMCommon.RedisAddr)
	if err != nil {
		log.Fatalln(err)
		return err
	}

	defer redisConn.Close()

	topic := "lianmi/cloud/dispatcher"
	var localUserName string
	var localDeviceID string
	localUserName, _ = redis.String(redisConn.Do("GET", "LocalUserName"))
	localDeviceID, _ = redis.String(redisConn.Do("GET", "LocalDeviceID"))
	if localUserName == "" {
		log.Println("localUserName is  empty")
		return errors.New("localUserName is empty error")
	}
	if localDeviceID == "" {
		log.Println("localDeviceID is  empty")
		return errors.New("localDeviceID is empty error")
	}
	responseTopic := fmt.Sprintf("lianmi/cloud/%s", localDeviceID)

	//从本地redis里获取Token，注意： 在auth模块的登录，登录成功后，需要写入，这里则读取
	key := fmt.Sprintf("jwtToken:%s", localUserName)
	jwtToken, err := redis.String(redisConn.Do("GET", key))
	if err != nil {
		log.Println("Redis GET jwtToken:{LocalUserName}", err)
		return err
	}
	if jwtToken == "" {
		return errors.New("jwtToken is empty error")
	}

	taskId, _ := redis.Int(redisConn.Do("INCR", fmt.Sprintf("taksID:%s", localUserName)))
	taskIdStr := fmt.Sprintf("%d", taskId)
	pkey := fmt.Sprintf("ProductID:%s", localUserName)
	productId, _ := redis.String(redisConn.Do("GET", pkey))
	oProduct := &Order.Product{
		ProductId:   productId,
		Expire:      uint64(0),
		ProductName: "双色球",
		ProductType: Global.ProductType(9), //9-彩票
		SubType:     1,                     //双色球枚举
		ProductDesc: "最高一千万",
		ShortVideo:  "",
		Price:       float32(2.0),
		AllowCancel: true,
	}
	oProduct.ProductPics = append(oProduct.ProductPics, &Order.ProductPic{
		Large: "products/2ff867f849548257d20c003e9de44f98.jpeg", // 原图1
	})
	oProduct.ProductPics = append(oProduct.ProductPics, &Order.ProductPic{
		Large: "products/2ff867f849548257d20c003e9de44f98.jpeg", // 原图2
	})
	oProduct.ProductPics = append(oProduct.ProductPics, &Order.ProductPic{
		Large: "products/2ff867f849548257d20c003e9de44f98.jpeg", // 原图3
	})

	//商品内容图片数组
	oProduct.DescPics = append(oProduct.DescPics, "products/07cb349819583706fee9c08d03434a30.jpeg")
	oProduct.DescPics = append(oProduct.DescPics, "products/048c9822ca1c424080fcbc195abf9624.jpeg")
	oProduct.DescPics = append(oProduct.DescPics, "products/fb5e2fa4e971b0aa4d3d8937e60997c.jpg")
	oProduct.DescPics = append(oProduct.DescPics, "products/4a3f0fe2d58092e2a7f48ce8f58a3df7.jpeg")
	oProduct.DescPics = append(oProduct.DescPics, "products/cf14e9281e6f3819a2001c4b1bdc1301.jpeg")
	oProduct.DescPics = append(oProduct.DescPics, "products/0ffb4f3bc3d419affa6d8fe3efa7eb31.jpeg")

	req := &Order.UpdateProductReq{
		Product: oProduct,
	}

	content, _ := proto.Marshal(req)

	pb := &paho.Publish{
		Topic:   topic,
		QoS:     byte(1),
		Payload: content,
		Properties: &paho.PublishProperties{
			ResponseTopic: responseTopic,
			// User: map[string]string{
			// 	"jwtToken":        jwtToken,      // jwt令牌
			// 	"deviceId":        localDeviceID, // 设备号
			// 	"businessType":    "7",           // 业务号
			// 	"businessSubType": "3",           //  业务子号
			// 	"taskId":          taskIdStr,
			// 	"code":            "0",
			// 	"errormsg":        "",
			// },
		},
	}
	pb.Properties.User.Add("jwtToken", jwtToken)
	pb.Properties.User.Add("deviceId", localDeviceID)
	pb.Properties.User.Add("businessType", "7")
	pb.Properties.User.Add("businessSubType", "3")
	pb.Properties.User.Add("taskId", taskIdStr)
	pb.Properties.User.Add("code", "0")
	pb.Properties.User.Add("errormsg", "")

	var client *paho.Client
	var payloadCh chan []byte
	payloadCh = make(chan []byte, 0)

	client = business.CreateClient(payloadCh)

	cp := &paho.Connect{
		KeepAlive:  30,
		ClientID:   localDeviceID,
		CleanStart: true,
		Username:   "",
		Password:   []byte(""),
	}
	ca, err := client.Connect(context.Background(), cp)
	if err == nil {
		if ca.ReasonCode == 0 {
			subTopic := fmt.Sprintf("lianmi/cloud/device/%s", localDeviceID)
			if _, err := client.Subscribe(context.Background(), &paho.Subscribe{
				Subscriptions: map[string]paho.SubscribeOptions{
					subTopic: paho.SubscribeOptions{QoS: byte(1), NoLocal: true},
				},
			}); err != nil {
				log.Println("Failed to subscribe:", err)
			}
			log.Println("Subscribed succed: ", subTopic)
		}
	} else {
		log.Println("Failed to Connect mqtt server", err)
	}

	if _, err := client.Publish(context.Background(), pb); err != nil {
		log.Println("Failed to Publish:", err)
	} else {
		log.Println("Succeed Publish to mqtt broker:", topic)
	}

	//堵塞
	payload := <-payloadCh

	//解包负载 payload

	_ = payload

	log.Println("UpdateProduct is Done.")

	return nil

}

// 7-4 商品下架
func SoldoutProduct() error {

	redisConn, err := redis.Dial("tcp", LMCommon.RedisAddr)
	if err != nil {
		log.Fatalln(err)
		return err
	}

	defer redisConn.Close()

	topic := "lianmi/cloud/dispatcher"
	var localUserName string
	var localDeviceID string
	localUserName, _ = redis.String(redisConn.Do("GET", "LocalUserName"))
	localDeviceID, _ = redis.String(redisConn.Do("GET", "LocalDeviceID"))
	if localUserName == "" {
		log.Println("localUserName is  empty")
		return errors.New("localUserName is empty error")
	}
	if localDeviceID == "" {
		log.Println("localDeviceID is  empty")
		return errors.New("localDeviceID is empty error")
	}
	responseTopic := fmt.Sprintf("lianmi/cloud/%s", localDeviceID)

	//从本地redis里获取jwtToken，注意： 在auth模块的登录，登录成功后，需要写入，这里则读取
	key := fmt.Sprintf("jwtToken:%s", localUserName)
	jwtToken, err := redis.String(redisConn.Do("GET", key))
	if err != nil {
		log.Println("Redis GET jwtToken:{LocalUserName}", err)
		return err
	}
	if jwtToken == "" {
		return errors.New("jwtToken is empty error")
	}

	taskId, _ := redis.Int(redisConn.Do("INCR", fmt.Sprintf("taksID:%s", localUserName)))
	taskIdStr := fmt.Sprintf("%d", taskId)
	pkey := fmt.Sprintf("ProductID:%s", localUserName)
	productId, _ := redis.String(redisConn.Do("GET", pkey))
	req := &Order.SoldoutProductReq{}
	req.ProductIDs = append(req.ProductIDs, productId)

	content, _ := proto.Marshal(req)

	pb := &paho.Publish{
		Topic:   topic,
		QoS:     byte(1),
		Payload: content,
		Properties: &paho.PublishProperties{
			ResponseTopic: responseTopic,
			// User: map[string]string{
			// 	"jwtToken":        jwtToken,      // jwt令牌
			// 	"deviceId":        localDeviceID, // 设备号
			// 	"businessType":    "7",           // 业务号
			// 	"businessSubType": "4",           //  业务子号
			// 	"taskId":          taskIdStr,
			// 	"code":            "0",
			// 	"errormsg":        "",
			// },
		},
	}

	pb.Properties.User.Add("jwtToken", jwtToken)
	pb.Properties.User.Add("deviceId", localDeviceID)
	pb.Properties.User.Add("businessType", "7")
	pb.Properties.User.Add("businessSubType", "4")
	pb.Properties.User.Add("taskId", taskIdStr)
	pb.Properties.User.Add("code", "0")
	pb.Properties.User.Add("errormsg", "")

	var client *paho.Client
	var payloadCh chan []byte
	payloadCh = make(chan []byte, 0)

	client = business.CreateClient(payloadCh)

	cp := &paho.Connect{
		KeepAlive:  30,
		ClientID:   localDeviceID,
		CleanStart: true,
		Username:   "",
		Password:   []byte(""),
	}
	ca, err := client.Connect(context.Background(), cp)
	if err == nil {
		if ca.ReasonCode == 0 {
			subTopic := fmt.Sprintf("lianmi/cloud/device/%s", localDeviceID)
			if _, err := client.Subscribe(context.Background(), &paho.Subscribe{
				Subscriptions: map[string]paho.SubscribeOptions{
					subTopic: paho.SubscribeOptions{QoS: byte(1), NoLocal: true},
				},
			}); err != nil {
				log.Println("Failed to subscribe:", err)
			}
			log.Println("Subscribed succed: ", subTopic)
		}
	} else {
		log.Println("Failed to Connect mqtt server", err)
	}

	if _, err := client.Publish(context.Background(), pb); err != nil {
		log.Println("Failed to Publish:", err)
	} else {
		log.Println("Succeed Publish to mqtt broker:", topic)
	}

	//堵塞
	payload := <-payloadCh

	//解包负载 payload
	_ = payload

	log.Println("SoldoutProduct is Done.")

	return nil

}
