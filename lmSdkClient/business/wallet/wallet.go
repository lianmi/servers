package wallet

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"math/big"
	"net"
	"time"

	"github.com/eclipse/paho.golang/paho" //支持v5.0
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/golang/protobuf/proto"
	"github.com/gomodule/redigo/redis"
	Wallet "github.com/lianmi/servers/api/proto/wallet"
	LMCommon "github.com/lianmi/servers/lmSdkClient/common"
	"github.com/miguelmota/go-ethereum-hdwallet"
)

const (
	// mac id2的助记词
	mnemonic = "cloth have cage erase shrug slot album village surprise fence erode direct"
	//id3的助记词
	// mnemonic = "someone author recipe spider ready exile occur volume relax song inner inform"
	//服务端的id4助记词
	// mnemonic = "abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon about"

	//发布合约地址
	ERC20DeployContractAddress = ""
)

type KeyPair struct {
	PrivateKeyHex string //hex格式的私钥
	AddressHex    string //hex格式的地址
}

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

//根据叶子索引号获取到公私钥对
func GetKeyPairsFromLeafIndex(index uint64) *KeyPair {

	wallet, err := hdwallet.NewFromMnemonic(mnemonic)
	if err != nil {
		log.Fatalln(err)
		return nil
	}
	leaf := fmt.Sprintf("m/44'/60'/0'/0/%d", index)
	path := hdwallet.MustParseDerivationPath(leaf)
	account, err := wallet.Derive(path, true)
	if err != nil {
		log.Fatalln(err)
		return nil
	}
	// s.logger.Info(fmt.Sprintf("m/44'/60'/0'/0/%d", index), zap.String("Account address", account.Address.Hex()))

	privateKey, err := wallet.PrivateKey(account)
	if err != nil {
		log.Fatalln(err)
		return nil
	}
	privateKeyHex, err := wallet.PrivateKeyHex(account)
	if err != nil {
		log.Fatalln(err)
		return nil
	}
	// fmt.Printf("Private key m/44'/60'/0'/0/0 in hex: %s\n", privateKeyHex)
	// s.logger.Info(fmt.Sprintf("m/44'/60'/0'/0/%d", index), zap.String("Private key", privateKeyHex))

	publicKeyHex, _ := wallet.PublicKeyHex(account)
	if err != nil {
		log.Fatalln(err)
		return nil
	}

	// s.logger.Info(fmt.Sprintf("m/44'/60'/0'/0/%d", index), zap.String("Public key", publicKeyHex))

	_ = privateKey
	_ = publicKeyHex

	return &KeyPair{
		PrivateKeyHex: privateKeyHex,         //hex格式的私钥
		AddressHex:    account.Address.Hex(), //hex格式的地址
	}

}

//10-1
func RegisterWallet(walletAddress string) error {
	if walletAddress == "" {
		walletAddress = CreateHDWallet()
	}

	redisConn, err := redis.Dial("tcp", LMCommon.RedisAddr)
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
	conn, err := net.Dial("tcp", LMCommon.BrokerAddr)
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
				// 回包
				//解包负载 m.Payload
				var rsq Wallet.RegisterWalletRsp
				if err := proto.Unmarshal(m.Payload, &rsq); err != nil {
					log.Println("Protobuf Unmarshal Error", err)

				} else {

					log.Println("回包内容---------------------")
					log.Println("blockNumber: ", rsq.BlockNumber)
					log.Println("hash: ", rsq.Hash)
					log.Println("amountEth: ", rsq.AmountEth)
					log.Println("amountLNMC: ", rsq.AmountLNMC)
					log.Println("time: ", rsq.Time)

				}

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
	ticker := time.NewTicker(60 * time.Second) // 60s后退出
	for run == true {
		select {
		case <-ticker.C:
			run = false
			break
		}

	}
	log.Println("RegisterWallet is Done.")

	return nil

}

