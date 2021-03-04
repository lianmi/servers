package sync

import (
	"context"
	"errors"
	"fmt"
	"log"

	// "net"
	"github.com/lianmi/servers/lmSdkClient/business"

	"github.com/eclipse/paho.golang/paho" //支持v5.0
	"github.com/golang/protobuf/proto"
	"github.com/gomodule/redigo/redis"
	Sync "github.com/lianmi/servers/api/proto/syn"
	clientcommon "github.com/lianmi/servers/lmSdkClient/common"
)

func DoSyncEvent() error {
	redisConn, err := redis.Dial("tcp", clientcommon.RedisAddr)
	if err != nil {
		log.Fatalln(err)
		return err
	}

	defer redisConn.Close()

	req := &Sync.SyncEventReq{
		MyInfoAt:      0,
		FriendsAt:     0,
		FriendUsersAt: 0,
		TeamsAt:       0,
		TagsAt:        0,
		WatchAt:       0,
		ProductAt:     0,
	}
	content, _ := proto.Marshal(req)
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
		log.Println("Redis GET jwtToken:{localUserName}", err)
		return err
	}
	if jwtToken == "" {
		return errors.New("jwtToken is empty error")
	}

	taskId, _ := redis.Int(redisConn.Do("INCR", fmt.Sprintf("taksID:%s", localUserName)))
	taskIdStr := fmt.Sprintf("%d", taskId)

	pb := &paho.Publish{
		Topic:   topic,
		QoS:     byte(1),
		Payload: content,
		Properties: &paho.PublishProperties{
			ResponseTopic: responseTopic,
			User: map[string]string{
				"jwtToken":        jwtToken,      // jwt令牌
				"deviceId":        localDeviceID, // 设备号
				"businessType":    "6",           // 业务号
				"businessSubType": "1",           //  业务子号
				"taskId":          taskIdStr,
				"code":            "0",
				"errormsg":        "",
			},
		},
	}

	// pb.Properties.User.Add("jwtToken", jwtToken)
	// pb.Properties.User.Add("deviceId", localDeviceID)
	// pb.Properties.User.Add("businessType", "6")
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

	_ = payload

	log.Println("DoSyncEvent Done.")

	return nil

}
