package order

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"

	"github.com/eclipse/paho.golang/paho" //支持v5.0
	"github.com/golang/protobuf/proto"
	"github.com/gomodule/redigo/redis"

	Global "github.com/lianmi/servers/api/proto/global"
	Order "github.com/lianmi/servers/api/proto/order"
	LMCommon "github.com/lianmi/servers/lmSdkClient/common"
	clientcommon "github.com/lianmi/servers/lmSdkClient/common"
	"io/ioutil"
	"log"
	"time"
)

func NewTlsConfig() *tls.Config {
	certpool := x509.NewCertPool()
	ca, err := ioutil.ReadFile(clientcommon.CaPath + "/ca.crt")
	if err != nil {
		log.Fatalln(err.Error())
	} else {
		log.Println("ReadFile ok")
	}
	certpool.AppendCertsFromPEM(ca)
	clientKeyPair, err := tls.LoadX509KeyPair(clientcommon.CaPath+"/mqtt.lianmi.cloud.crt", clientcommon.CaPath+"/mqtt.lianmi.cloud.key")
	if err != nil {
		panic(err)
	} else {
		log.Println("LoadX509KeyPair ok")
	}
	return &tls.Config{
		RootCAs:            certpool,
		ClientAuth:         tls.NoClientCert,
		ClientCAs:          nil,
		InsecureSkipVerify: true,
		Certificates:       []tls.Certificate{clientKeyPair},
	}
}

// 7-1  查询某个商户的所有商品信息
func QueryProducts() error {

	//TODO
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
	var adminName string
	var adminDeviceID string
	adminName, _ = redis.String(redisConn.Do("GET", "AdminName"))
	adminDeviceID, _ = redis.String(redisConn.Do("GET", "AdminDeviceID"))
	if adminName == "" {
		log.Println("adminName is empty")
		return errors.New("adminName is empty error")
	}
	if adminDeviceID == "" {
		log.Println("adminDeviceID is empty")
		return errors.New("adminDeviceID is empty error")
	}

	responseTopic := fmt.Sprintf("lianmi/cloud/%s", adminDeviceID)

	//从本地redis里获取jwtToken，注意： 在auth模块的登录，登录成功后，需要写入，这里则读取
	key := fmt.Sprintf("jwtToken:%s", adminName)
	jwtToken, err := redis.String(redisConn.Do("GET", key))
	if err != nil {
		log.Println("Redis GET jwtToken:{adminName}", err)
		return err
	}
	if jwtToken == "" {
		return errors.New("jwtToken is empty error")
	}

	taskId, _ := redis.Int(redisConn.Do("INCR", fmt.Sprintf("taksID:%s", adminName)))
	taskIdStr := fmt.Sprintf("%d", taskId)

	req := &Order.AddProductReq{
		Product: &Order.Product{
			Expire:            uint64(0),
			ProductName:       "福彩3D",
			ProductType:       Global.ProductType(8), //8-彩票
			ProductDesc:       "最新玩法福彩3D",
			ProductPic1Small:  "",
			ProductPic1Middle: "",
			ProductPic1Large:  "/product/215b66d14111da360261206e348c3223.jpg", // 原图
			ProductPic2Small:  "",
			ProductPic2Middle: "",
			ProductPic2Large:  "",
			ProductPic3Small:  "",
			ProductPic3Middle: "",
			ProductPic3Large:  "",
			Thumbnail:         "",
			ShortVideo:        "",
			Price:             float32(2.0),
			AllowCancel:       true,
		},
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
				"deviceId":        adminDeviceID, // 设备号
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

			if code == "200" {
				log.Println("Wallet register succeed")
				// 回包
				//解包负载 m.Payload
				var rsq Order.AddProductRsp
				if err := proto.Unmarshal(m.Payload, &rsq); err != nil {
					log.Println("Protobuf Unmarshal Error", err)

				} else {

					log.Println("回包内容 AddProductRsp ---------------------")
					log.Println("商品ID productID: ", rsq.ProductID)

					//

				}

			} else {
				log.Println("AddProduct failed")
			}

		}),
		Conn: conn,
	})

	cp := &paho.Connect{
		KeepAlive:  30,
		ClientID:   adminDeviceID,
		CleanStart: true,
		Username:   "",
		Password:   []byte(""),
	}
	ca, err := client.Connect(context.Background(), cp)
	if err == nil {
		if ca.ReasonCode == 0 {
			subTopic := fmt.Sprintf("lianmi/cloud/device/%s", adminDeviceID)
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
	ticker := time.NewTicker(5 * time.Second) // 5s后退出
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
