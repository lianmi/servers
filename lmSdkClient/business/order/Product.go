package order

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"

	"github.com/eclipse/paho.golang/paho" //支持v5.0
	"github.com/golang/protobuf/proto"
	"github.com/gomodule/redigo/redis"
	"github.com/lianmi/servers/util/array"

	Global "github.com/lianmi/servers/api/proto/global"
	Order "github.com/lianmi/servers/api/proto/order"
	LMCommon "github.com/lianmi/servers/lmSdkClient/common"
	clientcommon "github.com/lianmi/servers/lmSdkClient/common"
	"log"
	"time"
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
			CorrelationData: []byte(jwtToken), //jwt令牌
			ResponseTopic:   responseTopic,
			User: map[string]string{
				"deviceId":        localDeviceID, // 设备号
				"businessType":    "7",           // 业务号
				"businessSubType": "1",           // 业务子号
				"taskId":          taskIdStr,
				"code":            "0",
				"errormsg":        "",
			},
		},
	}

	//Connect mqtt broker using ssl
	tlsConfig := NewTlsConfig()
	conn, err := tls.Dial("tcp", clientcommon.BrokerAddr, tlsConfig)
	if err != nil {
		log.Fatalf("Failed to connect to %s: %s", clientcommon.BrokerAddr, err)
	}

	// Create paho client.
	client := paho.NewClient(paho.ClientConfig{
		Router: paho.NewSingleHandlerRouter(func(m *paho.Publish) {
			log.Println("Incoming mqtt broker message")

			topic := m.Topic
			jwtToken := string(m.Properties.CorrelationData)
			deviceId := m.Properties.User["deviceId"]
			businessTypeStr := m.Properties.User["businessType"]
			businessSubTypeStr := m.Properties.User["businessSubType"]
			taskIdStr := m.Properties.User["taskId"]
			code := m.Properties.User["code"]

			log.Println("topic: ", topic)
			log.Println("jwtToken: ", jwtToken)
			log.Println("deviceId: ", deviceId)
			log.Println("businessType: ", businessTypeStr)
			log.Println("businessSubType: ", businessSubTypeStr)
			log.Println("taskId: ", taskIdStr)
			log.Println("code: ", code)

			if businessTypeStr == "7" && businessSubTypeStr == "1" {
				if code == "200" {
					log.Println("response succeed")
					// 回包
					//解包负载 m.Payload
					var rsq Order.QueryProductsRsp
					if err := proto.Unmarshal(m.Payload, &rsq); err != nil {
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

				} else {
					log.Println("QueryProducts failed")
				}
			}

		}),
		Conn: conn,
	})

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

	run := true
	ticker := time.NewTicker(30 * time.Second) // 30s后退出
	for run == true {
		select {
		case <-ticker.C:
			run = false
			break
		}

	}
	log.Println("QueryProducts is Done.")

	return nil
}