//10-2 充值
func Deposit(rechargeAmount float64) error {
	if rechargeAmount < 0 {
		return errors.New("rechargeAmount must gather than 0")
	}

	redisConn, err := redis.Dial("tcp", LMCommon.RedisAddr)
	if err != nil {
		log.Fatalln(err)
		return err
	}

	defer redisConn.Close()

	req := &Wallet.DepositReq{
		PaymentType:    1,
		RechargeAmount: rechargeAmount,
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
				"businessSubType": "2",           //  业务子号
				"taskId":          taskIdStr,
				"code":            "0",
				"errormsg":        "",
			},
		},
	}

	//send req to mqtt
	conn, err := net.Dial("tcp", LMCommon.BrokerAddr)
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
				log.Println("Deposit succeed")
				// 回包
				//解包负载 m.Payload
				var rsq Wallet.DepositRsp
				if err := proto.Unmarshal(m.Payload, &rsq); err != nil {
					log.Println("Protobuf Unmarshal Error", err)

				} else {

					log.Println("回包内容---------------------")
					log.Println("blockNumber: ", rsq.BlockNumber)
					log.Println("hash: ", rsq.Hash)
					log.Println("AmountLNMC: ", rsq.AmountLNMC)
					log.Println("Time: ", rsq.Time)

				}

			} else {
				log.Println("Deposit failed")
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
	ticker := time.NewTicker(60 * time.Second) // 60s后退出
	for run == true {
		select {
		case <-ticker.C:
			run = false
			break
		}

	}
	log.Println("Deposit is Done.")

	return nil

}

//传入Rawtx， 进行签名, 构造一个已经签名的hex裸交易
//
func buildTx(rawDesc *Wallet.RawDesc, privKeyHex string) (string, error) {
	//A私钥
	privateKey, err := crypto.HexToECDSA(privKeyHex)
	if err != nil {
		log.Fatal(err)
	}

	//注意，这里需要填写发币合约地址
	tokenAddress := common.HexToAddress(rawDesc.ContractAddress)

	//构造代币转账的交易裸数据
	tx := types.NewTransaction(
		rawDesc.Nonce,
		tokenAddress,
		big.NewInt(int64(rawDesc.Value)),
		rawDesc.GasLimit,
		big.NewInt(int64(rawDesc.GasPrice)),
		rawDesc.Txdata,
	)

	//对裸交易数据签名
	signedTx, err := types.SignTx(tx, types.NewEIP155Signer(big.NewInt(int64(rawDesc.ChainID))), privateKey)
	if err != nil {
		log.Fatal(err)
		return "", err
	}

	ts := types.Transactions{signedTx}
	rawTxBytes := ts.GetRlp(0)
	log.Println("length: ", len(rawTxBytes))
	signedTxToTarget := hex.EncodeToString(rawTxBytes)

	log.Println("signedTxToTarget:", signedTxToTarget)
	return signedTxToTarget, nil
}

