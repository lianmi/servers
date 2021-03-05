/*
./lmSdkClient order BuyVipUser  -p ada166df-bb9f-4274-ab8d-e369a68d69ce -I 9.9
*/

package order

import (
	"context"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/lianmi/servers/lmSdkClient/business"

	"github.com/eclipse/paho.golang/paho" //支持v5.0
	"github.com/golang/protobuf/proto"
	"github.com/gomodule/redigo/redis"

	// "github.com/lianmi/servers/util/array"
	"github.com/lianmi/servers/internal/pkg/models"

	"log"

	Global "github.com/lianmi/servers/api/proto/global"
	Msg "github.com/lianmi/servers/api/proto/msg"
	Order "github.com/lianmi/servers/api/proto/order"
	clientcommon "github.com/lianmi/servers/lmSdkClient/common"
	"github.com/lianmi/servers/util/array"
)

//向商户id3 购买Vip会员
func BuyVipUser(price float64, orderID, productID string) error {
	var attach string
	redisConn, err := redis.Dial("tcp", clientcommon.RedisAddr)
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
	vipUser := models.VipUser{
		PayType: 3, //包月
	}

	attachBase := new(models.AttachBase)
	attachBase.BodyType = 99 //约定99为购买Vip会员的type
	bodyTemp, _ := vipUser.ToJson()
	attachBase.Body = base64.StdEncoding.EncodeToString([]byte(bodyTemp)) //约定，购买会员 及 服务费的attach的body部分，都是base64

	attach, _ = attachBase.ToJson()
	attachHex := hex.EncodeToString([]byte(attach)) //hex

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
		OrderTotalAmount: price,
		// json 格式的内容 , 由 ui 层处理 sdk 仅透传
		// 传输会进过sdk 处理
		Attach: attachHex, //hex
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
		Uuid:  "buyvipuser",                  //本地uuid
		Body:  orderContent,                  //订单包体
	}

	content, _ := proto.Marshal(msgReq)

	pb := &paho.Publish{
		Topic:      topic,
		QoS:        byte(2),
		Payload:    content,
		Properties: business.GeneProps(responseTopic, jwtToken, localDeviceID, "5", "1", taskIdStr, "0", ""),
	}

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
		return err

	} else {

		log.Println("购买Vip会员下单 回包内容 ---------------------")

		//解包body
		var orderProductBody = new(Order.OrderProductBody)
		if err := proto.Unmarshal(rsq.Body, orderProductBody); err != nil {
			log.Println("Protobuf Unmarshal Error", err)
		} else {
			array.PrintPretty(orderProductBody)
			attachData, _ := hex.DecodeString(orderProductBody.Attach) //反hex
			attachBase := new(models.AttachBase)
			if err := json.Unmarshal([]byte(attachData), attachBase); err != nil {
				log.Println("解包attachData failed, error: ", err)
				return err
			}
			if attachBase.BodyType == 99 {
				// log.Println("attach解析 payType:", vu.PayType)
				// if orderProductBody.State == 4 {
				//OS_Taked        = 4;     /**< 已接单*/
				vu := new(models.VipUser)
				bodyData, err := base64.StdEncoding.DecodeString(attachBase.Body)
				if err != nil {
					log.Println("base64.StdEncoding.DecodeString failed, error: ", err)
					return err
				}
				vu, err = models.VipUserFromJson(bodyData)
				if err != nil {
					log.Println("VipUserFromJson failed, error: ", err)
					return err
				}
				log.Printf("已接单, Vip会员类型: %d,  价格是: %f, 下一步请发起支付", vu.PayType, vu.Price)
			}

		}

	}

	log.Println("BuyVipUser is Done.")

	return nil
}
