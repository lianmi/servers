package wallet

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"math/big"

	"github.com/lianmi/servers/lmSdkClient/business"

	"github.com/eclipse/paho.golang/paho" //支持v5.0
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/golang/protobuf/proto"
	"github.com/gomodule/redigo/redis"
	Global "github.com/lianmi/servers/api/proto/global"
	Wallet "github.com/lianmi/servers/api/proto/wallet"
	"github.com/lianmi/servers/internal/pkg/blockchain/hdwallet"
	clientcommon "github.com/lianmi/servers/lmSdkClient/common"
)

const (
// 用户账号id1的助记词
// mnemonic = "cloth have cage erase shrug slot album village surprise fence erode direct"
//id3的助记词
// mnemonic = "someone author recipe spider ready exile occur volume relax song inner inform"
//服务端的id4助记词
// mnemonic = "abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon about"

)

type KeyPair struct {
	PrivateKeyHex string //hex格式的私钥
	AddressHex    string //hex格式的地址
}

//创建 用户 HD钱包
func CreateHDWallet(localUserName string) string {
	var mnemonic string
	redisConn, err := redis.Dial("tcp", clientcommon.RedisAddr)
	if err != nil {
		log.Fatalln(err)
		return ""
	}

	defer redisConn.Close()

	if localUserName == "id1" {
		// 用户账号id1的助记词
		mnemonic = "cloth have cage erase shrug slot album village surprise fence erode direct"
	} else {
		mnemonic, err = hdwallet.NewMnemonic(128)
		if err != nil {
			log.Println(err)
			return ""
		}

	}
	log.Println(mnemonic)

	// 保存到redis
	_, err = redisConn.Do("SET", fmt.Sprintf("Mnemonic:%s", localUserName), mnemonic)
	if err != nil {
		log.Fatalln(err)
		return ""
	}

	// mnemonic := clientcommon.MnemonicServer // "element urban soda endless beach celery scheme wet envelope east glory retire"
	// log.Println("mnemonic:", mnemonic)

	wallet, err := hdwallet.NewFromMnemonic(mnemonic, clientcommon.SeedPassword)
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
func GetKeyPairsFromLeafIndex(mnemonic string, index uint64) *KeyPair {

	wallet, err := hdwallet.NewFromMnemonic(mnemonic, clientcommon.SeedPassword)
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

//10-1 注册钱包
func RegisterWallet() error {
	var walletAddress string

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

	walletAddress = CreateHDWallet(localUserName)

	req := &Wallet.RegisterWalletReq{
		WalletAddress: walletAddress,
	}
	content, _ := proto.Marshal(req)

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
				"businessType":    "10",          // 业务号
				"businessSubType": "1",           // 业务子号
				"taskId":          taskIdStr,
				"code":            "0",
				"errormsg":        "",
			},
		},
	}

	// pb.Properties.User.Add("jwtToken", jwtToken)
	// pb.Properties.User.Add("deviceId", localDeviceID)
	// pb.Properties.User.Add("businessType", "10")
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
	var rsq Wallet.RegisterWalletRsp
	if err := proto.Unmarshal(payload, &rsq); err != nil {
		log.Println("Protobuf Unmarshal Error", err)

	} else {

		log.Println("回包内容---------------------")
		log.Println("blockNumber: ", rsq.BlockNumber)
		log.Println("hash: ", rsq.Hash)
		log.Println("amountEth: ", rsq.AmountEth)
		log.Println("amountLNMC: ", rsq.AmountLNMC)
		log.Println("time: ", rsq.Time)

	}
	log.Println("RegisterWallet is Done.")

	return nil

}

//10-2 充值 -- 此接口已经废弃
func Deposit(rechargeAmount float64) error {
	if rechargeAmount < 0 {
		return errors.New("rechargeAmount must gather than 0")
	}

	redisConn, err := redis.Dial("tcp", clientcommon.RedisAddr)
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
			ResponseTopic: responseTopic,
			User: map[string]string{
				"jwtToken":        jwtToken,      // jwt令牌
				"deviceId":        localDeviceID, // 设备号
				"businessType":    "10",          // 业务号
				"businessSubType": "2",           // 业务子号
				"taskId":          taskIdStr,
				"code":            "0",
				"errormsg":        "",
			},
		},
	}

	// pb.Properties.User.Add("jwtToken", jwtToken)
	// pb.Properties.User.Add("deviceId", localDeviceID)
	// pb.Properties.User.Add("businessType", "10")
	// pb.Properties.User.Add("businessSubType", "2")
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
	var rsq Wallet.DepositRsp
	if err := proto.Unmarshal(payload, &rsq); err != nil {
		log.Println("Protobuf Unmarshal Error", err)

	} else {

		log.Println("回包内容---------------------")
		// log.Println("blockNumber: ", rsq.BlockNumber)
		// log.Println("hash: ", rsq.Hash)
		// log.Println("AmountLNMC: ", rsq.AmountLNMC)
		// log.Println("Time: ", rsq.Time)

	}
	log.Println("Deposit is Done.")

	return nil

}