// 7-2 商品上架
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
		ProductName: "福彩3D",
		ProductType: Global.ProductType(8), //8-彩票
		ProductDesc: "最新玩法福彩3D",
		ShortVideo:  "",
		Price:       float32(2.0),
		AllowCancel: true,
	}
	oProduct.ProductPics = append(oProduct.ProductPics, &Order.ProductPic{
		Large: "products/215b66d14111da360261206e348c3223.jpg", // 原图1
	})
	oProduct.ProductPics = append(oProduct.ProductPics, &Order.ProductPic{
		Large: "products/215b66d14111da360261206e348c3223.jpg", // 原图2
	})
	oProduct.ProductPics = append(oProduct.ProductPics, &Order.ProductPic{
		Large: "products/215b66d14111da360261206e348c3223.jpg", // 原图3
	})

	//商品内容图片数组
	oProduct.DescPics = append(oProduct.DescPics, "products/215b66d14111da360261206e348c3223.jpg")
	oProduct.DescPics = append(oProduct.DescPics, "products/215b66d14111da360261206e348c3223.jpg")
	oProduct.DescPics = append(oProduct.DescPics, "products/215b66d14111da360261206e348c3223.jpg")
	oProduct.DescPics = append(oProduct.DescPics, "products/215b66d14111da360261206e348c3223.jpg")
	oProduct.DescPics = append(oProduct.DescPics, "products/215b66d14111da360261206e348c3223.jpg")
	oProduct.DescPics = append(oProduct.DescPics, "products/215b66d14111da360261206e348c3223.jpg")
	oProduct.DescPics = append(oProduct.DescPics, "products/215b66d14111da360261206e348c3223.jpg")

	req := &Order.AddProductReq{
		Product:         oProduct,
		OrderType:       Global.OrderType(1), //1- 正常 2-任务抢单类型 3-竞猜类
		OpkBusinessUser: "",
		Expire:          0,
	}

	content, _ := proto.Marshal(req)

	pb := &paho.Publish{
		Topic:   topic,
		QoS:     byte(1),
		Payload: content,
		Properties: &paho.PublishProperties{
			CorrelationData: []byte(jwtToken), //jwt令牌
			ResponseTopic:   responseTopic,
			User: map[string]string{
				"deviceId":        localDeviceID, // 设备号
				"businessType":    "7",           // 业务号
				"businessSubType": "2",           //  业务子号
				"taskId":          taskIdStr,
				"code":            "0",
				"errormsg":        "",
			},
		},
	}

	//Connect mqtt broker using ssl
	tlsConfig := NewTlsConfig()
	conn, err := tls.Dial("tcp", clientcommon.BrokerAddr, tlsConfig)
	if err != nil {
		log.Fatalf("Failed to connect to %s: %s", clientcommon.BrokerAddr, err)
	}

	// Create paho client.
	client := paho.NewClient(paho.ClientConfig{
		Router: paho.NewSingleHandlerRouter(func(m *paho.Publish) {
			log.Println("Incoming mqtt broker message")

			topic := m.Topic
			jwtToken := string(m.Properties.CorrelationData)
			deviceId := m.Properties.User["deviceId"]
			businessTypeStr := m.Properties.User["businessType"]
			businessSubTypeStr := m.Properties.User["businessSubType"]
			taskIdStr := m.Properties.User["taskId"]
			code := m.Properties.User["code"]

			log.Println("topic: ", topic)
			log.Println("jwtToken: ", jwtToken)
			log.Println("deviceId: ", deviceId)
			log.Println("businessType: ", businessTypeStr)
			log.Println("businessSubType: ", businessSubTypeStr)
			log.Println("taskId: ", taskIdStr)
			log.Println("code: ", code)

			if businessTypeStr == "7" && businessSubTypeStr == "2" {
				if code == "200" {
					log.Println("response succeed")
					// 回包
					//解包负载 m.Payload
					var rsq Order.AddProductRsp
					if err := proto.Unmarshal(m.Payload, &rsq); err != nil {
						log.Println("Protobuf Unmarshal Error", err)

					} else {

						log.Println("回包内容 AddProductRsp ---------------------")
						log.Println("商品ID productID: ", rsq.ProductID)
						pkey := fmt.Sprintf("ProductID:%s", localUserName)
						redisConn.Do("SET", pkey, rsq.ProductID)

						//

					}

				} else {
					log.Println("AddProduct failed")
				}
			}

		}),
		Conn: conn,
	})

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

	run := true
	ticker := time.NewTicker(30 * time.Second) // 30s后退出
	for run == true {
		select {
		case <-ticker.C:
			run = false
			break
		}

	}
	log.Println("AddProduct is Done.")

	return nil

}