/*
10-3 发起转账
*/
func PreTransfer(orderID, targetUserName string, amount float64) error {

	if amount < 0 {
		return errors.New("amount must gather than 0")
	} else {
		log.Println("amount:", amount)
	}

	redisConn, err := redis.Dial("tcp", LMCommon.RedisAddr)
	if err != nil {
		log.Fatalln(err)
		return err
	}

	defer redisConn.Close()

	req := &Wallet.PreTransferReq{
		OrderID:        orderID,
		TargetUserName: targetUserName,
		Amount:         amount,
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
				"businessSubType": "3",           //  业务子号
				"taskId":          taskIdStr,
				"code":            "0",
				"errormsg":        "",
			},
		},
	}

	//send req to mqtt
	conn, err := net.Dial("tcp", LMCommon.BrokerAddr)
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
				// 回包
				//解包负载 m.Payload
				var rsq Wallet.PreTransferRsp
				if err := proto.Unmarshal(m.Payload, &rsq); err != nil {
					log.Println("Protobuf Unmarshal Error", err)

				} else {
					rsq.RawDescToTarget.ChainID = 1
					log.Println("10-3 发起转账 回包内容 : \n RawDescToTarget---------------------")
					log.Println("")
					log.Println("RawDescToTarget---------------------")
					log.Println("RawDescToTarget.Nonce: ", rsq.RawDescToTarget.Nonce)
					log.Println("RawDescToTarget.GasPrice: ", rsq.RawDescToTarget.GasPrice)
					log.Println("RawDescToTarget.GasLimit: ", rsq.RawDescToTarget.GasLimit)
					log.Println("RawDescToTarget.ChainID: ", rsq.RawDescToTarget.ChainID)
					log.Println("RawDescToTarget.Value: ", rsq.RawDescToTarget.Value)
					log.Println("RawDescToTarget.TxHash: ", rsq.RawDescToTarget.TxHash)
					log.Println("RawDescToTarget.Txdata: ", hex.EncodeToString(rsq.RawDescToTarget.Txdata))
					log.Println("RawDescToTarget.ToWalletAddress: ", rsq.RawDescToTarget.ToWalletAddress)
					log.Println("RawDescToTarget.ContractAddress: ", rsq.RawDescToTarget.ContractAddress)
					log.Println("")
					log.Println("Time: ", rsq.Time)

					log.Println("=======")

					// rawTxToTarget := &models.RawDesc{
					// 	Nonce: rsq.RawDescToTarget.Nonce,
					// 	// gas价格
					// 	GasPrice: rsq.RawDescToTarget.GasPrice,
					// 	// 最低gas
					// 	GasLimit: rsq.RawDescToTarget.GasLimit,
					// 	//链id
					// 	ChainID: rsq.RawDescToTarget.ChainID,
					// 	// 交易数据
					// 	Txdata: rsq.RawDescToTarget.Txdata,
					// 	//ether，设为0
					// 	Value: 0,
					// 	//交易哈希
					// 	TxHash: rsq.RawDescToTarget.TxHash,
					// 	//发币合约地址
					// 	ContractAddress: rsq.RawDescToTarget.ToWalletAddress,
					// }

					// log.Println(rawTxToTarget)

					//privKeyHex 是用户自己的私钥，约定为第0号叶子的子私钥
					privKeyHex := GetKeyPairsFromLeafIndex(0).PrivateKeyHex
					log.Println("privKeyHex", privKeyHex)

					signedTxToTarget, err := buildTx(rsq.RawDescToTarget, privKeyHex)
					if err != nil {
						log.Fatalln(err)

					}
					// _ = rawTxHex

					//TODO 调用10-4 确认转账

					go func() {
						err = ConfirmTransfer(orderID, targetUserName, signedTxToTarget)
						if err != nil {
							log.Fatalln(err)
						}
					}()

				}

			} else {
				log.Println("Cmd failed, code: ", code)
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
	ticker := time.NewTicker(60 * time.Second) // 60s后退出
	for run == true {
		select {
		case <-ticker.C:
			run = false
			break
		}

	}
	log.Println("Cmd is Done.")

	return nil

}

//10-4 确认转账
func ConfirmTransfer(orderID, targetUsername string, signedTxToTarget string) error {

	redisConn, err := redis.Dial("tcp", LMCommon.RedisAddr)
	if err != nil {
		log.Fatalln(err)
		return err
	}

	defer redisConn.Close()

	req := &Wallet.ConfirmTransferReq{
		OrderID:          orderID,
		TargetUserName:   targetUsername,
		SignedTxToTarget: signedTxToTarget,
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
				"businessSubType": "4",           //  业务子号
				"taskId":          taskIdStr,
				"code":            "0",
				"errormsg":        "",
			},
		},
	}

	//send req to mqtt
	conn, err := net.Dial("tcp", LMCommon.BrokerAddr)
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
				// 回包
				//解包负载 m.Payload
				var rsq Wallet.ConfirmTransferRsp
				if err := proto.Unmarshal(m.Payload, &rsq); err != nil {
					log.Println("Protobuf Unmarshal Error", err)

				} else {

					log.Println("10-4 确认转账回包内容 : ---------------------")
					log.Println("blockNumber: ", rsq.BlockNumber)
					log.Println("hash: ", rsq.Hash)
					log.Println("time: ", rsq.Time)

				}

			} else {
				log.Println("Cmd failed, code: ", code)
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
	ticker := time.NewTicker(60 * time.Second) // 60s后退出
	for run == true {
		select {
		case <-ticker.C:
			run = false
			break
		}

	}
	log.Println("Cmd is Done.")

	return nil

}

//10-5 查询账号余额
func Balance() error {

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
		Payload: nil, //不需要包体
		Properties: &paho.PublishProperties{
			CorrelationData: []byte(jwtToken), //jwt令牌
			ResponseTopic:   responseTopic,
			User: map[string]string{
				"deviceId":        localDeviceID, // 设备号
				"businessType":    "10",          // 业务号
				"businessSubType": "5",           // 业务子号
				"taskId":          taskIdStr,
				"code":            "0",
				"errormsg":        "",
			},
		},
	}

	//send req to mqtt
	conn, err := net.Dial("tcp", LMCommon.BrokerAddr)
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
				// 回包
				//解包负载 m.Payload
				var rsq Wallet.BalanceRsp
				if err := proto.Unmarshal(m.Payload, &rsq); err != nil {
					log.Println("Protobuf Unmarshal Error", err)

				} else {
					log.Println("10-5 查询账号余额,  回包内容---------------------")
					log.Println("amountLNMC: ", rsq.AmountLNMC)
					log.Println("amountETH: ", rsq.AmountETH)
					log.Println("time: ", rsq.Time)

				}

			} else {
				log.Println("Failed, server return an error")
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
	ticker := time.NewTicker(60 * time.Second) // 60s后退出
	for run == true {
		select {
		case <-ticker.C:
			run = false
			break
		}

	}
	log.Println("Cmd is Done.")

	return nil

}

