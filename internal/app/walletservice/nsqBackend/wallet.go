package nsqBackend

import (
	"encoding/json"
	"fmt"
	"net/http"
	// "time"

	"github.com/golang/protobuf/proto"
	"github.com/gomodule/redigo/redis"

	Wallet "github.com/lianmi/servers/api/proto/wallet"
	"github.com/lianmi/servers/internal/pkg/models"

	"go.uber.org/zap"
)

/*
10-1 钱包账号注册
1. 钱包账号注册，生成助记词，并设置支付密码
2. 用户支付的时候，输入6位支付密码，就代表用私钥签名
3. 调用WalletSDK的接口生成私钥、公钥及地址，然后将发送到服务端，服务端在链上创建用户的私人钱包。
4. 用户通过助记词生成的私钥， 需要加密后保存在本地数据库里，以便随时进行签名

*/
func (nc *NsqClient) HandleRegisterWallet(msg *models.Message) error {
	var err error
	errorCode := 200
	var errorMsg string

	redisConn := nc.redisPool.Get()
	defer redisConn.Close()

	username := msg.GetUserName()
	// token := msg.GetJwtToken()
	deviceID := msg.GetDeviceID()

	nc.logger.Info("HandleRegisterPreKeys start...",
		zap.String("username", username),
		zap.String("DeviceId", deviceID))

	//取出当前设备的os， clientType， logonAt
	curDeviceHashKey := fmt.Sprintf("devices:%s:%s", username, deviceID)
	isMaster, _ := redis.Bool(redisConn.Do("HGET", curDeviceHashKey, "ismaster"))
	curOs, _ := redis.String(redisConn.Do("HGET", curDeviceHashKey, "os"))
	curClientType, _ := redis.Int(redisConn.Do("HGET", curDeviceHashKey, "clientType"))
	curLogonAt, _ := redis.Uint64(redisConn.Do("HGET", curDeviceHashKey, "logonAt"))

	nc.logger.Debug("HandleRegisterWallet",
		zap.Bool("isMaster", isMaster),
		zap.String("username", username),
		zap.String("deviceID", deviceID),
		zap.String("curOs", curOs),
		zap.Int("curClientType", curClientType),
		zap.Uint64("curLogonAt", curLogonAt))

	//打开msg里的负载， 获取请求参数
	body := msg.GetContent()
	//解包body
	var req Wallet.RegisterWalletReq
	if err := proto.Unmarshal(body, &req); err != nil {
		nc.logger.Error("Protobuf Unmarshal Error", zap.Error(err))
		errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
		errorMsg = fmt.Sprintf("Protobuf Unmarshal Error: %s", err.Error())
		goto COMPLETE

	} else {
		nc.logger.Debug("RegisterWallet payload",
			zap.Int("index", int(req.GetIndex())),
			zap.String("walletAddress", req.GetWalletAddress()),
		)

		if req.GetIndex() < 0 {
			errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
			errorMsg = fmt.Sprintf("Index must >= 0")
			goto COMPLETE
		}
		if req.GetWalletAddress() == "" {
			errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
			errorMsg = fmt.Sprintf("WalletAddress must not empty")
			goto COMPLETE
		}

		//检测用户合法性

		//检测钱包地址是否合法

		//保存到MySQL

		//保存到redis

		//给用户钱包发送0.01个代币及1个gas

	}

COMPLETE:
	msg.SetCode(int32(errorCode)) //状态码
	if errorCode == 200 {
		//
		msg.FillBody(nil)
	} else {
		msg.SetErrorMsg([]byte(errorMsg)) //错误提示
		msg.FillBody(nil)
	}

	//处理完成，向dispatcher发送
	topic := msg.GetSource() + ".Frontend"
	rawData, _ := json.Marshal(msg)
	if err := nc.Producer.Public(topic, rawData); err == nil {
		nc.logger.Info(" Message succeed send to ProduceChannel", zap.String("topic", topic))
	} else {
		nc.logger.Error("Failed to send  message to ProduceChannel", zap.Error(err))
	}
	_ = err
	return nil
}