//传入Rawtx， 进行签名, 构造一个已经签名的hex裸交易
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
		big.NewInt(0),
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
10-3 发起预转账
*/
func PreTransfer(orderID, targetUserName string, amount float64) error {

	if amount < 0 {
		return errors.New("amount must gather than 0")
	} else {
		log.Println("amount:", amount)
	}

	redisConn, err := redis.Dial("tcp", clientcommon.RedisAddr)
	if err != nil {
		log.Fatalln(err)
		return err
	}

	defer redisConn.Close()

	req := &Wallet.PreTransferReq{
		OrderID:        orderID,
		TargetUserName: targetUserName,
		Amount:         amount,
		Content:        "test transfer",
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
	mnemonic, err := redis.String(redisConn.Do("GET", fmt.Sprintf("Mnemonic:%s", localUserName)))
	if mnemonic == "" {
		log.Println(localUserName, " mnemonic is  empty")
		return errors.New("mnemonic is empty error")
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
				"businessType":    "10",          // 业务号
				"businessSubType": "3",           //  业务子号
				"taskId":          taskIdStr,
				"code":            "0",
				"errormsg":        "",
			},
		},
	}

	// pb.Properties.User.Add("jwtToken", jwtToken)
	// pb.Properties.User.Add("deviceId", localDeviceID)
	// pb.Properties.User.Add("businessType", "10")
	// pb.Properties.User.Add("businessSubType", "3")
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
	var rsq Wallet.PreTransferRsp
	if err := proto.Unmarshal(payload, &rsq); err != nil {
		log.Println("Protobuf Unmarshal Error", err)

	} else {
		rsq.RawDescToTarget.ChainID = 1
		log.Println("10-3 发起转账 回包内容 : \n RawDescToTarget---------------------")
		log.Println("")
		log.Println("RawDescToTarget---------------------")
		log.Println("RawDescToTarget.Uuid: ", rsq.Uuid)
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

		//privKeyHex 是用户自己的私钥，约定为第0号叶子的子私钥
		privKeyHex := GetKeyPairsFromLeafIndex(mnemonic, 0).PrivateKeyHex
		log.Println("privKeyHex", privKeyHex)

		signedTxToTarget, err := buildTx(rsq.RawDescToTarget, privKeyHex)
		if err != nil {
			log.Fatalln(err)

		}
		// _ = rawTxHex

		//TODO 调用10-4 确认转账

		go func() {
			err = ConfirmTransfer(rsq.Uuid, signedTxToTarget)
			if err != nil {
				log.Fatalln(err)
			}
		}()

	}

	log.Println("Cmd is Done.")

	return nil

}

//10-4 确认转账
func ConfirmTransfer(uuid string, signedTxToTarget string) error {

	redisConn, err := redis.Dial("tcp", clientcommon.RedisAddr)
	if err != nil {
		log.Fatalln(err)
		return err
	}

	defer redisConn.Close()

	req := &Wallet.ConfirmTransferReq{
		Uuid:             uuid,
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
			ResponseTopic: responseTopic,
			User: map[string]string{
				"jwtToken":        jwtToken,      // jwt令牌
				"deviceId":        localDeviceID, // 设备号
				"businessType":    "10",          // 业务号
				"businessSubType": "4",           //  业务子号
				"taskId":          taskIdStr,
				"code":            "0",
				"errormsg":        "",
			},
		},
	}

	// pb.Properties.User.Add("jwtToken", jwtToken)
	// pb.Properties.User.Add("deviceId", localDeviceID)
	// pb.Properties.User.Add("businessType", "10")
	// pb.Properties.User.Add("businessSubType", "4")
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
	var rsq Wallet.ConfirmTransferRsp
	if err := proto.Unmarshal(payload, &rsq); err != nil {
		log.Println("Protobuf Unmarshal Error", err)

	} else {

		log.Println("10-4 确认转账回包内容 : ---------------------")
		log.Println("blockNumber: ", rsq.BlockNumber)
		log.Println("hash: ", rsq.Hash)
		log.Println("time: ", rsq.Time)

	}

	log.Println("Cmd is Done.")

	return nil

}

