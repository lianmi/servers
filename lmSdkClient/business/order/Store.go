/*
商户及订单相关
对应文档的第九章
*/

package order

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"

	"github.com/eclipse/paho.golang/paho" //支持v5.0
	"github.com/golang/protobuf/proto"
	"github.com/gomodule/redigo/redis"

	Global "github.com/lianmi/servers/api/proto/global"
	Order "github.com/lianmi/servers/api/proto/order"
	LMCommon "github.com/lianmi/servers/lmSdkClient/common"
	clientcommon "github.com/lianmi/servers/lmSdkClient/common"
	"log"
	"time"
)

//9-1 商户上传订单DH加密公钥
func RegisterPreKeys() error {

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

	req := &Order.RegisterPreKeysReq{}
	req.PreKeys = append(req.PreKeys, "62C76C2FC49107A3D512F44082C7A845C20F8B9B22451A9921BE1CCC1E14FA19")
	req.PreKeys = append(req.PreKeys, "D139AFFB1FC86CF07C3D9CA119BF3702A3D26557BEC1CBF170556933A62FB92A")
	req.PreKeys = append(req.PreKeys, "D1E702D92C9A5E1A7C3FEE0F6998DA95E08938B5C185B98138105EEED3985A54")
	req.PreKeys = append(req.PreKeys, "04BBC68B78808AC7C1735900F52441613EA414220E2679BCA03234F76B4E585F")
	req.PreKeys = append(req.PreKeys, "8A4685819A978E04F0E15111BE6370EC36128185591CFA52F7ED800FD5B1560B")
	req.PreKeys = append(req.PreKeys, "961B09FBA687C8409DC42DC23ACEF3CB9885643E92E04A3CAEFFACB93536B304")
	req.PreKeys = append(req.PreKeys, "CB56FD536579A97E1E15B7F8EA2BA2DB9C8F57A57C6B17BB3D94EC4F68B6313E")
	req.PreKeys = append(req.PreKeys, "282BFB58DE91EBA200D3670746F0DB4F98AF56B1BDC7B297FBD0504B1A322D33")
	req.PreKeys = append(req.PreKeys, "49A32DAA6CF3DDF9E38FC4ADAF4179D66F4C6C3934F2B87C3E91BA9CCE031A72")
	req.PreKeys = append(req.PreKeys, "AC7C472C8C3B60B1B48D132A0A02D1D12C95734D6CA2908D8EBEFD4880D4D13D")
	req.PreKeys = append(req.PreKeys, "5DEB22DB0A6BAB9471D3A7B34306B3EF284369997B6FD09147E5AB33847B0245")
	req.PreKeys = append(req.PreKeys, "65A5ABEFFC862E236CFCA516ABFD571E41129DD4B3E5F51304D6DDC254681E3F")
	req.PreKeys = append(req.PreKeys, "8B2E07F03C153686B1A50516C6A52B6EEA59162832434E95C6B19CF2F9E9FD40")
	req.PreKeys = append(req.PreKeys, "ADA4D67EDB341258BF891A5B56609BC251EFBBFB1FEA8EDC8FA6E4FF63BB7B30")
	req.PreKeys = append(req.PreKeys, "1FD921C1BB6DFD980A7DE1D2B5DADED2C2B06E5BA1E4DE34907DA0D80896DA62")
	req.PreKeys = append(req.PreKeys, "3E258A22BC305A842E8B0F89567DC4D1ED3B6FB9774DF6B30856648AEE98111F")

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

			if businessTypeStr == "9" && businessSubTypeStr == "1" {
				if code == "200" {
					log.Println("response succeed")

				} else {
					log.Println("RegisterPreKeys failed")
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
	log.Println("RegisterPreKeys is Done.")

	return nil
}

//9-2 获取网点OPK公钥及订单ID, 只允许普通用户获取
func GetPreKeyOrderID(productId string) error {

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

	if productId == "" {
		pkey := fmt.Sprintf("ProductID:%s", localUserName)
		productId, _ = redis.String(redisConn.Do("GET", pkey))

	}

	req := &Order.GetPreKeyOrderIDReq{
		UserName:  "id3",
		OrderType: Global.OrderType(1),
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
				"businessType":    "9",           // 业务号
				"businessSubType": "2",           // 业务子号
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

					log.Println("回包内容 GetPreKeyOrderIDRsp ---------------------")
					log.Println("商户的账号id userName: ", rsq.UserName)
					log.Println("商品ID productID: ", rsq.ProductID)
					log.Println("订单类型 orderType: ", int(rsq.OrderType))
					log.Println("一次性公钥, hex方式 pubKey: ", rsq.PubKey)
					log.Println("订单ID orderID: ", rsq.OrderID)

					okey := fmt.Sprintf("OrderID:%s", localUserName)
					redisConn.Do("SET", okey, rsq.OrderID)

					//

				}

			} else {
				log.Println("GetPreKeyOrderID failed")
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
	log.Println("GetPreKeyOrderID is Done.")

	return nil
}

//9-3 下单
func AddOrder(orderID, productID string) error {

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

	if orderID == "" {
		okey := fmt.Sprintf("OrderID:%s", localUserName)
		orderID, _ = redis.String(redisConn.Do("GET", okey))

	}

	if productID == "" {
		pkey := fmt.Sprintf("ProductID:%s", localUserName)
		productID, _ = redis.String(redisConn.Do("GET", pkey))

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
		Attach: "",
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
				"businessSubType": "2",           // 业务子号
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
