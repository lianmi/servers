/*
商户及订单相关
对应文档的第九章
*/

package order

import (
	"context"
	"errors"
	"fmt"

	"github.com/lianmi/servers/lmSdkClient/business"
	clientcommon "github.com/lianmi/servers/lmSdkClient/common"

	"github.com/eclipse/paho.golang/paho" //支持v5.0
	"github.com/golang/protobuf/proto"
	"github.com/gomodule/redigo/redis"

	"log"

	Global "github.com/lianmi/servers/api/proto/global"
	Order "github.com/lianmi/servers/api/proto/order"
)

//9-2 获取网点OPK公钥及订单ID, 只允许普通用户获取
func GetPreKeyOrderID(productId string) error {

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
		Topic:      topic,
		QoS:        byte(1),
		Payload:    content,
		Properties: business.GeneProps(responseTopic, jwtToken, localDeviceID, "9", "2", taskIdStr, "0", ""),
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

	var rsq Order.GetPreKeyOrderIDRsp
	if err := proto.Unmarshal(payload, &rsq); err != nil {
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

	}

	log.Println("GetPreKeyOrderID is Done.")

	return nil
}
