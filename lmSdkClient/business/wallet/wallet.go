package wallet

import (
	"context"
	"errors"
	"fmt"
	"net"
	"time"

	"github.com/golang/protobuf/proto"
	Wallet "github.com/lianmi/servers/api/proto/wallet"
	"github.com/lianmi/servers/lmSdkClient/common"
	"log"

	"github.com/eclipse/paho.golang/paho" //支持v5.0
	"github.com/gomodule/redigo/redis"
	"github.com/miguelmota/go-ethereum-hdwallet"
)

const (
	mnemonic = "cloth have cage erase shrug slot album village surprise fence erode direct"
)

//创建 用户 HD钱包
func CreateHDWallet() string {

	// mnemonic := LMCommon.MnemonicServer // "element urban soda endless beach celery scheme wet envelope east glory retire"
	log.Println("mnemonic:", mnemonic)

	wallet, err := hdwallet.NewFromMnemonic(mnemonic)
	if err != nil {
		log.Fatalln(err)
		return ""
	}

	path := hdwallet.MustParseDerivationPath("m/44'/60'/0'/0/0")
	account, err := wallet.Derive(path, true)
	if err != nil {
		log.Fatalln(err)
		return ""
	}
	log.Printf("Address m/44'/60'/0'/0/0 in hex: %s\n", account.Address.Hex())
	privateKey, err := wallet.PrivateKey(account)
	if err != nil {
		log.Fatalln(err)
		return ""
	}
	privateKeyHex, err := wallet.PrivateKeyHex(account)
	if err != nil {
		log.Fatalln(err)
		return ""
	}
	fmt.Printf("Private key m/44'/60'/0'/0/0 in hex: %s\n", privateKeyHex)

	publicKeyHex, _ := wallet.PublicKeyHex(account)
	if err != nil {
		log.Fatalln(err)
		return ""
	}
	fmt.Printf("Public key m/44'/60'/0'/0/0 in hex: %s\n", publicKeyHex)

	_ = privateKey

	return account.Address.Hex()

}

//10-1
func RegisterWallet(walletAddress string) error {
	if walletAddress == "" {
		walletAddress = CreateHDWallet()
	}

	redisConn, err := redis.Dial("tcp", common.RedisAddr)
	if err != nil {
		log.Fatalln(err)
		return err
	}

	defer redisConn.Close()

	req := &Wallet.RegisterWalletReq{
		WalletAddress: walletAddress,
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
			CorrelationData: []byte(jwtToken), //jwt令牌
			ResponseTopic:   responseTopic,
			User: map[string]string{
				"deviceId":        localDeviceID, // 设备号
				"businessType":    "10",          // 业务号
				"businessSubType": "1",           //  业务子号
				"taskId":          taskIdStr,
				"code":            "0",
				"errormsg":        "",
			},
		},
	}

	//send req to mqtt
	conn, err := net.Dial("tcp", common.BrokerAddr)
	if err != nil {
		// mc.logger.Error("Client dial error ", zap.String("BrokerServer", mc.Addr), zap.Error(err))
		return errors.New("BrokerServer dial error")
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

			} else {
				log.Println("Wallet register failed")
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
	ticker := time.NewTicker(5 * time.Second) // 5s后退出
	for run == true {
		select {
		case <-ticker.C:
			run = false
			break
		}

	}
	log.Println("RegisterWallet Done.")

	return nil

}
