package nsqBackend

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/gomodule/redigo/redis"

	Wallet "github.com/lianmi/servers/api/proto/wallet"
	LMCommon "github.com/lianmi/servers/internal/common"
	"github.com/lianmi/servers/internal/pkg/models"

	"go.uber.org/zap"
)

/*
10-1 钱包账号注册
1. 钱包账号注册流程，SDK生成助记词，并设置支付密码
2. 用户支付的时候，输入6位支付密码，就代表用私钥签名
3. 调用WalletSDK的接口生成私钥、公钥及地址，然后将发送第0号叶子的地址到服务端，服务端在链上创建用户的私人钱包。
4. 用户通过助记词生成的私钥， 需要加密后保存在本地数据库里，以便随时进行签名

*/
func (nc *NsqClient) HandleRegisterWallet(msg *models.Message) error {
	var err error
	errorCode := 200
	var errorMsg string
	var blockNumber uint64 //区块高度
	var hash string

	redisConn := nc.redisPool.Get()
	defer redisConn.Close()

	username := msg.GetUserName()
	// token := msg.GetJwtToken()
	deviceID := msg.GetDeviceID()

	nc.logger.Info("HandleRegisterWallet start...",
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
			zap.String("walletAddress", req.GetWalletAddress()),
		)

		if req.GetWalletAddress() == "" {

			nc.logger.Warn("钱包地址为空 ", zap.String("WalletAddress", req.GetWalletAddress()))
			errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
			errorMsg = fmt.Sprintf("WalletAddress must not empty")
			goto COMPLETE
		}

		//检测钱包地址是否合法
		if nc.ethService.CheckIsvalidAddress(req.GetWalletAddress()) == false {
			nc.logger.Warn("非法钱包地址", zap.String("WalletAddress", req.GetWalletAddress()))
			errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
			errorMsg = fmt.Sprintf("WalletAddress is not valid")
			goto COMPLETE
		}

		//检测是否已经注册过了，不能重复注册
		if isExists, err := redis.Bool(redisConn.Do("HEXISTS", fmt.Sprintf("userWallet:%s", username), "WalletAddress")); err != nil {
			errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
			errorMsg = fmt.Sprintf("HEXISTS error")
			goto COMPLETE
		} else {
			if isExists {
				nc.logger.Warn("钱包地址已经注册过了，不能重复注册", zap.String("WalletAddress", req.GetWalletAddress()))
				errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
				errorMsg = fmt.Sprintf("Wallet had registered")
				goto COMPLETE
			}
		}

		//给用户钱包发送3000000个gas
		if blockNumber, hash, err = nc.ethService.TransferEthToOtherAccount(req.GetWalletAddress(), LMCommon.GASLIMIT); err != nil {
			errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
			errorMsg = fmt.Sprintf("Wallet register error")
			goto COMPLETE
		}

		//保存到MySQL 表 UserWallet
		ethAmountString := fmt.Sprintf("%d", LMCommon.GASLIMIT)

		if err := nc.SaveUserWallet(username, req.GetWalletAddress(), ethAmountString); err != nil {
			nc.logger.Error("SaveUserWallet ", zap.Error(err))
		}

		//保存到redis
		redisConn.Do("HSET",
			fmt.Sprintf("userWallet:%s", username),
			"WalletAddress",
			req.GetWalletAddress())

		redisConn.Do("HSET",
			fmt.Sprintf("userWallet:%s", username),
			"EthAmount",
			LMCommon.GASLIMIT)

		redisConn.Do("HSET",
			fmt.Sprintf("userWallet:%s", username),
			"LNMCAmount",
			0)

	}