//10-13 签到
func UserSignIn() error {

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
		Payload: nil, //不需要包体
		Properties: &paho.PublishProperties{
			CorrelationData: []byte(jwtToken), //jwt令牌
			ResponseTopic:   responseTopic,
			User: map[string]string{
				"deviceId":        localDeviceID, // 设备号
				"businessType":    "10",          // 业务号
				"businessSubType": "13",          // 业务子号
				"taskId":          taskIdStr,
				"code":            "0",
				"errormsg":        "",
			},
		},
	}

	//send req to mqtt
	conn, err := net.Dial("tcp", LMCommon.BrokerAddr)
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
				log.Println("UserSignIn succeed")
				// 回包
				//解包负载 m.Payload
				var rsq Wallet.UserSignInRsp
				if err := proto.Unmarshal(m.Payload, &rsq); err != nil {
					log.Println("Protobuf Unmarshal Error", err)

				} else {
					log.Println("10-13 签到,  回包内容---------------------")
					log.Println("count: ", rsq.Count)
					log.Println("totalSignIn: ", rsq.TotalSignIn)

				}

			} else {
				log.Println("UserSignIn failed")
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
	ticker := time.NewTicker(60 * time.Second) // 60s后退出
	for run == true {
		select {
		case <-ticker.C:
			run = false
			break
		}

	}
	log.Println("UserSignIn is Done.")

	return nil

}

/*
10-6 发起提现预审核
*/
func PreWithDraw(amount float64, smscode, bank, bankCard, cardOwner string) error {

	if amount < 0 {
		return errors.New("amount must gather than 0")
	} else {
		log.Println("amount:", amount)
	}

	if smscode == "" {
		smscode = "123456"
	}

	redisConn, err := redis.Dial("tcp", LMCommon.RedisAddr)
	if err != nil {
		log.Fatalln(err)
		return err
	}

	defer redisConn.Close()

	req := &Wallet.PreWithDrawReq{
		Amount:    amount,
		Smscode:   smscode,
		Bank:      bank,
		BankCard:  bankCard,
		CardOwner: cardOwner,
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
				"businessSubType": "6",           // 业务子号
				"taskId":          taskIdStr,
				"code":            "0",
				"errormsg":        "",
			},
		},
	}

	//send req to mqtt
	conn, err := net.Dial("tcp", LMCommon.BrokerAddr)
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
				// 回包
				//解包负载 m.Payload
				var rsq Wallet.PreWithDrawRsp
				if err := proto.Unmarshal(m.Payload, &rsq); err != nil {
					log.Println("Protobuf Unmarshal Error", err)

				} else {

					log.Println("10-6 发起提现预审核 回包内容 : \n rawDescToPlatform---------------------")
					// log.Println(rsq.Time)
					// log.Println(rsq.RawDescToPlatform)

					// rawTxToPlatform := &models.RawDesc{
					// 	Nonce: rsq.RawDescToPlatform.Nonce,
					// 	// gas价格
					// 	GasPrice: rsq.RawDescToPlatform.GasPrice,
					// 	// 最低gas
					// 	GasLimit: rsq.RawDescToPlatform.GasLimit,
					// 	//链id
					// 	ChainID: rsq.RawDescToPlatform.ChainID,
					// 	// 交易数据
					// 	Txdata: rsq.RawDescToPlatform.Txdata,
					// 	//ether，设为0
					// 	Value: 0,
					// 	//交易哈希
					// 	TxHash: rsq.RawDescToPlatform.TxHash,
					// 	//发币合约地址
					// 	ContractAddress: rsq.RawDescToPlatform.ToWalletAddress,
					// }

					// log.Println(rawTxToPlatform)

					//privKeyHex 是用户自己的私钥，约定为第0号叶子的子私钥
					privKeyHex := GetKeyPairsFromLeafIndex(0).PrivateKeyHex

					rawTxHex, err := buildTx(rsq.RawDescToPlatform, privKeyHex)
					if err != nil {
						log.Fatalln(err)

					}

					//TODO 调用10-7 确认转账

					go func() {
						err = WithDraw(rsq.WithdrawUUID, rawTxHex)
						if err != nil {
							log.Fatalln(err)
						}
					}()

				}

			} else {
				log.Println("Cmd failed, code: ", code)
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
	ticker := time.NewTicker(60 * time.Second) // 60s后退出
	for run == true {
		select {
		case <-ticker.C:
			run = false
			break
		}

	}
	log.Println("Cmd is Done.")

	return nil

}

