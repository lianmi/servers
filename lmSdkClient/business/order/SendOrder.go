package order

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"

	"github.com/eclipse/paho.golang/paho" //支持v5.0
	"github.com/golang/protobuf/proto"
	"github.com/gomodule/redigo/redis"
	// "github.com/lianmi/servers/util/array"

	Global "github.com/lianmi/servers/api/proto/global"
	Order "github.com/lianmi/servers/api/proto/order"
	LMCommon "github.com/lianmi/servers/lmSdkClient/common"
	clientcommon "github.com/lianmi/servers/lmSdkClient/common"
	"log"
	"time"
)

//9-3 下单
func AddOrder(orderID, productID string) error {
	var attach string
	// var imageFile = "/Users/mac/developments/lianmi/产品/图片/双色球彩票.jpeg"
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
	} else {
		log.Println("localUserName:", localUserName)
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

	if orderID == "" {
		okey := fmt.Sprintf("OrderID:%s", localUserName)
		orderID, _ = redis.String(redisConn.Do("GET", okey))
		log.Println("orderID:", orderID)
	}

	if productID == "" {
		pkey := fmt.Sprintf("ProductID:%s", localUserName)
		productID, _ = redis.String(redisConn.Do("GET", pkey))
		if productID == "" {
			return errors.New("productID is empty error")
		}

		log.Println("productID:", productID)
	}

	ssqOrder := ShuangSeQiuOrder{
		BuyUser:          "id1",     //买家
		BusinessUsername: "id3",     //商户注册账号
		ProductID:        productID, //商品ID
		OrderID:          orderID,   //订单ID
		// LotteryPicHash:   "",                          //彩票拍照的照片md5
		// LotteryPicURL:    "",                          //彩票拍照的照片原图 oss objectID
		TicketType:  1,                           //彩票投注类型，1-单式\2-复式\3-胆拖
		Count:       1,                           //总注数
		Cost:        2.0,                         //花费, 每注2元, 乘以总注数
		TxHash:      "",                          //上链的哈希
		BlockNumber: 0,                           //区块高度
		CreatedAt:   time.Now().UnixNano() / 1e6, //创建订单的时刻，服务端为准
	}

	//单式
	ssqOrder.Straws = append(ssqOrder.Straws, &ShuangSeQiu{
		DantuoBall: nil,
		RedBall:    []int{1, 2, 3, 4, 5, 6},
		BlueBall:   []int{9},
	})

	attach, err = ssqOrder.ToJson()
	if err != nil {
		return errors.New("ssqOrder.ToJson error")
	}

	req := &Order.OrderProductBody{
		OrderID:   orderID,
		ProductID: productID,
		//买家账号
		BuyUser: "id1",
		//买家的协商公钥
		OpkBuyUser: "",
		//商户账号
		BusinessUser: "id3",
		//商户的协商公钥
		OpkBusinessUser: "",
		// 订单的总金额, 支付的时候以这个金额 计算
		OrderTotalAmount: 2.00,
		// json 格式的内容 , 由 ui 层处理 sdk 仅透传
		// 传输会进过sdk 处理
		Attach: attach,
		// 透传信息 , 不加密 ，直接传过去 不处理
		Userdata: nil,
		//订单的状态;
		State: Global.OrderState_OS_Prepare,
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
				"businessType":    "9",           // 业务号
				"businessSubType": "3",           // 业务子号
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

			if code == "200" {
				log.Println("response succeed")
				// 回包
				//解包负载 m.Payload
				var rsq Order.GetPreKeyOrderIDRsp
				if err := proto.Unmarshal(m.Payload, &rsq); err != nil {
					log.Println("Protobuf Unmarshal Error", err)

				} else {

					log.Println("回包内容 ---------------------")

					//

				}

			} else {
				log.Println("AddOrder failed")
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
	log.Println("AddOrder is Done.")

	return nil
}