COMPLETE:
	msg.SetCode(int32(errorCode)) //状态码
	if errorCode == 200 {
		//RegisterWalletRsp
		rsp := &Wallet.RegisterWalletRsp{
			BlockNumber: blockNumber,
			Hash:        hash,
			AmountEth:   LMCommon.GASLIMIT,
			AmountLNMC:  0,
			Time:        uint64(time.Now().UnixNano() / 1e6), // 当前时间
		}
		data, _ := proto.Marshal(rsp)
		msg.FillBody(data) //网络包的body，承载真正的业务数据
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

/*
10-2 充值
从支付宝、微信、银行卡、信用卡等转入平台账号，平台账号收到款项后再转相应数量的代币到用户钱包地址,此地址对应钱包注册传上来的地址
*/
//TODO 等注册好支付宝或微信支付后再完善，目前测试阶段，直接充值
func (nc *NsqClient) HandleDeposit(msg *models.Message) error {
	var err error
	errorCode := 200
	var errorMsg string
	var walletAddress string //用户钱包地址
	var amountLNMC int64     //用户当前代币数量
	// var accountLNMCBalance uint64 //用户充值之后的代币数量
	var blockNumber uint64
	var hash string
	var amountAfter uint64 //用户充值之后的代币数量

	redisConn := nc.redisPool.Get()
	defer redisConn.Close()

	username := msg.GetUserName()
	// token := msg.GetJwtToken()
	deviceID := msg.GetDeviceID()

	nc.logger.Info("HandleDeposit start...",
		zap.String("username", username),
		zap.String("DeviceId", deviceID))

	//取出当前设备的os， clientType， logonAt
	curDeviceHashKey := fmt.Sprintf("devices:%s:%s", username, deviceID)
	isMaster, _ := redis.Bool(redisConn.Do("HGET", curDeviceHashKey, "ismaster"))
	curOs, _ := redis.String(redisConn.Do("HGET", curDeviceHashKey, "os"))
	curClientType, _ := redis.Int(redisConn.Do("HGET", curDeviceHashKey, "clientType"))
	curLogonAt, _ := redis.Uint64(redisConn.Do("HGET", curDeviceHashKey, "logonAt"))

	nc.logger.Debug("HandleDeposit",
		zap.Bool("isMaster", isMaster),
		zap.String("username", username),
		zap.String("deviceID", deviceID),
		zap.String("curOs", curOs),
		zap.Int("curClientType", curClientType),
		zap.Uint64("curLogonAt", curLogonAt))

	//打开msg里的负载， 获取请求参数
	body := msg.GetContent()
	//解包body
	var req Wallet.DepositReq
	if err := proto.Unmarshal(body, &req); err != nil {
		nc.logger.Error("Protobuf Unmarshal Error", zap.Error(err))
		errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
		errorMsg = fmt.Sprintf("Protobuf Unmarshal Error: %s", err.Error())
		goto COMPLETE

	} else {
		nc.logger.Debug("Deposit payload",
			zap.String("username", username),
			zap.Int("paymentType", int(req.GetPaymentType())),
			zap.Float64("RechargeAmount", req.GetRechargeAmount()),
		)

		if req.GetRechargeAmount() <= 0 {

			nc.logger.Warn("充值金额错误，必须大于0 ", zap.Float64("RechargeAmount", req.GetRechargeAmount()))
			errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
			errorMsg = fmt.Sprintf("RechargeAmount must gather than 0")
			goto COMPLETE
		}

		//检测钱包是否注册, 如果没注册， 则不能充值
		if isExists, err := redis.Bool(redisConn.Do("HEXISTS", fmt.Sprintf("userWallet:%s", username), "WalletAddress")); err != nil {
			errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
			errorMsg = fmt.Sprintf("HEXISTS error")
			goto COMPLETE
		} else {
			if !isExists {
				nc.logger.Warn("钱包没注册，不能充值", zap.String("username", username))
				errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
				errorMsg = fmt.Sprintf("Wallet had not registered")
				goto COMPLETE
			}
		}

		walletAddress, err = redis.String(redisConn.Do("HGET", fmt.Sprintf("userWallet:%s", username), "WalletAddress"))
		if err != nil {
			errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
			errorMsg = fmt.Sprintf("HGET error")
			goto COMPLETE
		}

		if nc.ethService.CheckIsvalidAddress(walletAddress) == false {
			nc.logger.Warn("非法钱包地址", zap.String("WalletAddress", walletAddress))
			errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
			errorMsg = fmt.Sprintf("WalletAddress is not valid")
			goto COMPLETE
		}

		amountLNMC, err = redis.Int64(redisConn.Do("HGET", fmt.Sprintf("userWallet:%s", username), "LNMCAmount"))
		if err != nil {
			errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
			errorMsg = fmt.Sprintf("HGET error")
			goto COMPLETE
		}
		nc.logger.Info("钱包信息",
			zap.String("username", username),
			zap.String("walletAddress", walletAddress),
			zap.Int64("当前余额 amountLNMC", amountLNMC),
		)

		//TODO 核对是否支付成功，必须与第三方支付对接后才能完善

		//调用eth接口， 给用户钱包充值连米币
		amount := int64(req.GetRechargeAmount() * 100)
		blockNumber, hash, amountAfter, err = nc.ethService.TransferLNMCFromLeaf1ToNormalAddress(walletAddress, amount)
		if err != nil {
			errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
			errorMsg = fmt.Sprintf("Deposit error")
			goto COMPLETE
		}

		//保存充值记录到 MySQL
		lnmcDepositHistory := &models.LnmcDepositHistory{
			Username:         username,
			WalletAddress:    walletAddress,
			AmountLNMCBefore: amountLNMC,
			DepositAmount:    amount, //以分为单位
			PaymentType:      int(req.GetPaymentType()),
			AmountLNMCAfter:  int64(amountAfter),
		}
		if err := nc.SaveDepositHistory(lnmcDepositHistory); err != nil {
			nc.logger.Error("SaveDepositHistory failed", zap.Error(err))
		}

		redisConn.Do("HSET",
			fmt.Sprintf("userWallet:%s", username),
			"LNMCAmount",
			amountAfter)
	}

COMPLETE:
	msg.SetCode(int32(errorCode)) //状态码
	if errorCode == 200 {
		rsp := &Wallet.DepositRsp{
			BlockNumber: blockNumber,
			Hash:        hash,
			AmountLNMC:  amountAfter,
			Time:        uint64(time.Now().UnixNano() / 1e6), // 当前时间
		}
		data, _ := proto.Marshal(rsp)
		msg.FillBody(data) //网络包的body，承载真正的业务数据
	} else {
		msg.SetErrorMsg([]byte(errorMsg)) //错误提示
		msg.FillBody(nil)
	}

	//处理完成，向dispatcher发送
	topic := msg.GetSource() + ".Frontend"
	rawData, _ := json.Marshal(msg)
	if err := nc.Producer.Public(topic, rawData); err == nil {
		nc.logger.Info("Message succeed send to ProduceChannel", zap.String("topic", topic))
	} else {
		nc.logger.Error("Failed to send  message to ProduceChannel", zap.Error(err))
	}
	_ = err
	return nil
}
