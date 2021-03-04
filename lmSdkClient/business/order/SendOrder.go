package order

import (
	"context"
	"errors"

	"github.com/lianmi/servers/lmSdkClient/business"

	"fmt"

	"github.com/eclipse/paho.golang/paho" //支持v5.0
	"github.com/golang/protobuf/proto"
	"github.com/gomodule/redigo/redis"

	// "github.com/lianmi/servers/util/array"
	"log"
	"time"

	Global "github.com/lianmi/servers/api/proto/global"
	Msg "github.com/lianmi/servers/api/proto/msg"
	Order "github.com/lianmi/servers/api/proto/order"
	"github.com/lianmi/servers/internal/pkg/models"
	LMCommon "github.com/lianmi/servers/lmSdkClient/common"
	"github.com/lianmi/servers/util/array"
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
		DantuoBalls: nil,
		RedBalls:    []int{1, 2, 3, 4, 5, 6},
		BlueBalls:   []int{9},
	})

	bodyJsonData, _ := ssqOrder.ToJson()
	attachBase := new(models.AttachBase)
	attachBase.BodyType = 9 //约定9为购买双色球
	attachBase.Body = bodyJsonData
	attach, err = attachBase.ToJson()
	if err != nil {
		return errors.New("ssqOrder.ToJson error")
	}

	req := &Order.OrderProductBody{
		OrderID:   orderID,
		ProductID: productID,
		//买家账号
		BuyUser: "id1",
		//买家的协商公钥, 必填
		OpkBuyUser: "bbbbb",
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

	orderContent, _ := proto.Marshal(req)
	msgReq := &Msg.SendMsgReq{
		Scene: Msg.MessageScene_MsgScene_C2C, //个人
		Type:  Msg.MessageType_MsgType_Order, //订单消息
		To:    "id3",                         //商户
		Uuid:  "aaaaaa",                      //本地uuid，随机
		Body:  orderContent,                  //订单包体
	}

	content, _ := proto.Marshal(msgReq)

	pb := &paho.Publish{
		Topic:   topic,
		QoS:     byte(1),
		Payload: content,
		Properties: &paho.PublishProperties{
			ResponseTopic: responseTopic,
			User: map[string]string{
				"jwtToken":        jwtToken,      // jwt令牌
				"deviceId":        localDeviceID, // 设备号
				"businessType":    "5",           // 业务号
				"businessSubType": "1",           // 业务子号
				"taskId":          taskIdStr,
				"code":            "0",
				"errormsg":        "",
			},
		},
	}
	// pb.Properties.User.Add("jwtToken", jwtToken)
	// pb.Properties.User.Add("deviceId", localDeviceID)
	// pb.Properties.User.Add("businessType", "5")
	// pb.Properties.User.Add("businessSubType", "1")
	// pb.Properties.User.Add("taskId", taskIdStr)
	// pb.Properties.User.Add("code", "0")
	// pb.Properties.User.Add("errormsg", "")

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
	var rsq Msg.RecvMsgEventRsp
	if err := proto.Unmarshal(payload, &rsq); err != nil {
		log.Println("Protobuf Unmarshal Error", err)

	} else {

		log.Println("回包内容 ---------------------")

		array.PrintPretty(rsq)

	}

	log.Println("AddOrder is Done.")

	return nil
}
