package chat

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/lianmi/servers/lmSdkClient/business"

	"github.com/eclipse/paho.golang/paho" //支持v5.0
	"github.com/golang/protobuf/proto"
	"github.com/gomodule/redigo/redis"

	Msg "github.com/lianmi/servers/api/proto/msg"
	clientcommon "github.com/lianmi/servers/lmSdkClient/common"
)

// 5-12 获取阿里云临时令牌
func GetOssToken(isPrivate bool) error {

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
		log.Println("localDeviceID is empty")
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

	req := &Msg.GetOssTokenReq{
		IsPrivate: isPrivate,
	}

	content, _ := proto.Marshal(req)

	pb := &paho.Publish{
		Topic:   topic,
		QoS:     byte(2),
		Payload: content,
		Properties: &paho.PublishProperties{
			ResponseTopic: responseTopic,
			User: map[string]string{
				"jwtToken":        jwtToken,      // jwt令牌
				"deviceId":        localDeviceID, // 设备号
				"businessType":    "5",           // 业务号
				"businessSubType": "12",          //  业务子号
				"taskId":          taskIdStr,
				"code":            "0",
				"errormsg":        "",
			},
		},
	}
	// pb.Properties.User.Add("jwtToken", jwtToken)
	// pb.Properties.User.Add("deviceId", localDeviceID)
	// pb.Properties.User.Add("businessType", "5")
	// pb.Properties.User.Add("businessSubType", "12")
	// pb.Properties.User.Add("taskId", taskIdStr)
	// pb.Properties.User.Add("code", "0")
	// pb.Properties.User.Add("errormsg", string(msg.GetErrorMsg()))
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
		return err
	} else {
		log.Println("Succeed Publish to mqtt broker:", topic)
	}

	//堵塞
	payload := <-payloadCh

	//解包负载 payload
	var rsq Msg.GetOssTokenRsp
	if err := proto.Unmarshal(payload, &rsq); err != nil {
		log.Println("Protobuf Unmarshal Error", err)

	} else {

		log.Println("回包内容---------------------")
		log.Println(rsq)
		log.Println("资源服务器地址 EndPoint: ", rsq.EndPoint)
		log.Println("空间名称 BucketName: ", rsq.BucketName)
		log.Println("Bucket访问凭证 AccessKeyId: ", rsq.AccessKeyId)
		log.Println("Bucket访问密钥 AccessKeySecret: ", rsq.AccessKeySecret)
		log.Println("安全凭证 SecurityToken: ", rsq.SecurityToken)
		log.Println("oss的文件目录 Directory: ", rsq.Directory)

		//保存到redis
		redisConn.Do("SET", "OSSEndPoint", rsq.EndPoint)
		redisConn.Do("SET", "OSSBucketName", rsq.BucketName)
		redisConn.Do("SET", "OSSAccessKeyId", rsq.AccessKeyId)
		redisConn.Do("SET", "OSSAccessKeySecret", rsq.AccessKeySecret)
		redisConn.Do("SET", "OSSSecurityToken", rsq.SecurityToken)
		redisConn.Do("SET", "OSSDirectory", rsq.Directory)
		/*
			2020/11/23 15:24:28 资源服务器地址 EndPoint:  https://oss-cn-hangzhou.aliyuncs.com
			2020/11/23 15:24:28 空间名称 BucketName:  lianmi-ipfs
			2020/11/23 15:24:28 Bucket访问凭证 AccessKeyId:  STS.NToa6SbpTV9XNNhjCG68FZWiB
			2020/11/23 15:24:28 Bucket访问密钥 AccessKeySecret:  5EgiwHWw5YQojjiobyiwLxB49Hi2X5YUXQh134DtQAZ
			2020/11/23 15:24:28 安全凭证 SecurityToken:  CAIS8QF1q6Ft5B2yfSjIr5faKoznj6914fuzTGjZjkMSOrdqtZLCoDz2IH1Fe3ZtBu0Wvv42mGhR6vcblq94T55IQ1CcmyvJJyMRo22beIPkl5Gfz95t0e+IewW6Dxr8w7WhAYHQR8/cffGAck3NkjQJr5LxaTSlWS7OU/TL8+kFCO4aRQ6ldzFLKc5LLw950q8gOGDWKOymP2yB4AOSLjIx4FEk1T8hufngnpPBtEWFtjCglL9J/baWC4O/csxhMK14V9qIx+FsfsLDqnUNukcVqfgr3PweoGuf543MWkM14g2IKPfM9tpmIAJjdgmMmRj3JgeWGoABacemwmaJvPS4R/oV5wbS2QS7xZTnEU1HFDqNyFsP+QdhQTrRD/h1Utlg2z1+xcZr6J54nVO8xTH1pshEPlw3MBnsHW3Jq31NQHdPppMoE5d0Qd1aMnlFgC+pQUNu5n1TyxU8BVCfHFT62EhT+EZz6ugpQ1LmQh1/a35zlCOo6oQ=
			2020/11/23 15:24:28 oss的文件目录 Directory:  2020/11/23/
		*/

		//

	}

	log.Println("GetOssToken is Done.")

	return nil

}