//10-5 查询账号余额
func Balance() error {

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
		log.Println("Redis GET jwtToken:{localUserName}", err)
		return err
	}
	if jwtToken == "" {
		return errors.New("jwtToken is empty error")
	}

	log.Println("jwtToken:", jwtToken)

	taskId, _ := redis.Int(redisConn.Do("INCR", fmt.Sprintf("taksID:%s", localUserName)))
	taskIdStr := fmt.Sprintf("%d", taskId)

	pb := &paho.Publish{
		Topic:   topic,
		QoS:     byte(1),
		Payload: nil, //不需要包体
		Properties: &paho.PublishProperties{
			ResponseTopic: responseTopic,
			User: map[string]string{
				"jwtToken":        jwtToken,      // jwt令牌
				"deviceId":        localDeviceID, // 设备号
				"businessType":    "10",          // 业务号
				"businessSubType": "5",           // 业务子号
				"taskId":          taskIdStr,
				"code":            "0",
				"errormsg":        "",
			},
		},
	}

	// pb.Properties.User.Add("jwtToken", jwtToken)
	// pb.Properties.User.Add("deviceId", localDeviceID)
	// pb.Properties.User.Add("businessType", "10")
	// pb.Properties.User.Add("businessSubType", "5")
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
	var rsq Wallet.BalanceRsp
	if err := proto.Unmarshal(payload, &rsq); err != nil {
		log.Println("Protobuf Unmarshal Error", err)

	} else {
		log.Println("10-5 查询账号余额,  回包内容---------------------")
		log.Println("username: ", localUserName)
		log.Println("amountLNMC: ", rsq.AmountLNMC)
		log.Println("amountETH: ", rsq.AmountETH)
		log.Println("time: ", rsq.Time)

	}

	log.Println("Cmd is Done.")

	return nil

}

//10-13 签到
func UserSignIn() error {

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
			ResponseTopic: responseTopic,
			User: map[string]string{
				"jwtToken":        jwtToken,      // jwt令牌
				"deviceId":        localDeviceID, // 设备号
				"businessType":    "10",          // 业务号
				"businessSubType": "13",          // 业务子号
				"taskId":          taskIdStr,
				"code":            "0",
				"errormsg":        "",
			},
		},
	}

	// pb.Properties.User.Add("jwtToken", jwtToken)
	// pb.Properties.User.Add("deviceId", localDeviceID)
	// pb.Properties.User.Add("businessType", "10")
	// pb.Properties.User.Add("businessSubType", "13")
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
	var rsq Wallet.UserSignInRsp
	if err := proto.Unmarshal(payload, &rsq); err != nil {
		log.Println("Protobuf Unmarshal Error", err)

	} else {
		log.Println("10-13 签到,  回包内容---------------------")
		log.Println("username: ", localUserName)
		log.Println("count: ", rsq.Count)
		log.Println("totalSignIn: ", rsq.TotalSignIn)

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

	redisConn, err := redis.Dial("tcp", clientcommon.RedisAddr)
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
	mnemonic, err := redis.String(redisConn.Do("GET", fmt.Sprintf("Mnemonic:%s", localUserName)))
	if mnemonic == "" {
		log.Println("mnemonic is  empty")
		return errors.New("mnemonic is empty error")
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
				"businessType":    "10",          // 业务号
				"businessSubType": "6",           // 业务子号
				"taskId":          taskIdStr,
				"code":            "0",
				"errormsg":        "",
			},
		},
	}

	// pb.Properties.User.Add("jwtToken", jwtToken)
	// pb.Properties.User.Add("deviceId", localDeviceID)
	// pb.Properties.User.Add("businessType", "10")
	// pb.Properties.User.Add("businessSubType", "6")
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
	var rsq Wallet.PreWithDrawRsp
	if err := proto.Unmarshal(payload, &rsq); err != nil {
		log.Println("Protobuf Unmarshal Error", err)

	} else {

		log.Println("10-6 发起提现预审核 回包内容 : \n rawDescToPlatform---------------------")
		log.Println("username: ", localUserName)
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
		privKeyHex := GetKeyPairsFromLeafIndex(mnemonic, 0).PrivateKeyHex

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

	log.Println("Cmd is Done.")

	return nil

}