// 7-3 商品编辑更新
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
		ProductName: "福彩3D",
		ProductType: Global.ProductType(8), //8-彩票
		ProductDesc: "最新玩法福彩3D",
		ShortVideo:  "",
		Price:       float32(2.0),
		AllowCancel: true,
	}
	oProduct.ProductPics = append(oProduct.ProductPics, &Order.ProductPic{
		Large: "products/215b66d14111da360261206e348c3223.jpg", // 原图1
	})
	oProduct.ProductPics = append(oProduct.ProductPics, &Order.ProductPic{
		Large: "products/215b66d14111da360261206e348c3223.jpg", // 原图2
	})
	oProduct.ProductPics = append(oProduct.ProductPics, &Order.ProductPic{
		Large: "products/215b66d14111da360261206e348c3223.jpg", // 原图3
	})

	//商品内容图片数组
	oProduct.DescPics = append(oProduct.DescPics, "products/215b66d14111da360261206e348c3223.jpg")
	oProduct.DescPics = append(oProduct.DescPics, "products/215b66d14111da360261206e348c3223.jpg")
	oProduct.DescPics = append(oProduct.DescPics, "products/215b66d14111da360261206e348c3223.jpg")
	oProduct.DescPics = append(oProduct.DescPics, "products/215b66d14111da360261206e348c3223.jpg")
	oProduct.DescPics = append(oProduct.DescPics, "products/215b66d14111da360261206e348c3223.jpg")
	oProduct.DescPics = append(oProduct.DescPics, "products/215b66d14111da360261206e348c3223.jpg")
	oProduct.DescPics = append(oProduct.DescPics, "products/215b66d14111da360261206e348c3223.jpg")

	req := &Order.UpdateProductReq{
		Product:   oProduct,
		OrderType: Global.OrderType(1), //1- 正常 2-任务抢单类型 3-竞猜类
		Expire:    0,
	}

	content, _ := proto.Marshal(req)

	pb := &paho.Publish{
		Topic:   topic,
		QoS:     byte(1),
		Payload: content,
		Properties: &paho.PublishProperties{
			CorrelationData: []byte(jwtToken), //jwt令牌
			ResponseTopic:   responseTopic,
			User: map[string]string{
				"deviceId":        localDeviceID, // 设备号
				"businessType":    "7",           // 业务号
				"businessSubType": "3",           //  业务子号
				"taskId":          taskIdStr,
				"code":            "0",
				"errormsg":        "",
			},
		},
	}

	//Connect mqtt broker using ssl
	tlsConfig := NewTlsConfig()
	conn, err := tls.Dial("tcp", clientcommon.BrokerAddr, tlsConfig)
	if err != nil {
		log.Fatalf("Failed to connect to %s: %s", clientcommon.BrokerAddr, err)
	}

	// Create paho client.
	client := paho.NewClient(paho.ClientConfig{
		Router: paho.NewSingleHandlerRouter(func(m *paho.Publish) {
			log.Println("Incoming mqtt broker message")

			topic := m.Topic
			jwtToken := string(m.Properties.CorrelationData)
			deviceId := m.Properties.User["deviceId"]
			businessTypeStr := m.Properties.User["businessType"]
			businessSubTypeStr := m.Properties.User["businessSubType"]
			taskIdStr := m.Properties.User["taskId"]
			code := m.Properties.User["code"]

			log.Println("topic: ", topic)
			log.Println("jwtToken: ", jwtToken)
			log.Println("deviceId: ", deviceId)
			log.Println("businessType: ", businessTypeStr)
			log.Println("businessSubType: ", businessSubTypeStr)
			log.Println("taskId: ", taskIdStr)
			log.Println("code: ", code)

			if businessTypeStr == "7" && businessSubTypeStr == "3" {
				if code == "200" {
					log.Println("response succeed")

				} else {
					log.Println("UpdateProduct failed")
				}
			}

		}),
		Conn: conn,
	})

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

	run := true
	ticker := time.NewTicker(30 * time.Second) // 30s后退出
	for run == true {
		select {
		case <-ticker.C:
			run = false
			break
		}

	}
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
	req := &Order.SoldoutProductReq{
		ProductID: productId,
	}

	content, _ := proto.Marshal(req)

	pb := &paho.Publish{
		Topic:   topic,
		QoS:     byte(1),
		Payload: content,
		Properties: &paho.PublishProperties{
			CorrelationData: []byte(jwtToken), //jwt令牌
			ResponseTopic:   responseTopic,
			User: map[string]string{
				"deviceId":        localDeviceID, // 设备号
				"businessType":    "7",           // 业务号
				"businessSubType": "4",           //  业务子号
				"taskId":          taskIdStr,
				"code":            "0",
				"errormsg":        "",
			},
		},
	}

	//Connect mqtt broker using ssl
	tlsConfig := NewTlsConfig()
	conn, err := tls.Dial("tcp", clientcommon.BrokerAddr, tlsConfig)
	if err != nil {
		log.Fatalf("Failed to connect to %s: %s", clientcommon.BrokerAddr, err)
	}

	// Create paho client.
	client := paho.NewClient(paho.ClientConfig{
		Router: paho.NewSingleHandlerRouter(func(m *paho.Publish) {
			log.Println("Incoming mqtt broker message")

			topic := m.Topic
			jwtToken := string(m.Properties.CorrelationData)
			deviceId := m.Properties.User["deviceId"]
			businessTypeStr := m.Properties.User["businessType"]
			businessSubTypeStr := m.Properties.User["businessSubType"]
			taskIdStr := m.Properties.User["taskId"]
			code := m.Properties.User["code"]

			log.Println("topic: ", topic)
			log.Println("jwtToken: ", jwtToken)
			log.Println("deviceId: ", deviceId)
			log.Println("businessType: ", businessTypeStr)
			log.Println("businessSubType: ", businessSubTypeStr)
			log.Println("taskId: ", taskIdStr)
			log.Println("code: ", code)

			if businessTypeStr == "7" && businessSubTypeStr == "4" {
				if code == "200" {
					log.Println("response succeed")

				} else {
					log.Println("SoldoutProduct failed")
				}
			}

		}),
		Conn: conn,
	})

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

	run := true
	ticker := time.NewTicker(30 * time.Second) // 30s后退出
	for run == true {
		select {
		case <-ticker.C:
			run = false
			break
		}

	}
	log.Println("SoldoutProduct is Done.")

	return nil

}