//10-7 确认提现
func WithDraw(withdrawUUID, signedTxToPlatform string) error {

	redisConn, err := redis.Dial("tcp", LMCommon.RedisAddr)
	if err != nil {
		log.Fatalln(err)
		return err
	}

	defer redisConn.Close()

	req := &Wallet.WithDrawReq{
		WithdrawUUID:       withdrawUUID,
		SignedTxToPlatform: signedTxToPlatform,
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
				"businessSubType": "7",           // 业务子号
				"taskId":          taskIdStr,
				"code":            "0",
				"errormsg":        "",
			},
		},
	}

	//send req to mqtt
	conn, err := net.Dial("tcp", LMCommon.BrokerAddr)
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
				// 回包
				//解包负载 m.Payload
				var rsq Wallet.WithDrawRsp
				if err := proto.Unmarshal(m.Payload, &rsq); err != nil {
					log.Println("Protobuf Unmarshal Error", err)

				} else {

					log.Println("10-7 确认提现回包内容 : ---------------------")
					log.Println("blockNumber: ", rsq.BlockNumber)
					log.Println("hash: ", rsq.Hash)
					log.Println("balanceLNMC: ", rsq.BalanceLNMC)
					log.Println("time: ", rsq.Time)

				}

			} else {
				log.Println("Cmd failed, code: ", code)
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
	ticker := time.NewTicker(60 * time.Second) // 60s后退出
	for run == true {
		select {
		case <-ticker.C:
			run = false
			break
		}

	}
	log.Println("Cmd is Done.")

	return nil

}

//10-10 同步充值历史
func DoSyncDepositHistoryPage(depositRecharge int32, startAt, endAt int64, page, pageSize int32) error {

	redisConn, err := redis.Dial("tcp", LMCommon.RedisAddr)
	if err != nil {
		log.Fatalln(err)
		return err
	}

	defer redisConn.Close()

	if page == 0 {
		page = 1
	}
	if pageSize == 0 {
		pageSize = 100
	}
	req := &Wallet.SyncDepositHistoryPageReq{
		DepositRecharge: 0, //TODO
		StartAt:         uint64(startAt),
		EndAt:           uint64(endAt),
		Page:            page,
		PageSize:        pageSize,
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
				"businessSubType": "10",          // 业务子号
				"taskId":          taskIdStr,
				"code":            "0",
				"errormsg":        "",
			},
		},
	}

	//send req to mqtt
	conn, err := net.Dial("tcp", LMCommon.BrokerAddr)
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
				// 回包
				//解包负载 m.Payload
				var rsq Wallet.SyncDepositHistoryPageRsp
				if err := proto.Unmarshal(m.Payload, &rsq); err != nil {
					log.Println("Protobuf Unmarshal Error", err)

				} else {

					log.Println("10-10 同步充值历史 回包内容 : \n ---------------------")
					log.Println("total", rsq.Total)
					log.Println(rsq.Deposits)

				}

			} else {
				log.Println("Cmd failed, code: ", code)
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
	ticker := time.NewTicker(60 * time.Second) // 60s后退出
	for run == true {
		select {
		case <-ticker.C:
			run = false
			break
		}

	}
	log.Println("Cmd is Done.")

	return nil

}