//10-7 确认提现
func WithDraw(withdrawUUID, signedTxToPlatform string) error {

	redisConn, err := redis.Dial("tcp", clientcommon.RedisAddr)
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
			ResponseTopic: responseTopic,
			User: map[string]string{
				"jwtToken":        jwtToken,      // jwt令牌
				"deviceId":        localDeviceID, // 设备号
				"businessType":    "10",          // 业务号
				"businessSubType": "7",           // 业务子号
				"taskId":          taskIdStr,
				"code":            "0",
				"errormsg":        "",
			},
		},
	}

	// pb.Properties.User.Add("jwtToken", jwtToken)
	// pb.Properties.User.Add("deviceId", localDeviceID)
	// pb.Properties.User.Add("businessType", "10")
	// pb.Properties.User.Add("businessSubType", "7")
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
	var rsq Wallet.WithDrawRsp
	if err := proto.Unmarshal(payload, &rsq); err != nil {
		log.Println("Protobuf Unmarshal Error", err)

	} else {

		log.Println("10-7 确认提现回包内容 : ---------------------")
		log.Println("username: ", localUserName)
		log.Println("blockNumber: ", rsq.BlockNumber)
		log.Println("hash: ", rsq.Hash)
		log.Println("balanceLNMC: ", rsq.BalanceLNMC)
		log.Println("time: ", rsq.Time)

	}

	log.Println("Cmd is Done.")

	return nil

}

//10-10 同步充值历史
func DoSyncDepositHistoryPage(depositRecharge int32, startAt, endAt int64, page, pageSize int32) error {

	redisConn, err := redis.Dial("tcp", clientcommon.RedisAddr)
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
			ResponseTopic: responseTopic,
			User: map[string]string{
				"jwtToken":        jwtToken,      // jwt令牌
				"deviceId":        localDeviceID, // 设备号
				"businessType":    "10",          // 业务号
				"businessSubType": "10",          // 业务子号
				"taskId":          taskIdStr,
				"code":            "0",
				"errormsg":        "",
			},
		},
	}

	// pb.Properties.User.Add("jwtToken", jwtToken)
	// pb.Properties.User.Add("deviceId", localDeviceID)
	// pb.Properties.User.Add("businessType", "10")
	// pb.Properties.User.Add("businessSubType", "10")
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
	var rsq Wallet.SyncDepositHistoryPageRsp
	if err := proto.Unmarshal(payload, &rsq); err != nil {
		log.Println("Protobuf Unmarshal Error", err)

	} else {

		log.Println("10-10 同步充值历史 回包内容 : \n ---------------------")
		log.Println("username: ", localUserName)
		log.Println("total", rsq.Total)
		log.Println(rsq.Deposits)

	}
	log.Println("Cmd is Done.")

	return nil

}

//10-11 同步提现历史
func DoSyncWithdrawHistoryPage(startAt, endAt int64, page, pageSize int32) error {

	redisConn, err := redis.Dial("tcp", clientcommon.RedisAddr)
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
			ResponseTopic: responseTopic,
			User: map[string]string{
				"jwtToken":        jwtToken,      // jwt令牌
				"deviceId":        localDeviceID, // 设备号
				"businessType":    "10",          // 业务号
				"businessSubType": "11",          // 业务子号
				"taskId":          taskIdStr,
				"code":            "0",
				"errormsg":        "",
			},
		},
	}

	// pb.Properties.User.Add("jwtToken", jwtToken)
	// pb.Properties.User.Add("deviceId", localDeviceID)
	// pb.Properties.User.Add("businessType", "10")
	// pb.Properties.User.Add("businessSubType", "11")
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
	var rsq Wallet.SyncWithdrawHistoryPageRsp
	if err := proto.Unmarshal(payload, &rsq); err != nil {
		log.Println("Protobuf Unmarshal Error", err)

	} else {

		log.Println("10-11 同步提现历史 回包内容 : \n ---------------------")
		log.Println("username: ", localUserName)
		log.Println("total", rsq.Total)
		log.Println(rsq.Withdraws)

	}

	log.Println("Cmd is Done.")

	return nil

}