//10-11 同步提现历史
func DoSyncWithdrawHistoryPage(startAt, endAt int64, page, pageSize int32) error {

	redisConn, err := redis.Dial("tcp", LMCommon.RedisAddr)
	if err != nil {
		log.Fatalln(err)
		return err
	}

	defer redisConn.Close()

	if page == 0 {
		page = 1
	}
	if pageSize == 0 {
		pageSize = 100
	}
	req := &Wallet.SyncWithdrawHistoryPageReq{
		StartAt:  uint64(startAt),
		EndAt:    uint64(endAt),
		Page:     page,
		PageSize: pageSize,
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
				"businessSubType": "11",          // 业务子号
				"taskId":          taskIdStr,
				"code":            "0",
				"errormsg":        "",
			},
		},
	}

	//send req to mqtt
	conn, err := net.Dial("tcp", LMCommon.BrokerAddr)
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
				// 回包
				//解包负载 m.Payload
				var rsq Wallet.SyncWithdrawHistoryPageRsp
				if err := proto.Unmarshal(m.Payload, &rsq); err != nil {
					log.Println("Protobuf Unmarshal Error", err)

				} else {

					log.Println("10-11 同步提现历史 回包内容 : \n ---------------------")
					log.Println("total", rsq.Total)
					log.Println(rsq.Withdraws)

				}

			} else {
				log.Println("Cmd failed, code: ", code)
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
	ticker := time.NewTicker(60 * time.Second) // 60s后退出
	for run == true {
		select {
		case <-ticker.C:
			run = false
			break
		}

	}
	log.Println("Cmd is Done.")

	return nil

}

//10-9 同步收款历史
func DoSyncCollectionHistoryPage(fromUsername string, startAt, endAt int64, page, pageSize int32) error {

	redisConn, err := redis.Dial("tcp", LMCommon.RedisAddr)
	if err != nil {
		log.Fatalln(err)
		return err
	}

	defer redisConn.Close()

	if page == 0 {
		page = 1
	}
	if pageSize == 0 {
		pageSize = 100
	}
	req := &Wallet.SyncCollectionHistoryPageReq{
		FromUsername: fromUsername,
		StartAt:      uint64(startAt),
		EndAt:        uint64(endAt),
		Page:         page,
		PageSize:     pageSize,
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
				"businessSubType": "9",           // 业务子号
				"taskId":          taskIdStr,
				"code":            "0",
				"errormsg":        "",
			},
		},
	}

	//send req to mqtt
	conn, err := net.Dial("tcp", LMCommon.BrokerAddr)
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
				// 回包
				//解包负载 m.Payload
				var rsq Wallet.SyncCollectionHistoryPageRsp
				if err := proto.Unmarshal(m.Payload, &rsq); err != nil {
					log.Println("Protobuf Unmarshal Error", err)

				} else {

					log.Println("10-9 同步收款历史 回包内容 : \n ---------------------")
					log.Println("total", rsq.Total)
					log.Println(rsq.Collections)

				}

			} else {
				log.Println("Cmd failed, code: ", code)
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
	ticker := time.NewTicker(60 * time.Second) // 60s后退出
	for run == true {
		select {
		case <-ticker.C:
			run = false
			break
		}

	}
	log.Println("Cmd is Done.")

	return nil

}

//10-12 同步转账历史
func DoSyncTransferHistoryPage(startAt, endAt int64, page, pageSize int32) error {

	redisConn, err := redis.Dial("tcp", LMCommon.RedisAddr)
	if err != nil {
		log.Fatalln(err)
		return err
	}

	defer redisConn.Close()

	if page == 0 {
		page = 1
	}
	if pageSize == 0 {
		pageSize = 100
	}
	req := &Wallet.SyncTransferHistoryPageReq{
		StartAt:  uint64(startAt),
		EndAt:    uint64(endAt),
		Page:     page,
		PageSize: pageSize,
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
				"businessSubType": "12",          // 业务子号
				"taskId":          taskIdStr,
				"code":            "0",
				"errormsg":        "",
			},
		},
	}

	//send req to mqtt
	conn, err := net.Dial("tcp", LMCommon.BrokerAddr)
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
				// 回包
				//解包负载 m.Payload
				var rsq Wallet.SyncTransferHistoryPageRsp
				if err := proto.Unmarshal(m.Payload, &rsq); err != nil {
					log.Println("Protobuf Unmarshal Error", err)

				} else {

					log.Println("10-12 同步转账历史 回包内容 : \n ---------------------")
					log.Println("total", rsq.Total)
					log.Println(rsq.Transfers)

				}

			} else {
				log.Println("Cmd failed, code: ", code)
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
	ticker := time.NewTicker(60 * time.Second) // 60s后退出
	for run == true {
		select {
		case <-ticker.C:
			run = false
			break
		}

	}
	log.Println("Cmd is Done.")

	return nil

}