//10-9 同步收款历史
func DoSyncCollectionHistoryPage(fromUsername string, startAt, endAt int64, page, pageSize int32) error {

	redisConn, err := redis.Dial("tcp", clientcommon.RedisAddr)
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
			ResponseTopic: responseTopic,
			User: map[string]string{
				"jwtToken":        jwtToken,      // jwt令牌
				"deviceId":        localDeviceID, // 设备号
				"businessType":    "10",          // 业务号
				"businessSubType": "9",           // 业务子号
				"taskId":          taskIdStr,
				"code":            "0",
				"errormsg":        "",
			},
		},
	}

	// pb.Properties.User.Add("jwtToken", jwtToken)
	// pb.Properties.User.Add("deviceId", localDeviceID)
	// pb.Properties.User.Add("businessType", "10")
	// pb.Properties.User.Add("businessSubType", "9")
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
	var rsq Wallet.SyncCollectionHistoryPageRsp
	if err := proto.Unmarshal(payload, &rsq); err != nil {
		log.Println("Protobuf Unmarshal Error", err)

	} else {

		log.Println("10-9 同步收款历史 回包内容 : \n ---------------------")
		log.Println("username: ", localUserName)
		log.Println("total", rsq.Total)
		log.Println(rsq.Collections)

	}
	log.Println("Cmd is Done.")

	return nil

}

//10-12 同步转账历史
func DoSyncTransferHistoryPage(startAt, endAt int64, page, pageSize int32) error {

	redisConn, err := redis.Dial("tcp", clientcommon.RedisAddr)
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
			ResponseTopic: responseTopic,
			User: map[string]string{
				"jwtToken":        jwtToken,      // jwt令牌
				"deviceId":        localDeviceID, // 设备号
				"businessType":    "10",          // 业务号
				"businessSubType": "12",          // 业务子号
				"taskId":          taskIdStr,
				"code":            "0",
				"errormsg":        "",
			},
		},
	}

	// pb.Properties.User.Add("jwtToken", jwtToken)
	// pb.Properties.User.Add("deviceId", localDeviceID)
	// pb.Properties.User.Add("businessType", "10")
	// pb.Properties.User.Add("businessSubType", "12")
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
	var rsq Wallet.SyncTransferHistoryPageRsp
	if err := proto.Unmarshal(payload, &rsq); err != nil {
		log.Println("Protobuf Unmarshal Error", err)

	} else {

		log.Println("10-12 同步转账历史 回包内容 : \n ---------------------")
		log.Println("username: ", localUserName)
		log.Println("total", rsq.Total)
		log.Println(rsq.Transfers)

	}
	log.Println("Cmd is Done.")

	return nil

}

//10-14查询交易哈希详情
func DoTxHashInfo(txType int32, txHash string) error {

	redisConn, err := redis.Dial("tcp", clientcommon.RedisAddr)
	if err != nil {
		log.Fatalln(err)
		return err
	}

	defer redisConn.Close()

	if txType == 0 {
		txType = 1
	}

	req := &Wallet.TxHashInfoReq{
		TxType: Global.TransactionType(txType),
		TxHash: txHash,
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
				"businessType":    "10",          // 业务号
				"businessSubType": "14",          // 业务子号
				"taskId":          taskIdStr,
				"code":            "0",
				"errormsg":        "",
			},
		},
	}

	// pb.Properties.User.Add("jwtToken", jwtToken)
	// pb.Properties.User.Add("deviceId", localDeviceID)
	// pb.Properties.User.Add("businessType", "10")
	// pb.Properties.User.Add("businessSubType", "14")
	// pb.Properties.User.Add("taskId", taskIdStr)
	// pb.Properties.User.Add("code", "0")
	// pb.Properties.User.Add("errormsg", "")

	/*
		//send req to mqtt
		//利用TLS协议连接broker
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
				jwtToken := m.Properties.User["jwtToken"] // Add by lishijia  for flutter mqtt
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


				} else {
					log.Println("Cmd failed, code: ", code)
				}

			}),
			Conn: conn,
		})

	*/

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
	var rsq Wallet.TxHashInfoRsp
	if err := proto.Unmarshal(payload, &rsq); err != nil {
		log.Println("Protobuf Unmarshal Error", err)

	} else {

		log.Println("10-14 查询交易哈希详情 回包内容 : \n ---------------------")
		log.Println("username: ", localUserName)
		log.Println("blockNumber", rsq.BlockNumber)
		log.Println("gas", rsq.Gas)
		log.Println("nonce", rsq.Nonce)
		log.Println("data", rsq.Data)
		log.Println("to", rsq.To)

	}

	log.Println("Cmd is Done.")

	return nil

}
