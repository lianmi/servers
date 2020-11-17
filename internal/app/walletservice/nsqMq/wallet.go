package nsqMq

import (
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/gomodule/redigo/redis"

	Global "github.com/lianmi/servers/api/proto/global"
	Order "github.com/lianmi/servers/api/proto/order"
	Wallet "github.com/lianmi/servers/api/proto/wallet"
	LMCommon "github.com/lianmi/servers/internal/common"
	"github.com/lianmi/servers/internal/pkg/models"
	"github.com/lianmi/servers/util/dateutil"
	uuid "github.com/satori/go.uuid"

	"go.uber.org/zap"
)

//检测校验码是否正确
func (nc *NsqClient) CheckSmsCode(mobile, smscode string) bool {
	if mobile == "" || smscode == "" {
		return false
	}
	var err error
	var isExists bool

	redisConn := nc.redisPool.Get()
	defer redisConn.Close()
	key := fmt.Sprintf("smscode:%s", mobile)

	if isExists, err = redis.Bool(redisConn.Do("EXISTS", key)); err != nil {
		nc.logger.Error("redisConn GET smscode Error", zap.Error(err))
		return false
	} else {
		if !isExists {
			nc.logger.Warn("isExists=false, smscode is expire", zap.String("key", key))
			return false
		} else {
			if smscodeInRedis, err := redis.String(redisConn.Do("GET", key)); err != nil {
				nc.logger.Error("redisConn GET smscode Error", zap.Error(err))
				return false
			} else {
				nc.logger.Info("redisConn GET smscode ok ", zap.String("smscodeInRedis", smscodeInRedis))
				return smscodeInRedis == smscode
			}
		}
	}
	return false

}

//
// 10-1 钱包账号注册
// 1. 钱包账号注册流程，SDK生成助记词，并设置支付密码
// 2. 用户支付的时候，输入6位支付密码，就代表用私钥签名
// 3. 调用WalletSDK的接口生成私钥、公钥及地址，然后将发送第0号叶子的地址到服务端，服务端在链上创建用户的私人钱包。
// 4. 用户通过助记词生成的私钥， 需要加密后保存在本地数据库里，以便随时进行签名
// 5. 用来做证明人的系统HD钱包叶子也是与用户一一对应，系统的HD钱包的叶子递增, 也就是说每个用户的多签证明人，对应一个系统HD叶子索引号
// 6. 为实现中转账号能够有足够的gas，对应一个系统HD叶子索引号需要在注册后就转1个eth进去
//
func (nc *NsqClient) HandleRegisterWallet(msg *models.Message) error {
	var err error
	errorCode := 200
	var errorMsg string
	var newBip32Index uint64 //自增的平台HD钱包派生索引号
	var blockNumber uint64
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

		//HD派生出一个叶子索引号，与此用户一一对应
		//平台HD钱包利用bip32派生一个子私钥及子地址，作为证明人 - B签
		newBip32Index, err = redis.Uint64(redisConn.Do("INCR", "Bip32Index"))
		newKeyPair := nc.ethService.GetKeyPairsFromLeafIndex(newBip32Index)

		nc.logger.Info("平台HD钱包利用bip32派生一个子私钥及子地址",
			zap.String("username", username),
			zap.Uint64("newBip32Index", newBip32Index),
			zap.String("PrivateKeyHex", newKeyPair.PrivateKeyHex),
			zap.String("AddressHex", newKeyPair.AddressHex),
		)

		//给叶子发送 1 个ether 以便作为中转账号的时候，可以对商户转账或对买家退款 有足够的gas
		if blockNumber, hash, err = nc.ethService.TransferEthToOtherAccount(newKeyPair.AddressHex, LMCommon.ETHER); err != nil {
			errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
			errorMsg = fmt.Sprintf("Wallet register while TransferEthToOtherAccount error")
			goto COMPLETE
		}

		//给用户钱包发送10000000个gas
		if blockNumber, hash, err = nc.ethService.TransferEthToOtherAccount(req.GetWalletAddress(), 2*LMCommon.GASLIMIT); err != nil {
			errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
			errorMsg = fmt.Sprintf("Wallet register while TransferEthToOtherAccount error")
			goto COMPLETE
		}

		//保存到MySQL 表 UserWallet
		ethAmountString := fmt.Sprintf("%d", LMCommon.GASLIMIT)

		if err := nc.Repository.SaveUserWallet(username, req.GetWalletAddress(), ethAmountString); err != nil {
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
			0) //LMCommon.GASLIMIT

		redisConn.Do("HSET",
			fmt.Sprintf("userWallet:%s", username),
			"LNMCAmount",
			0)

		// 保存叶子Index
		redisConn.Do("HSET",
			fmt.Sprintf("userWallet:%s", username),
			"Bip32Index",
			newBip32Index)
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
		nc.logger.Error("Failed to send message to ProduceChannel", zap.Error(err))
	}
	_ = err
	return nil
}

// 10-2 充值
// 从支付宝、微信、银行卡、信用卡等转入平台账号，平台账号收到款项后再转相应数量的代币到用户钱包地址,此地址对应钱包注册传上来的地址

//TODO 等注册好支付宝或微信支付后再完善，目前测试阶段，直接充值
func (nc *NsqClient) HandleDeposit(msg *models.Message) error {
	var err error
	errorCode := 200
	var errorMsg string
	var walletAddress string //用户钱包地址
	var balanceLNMC uint64   //用户当前代币余额
	var blockNumber uint64
	var hash string
	var balanceAfter uint64 //用户充值之后的代币数量

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

		balanceLNMC, err = nc.ethService.GetLNMCTokenBalance(walletAddress)
		if err != nil {
			errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
			errorMsg = fmt.Sprintf("HGET error")
			goto COMPLETE
		}
		nc.logger.Info("充值之前钱包信息",
			zap.String("username", username),
			zap.String("walletAddress", walletAddress),
			zap.Uint64("当前余额 balanceLNMC", balanceLNMC),
		)

		//TODO 核对是否支付成功，必须与第三方支付对接后才能完善

		//调用eth接口， 给用户钱包充值连米币
		amount := int64(req.GetRechargeAmount() * 100)
		blockNumber, hash, balanceAfter, err = nc.ethService.TransferLNMCFromLeaf1ToNormalAddress(walletAddress, amount)
		if err != nil {
			errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
			errorMsg = fmt.Sprintf("Deposit error")
			goto COMPLETE
		}

		//保存充值记录到 MySQL
		lnmcDepositHistory := &models.LnmcDepositHistory{
			Username:          username,
			WalletAddress:     walletAddress,
			BalanceLNMCBefore: int64(balanceLNMC),
			RechargeAmount:    req.GetRechargeAmount(), //充值金额，单位是人民币
			PaymentType:       int(req.GetPaymentType()),

			BalanceLNMCAfter: int64(balanceAfter),
			BlockNumber:      blockNumber,
			TxHash:           hash,
		}

		nc.Repository.SaveDepositHistory(lnmcDepositHistory)

		//更新redis里用户钱包的代币余额
		redisConn.Do("HSET",
			fmt.Sprintf("userWallet:%s", username),
			"LNMCAmount",
			balanceAfter)

		nc.logger.Info("充值之后钱包信息",
			zap.String("username", username),
			zap.String("walletAddress", walletAddress),
			zap.Uint64("当前余额", balanceAfter),
			zap.Uint64("增加的数量", balanceAfter-uint64(balanceLNMC)),
		)
	}

COMPLETE:
	msg.SetCode(int32(errorCode)) //状态码
	if errorCode == 200 {
		rsp := &Wallet.DepositRsp{
			BlockNumber: blockNumber,
			Hash:        hash,
			AmountLNMC:  balanceAfter,
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
		nc.logger.Error("Failed to send message to ProduceChannel", zap.Error(err))
	}
	_ = err
	return nil
}

// 10-3 发起转账
// 1. 用户下单需要支付（订单ID非空的时候）或者仅仅是用户之间转账时，向服务端发起一个转账申请, 接收者也必须开通钱包 。
// 2. 服务端收到请求后，判断发起方的余额是否足够支付
// 3. 服务端构造Tx裸交易数据，当订单ID非空的时候， 目标接收者是用户

func (nc *NsqClient) HandlePreTransfer(msg *models.Message) error {
	var err error
	errorCode := 200
	var errorMsg string
	var data []byte

	var walletAddress string   //用户钱包地址
	var toWalletAddress string //接收者钱包地址, 订单id及普通转账需要不同的钱包地址

	var balanceLNMC uint64 //用户当前代币数量
	var blockNumber uint64
	var hash string
	var newBip32Index uint64 //自增的平台HD钱包派生索引号

	var amountLNMC uint64 //本次转账的代币数量,  等于amount * 100
	var balanceETH uint64 //当前用户的Eth余额

	var toUsername string

	redisConn := nc.redisPool.Get()
	defer redisConn.Close()

	username := msg.GetUserName()
	// token := msg.GetJwtToken()
	deviceID := msg.GetDeviceID()

	nc.logger.Info("HandlePreTransfer start...",
		zap.String("username", username),
		zap.String("DeviceId", deviceID))

	//取出当前设备的os， clientType， logonAt
	curDeviceHashKey := fmt.Sprintf("devices:%s:%s", username, deviceID)
	isMaster, _ := redis.Bool(redisConn.Do("HGET", curDeviceHashKey, "ismaster"))
	curOs, _ := redis.String(redisConn.Do("HGET", curDeviceHashKey, "os"))
	curClientType, _ := redis.Int(redisConn.Do("HGET", curDeviceHashKey, "clientType"))
	curLogonAt, _ := redis.Uint64(redisConn.Do("HGET", curDeviceHashKey, "logonAt"))

	nc.logger.Debug("HandlePreTransfer",
		zap.Bool("isMaster", isMaster),
		zap.String("username", username),
		zap.String("deviceID", deviceID),
		zap.String("curOs", curOs),
		zap.Int("curClientType", curClientType),
		zap.Uint64("curLogonAt", curLogonAt))

	//打开msg里的负载， 获取请求参数
	body := msg.GetContent()
	//解包body
	var req Wallet.PreTransferReq
	if err := proto.Unmarshal(body, &req); err != nil {
		nc.logger.Error("Protobuf Unmarshal Error", zap.Error(err))
		errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
		errorMsg = fmt.Sprintf("Protobuf Unmarshal Error: %s", err.Error())
		goto COMPLETE

	} else {
		nc.logger.Debug("PreTransferReq payload",
			zap.String("username", username),
			zap.String("orderID", req.GetOrderID()),
			zap.String("targetUserName", req.GetTargetUserName()),
			zap.Float64("amount", req.GetAmount()),  //人民币格式 ，有小数点
			zap.String("content", req.GetContent()), //附言
		)

		if req.GetOrderID() != "" && req.GetTargetUserName() != "" {
			nc.logger.Warn("订单ID与收款方的用户账号只能两者选一 ", zap.String("orderID", req.GetOrderID()), zap.String("targetUserName", req.GetTargetUserName()))
			errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
			errorMsg = fmt.Sprintf("orderID and targetUserName, 2 choice  1")
			goto COMPLETE
		}
		if req.GetOrderID() == "" && req.GetTargetUserName() == "" {
			nc.logger.Warn("订单ID与收款方的用户账号不能都是空 ", zap.String("orderID", req.GetOrderID()), zap.String("targetUserName", req.GetTargetUserName()))
			errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
			errorMsg = fmt.Sprintf("orderID and targetUserName cannot both empty")
			goto COMPLETE
		}

		if req.GetAmount() <= 0 {

			nc.logger.Warn("金额错误，必须大于0 ", zap.Float64("amount", req.GetAmount()))
			errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
			errorMsg = fmt.Sprintf("amount must gather than 0")
			goto COMPLETE
		}

		//从redis里获取用户对应的叶子编号，作为证明人 - B签
		newBip32Index, err = redis.Uint64(redisConn.Do("HGET", fmt.Sprintf("userWallet:%s", username), "Bip32Index"))
		newKeyPair := nc.ethService.GetKeyPairsFromLeafIndex(newBip32Index)

		if req.GetTargetUserName() != "" {
			toUsername = req.GetTargetUserName()
			//检测钱包是否注册, 如果没注册， 则不能转账
			if isExists, err := redis.Bool(redisConn.Do("HEXISTS", fmt.Sprintf("userWallet:%s", username), "WalletAddress")); err != nil {
				errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
				errorMsg = fmt.Sprintf("HEXISTS error")
				goto COMPLETE
			} else {
				if !isExists {
					nc.logger.Warn("钱包没注册，不能转账", zap.String("username", username))
					errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
					errorMsg = fmt.Sprintf("Wallet had not registered")
					goto COMPLETE
				}
			}

			//检测接收者钱包是否注册, 如果没注册， 则不能转账
			if isExists, err := redis.Bool(redisConn.Do("HEXISTS", fmt.Sprintf("userWallet:%s", toUsername), "WalletAddress")); err != nil {
				errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
				errorMsg = fmt.Sprintf("HEXISTS error")
				goto COMPLETE
			} else {
				if !isExists {
					nc.logger.Warn("钱包没注册，不能转账", zap.String("TargetUserName", toUsername))
					errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
					errorMsg = fmt.Sprintf("Target Wallet had not registered")
					goto COMPLETE
				}
			}

			toWalletAddress, err = redis.String(redisConn.Do("HGET", fmt.Sprintf("userWallet:%s", toUsername), "WalletAddress"))
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

		} else if req.GetOrderID() != "" {

			//根据orderID找到目标用户账号id
			orderIDKey := fmt.Sprintf("Order:%s", req.GetOrderID())
			businessUser, err := redis.String(redisConn.Do("HGET", orderIDKey, "BusinessUser"))
			if err != nil {
				nc.logger.Error("从Redis里取出此 Order 对应的usinessUser Error", zap.String("orderIDKey", orderIDKey), zap.Error(err))
			}

			toUsername = businessUser
			toWalletAddress = newKeyPair.AddressHex //中转账号

			//将redis里的订单信息哈希表状态字段设置为 OS_Paying
			_, err = redisConn.Do("HSET", orderIDKey, "State", int(Global.OrderState_OS_Paying))

		}

		if toUsername == "" {
			nc.logger.Error("严重错误, toUsername为空")
			errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
			errorMsg = fmt.Sprintf("Error:  toUsername is empty")
			goto COMPLETE
		}

		walletAddress, err = redis.String(redisConn.Do("HGET", fmt.Sprintf("userWallet:%s", username), "WalletAddress"))
		if err != nil {
			errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
			errorMsg = fmt.Sprintf("HGET error")
			goto COMPLETE
		}

		nc.logger.Info("用户对应的叶子编号、子私钥及子地址",
			zap.String("username", username),
			zap.Uint64("newBip32Index", newBip32Index),
			zap.String("PrivateKeyHex", newKeyPair.PrivateKeyHex),
			zap.String("AddressHex", newKeyPair.AddressHex),
		)

		//当前用户的代币余额
		balanceLNMC, err = nc.ethService.GetLNMCTokenBalance(walletAddress)
		if err != nil {
			errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
			errorMsg = fmt.Sprintf("HGET error")
			goto COMPLETE
		}

		//由于amout是人民币，以元为单位，因此，需要乘以100
		amountLNMC = uint64(req.GetAmount() * 100)

		//当前用户的Eth余额
		balanceETH, err = nc.ethService.GetWeiBalance(walletAddress)

		nc.logger.Info("当前用户的钱包信息",
			zap.String("username", username),
			zap.String("walletAddress", walletAddress),
			zap.Uint64("当前代币余额 balanceLNMC", balanceLNMC),
			zap.Uint64("当前ETH余额 balanceETH", balanceETH),
			zap.Uint64("转账代币数量  amountLNMC", amountLNMC),
		)
		if balanceETH < LMCommon.GASLIMIT {
			nc.logger.Warn("gas余额不足")
			errorCode = http.StatusPaymentRequired       //错误码， 402
			errorMsg = fmt.Sprintf("Not sufficient gas") //  余额不足
			goto COMPLETE
		}

		//判断是否有足够代币数量
		if balanceLNMC < amountLNMC {
			nc.logger.Warn("余额不足",
				zap.String("username", username),
				zap.String("walletAddress", walletAddress),
				zap.Uint64("当前代币余额 balanceLNMC", balanceLNMC),
				zap.Uint64("当前ETH余额 balanceETH", balanceETH),
				zap.Uint64("转账代币数量  amountLNMC", amountLNMC),
				zap.Uint64("缺失数量", amountLNMC-balanceLNMC),
			)
			errorCode = http.StatusBadRequest              //错误码， 400
			errorMsg = fmt.Sprintf("Not sufficient funds") //  余额不足
			goto COMPLETE
		}

		//保存预审核转账记录到 MySQL
		lnmcTransferHistory := &models.LnmcTransferHistory{
			Username:          username,         //发起支付
			ToUsername:        toUsername,       //如果是普通转账，toUsername非空
			OrderID:           req.GetOrderID(), //如果是订单支付 ，非空
			WalletAddress:     walletAddress,    //发起方钱包账户
			ToWalletAddress:   toWalletAddress,  //接收者钱包账户
			BalanceLNMCBefore: balanceLNMC,      //发送方用户在转账时刻的连米币数量
			AmountLNMC:        amountLNMC,       //本次转账的用户连米币数量
			Bip32Index:        newBip32Index,    //平台HD钱包Bip32派生索引号
			State:             0,                //执行状态，0-默认未执行，1-A签，2-全部完成
			Content:           req.GetContent(), //附言
		}
		nc.Repository.SaveLnmcTransferHistory(lnmcTransferHistory)

		//发起者钱包账户向接收者账户转账，由于服务端没有发起者的私钥，所以只能生成裸交易，让发起者签名后才能向接收者账户转账
		tokens := int64(amountLNMC)
		rawDescToTarget, err := nc.ethService.GenerateTransferLNMCTokenTx(walletAddress, toWalletAddress, tokens)
		if err != nil {
			nc.logger.Error("构造发起者向接收者转账的交易 失败", zap.String("walletAddress", walletAddress), zap.String("toWalletAddress", toWalletAddress), zap.Error(err))
			errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
			errorMsg = fmt.Sprintf("Generate TransferLNMCTokenTx error")
			goto COMPLETE
		}

		rsp := &Wallet.PreTransferRsp{
			RawDescToTarget: &Wallet.RawDesc{
				ContractAddress: rawDescToTarget.ContractAddress, //发币智能合约地址
				ToWalletAddress: toWalletAddress,                 //接收者钱包地址
				Nonce:           rawDescToTarget.Nonce,
				GasPrice:        rawDescToTarget.GasPrice,
				GasLimit:        rawDescToTarget.GasLimit,
				ChainID:         rawDescToTarget.ChainID,
				Txdata:          rawDescToTarget.Txdata,
				Value:           amountLNMC, //要转账的代币数量
				TxHash:          rawDescToTarget.TxHash,
			},

			Time: uint64(time.Now().UnixNano() / 1e6), // 当前时间
		}
		data, _ = proto.Marshal(rsp)

		//保存预审核转账记录到 redis
		_, err = redisConn.Do("HMSET",
			fmt.Sprintf("PreTransfer:%s:%s", username, toWalletAddress),
			"Username", username,
			"OrderID", req.GetOrderID(),
			"ToUsername", toUsername,
			"WalletAddress", walletAddress,
			"ToWalletAddress", toWalletAddress,
			"AmountLNMC", amountLNMC,
			"BalanceLNMCBefore", balanceLNMC,
			"Bip32Index", newBip32Index,
			"BlockNumber", blockNumber,
			"Hash", hash,
			"State", 0,
			"Content", req.GetContent(),
			"CreateAt", uint64(time.Now().UnixNano()/1e6),
		)

	}

COMPLETE:
	msg.SetCode(int32(errorCode)) //状态码
	if errorCode == 200 {

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
		nc.logger.Error("Failed to send message to ProduceChannel", zap.Error(err))
	}
	_ = err
	return nil
}

// 10-4 确认转账
// 1. 发起方收到服务端的预审核10-3回包后 ，需要对返回的裸交易哈希进行签名(A签)
// 2. 服务端收到后， 如果是普通转账，则进行B签, 并广播到链上，完成转账， 接收方将收到代币
// 3. 与9-11 的区别是请求参数没有携带 订单id

func (nc *NsqClient) HandleConfirmTransfer(msg *models.Message) error {
	var err error
	errorCode := 200
	var errorMsg string
	var data []byte

	var walletAddress string   //用户钱包地址
	var toWalletAddress string // 接收者钱包地址
	var toUsername string      // 接收者用户账号

	var newBip32Index uint64 //自增的平台HD钱包派生索引号

	var balanceLNMC uint64    //用户当前代币数量
	var toBalanceLNMC uint64  //接收者在转账之前的代币数量
	var amountLNMC uint64     //本次转账的代币数量, 无小数点
	var balanceAfter uint64   //转账之后的代币数量, 无小数点
	var toBalanceAfter uint64 //接收者在AB签名后的代币数量
	var content string        //附言

	redisConn := nc.redisPool.Get()
	defer redisConn.Close()

	username := msg.GetUserName()
	// token := msg.GetJwtToken()
	deviceID := msg.GetDeviceID()

	nc.logger.Info("HandleConfirmTransfer start...",
		zap.String("username", username),
		zap.String("DeviceId", deviceID))

	//取出当前设备的os， clientType， logonAt
	curDeviceHashKey := fmt.Sprintf("devices:%s:%s", username, deviceID)
	isMaster, _ := redis.Bool(redisConn.Do("HGET", curDeviceHashKey, "ismaster"))
	curOs, _ := redis.String(redisConn.Do("HGET", curDeviceHashKey, "os"))
	curClientType, _ := redis.Int(redisConn.Do("HGET", curDeviceHashKey, "clientType"))
	curLogonAt, _ := redis.Uint64(redisConn.Do("HGET", curDeviceHashKey, "logonAt"))

	nc.logger.Debug("HandleConfirmTransfer",
		zap.Bool("isMaster", isMaster),
		zap.String("username", username),
		zap.String("deviceID", deviceID),
		zap.String("curOs", curOs),
		zap.Int("curClientType", curClientType),
		zap.Uint64("curLogonAt", curLogonAt))

	//打开msg里的负载， 获取请求参数
	body := msg.GetContent()
	//解包body
	var req Wallet.ConfirmTransferReq
	if err := proto.Unmarshal(body, &req); err != nil {
		nc.logger.Error("Protobuf Unmarshal Error", zap.Error(err))
		errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
		errorMsg = fmt.Sprintf("Protobuf Unmarshal Error: %s", err.Error())
		goto COMPLETE

	} else {
		nc.logger.Debug("ConfirmTransferReq payload",
			zap.String("username", username),
			zap.String("orderID", req.GetOrderID()),
			zap.String("targetUserName", req.GetTargetUserName()),
			zap.String("SignedTxToTarget", req.GetSignedTxToTarget()), //签名后的Tx(A签) hex
		)

		if req.GetOrderID() != "" && req.GetTargetUserName() != "" {

			nc.logger.Warn("orderID 目标地址不能同时为非空", zap.String("targetUserName", req.GetTargetUserName()))
			errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
			errorMsg = fmt.Sprintf("orderID targetUserName cannot both not empty")
			goto COMPLETE
		}
		if req.GetOrderID() == "" && req.GetTargetUserName() == "" {

			nc.logger.Warn("orderID 目标地址不能同时为空", zap.String("targetUserName", req.GetTargetUserName()))
			errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
			errorMsg = fmt.Sprintf("orderID targetUserName cannot both empty")
			goto COMPLETE
		}

		if len(req.SignedTxToTarget) == 0 {

			nc.logger.Warn("SignedTxToTarget不能为空")
			errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
			errorMsg = fmt.Sprintf("SignedTxToTarget is empty")
			goto COMPLETE
		}

		//检测钱包是否注册, 如果没注册， 则不能转账
		if isExists, err := redis.Bool(redisConn.Do("HEXISTS", fmt.Sprintf("userWallet:%s", username), "WalletAddress")); err != nil {
			errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
			errorMsg = fmt.Sprintf("HEXISTS error")
			goto COMPLETE
		} else {
			if !isExists {
				nc.logger.Warn("钱包没注册，不能转账", zap.String("username", username))
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

		//当前用户的代币余额
		balanceLNMC, err = nc.ethService.GetLNMCTokenBalance(walletAddress)
		if err != nil {
			errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
			errorMsg = fmt.Sprintf("HGET error")
			goto COMPLETE
		}
		nc.logger.Info("当前用户(发送者)的钱包信息",
			zap.String("username", username),
			zap.String("walletAddress", walletAddress),
			zap.Uint64("代币当前余额", balanceLNMC),
		)

		//平台HD钱包利用bip32派生一个子私钥及子地址，作为证明人 - B签
		newBip32Index, err = redis.Uint64(redisConn.Do("HGET", fmt.Sprintf("userWallet:%s", username), "Bip32Index"))
		if err != nil {
			errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
			errorMsg = fmt.Sprintf("HGET Bip32Index error")
			goto COMPLETE
		}
		if newBip32Index == 0 {
			errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
			errorMsg = fmt.Sprintf("Bip32Index is 0")
			goto COMPLETE
		}
		newKeyPair := nc.ethService.GetKeyPairsFromLeafIndex(newBip32Index)

		nc.logger.Info("平台HD钱包利用bip32派生一个子私钥及子地址",
			zap.String("username", username),
			zap.Uint64("newBip32Index", newBip32Index),
			zap.String("PrivateKeyHex", newKeyPair.PrivateKeyHex),
			zap.String("AddressHex", newKeyPair.AddressHex),
		)

		if req.GetTargetUserName() != "" {
			toUsername = req.GetTargetUserName()
			toWalletAddress, err = redis.String(redisConn.Do("HGET", fmt.Sprintf("userWallet:%s", toUsername), "WalletAddress"))
			if nc.ethService.CheckIsvalidAddress(toWalletAddress) == false {
				nc.logger.Warn("非法钱包地址", zap.String("toWalletAddress", toWalletAddress))
				errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
				errorMsg = fmt.Sprintf("toWalletAddress is not valid")
				goto COMPLETE
			}

		} else if req.GetOrderID() != "" {
			toWalletAddress = newKeyPair.AddressHex
			toUsername, err = redis.String(redisConn.Do("HGET", fmt.Sprintf("PreTransfer:%s:%s", username, toWalletAddress), "ToUsername"))
			if err != nil {
				errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
				errorMsg = fmt.Sprintf("HGET Bip32Index error")
				goto COMPLETE
			}
			if toUsername == "" {
				orderIDKey := fmt.Sprintf("Order:%s", req.GetOrderID())
				businessUser, err := redis.String(redisConn.Do("HGET", orderIDKey, "BusinessUser"))
				if err != nil {
					nc.logger.Error("从Redis里取出此 Order 对应的usinessUser Error", zap.String("orderIDKey", orderIDKey), zap.Error(err))
				}

				toUsername = businessUser
			}
		}

		if toUsername == "" {
			nc.logger.Error("严重错误, toUsername为空")
			errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
			errorMsg = fmt.Sprintf("Error:  toUsername is empty")
			goto COMPLETE
		}

		//本次转账的代币数量
		amountLNMC, err = redis.Uint64(redisConn.Do("HGET", fmt.Sprintf("PreTransfer:%s:%s", username, toWalletAddress), "AmountLNMC"))
		if err != nil {
			errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
			errorMsg = fmt.Sprintf("HGET AmountLNMC error")
			goto COMPLETE
		}

		toBalanceLNMC, err = nc.ethService.GetLNMCTokenBalance(toWalletAddress)

		//附言
		content, err = redis.String(redisConn.Do("HGET", fmt.Sprintf("PreTransfer:%s:%s", username, toWalletAddress), "Content"))
		if err != nil {
			errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
			errorMsg = fmt.Sprintf("HGET Bip32Index error")
			goto COMPLETE
		}

		//调用eth接口，将发起方签名的转到目标接收者的交易数据广播到链上- A签
		blockNumber, hash, err := nc.ethService.SendSignedTxToGeth(req.GetSignedTxToTarget())
		if err != nil {
			errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
			errorMsg = fmt.Sprintf("A签 SendSignedTxToGeth error")
			goto COMPLETE
		} else {
			nc.logger.Info("发起方转到目标接收者的交易数据广播到链上  A签成功 ",
				zap.String("username", username),
				zap.String("toUsername", toUsername),
				zap.String("toWalletAddress", toWalletAddress),
				zap.Uint64("blockNumber", blockNumber),
				zap.String("hash", hash),
			)

			// 获取发送者链上代币余额
			balanceAfter, err = nc.ethService.GetLNMCTokenBalance(walletAddress)
			if err != nil {
				errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
				errorMsg = fmt.Sprintf("GetLNMCTokenBalance error")
				goto COMPLETE
			}
			nc.logger.Info("获取发送者链上代币余额",
				zap.String("username", username),
				zap.String("walletAddress", walletAddress),
				zap.Uint64("balanceAfter", balanceAfter),
			)
			//更新Redis里用户钱包的代币数量
			redisConn.Do("HSET",
				fmt.Sprintf("userWallet:%s", username),
				"LNMCAmount",
				balanceAfter)
		}

		//更新转账记录到 MySQL
		lnmcTransferHistory := &models.LnmcTransferHistory{
			Username:          username,         //发起支付
			ToUsername:        toUsername,       //如果是普通转账，toUsername非空
			OrderID:           req.GetOrderID(), //如果是订单支付 ，非空
			WalletAddress:     walletAddress,    // 发起方钱包账户
			BalanceLNMCBefore: balanceLNMC,      //发送方用户在转账时刻的连米币数量
			AmountLNMC:        amountLNMC,       //本次转账的用户连米币数量
			BalanceLNMCAfter:  balanceAfter,     //发送方用户在转账之后的连米币数量
			Bip32Index:        newBip32Index,    //平台HD钱包Bip32派生索引号
			State:             1,                //执行状态，0-默认未执行，1-A签，2-全部完成
			BlockNumber:       blockNumber,
			TxHash:            hash,
		}
		nc.Repository.UpdateLnmcTransferHistory(lnmcTransferHistory)

		//更新转账记录到 redis  HSET
		_, err = redisConn.Do("HSET",
			fmt.Sprintf("PreTransfer:%s:%s", username, toWalletAddress),
			"State", 1,
		)

		_, err = redisConn.Do("HSET",
			fmt.Sprintf("PreTransfer:%s:%s", username, toWalletAddress),
			"SignedTx", req.GetSignedTxToTarget(),
		)
		_, err = redisConn.Do("HSET",
			fmt.Sprintf("PreTransfer:%s:%s", username, toWalletAddress),
			"BlockNumber", blockNumber,
		)
		_, err = redisConn.Do("HSET",
			fmt.Sprintf("PreTransfer:%s:%s", username, toWalletAddress),
			"Hash", hash,
		)

		if req.GetTargetUserName() != "" {
			//更新接收者的收款历史记录
			//刷新接收者redis里的代币数量
			toBalanceAfter, err = nc.ethService.GetLNMCTokenBalance(toWalletAddress)
			if err != nil {
				errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
				errorMsg = fmt.Sprintf("toBalanceAfter, GetLNMCTokenBalance error")
				goto COMPLETE
			}
			redisConn.Do("HSET",
				fmt.Sprintf("userWallet:%s", toUsername),
				"LNMCAmount",
				toBalanceAfter)

			nc.logger.Info("发送者转账之后钱包信息",
				zap.String("username", username),
				zap.String("walletAddress", walletAddress),
				zap.Uint64("发起者之前前余额", balanceLNMC),
				zap.Uint64("发起者现在余额", balanceAfter),
			)
			nc.logger.Info("接收者转账之后钱包信息",
				zap.String("username", username),
				zap.String("walletAddress", walletAddress),
				zap.Uint64("接收者之前余额", toBalanceAfter),
				zap.Uint64("接收者现在余额", toBalanceAfter),
			)

			//代币减少的数量
			exchangeLNMC := balanceLNMC - balanceAfter
			addedLNMC := toBalanceAfter - toBalanceLNMC

			if exchangeLNMC != addedLNMC {
				nc.logger.Error("转账发生严重错误, 发送者代币减少的数量不等于接收者增加的数量")
				errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
				errorMsg = fmt.Sprintf("严重错误")
				goto COMPLETE
			}

			//更新接收者的收款历史表
			lnmcCollectionHistory := &models.LnmcCollectionHistory{
				FromUsername:      username,        //发送者
				FromWalletAddress: walletAddress,   //发送者钱包地址
				ToUsername:        toUsername,      //接收者
				ToWalletAddress:   toWalletAddress, //接收者钱包地址
				BalanceLNMCBefore: toBalanceLNMC,   //接收方用户在转账时刻的连米币数量
				AmountLNMC:        amountLNMC,      //本次转账的用户连米币数量
				BalanceLNMCAfter:  toBalanceAfter,  //接收方用户在转账之后的连米币数量
				Bip32Index:        newBip32Index,   //平台HD钱包Bip32派生索引号
				BlockNumber:       blockNumber,
				TxHash:            hash,
			}
			nc.Repository.SaveCollectionHistory(lnmcCollectionHistory)

		}

		// 10-16 连米币到账通知事件
		lnmcReceivedEventRspData, _ := proto.Marshal(&Wallet.LNMCReceivedEventRsp{
			BlockNumber: blockNumber,                         // 区块高度
			Hash:        hash,                                // 交易哈希hex
			AmountLNMC:  amountLNMC,                          //本次转账接收到的连米币数量
			Content:     content,                             //附言
			Time:        uint64(time.Now().UnixNano() / 1e6), //到账时间
		})
		go nc.BroadcastSpecialMsgToAllDevices(lnmcReceivedEventRspData, uint32(Global.BusinessType_Wallet), uint32(Global.WalletSubType_LNMCReceivedEvent), toUsername)

		//9-12，通知双方
		if req.GetOrderID() != "" {

			//将redis里的订单信息哈希表状态字段设置为 OS_IsPayed
			orderIDKey := fmt.Sprintf("Order:%s", req.GetOrderID())
			_, err = redisConn.Do("HSET", orderIDKey, "State", int(Global.OrderState_OS_IsPayed))
			_, err = redisConn.Do("HSET", orderIDKey, "IsPayed", LMCommon.REDISTRUE)
			if err != nil {
				nc.logger.Error("将redis里的订单信息哈希表状态字段设置为 OS_IsPayed发生严重错误", zap.Error(err), zap.String("orderIDKey", orderIDKey))
				errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
				errorMsg = fmt.Sprintf("HSET IsPayed error: %s", orderIDKey)
				goto COMPLETE
			}

			// 9-12 支付订单完成的事件
			orderPayDoneEventRsp := &Order.OrderPayDoneEventRsp{
				OrderID:      req.GetOrderID(),
				FromUsername: username,
				Amount:       amountLNMC,
				BlockNumber:  blockNumber,
				Hash:         hash,
				Time:         uint64(time.Now().UnixNano() / 1e6),
			}
			payData, _ := proto.Marshal(orderPayDoneEventRsp)
			//向接收者推送 9-12 订单支付完成的事件
			nc.logger.Debug("向接收者推送 9-12 订单支付完成的事件", zap.String("toUsername", toUsername), zap.String("orderID", req.GetOrderID()))
			go nc.BroadcastSpecialMsgToAllDevices(payData, uint32(Global.BusinessType_Order), uint32(Global.OrderSubType_OrderPayDoneEvent), toUsername)

			//向支付发起者推送 9-12 支付订单完成的事件
			nc.logger.Debug("向支付发起者推送 9-12 订单支付完成的事件", zap.String("username", username))
			go nc.BroadcastSpecialMsgToAllDevices(payData, uint32(Global.BusinessType_Order), uint32(Global.OrderSubType_OrderPayDoneEvent), username)

			//刷新接收者redis里的代币数量
			toBalanceAfter, _ := nc.ethService.GetLNMCTokenBalance(toWalletAddress)
			redisConn.Do("HSET",
				fmt.Sprintf("userWallet:%s", toUsername),
				"LNMCAmount",
				toBalanceAfter)

		}

		rsp := &Wallet.ConfirmTransferRsp{
			BlockNumber: blockNumber,
			Hash:        hash,
			Time:        uint64(time.Now().UnixNano() / 1e6),
		}
		data, _ = proto.Marshal(rsp)

	}

COMPLETE:
	msg.SetCode(int32(errorCode)) //状态码
	if errorCode == 200 {
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
		nc.logger.Error("Failed to send message to ProduceChannel", zap.Error(err))
	}
	_ = err
	return nil
}

// 10-5 查询账号余额
// 查询链上账号余额， 包括连米币及以太币, 将查询到的余额更新到redis

func (nc *NsqClient) HandleBalance(msg *models.Message) error {
	var err error
	errorCode := 200
	var errorMsg string
	var data []byte

	var walletAddress string //用户钱包地址
	var balanceLNMC uint64   //用户当前代币数量
	var balanceETH uint64    //用户当前以太币数量

	redisConn := nc.redisPool.Get()
	defer redisConn.Close()

	username := msg.GetUserName()
	// token := msg.GetJwtToken()
	deviceID := msg.GetDeviceID()

	nc.logger.Info("HandleBalance start...",
		zap.String("username", username),
		zap.String("DeviceId", deviceID))

	//取出当前设备的os， clientType， logonAt
	curDeviceHashKey := fmt.Sprintf("devices:%s:%s", username, deviceID)
	isMaster, _ := redis.Bool(redisConn.Do("HGET", curDeviceHashKey, "ismaster"))
	curOs, _ := redis.String(redisConn.Do("HGET", curDeviceHashKey, "os"))
	curClientType, _ := redis.Int(redisConn.Do("HGET", curDeviceHashKey, "clientType"))
	curLogonAt, _ := redis.Uint64(redisConn.Do("HGET", curDeviceHashKey, "logonAt"))

	nc.logger.Debug("HandleBalance",
		zap.Bool("isMaster", isMaster),
		zap.String("username", username),
		zap.String("deviceID", deviceID),
		zap.String("curOs", curOs),
		zap.Int("curClientType", curClientType),
		zap.Uint64("curLogonAt", curLogonAt))

	//打开msg里的负载， 获取请求参数
	body := msg.GetContent()
	//解包body
	var req Wallet.ConfirmTransferReq
	if err := proto.Unmarshal(body, &req); err != nil {
		nc.logger.Error("Protobuf Unmarshal Error", zap.Error(err))
		errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
		errorMsg = fmt.Sprintf("Protobuf Unmarshal Error: %s", err.Error())
		goto COMPLETE

	} else {

		//检测钱包是否注册, 如果没注册， 则不能转账
		if isExists, err := redis.Bool(redisConn.Do("HEXISTS", fmt.Sprintf("userWallet:%s", username), "WalletAddress")); err != nil {
			errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
			errorMsg = fmt.Sprintf("HEXISTS error")
			goto COMPLETE
		} else {
			if !isExists {
				nc.logger.Warn("钱包没注册，不能转账", zap.String("username", username))
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
		//当前用户的Eth余额
		balanceETH, err = nc.ethService.GetWeiBalance(walletAddress)
		if err != nil {
			errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
			errorMsg = fmt.Sprintf("GetWeiBalance error")
			goto COMPLETE
		}

		//当前用户的代币余额
		balanceLNMC, err = nc.ethService.GetLNMCTokenBalance(walletAddress)
		if err != nil {
			errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
			errorMsg = fmt.Sprintf("GetLNMCTokenBalance error")
			goto COMPLETE
		}
		redisConn.Do("HSET",
			fmt.Sprintf("userWallet:%s", username),
			"EthAmount",
			balanceETH)
		redisConn.Do("HSET",
			fmt.Sprintf("userWallet:%s", username),
			"LNMCAmount",
			balanceLNMC)

		nc.logger.Info("当前用户的钱包信息",
			zap.String("username", username),
			zap.String("walletAddress", walletAddress),
			zap.Uint64("当前Eth余额 balanceETH", balanceETH),
			zap.Uint64("当前代币余额 balanceLNMC", balanceLNMC),
		)

		rsp := &Wallet.BalanceRsp{
			AmountLNMC: balanceLNMC,
			AmountETH:  balanceETH,
			Time:       uint64(time.Now().UnixNano() / 1e6),
		}
		data, _ = proto.Marshal(rsp)

	}

COMPLETE:
	msg.SetCode(int32(errorCode)) //状态码
	if errorCode == 200 {
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
		nc.logger.Error("Failed to send message to ProduceChannel", zap.Error(err))
	}
	_ = err
	return nil
}

// 10-6 发起提现预审核
// 用户先向服务端发起提现预审核，服务端校验用户是否合法及有足够的余额提现后，回包Tx交易裸数据，用户需要进行签名

func (nc *NsqClient) HandlePreWithDraw(msg *models.Message) error {
	var err error
	errorCode := 200
	var errorMsg string
	var data []byte

	var walletAddress string //用户钱包地址

	var balanceLNMC uint64 //用户当前代币数量
	var balanceETH uint64
	var amountLNMC uint64 //本次提现的代币数量,  等于amount * 100
	var feeF float64      //佣金
	var fee uint64        //佣金

	var withdrawUUID string //提现单编号，UUID

	redisConn := nc.redisPool.Get()
	defer redisConn.Close()

	username := msg.GetUserName()
	// token := msg.GetJwtToken()
	deviceID := msg.GetDeviceID()

	nc.logger.Info("HandlePreWithDraw start...",
		zap.String("username", username),
		zap.String("DeviceId", deviceID))

	//取出当前设备的os， clientType， logonAt
	curDeviceHashKey := fmt.Sprintf("devices:%s:%s", username, deviceID)
	isMaster, _ := redis.Bool(redisConn.Do("HGET", curDeviceHashKey, "ismaster"))
	curOs, _ := redis.String(redisConn.Do("HGET", curDeviceHashKey, "os"))
	curClientType, _ := redis.Int(redisConn.Do("HGET", curDeviceHashKey, "clientType"))
	curLogonAt, _ := redis.Uint64(redisConn.Do("HGET", curDeviceHashKey, "logonAt"))

	nc.logger.Debug("HandlePreWithDraw",
		zap.Bool("isMaster", isMaster),
		zap.String("username", username),
		zap.String("deviceID", deviceID),
		zap.String("curOs", curOs),
		zap.Int("curClientType", curClientType),
		zap.Uint64("curLogonAt", curLogonAt))

	//打开msg里的负载， 获取请求参数
	body := msg.GetContent()
	//解包body
	var req Wallet.PreWithDrawReq
	if err := proto.Unmarshal(body, &req); err != nil {
		nc.logger.Error("Protobuf Unmarshal Error", zap.Error(err))
		errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
		errorMsg = fmt.Sprintf("Protobuf Unmarshal Error: %s", err.Error())
		goto COMPLETE

	} else {
		nc.logger.Debug("PreWithDrawReq payload",
			zap.String("username", username),
			zap.Float64("amount", req.GetAmount()),      //人民币格式 ，有小数点
			zap.String("smscode", req.GetSmscode()),     //手机短信验证码
			zap.String("bank", req.GetBank()),           //银行
			zap.String("bankCard", req.GetBankCard()),   //银行卡号
			zap.String("cardOwner", req.GetCardOwner()), //银行开卡人姓名
		)

		//smscode是否正确
		mobile, err := redis.String(redisConn.Do("HGET", fmt.Sprintf("userData:%s", username), "Mobile"))
		if nc.CheckSmsCode(mobile, req.GetSmscode()) == false {

			nc.logger.Error("手机验证码错误", zap.String("smscode", req.GetSmscode()))
			errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
			errorMsg = fmt.Sprintf("sms code error")
			goto COMPLETE
		}

		if req.GetAmount() <= 0 {

			nc.logger.Warn("金额错误，必须大于0 ", zap.Float64("amount", req.GetAmount()))
			errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
			errorMsg = fmt.Sprintf("amount must gather than 0")
			goto COMPLETE
		}

		//检测钱包是否注册, 如果没注册， 则不能转账
		if isExists, err := redis.Bool(redisConn.Do("HEXISTS", fmt.Sprintf("userWallet:%s", username), "WalletAddress")); err != nil {
			errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
			errorMsg = fmt.Sprintf("HEXISTS error")
			goto COMPLETE
		} else {
			if !isExists {
				nc.logger.Warn("钱包没注册，不能转账", zap.String("username", username))
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

		//当前用户的链上代币余额
		balanceLNMC, err = nc.ethService.GetLNMCTokenBalance(walletAddress)
		if err != nil {
			errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
			errorMsg = fmt.Sprintf("HGET error")
			goto COMPLETE
		}

		//当前用户的Eth余额
		balanceETH, err = nc.ethService.GetWeiBalance(walletAddress)

		if balanceETH < LMCommon.GASLIMIT {
			nc.logger.Warn("gas余额不足")
			errorCode = http.StatusPaymentRequired       //错误码， 402
			errorMsg = fmt.Sprintf("Not sufficient gas") //  余额不足
			goto COMPLETE
		}
		nc.logger.Info("当前用户的钱包信息",
			zap.String("username", username),
			zap.String("walletAddress", walletAddress),
			zap.Uint64("当前ETH余额 balanceETH", balanceETH),
			zap.Uint64("当前余额 balanceLNMC", balanceLNMC),
		)

		amountLNMC = uint64(req.GetAmount() * 100)

		feeF = math.Ceil(float64(amountLNMC) * LMCommon.FEERATE)
		fee = uint64(feeF) //向上取整

		//amount是人民币格式（单位是 元），要转为int64
		if balanceLNMC < amountLNMC+fee {
			nc.logger.Warn("余额不足")
			errorCode = http.StatusBadRequest              //错误码， 400
			errorMsg = fmt.Sprintf("Not sufficient funds") //  余额不足
			goto COMPLETE
		} else {
			if balanceLNMC-amountLNMC-fee < LMCommon.BaseAmountLNMC {
				nc.logger.Warn("提现后需要保留至少1000个代币")
				errorCode = http.StatusBadRequest //错误码， 400
				errorMsg = fmt.Sprintf("Not sufficient base amount of LNMC")
				goto COMPLETE
			}
		}

		//约定，转账到第2号叶子，作为平台收款地址，提现用户需要往这地址转账
		withdrawKeyPair := nc.ethService.GetKeyPairsFromLeafIndex(LMCommon.WITHDRAWINDEX)

		nc.logger.Info("对应的叶子编号、子私钥及子地址",
			zap.String("username", username),
			zap.Uint64("Bip32Index", LMCommon.WITHDRAWINDEX),
			// zap.String("PrivateKeyHex", newKeyPair.PrivateKeyHex),
			zap.String("AddressHex", withdrawKeyPair.AddressHex),
		)

		//调用eth接口， 构造用户转账给平台方子地址的裸交易数据
		tokens := int64(amountLNMC + fee) //加上佣金
		rawDesc, err := nc.ethService.GenerateTransferLNMCTokenTx(walletAddress, withdrawKeyPair.AddressHex, tokens)
		if err != nil {
			nc.logger.Error("提现，构造用户转账给平台方子地址的裸交易数据 失败", zap.String("walletAddress", walletAddress), zap.String("Plaform Address", withdrawKeyPair.AddressHex), zap.Error(err))
			errorCode = http.StatusPaymentRequired //402
			errorMsg = fmt.Sprintf("Generate TransferLNMCTokenTx error")
			goto COMPLETE
		} else {
			nc.logger.Debug("提现，构造用户转账给平台方子地址的裸交易数据 成功",
				zap.String("walletAddress", walletAddress),
				zap.String("Plaform Address", withdrawKeyPair.AddressHex),
				zap.Uint64("rawDes.Nonce", rawDesc.Nonce),
				zap.Uint64("rawDes.GasPrice", rawDesc.GasPrice),
				zap.Uint64("rawDes.GasLimit", rawDesc.GasLimit),
				zap.Uint64("rawDes.ChainID", rawDesc.ChainID),
				zap.String("rawDes.TxHash", rawDesc.TxHash),
			)
		}

		// 生成UUID
		withdrawUUID = uuid.NewV4().String()

		//保存预审核提现记录到 MySQL
		lnmcWithdrawHistory := &models.LnmcWithdrawHistory{
			WithdrawUUID:      withdrawUUID,
			Username:          username,           //发起提现
			WalletAddress:     walletAddress,      //发起方钱包账户
			Bank:              req.GetBank(),      //银行名称
			BankCard:          req.GetBankCard(),  //银行卡号
			CardOwner:         req.GetCardOwner(), //银行卡持有人
			BalanceLNMCBefore: balanceLNMC,        //发送方用户在提现时刻的连米币数量
			AmountLNMC:        amountLNMC,         //本次提现的用户连米币数量
			Fee:               fee,                //佣金
			State:             0,                  //执行状态，0-默认未执行，1-A签，2-全部完成
			TxHash:            rawDesc.TxHash,     //哈希
		}

		nc.Repository.SaveLnmcWithdrawHistory(lnmcWithdrawHistory)

		rsp := &Wallet.PreWithDrawRsp{
			WithdrawUUID: withdrawUUID,
			RawDescToPlatform: &Wallet.RawDesc{
				ContractAddress: rawDesc.ContractAddress,    //发币智能合约地址
				ToWalletAddress: withdrawKeyPair.AddressHex, //接收者钱包地址
				Nonce:           rawDesc.Nonce,
				GasPrice:        rawDesc.GasPrice,
				GasLimit:        rawDesc.GasLimit,
				ChainID:         rawDesc.ChainID,
				Txdata:          rawDesc.Txdata,
				Value:           amountLNMC, //要提现的代币数量
				TxHash:          rawDesc.TxHash,
			},
			Time: uint64(time.Now().UnixNano() / 1e6), // 当前时间
		}
		data, _ = proto.Marshal(rsp)

	}

COMPLETE:
	msg.SetCode(int32(errorCode)) //状态码
	if errorCode == 200 {

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
		nc.logger.Error("Failed to send message to ProduceChannel", zap.Error(err))
	}
	_ = err
	return nil
}

// 10-7 确认提现
// 1. 服务端下发Tx交易裸数据，用户需要进行签名，并上报服务端在链上广播
// 2. 当链上打包后，平台方指定的地址收到用户提现的 代币后，则发起银行转账

func (nc *NsqClient) HandleWithDraw(msg *models.Message) error {
	var err error
	errorCode := 200
	var errorMsg string
	var data []byte

	var walletAddress string //用户钱包地址

	var blockNumber uint64
	var hash string

	var balanceLNMC uint64  //用户当前代币数量
	var amountLNMC uint64   //本次提现的代币数量,  等于amount * 100
	var balanceAfter uint64 //提现之后s的代币余额

	var balancePlatform uint64 //系统HD接收账号代币数量
	var withdrawUUID string

	redisConn := nc.redisPool.Get()
	defer redisConn.Close()

	username := msg.GetUserName()
	// token := msg.GetJwtToken()
	deviceID := msg.GetDeviceID()

	nc.logger.Info("HandleWithDraw start...",
		zap.String("username", username),
		zap.String("DeviceId", deviceID))

	//取出当前设备的os， clientType， logonAt
	curDeviceHashKey := fmt.Sprintf("devices:%s:%s", username, deviceID)
	isMaster, _ := redis.Bool(redisConn.Do("HGET", curDeviceHashKey, "ismaster"))
	curOs, _ := redis.String(redisConn.Do("HGET", curDeviceHashKey, "os"))
	curClientType, _ := redis.Int(redisConn.Do("HGET", curDeviceHashKey, "clientType"))
	curLogonAt, _ := redis.Uint64(redisConn.Do("HGET", curDeviceHashKey, "logonAt"))

	nc.logger.Debug("HandleWithDraw",
		zap.Bool("isMaster", isMaster),
		zap.String("username", username),
		zap.String("deviceID", deviceID),
		zap.String("curOs", curOs),
		zap.Int("curClientType", curClientType),
		zap.Uint64("curLogonAt", curLogonAt))

	//打开msg里的负载， 获取请求参数
	body := msg.GetContent()
	//解包body
	var req Wallet.WithDrawReq
	if err := proto.Unmarshal(body, &req); err != nil {
		nc.logger.Error("Protobuf Unmarshal Error", zap.Error(err))
		errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
		errorMsg = fmt.Sprintf("Protobuf Unmarshal Error: %s", err.Error())
		goto COMPLETE

	} else {
		nc.logger.Debug("WithDrawReq payload",
			zap.String("username", username),
			zap.String("withdrawUUID", req.GetWithdrawUUID()),
			zap.String("signedTxToPlatform", req.GetSignedTxToPlatform()), //签名后的转账到平台方的Tx交易数据. hex
		)
		withdrawUUID = req.GetWithdrawUUID()

		//检测钱包是否注册, 如果没注册， 则不能转账
		if isExists, err := redis.Bool(redisConn.Do("HEXISTS", fmt.Sprintf("userWallet:%s", username), "WalletAddress")); err != nil {
			errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
			errorMsg = fmt.Sprintf("HEXISTS error")
			goto COMPLETE
		} else {
			if !isExists {
				nc.logger.Warn("钱包没注册，不能转账", zap.String("username", username))
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

		//当前用户的代币余额
		balanceLNMC, err = nc.ethService.GetLNMCTokenBalance(walletAddress)
		if err != nil {
			errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
			errorMsg = fmt.Sprintf("HGET error")
			goto COMPLETE
		}
		nc.logger.Info("当前用户的钱包信息",
			zap.String("username", username),
			zap.String("walletAddress", walletAddress),
			zap.Uint64("代币当前余额", balanceLNMC),
		)

		//平台HD钱包第2号叶子
		withdrawKeyPair := nc.ethService.GetKeyPairsFromLeafIndex(LMCommon.WITHDRAWINDEX)

		//获取系统HD钱包第2号叶子的代币
		balancePlatform, err = nc.ethService.GetLNMCTokenBalance(withdrawKeyPair.AddressHex)
		if err != nil {
			errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
			errorMsg = fmt.Sprintf("HGET error")
			goto COMPLETE
		}
		nc.logger.Info("系统HD钱包第2号叶子钱包信息",
			zap.Uint64("Bip32Index", LMCommon.WITHDRAWINDEX),
			zap.String("HD walletAddress", withdrawKeyPair.AddressHex),
			zap.Uint64("系统HD钱包第2号叶子的代币当前余额", balancePlatform),
		)

		//调用eth接口，将签名后的交易数据广播到链上
		blockNumber, hash, err = nc.ethService.SendSignedTxToGeth(req.GetSignedTxToPlatform())
		if err != nil {
			errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
			errorMsg = fmt.Sprintf("SendSignedTxToGeth error")
			goto COMPLETE
		}

		// 获取用户链上代币余额
		balanceAfter, _ = nc.ethService.GetLNMCTokenBalance(walletAddress)
		exchangeLNMC := uint64(balanceLNMC) - balanceAfter
		nc.logger.Info("提现之后钱包信息",
			zap.String("username", username),
			zap.String("walletAddress", walletAddress),
			zap.Uint64("提现之前余额 balanceLNMC", balanceLNMC),
			zap.Uint64("当前余额 balanceAfter", balanceAfter),
			zap.Uint64("代币减少的数量 exchangeLNMC", exchangeLNMC),
		)

		//保存提现记录到 MySQL
		lnmcWithdrawHistory := &models.LnmcWithdrawHistory{
			WithdrawUUID:      withdrawUUID,
			Username:          username,      //发起支付
			WalletAddress:     walletAddress, //发起方钱包账户
			BalanceLNMCBefore: balanceLNMC,   //发送方用户在提现时刻的连米币数量
			AmountLNMC:        amountLNMC,    //本次提现的用户连米币数量
			BalanceLNMCAfter:  balanceAfter,  //本次提现之后的用户连米币数量
			State:             1,             //执行状态，0-默认未执行，1-A签，2-全部完成
			BlockNumber:       blockNumber,
			TxHash:            hash,
		}
		nc.Repository.UpdateLnmcWithdrawHistory(lnmcWithdrawHistory)

		//更新redis里用户钱包的代币余额
		redisConn.Do("HSET",
			fmt.Sprintf("userWallet:%s", username),
			"LNMCAmount",
			balanceAfter)

		//获取系统HD钱包第2号叶子的代币
		balancePlatform, err = nc.ethService.GetLNMCTokenBalance(withdrawKeyPair.AddressHex)
		if err != nil {
			errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
			errorMsg = fmt.Sprintf("HGET error")
			goto COMPLETE
		}
		nc.logger.Info("系统HD钱包第2号叶子钱包信息(打包之后)",
			zap.Uint64("Bip32Index", LMCommon.WITHDRAWINDEX),
			zap.String("HD walletAddress", withdrawKeyPair.AddressHex),
			zap.Uint64("系统HD钱包第2号叶子的代币当前余额(打包之后)", balancePlatform),
		)
		rsp := &Wallet.WithDrawRsp{
			BlockNumber: blockNumber,
			Hash:        hash,
			BalanceLNMC: balanceAfter,
			Time:        uint64(time.Now().UnixNano() / 1e6), // 当前时间
		}
		data, _ = proto.Marshal(rsp)

		//TODO, 发起向用户的银行卡转账，这里需要等第三方支付接口开通后再开发, 届时，需要人工审核

	}

COMPLETE:
	msg.SetCode(int32(errorCode)) //状态码
	if errorCode == 200 {

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
		nc.logger.Error("Failed to send message to ProduceChannel", zap.Error(err))
	}
	_ = err
	return nil
}

// 向目标用户账号的所有端推送传入的业务号及子号的消息， 接收端会触发对应事件
// 传参：
// 1. data 字节流
// 2. businessType 业务号
// 3. businessSubType 业务子号

func (nc *NsqClient) BroadcastSpecialMsgToAllDevices(data []byte, businessType, businessSubType uint32, toUser string) error {

	redisConn := nc.redisPool.Get()
	defer redisConn.Close()

	//向toUser所有端发送
	deviceListKey := fmt.Sprintf("devices:%s", toUser)
	deviceIDSliceNew, _ := redis.Strings(redisConn.Do("ZRANGEBYSCORE", deviceListKey, "-inf", "+inf"))
	//查询出当前在线所有主从设备
	for _, eDeviceID := range deviceIDSliceNew {

		targetMsg := &models.Message{}
		curDeviceKey := fmt.Sprintf("DeviceJwtToken:%s", eDeviceID)
		curJwtToken, _ := redis.String(redisConn.Do("GET", curDeviceKey))
		nc.logger.Debug("Redis GET ", zap.String("curDeviceKey", curDeviceKey), zap.String("curJwtToken", curJwtToken))

		targetMsg.UpdateID()
		//构建消息路由, 第一个参数是要处理的业务类型，后端服务器处理完成后，需要用此来拼接topic: {businessTypeName.Frontend}
		targetMsg.BuildRouter("Wallet", "", "Wallet.Frontend")

		targetMsg.SetJwtToken(curJwtToken)
		targetMsg.SetUserName(toUser)
		targetMsg.SetDeviceID(eDeviceID)
		// opkAlertMsg.SetTaskID(uint32(taskId))
		targetMsg.SetBusinessTypeName("Wallet")
		targetMsg.SetBusinessType(businessType)       //业务号
		targetMsg.SetBusinessSubType(businessSubType) //业务子号

		targetMsg.BuildHeader("WalletService", time.Now().UnixNano()/1e6)

		targetMsg.FillBody(data) //网络包的body，承载真正的业务数据

		targetMsg.SetCode(200) //成功的状态码

		//构建数据完成，向dispatcher发送
		topic := "Wallet.Frontend"
		rawData, _ := json.Marshal(targetMsg)
		if err := nc.Producer.Public(topic, rawData); err == nil {
			nc.logger.Info("BroadcastSpecialMsgToAllDevices: message succeed send to ProduceChannel", zap.String("topic", topic))
		} else {
			nc.logger.Error("BroadcastSpecialMsgToAllDevices: Failed to send message to ProduceChannel", zap.Error(err))
		}

		nc.logger.Info("Broadcast SpecialMsg To All Devices Succeed",
			zap.String("Username:", toUser),
			zap.String("DeviceID:", curDeviceKey),
			zap.Int64("Now", time.Now().UnixNano()/1e6))

	}

	return nil
}

// 10-9 同步收款历史
// 此接口 支持分页查询，默认是全量查询

func (nc *NsqClient) HandleSyncCollectionHistoryPage(msg *models.Message) error {
	var err error
	errorCode := 200
	var errorMsg string
	var data []byte
	var maps string
	var page, pageSize int
	var total uint64

	rsp := &Wallet.SyncCollectionHistoryPageRsp{
		Total:       0,
		Collections: make([]*Wallet.Collection, 0),
	}

	redisConn := nc.redisPool.Get()
	defer redisConn.Close()

	username := msg.GetUserName() //用户自己的账号
	// token := msg.GetJwtToken()
	deviceID := msg.GetDeviceID()

	nc.logger.Info("HandleSyncCollectionHistoryPage start...",
		zap.String("username", username),
		zap.String("deviceId", deviceID))

	//取出当前设备的os， clientType， logonAt
	curDeviceHashKey := fmt.Sprintf("devices:%s:%s", username, deviceID)
	isMaster, _ := redis.Bool(redisConn.Do("HGET", curDeviceHashKey, "ismaster"))
	curOs, _ := redis.String(redisConn.Do("HGET", curDeviceHashKey, "os"))
	curClientType, _ := redis.Int(redisConn.Do("HGET", curDeviceHashKey, "clientType"))
	curLogonAt, _ := redis.Uint64(redisConn.Do("HGET", curDeviceHashKey, "logonAt"))

	nc.logger.Debug("HandleSyncCollectionHistoryPage ",
		zap.Bool("isMaster", isMaster),
		zap.String("username", username),
		zap.String("deviceID", deviceID),
		zap.String("curOs", curOs),
		zap.Int("curClientType", curClientType),
		zap.Uint64("curLogonAt", curLogonAt))

	//打开msg里的负载， 获取请求参数
	body := msg.GetContent()

	//解包body
	req := &Wallet.SyncCollectionHistoryPageReq{}
	if err := proto.Unmarshal(body, req); err != nil {
		nc.logger.Error("Protobuf Unmarshal Error", zap.Error(err))
		errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
		errorMsg = fmt.Sprintf("Protobuf Unmarshal Error: %s", err.Error())
		goto COMPLETE

	} else {
		nc.logger.Debug("SyncCollectionHistoryPageReq payload",
			zap.String("FromUsername", req.FromUsername),
			zap.Uint64("StartAt", req.StartAt),
			zap.Uint64("EndAt", req.EndAt),
			zap.Int32("Page", req.Page),
			zap.Int32("PageSize", req.PageSize),
		)
		page = int(req.Page)
		if page == 0 {
			page = 1
		}

		pageSize = int(req.PageSize)
		if pageSize == 0 {
			pageSize = 100
		}

		// GetPages 分页返回数据
		if req.StartAt > 0 && req.EndAt > 0 {
			maps = fmt.Sprintf("created_at >= %d and created_at <= %d", req.StartAt, req.EndAt)
		}

		collections := nc.Repository.GetCollectionHistorys(username, req.FromUsername, page, pageSize, &total, maps)
		nc.logger.Debug("GetCollectionHistorys", zap.Uint64("total", total))

		rsp.Total = int32(total) //总页数
		for _, collection := range collections {
			nc.logger.Debug("for...range: collections",
				zap.Uint64("id", collection.ID),
				zap.String("FromUsername", collection.FromUsername),
				zap.String("ToUsername", collection.ToUsername),
			)
			rsp.Collections = append(rsp.Collections, &Wallet.Collection{
				Id:           collection.ID,                //ID
				CreatedAt:    uint64(collection.CreatedAt), //创建时间
				FromUsername: collection.FromUsername,      //发送方用户账号
				ToUsername:   collection.ToUsername,        //接收方的用户账号
				AmountLNMC:   collection.AmountLNMC,        //本次转账的用户连米币数量
				OrderID:      collection.OrderID,           //如果非空，则此次支付是对订单的支付，如果空，则为普通转账
				BlockNumber:  collection.BlockNumber,       //成功执行合约的所在区块高度
				Hash:         collection.TxHash,            //交易哈希
			})
		}

	}

COMPLETE:
	msg.SetCode(int32(errorCode)) //状态码
	if errorCode == 200 {
		data, _ = proto.Marshal(rsp)
		msg.FillBody(data)
	} else {
		msg.SetErrorMsg([]byte(errorMsg)) //错误提示
		msg.FillBody(nil)
	}

	//处理完成，向dispatcher发送
	topic := msg.GetSource() + ".Frontend"
	rawData, _ := json.Marshal(msg)
	if err := nc.Producer.Public(topic, rawData); err == nil {
		nc.logger.Info("Succeed send message to ProduceChannel", zap.String("topic", topic))
	} else {
		nc.logger.Error("Failed to send  message to ProduceChannel", zap.Error(err))
	}
	_ = err
	return nil

}

// 10-10 同步充值历史
// 此接口 支持分页查询，默认是全量查询

func (nc *NsqClient) HandleSyncDepositHistoryPage(msg *models.Message) error {
	var err error
	errorCode := 200
	var errorMsg string
	var data []byte
	var maps string
	var page, pageSize int
	var total uint64

	rsp := &Wallet.SyncDepositHistoryPageRsp{
		Total:    0,
		Deposits: make([]*Wallet.Deposit, 0),
	}

	redisConn := nc.redisPool.Get()
	defer redisConn.Close()

	username := msg.GetUserName() //用户自己的账号
	// token := msg.GetJwtToken()
	deviceID := msg.GetDeviceID()

	nc.logger.Info("HandleSyncDepositHistoryPage start...",
		zap.String("username", username),
		zap.String("deviceId", deviceID))

	//取出当前设备的os， clientType， logonAt
	curDeviceHashKey := fmt.Sprintf("devices:%s:%s", username, deviceID)
	isMaster, _ := redis.Bool(redisConn.Do("HGET", curDeviceHashKey, "ismaster"))
	curOs, _ := redis.String(redisConn.Do("HGET", curDeviceHashKey, "os"))
	curClientType, _ := redis.Int(redisConn.Do("HGET", curDeviceHashKey, "clientType"))
	curLogonAt, _ := redis.Uint64(redisConn.Do("HGET", curDeviceHashKey, "logonAt"))

	nc.logger.Debug("HandleSyncDepositHistoryPage ",
		zap.Bool("isMaster", isMaster),
		zap.String("username", username),
		zap.String("deviceID", deviceID),
		zap.String("curOs", curOs),
		zap.Int("curClientType", curClientType),
		zap.Uint64("curLogonAt", curLogonAt))

	//打开msg里的负载， 获取请求参数
	body := msg.GetContent()

	//解包body
	req := &Wallet.SyncDepositHistoryPageReq{}
	if err := proto.Unmarshal(body, req); err != nil {
		nc.logger.Error("Protobuf Unmarshal Error", zap.Error(err))
		errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
		errorMsg = fmt.Sprintf("Protobuf Unmarshal Error: %s", err.Error())
		goto COMPLETE

	} else {
		nc.logger.Debug("SyncDepositHistoryPageReq  payload",
			zap.Int("depositRecharge", int(req.DepositRecharge)),
			zap.Uint64("StartAt", req.StartAt),
			zap.Uint64("EndAt", req.EndAt),
			zap.Int32("Page", req.Page),
			zap.Int32("PageSize", req.PageSize),
		)
		page = int(req.Page)
		if page == 0 {
			page = 1
		}

		pageSize = int(req.PageSize)
		if pageSize == 0 {
			pageSize = 100
		}

		//  DR_100         = 1;     //100元
		//  DR_200         = 2;     //200元
		//  DR_300         = 3;     //300元
		//  DR_500         = 4;     //500元
		//  DR_1000        = 5;    //1000元
		//  DR_3000        = 6;    //3000元
		//  DR_10000       = 7;   //10000元

		var depositRecharge float64
		switch req.DepositRecharge {
		case 1:
			depositRecharge = 100.00
		case 2:
			depositRecharge = 200.00
		case 3:
			depositRecharge = 300.00
		case 4:
			depositRecharge = 500.00
		case 5:
			depositRecharge = 1000.00
		case 6:
			depositRecharge = 3000.00
		case 7:
			depositRecharge = 10000.00
		default:
			depositRecharge = 0.0

		}

		if req.StartAt > 0 && req.EndAt > 0 {

			maps = fmt.Sprintf("created_at >= %d and created_at <= %d and recharge_amount >= %f", req.StartAt, req.EndAt, depositRecharge)

		}

		deposits := nc.Repository.GetDepositHistorys(username, page, pageSize, &total, maps)
		nc.logger.Debug("GetDepositHistorys", zap.Uint64("total", total))

		rsp.Total = int32(total) //总页数
		for _, deposit := range deposits {
			nc.logger.Debug("for...range: deposits",
				zap.Uint64("id", deposit.ID),
				zap.String("Username", deposit.Username),
			)
			amountLNMC := uint64(deposit.RechargeAmount * 100)
			rsp.Deposits = append(rsp.Deposits, &Wallet.Deposit{
				Id:          deposit.ID,                //ID
				CreatedAt:   uint64(deposit.CreatedAt), //创建时间
				PaymentType: Global.ThirdPartyPaymentType(deposit.PaymentType),
				Recharge:    deposit.RechargeAmount,      //充值金额，单位是人民币
				AmountLNMC:  amountLNMC,                  //换算为连米币的数量, 无小数点
				BlockNumber: uint64(deposit.BlockNumber), //成功执行合约的所在区块高度
				Hash:        deposit.TxHash,              //交易哈希
			})
		}

	}

COMPLETE:
	msg.SetCode(int32(errorCode)) //状态码
	if errorCode == 200 {
		data, _ = proto.Marshal(rsp)
		msg.FillBody(data)
	} else {
		msg.SetErrorMsg([]byte(errorMsg)) //错误提示
		msg.FillBody(nil)
	}

	//处理完成，向dispatcher发送
	topic := msg.GetSource() + ".Frontend"
	rawData, _ := json.Marshal(msg)
	if err := nc.Producer.Public(topic, rawData); err == nil {
		nc.logger.Info("Succeed send message to ProduceChannel", zap.String("topic", topic))
	} else {
		nc.logger.Error("Failed to send  message to ProduceChannel", zap.Error(err))
	}
	_ = err
	return nil

}

// 10-11 同步提现历史
// 此接口 支持分页查询，默认是全量查询

func (nc *NsqClient) HandleSyncWithdrawHistoryPage(msg *models.Message) error {
	var err error
	errorCode := 200
	var errorMsg string
	var data []byte
	var maps string
	var page, pageSize int
	var total uint64

	rsp := &Wallet.SyncWithdrawHistoryPageRsp{
		Total:     0,
		Withdraws: make([]*Wallet.Withdraw, 0),
	}

	redisConn := nc.redisPool.Get()
	defer redisConn.Close()

	username := msg.GetUserName() //用户自己的账号
	// token := msg.GetJwtToken()
	deviceID := msg.GetDeviceID()

	nc.logger.Info("HandleSyncWithdrawHistoryPage start...",
		zap.String("username", username),
		zap.String("deviceId", deviceID))

	//取出当前设备的os， clientType， logonAt
	curDeviceHashKey := fmt.Sprintf("devices:%s:%s", username, deviceID)
	isMaster, _ := redis.Bool(redisConn.Do("HGET", curDeviceHashKey, "ismaster"))
	curOs, _ := redis.String(redisConn.Do("HGET", curDeviceHashKey, "os"))
	curClientType, _ := redis.Int(redisConn.Do("HGET", curDeviceHashKey, "clientType"))
	curLogonAt, _ := redis.Uint64(redisConn.Do("HGET", curDeviceHashKey, "logonAt"))

	nc.logger.Debug("HandleSyncWithdrawHistoryPage ",
		zap.Bool("isMaster", isMaster),
		zap.String("username", username),
		zap.String("deviceID", deviceID),
		zap.String("curOs", curOs),
		zap.Int("curClientType", curClientType),
		zap.Uint64("curLogonAt", curLogonAt))

	//打开msg里的负载， 获取请求参数
	body := msg.GetContent()

	//解包body
	req := &Wallet.SyncWithdrawHistoryPageReq{}
	if err := proto.Unmarshal(body, req); err != nil {
		nc.logger.Error("Protobuf Unmarshal Error", zap.Error(err))
		errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
		errorMsg = fmt.Sprintf("Protobuf Unmarshal Error: %s", err.Error())
		goto COMPLETE

	} else {
		nc.logger.Debug("SyncWithdrawHistoryPage  payload",
			zap.Uint64("StartAt", req.StartAt),
			zap.Uint64("EndAt", req.EndAt),
			zap.Int32("Page", req.Page),
			zap.Int32("PageSize", req.PageSize),
		)
		page = int(req.Page)
		if page == 0 {
			page = 1
		}

		pageSize = int(req.PageSize)
		if pageSize == 0 {
			pageSize = 100
		}
		if req.StartAt > 0 && req.EndAt > 0 {

			maps = fmt.Sprintf("ucreated_at >= %d and created_at <= %d ", req.StartAt, req.EndAt)

		}

		withdraws := nc.Repository.GetWithdrawHistorys(username, page, pageSize, &total, maps)
		nc.logger.Debug("GetWithdrawHistorys", zap.Uint64("total", total))

		rsp.Total = int32(total) //总页数
		for _, withdraw := range withdraws {
			nc.logger.Debug("for...range: deposits",
				zap.Uint64("id", withdraw.ID),
				zap.String("Username", withdraw.Username),
			)
			rsp.Withdraws = append(rsp.Withdraws, &Wallet.Withdraw{
				Id:          withdraw.ID,                //ID
				CreatedAt:   uint64(withdraw.CreatedAt), //创建时间
				Bank:        withdraw.Bank,
				BankCard:    withdraw.BankCard,
				CardOwner:   withdraw.CardOwner,
				State:       int32(withdraw.State),
				BlockNumber: uint64(withdraw.BlockNumber), //成功执行合约的所在区块高度
				Hash:        withdraw.TxHash,              //交易哈希
			})
		}

	}

COMPLETE:
	msg.SetCode(int32(errorCode)) //状态码
	if errorCode == 200 {
		data, _ = proto.Marshal(rsp)
		msg.FillBody(data)
	} else {
		msg.SetErrorMsg([]byte(errorMsg)) //错误提示
		msg.FillBody(nil)
	}

	//处理完成，向dispatcher发送
	topic := msg.GetSource() + ".Frontend"
	rawData, _ := json.Marshal(msg)
	if err := nc.Producer.Public(topic, rawData); err == nil {
		nc.logger.Info("Succeed send message to ProduceChannel", zap.String("topic", topic))
	} else {
		nc.logger.Error("Failed to send  message to ProduceChannel", zap.Error(err))
	}
	_ = err
	return nil

}

// 10-12 同步转账历史
// 此接口 支持分页查询，默认是全量查询

func (nc *NsqClient) HandleSyncTransferHistoryPage(msg *models.Message) error {
	var err error
	errorCode := 200
	var errorMsg string
	var data []byte
	var maps string
	var page, pageSize int
	var total uint64

	rsp := &Wallet.SyncTransferHistoryPageRsp{
		Total:     0,
		Transfers: make([]*Wallet.Transfer, 0),
	}

	redisConn := nc.redisPool.Get()
	defer redisConn.Close()

	username := msg.GetUserName() //用户自己的账号
	// token := msg.GetJwtToken()
	deviceID := msg.GetDeviceID()

	nc.logger.Info("HandleSyncTransferHistoryPage start...",
		zap.String("username", username),
		zap.String("deviceId", deviceID))

	//取出当前设备的os， clientType， logonAt
	curDeviceHashKey := fmt.Sprintf("devices:%s:%s", username, deviceID)
	isMaster, _ := redis.Bool(redisConn.Do("HGET", curDeviceHashKey, "ismaster"))
	curOs, _ := redis.String(redisConn.Do("HGET", curDeviceHashKey, "os"))
	curClientType, _ := redis.Int(redisConn.Do("HGET", curDeviceHashKey, "clientType"))
	curLogonAt, _ := redis.Uint64(redisConn.Do("HGET", curDeviceHashKey, "logonAt"))

	nc.logger.Debug("HandleSyncTransferHistoryPage",
		zap.Bool("isMaster", isMaster),
		zap.String("username", username),
		zap.String("deviceID", deviceID),
		zap.String("curOs", curOs),
		zap.Int("curClientType", curClientType),
		zap.Uint64("curLogonAt", curLogonAt))

	//打开msg里的负载， 获取请求参数
	body := msg.GetContent()

	//解包body
	req := &Wallet.SyncTransferHistoryPageReq{}
	if err := proto.Unmarshal(body, req); err != nil {
		nc.logger.Error("Protobuf Unmarshal Error", zap.Error(err))
		errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
		errorMsg = fmt.Sprintf("Protobuf Unmarshal Error: %s", err.Error())
		goto COMPLETE

	} else {
		nc.logger.Debug("SyncTransferHistoryPage payload",
			zap.Uint64("StartAt", req.StartAt),
			zap.Uint64("EndAt", req.EndAt),
			zap.Int32("Page", req.Page),
			zap.Int32("PageSize", req.PageSize),
		)
		page = int(req.Page)
		if page == 0 {
			page = 1
		}

		pageSize = int(req.PageSize)
		if pageSize == 0 {
			pageSize = 100
		}
		if req.StartAt > 0 && req.EndAt > 0 {

			maps = fmt.Sprintf("created_at >= %d and created_at <= %d ", req.StartAt, req.EndAt)

		}

		transfers := nc.Repository.GetTransferHistorys(username, page, pageSize, &total, maps)
		nc.logger.Debug("GetTransferHistorys", zap.Uint64("total", total))

		rsp.Total = int32(total) //总页数
		for _, transfer := range transfers {
			nc.logger.Debug("for...range: deposits",
				zap.Uint64("id", transfer.ID),
				zap.String("Username", transfer.Username),
			)
			rsp.Transfers = append(rsp.Transfers, &Wallet.Transfer{
				Id:          transfer.ID,                //ID
				CreatedAt:   uint64(transfer.CreatedAt), //创建时间
				ToUsername:  transfer.ToUsername,        //接收方
				AmountLNMC:  transfer.AmountLNMC,
				State:       int32(transfer.State),
				OrderID:     transfer.OrderID,
				BlockNumber: uint64(transfer.BlockNumber), //成功执行合约的所在区块高度
				Hash:        transfer.TxHash,              //交易哈希
			})
		}

	}

COMPLETE:
	msg.SetCode(int32(errorCode)) //状态码
	if errorCode == 200 {
		data, _ = proto.Marshal(rsp)
		msg.FillBody(data)
	} else {
		msg.SetErrorMsg([]byte(errorMsg)) //错误提示
		msg.FillBody(nil)
	}

	//处理完成，向dispatcher发送
	topic := msg.GetSource() + ".Frontend"
	rawData, _ := json.Marshal(msg)
	if err := nc.Producer.Public(topic, rawData); err == nil {
		nc.logger.Info("Succeed send message to ProduceChannel", zap.String("topic", topic))
	} else {
		nc.logger.Error("Failed to send  message to ProduceChannel", zap.Error(err))
	}
	_ = err
	return nil

}

// 10-13 签到
// 用户每天签到，每成功签到2次，送若干1千万wei的以太币

func (nc *NsqClient) HandleUserSignIn(msg *models.Message) error {
	var err error
	errorCode := 200
	var errorMsg string
	var data, awardData []byte
	var isExists bool
	var count int
	var total uint64
	var latestDate string
	var walletAddress string    //用户钱包地址
	var awardEth uint64         //奖励的eth
	var balanceEthBefore uint64 //奖励之前用户的eth数量 ，wei为单位
	var balanceEth uint64       //奖励之后的eth数量

	rsp := &Wallet.UserSignInRsp{}
	ethReceivedEventRsp := &Wallet.EthReceivedEventRsp{}

	redisConn := nc.redisPool.Get()
	defer redisConn.Close()

	username := msg.GetUserName() //用户自己的账号
	// token := msg.GetJwtToken()
	deviceID := msg.GetDeviceID()

	nc.logger.Info("HandleUserSignIn start...",
		zap.String("username", username),
		zap.String("deviceId", deviceID))

	//取出当前设备的os， clientType， logonAt
	curDeviceHashKey := fmt.Sprintf("devices:%s:%s", username, deviceID)
	isMaster, _ := redis.Bool(redisConn.Do("HGET", curDeviceHashKey, "ismaster"))
	curOs, _ := redis.String(redisConn.Do("HGET", curDeviceHashKey, "os"))
	curClientType, _ := redis.Int(redisConn.Do("HGET", curDeviceHashKey, "clientType"))
	curLogonAt, _ := redis.Uint64(redisConn.Do("HGET", curDeviceHashKey, "logonAt"))

	nc.logger.Debug("HandleUserSignIn",
		zap.Bool("isMaster", isMaster),
		zap.String("username", username),
		zap.String("deviceID", deviceID),
		zap.String("curOs", curOs),
		zap.Int("curClientType", curClientType),
		zap.Uint64("curLogonAt", curLogonAt))

	currDate := dateutil.GetDateString()
	key := fmt.Sprintf("userSignin:%s", username)

	isExists, err = redis.Bool(redisConn.Do("HEXISTS", key, "LatestDate"))
	if err != nil {
		nc.logger.Error("redisConn EXISTS Error", zap.Error(err))
		errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
		errorMsg = fmt.Sprintf("redis Error: %s", err.Error())
		goto COMPLETE

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

	//不存在则表示首次签到
	if !isExists {
		_, err = redisConn.Do("HMSET",
			key,
			"LatestDate", currDate,
			"Count", 1, //如果达到2次则奖励
			"Total", 1,
		)
		rsp.Count = 1
		rsp.TotalSignIn = 1

	} else {
		latestDate, _ = redis.String(redisConn.Do("HGET", key, "LatestDate"))
		if currDate == latestDate {
			nc.logger.Warn("每天只能签到一次")
			errorCode = http.StatusGone //错误码 410
			errorMsg = fmt.Sprintf("Today already signineed: %s", latestDate)
			goto COMPLETE
		}

		_, err = redisConn.Do("HSET", key, "LatestDate", currDate)
		count, _ = redis.Int(redisConn.Do("HINCRBY", key, "Count", 1))
		total, _ = redis.Uint64(redisConn.Do("HINCRBY", key, "Total", 1))
		rsp.Count = int32(count)
		rsp.TotalSignIn = total

		//如果count累计大于或等于2次，则奖励并重置为0
		if count == 2 {
			redisConn.Do("HSET", key, "Count", 0)
			go func() {
				balanceEthBefore, err = nc.ethService.GetWeiBalance(walletAddress)
				awardEth = uint64(LMCommon.AWARDGAS)
				nc.ethService.TransferEthToOtherAccount(walletAddress, int64(awardEth))
				balanceEth, err = nc.ethService.GetWeiBalance(walletAddress)
				nc.logger.Info("奖励ETH",
					zap.Uint64("奖励之前用户的eth数量", balanceEthBefore),
					zap.Uint64("奖励之后用户的eth数量", balanceEth),
				)

				//向用户推送10-15 ETH奖励到账通知事件
				ethReceivedEventRsp = &Wallet.EthReceivedEventRsp{
					AmountETH: awardEth,
					Time:      uint64(time.Now().UnixNano() / 1e6),
				}
				awardData, _ = proto.Marshal(ethReceivedEventRsp)

				nc.BroadcastSpecialMsgToAllDevices(awardData, uint32(Global.BusinessType_Wallet), uint32(Global.WalletSubType_EthReceivedEvent), username)

			}()
		}

	}

COMPLETE:
	msg.SetCode(int32(errorCode)) //状态码
	if errorCode == 200 {
		nc.logger.Info("UserSignInRsp回包",
			zap.Int("Count", count),
			zap.Uint64("Total", total),
		)
		data, _ = proto.Marshal(rsp)
		msg.FillBody(data)

	} else {
		msg.SetErrorMsg([]byte(errorMsg)) //错误提示
		msg.FillBody(nil)
	}

	//处理完成，向dispatcher发送
	topic := msg.GetSource() + ".Frontend"
	rawData, _ := json.Marshal(msg)
	if err := nc.Producer.Public(topic, rawData); err == nil {
		nc.logger.Info("Succeed send message to ProduceChannel", zap.String("topic", topic))
	} else {
		nc.logger.Error("Failed to send  message to ProduceChannel", zap.Error(err))
	}
	_ = err
	return nil

}

// 10-14查询交易哈希详情
// 用户每天签到，每成功签到2次，送若干1千万wei的以太币

func (nc *NsqClient) HandleTxHashInfo(msg *models.Message) error {
	var err error
	errorCode := 200
	var errorMsg string
	var hashInfo *models.HashInfo

	redisConn := nc.redisPool.Get()
	defer redisConn.Close()

	username := msg.GetUserName() //用户自己的账号
	// token := msg.GetJwtToken()
	deviceID := msg.GetDeviceID()

	nc.logger.Info("HandleTxHashInfo start...",
		zap.String("username", username),
		zap.String("deviceId", deviceID))

	//取出当前设备的os， clientType， logonAt
	curDeviceHashKey := fmt.Sprintf("devices:%s:%s", username, deviceID)
	isMaster, _ := redis.Bool(redisConn.Do("HGET", curDeviceHashKey, "ismaster"))
	curOs, _ := redis.String(redisConn.Do("HGET", curDeviceHashKey, "os"))
	curClientType, _ := redis.Int(redisConn.Do("HGET", curDeviceHashKey, "clientType"))
	curLogonAt, _ := redis.Uint64(redisConn.Do("HGET", curDeviceHashKey, "logonAt"))

	nc.logger.Debug("HandleTxHashInfo",
		zap.Bool("isMaster", isMaster),
		zap.String("username", username),
		zap.String("deviceID", deviceID),
		zap.String("curOs", curOs),
		zap.Int("curClientType", curClientType),
		zap.Uint64("curLogonAt", curLogonAt))

	//打开msg里的负载， 获取请求参数
	body := msg.GetContent()

	//解包body
	req := &Wallet.TxHashInfoReq{}
	if err := proto.Unmarshal(body, req); err != nil {
		nc.logger.Error("Protobuf Unmarshal Error", zap.Error(err))
		errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
		errorMsg = fmt.Sprintf("Protobuf Unmarshal Error: %s", err.Error())
		goto COMPLETE

	} else {
		nc.logger.Debug("TxHashInfo payload",
			zap.Int("txType", int(req.TxType)),
			zap.String("txHash", req.TxHash),
		)

		switch req.TxType {
		case Global.TransactionType_DT_Deposit: //充值
			depositInfo, err := nc.Repository.GetDepositInfo(req.TxHash)
			if err != nil {
				errorCode = http.StatusNotFound //错误码， 200是正常，其它是错误
				errorMsg = fmt.Sprintf("UnKnown transation hash")
				goto COMPLETE
			}
			hashInfo, err = nc.ethService.QueryTxInfoByHash(req.TxHash)
			hashInfo.BlockNumber = depositInfo.BlockNumber
			hashInfo.To = ""

		case Global.TransactionType_DT_WithDraw: //提现
			withdrawInfo, err := nc.Repository.GetWithdrawInfo(req.TxHash)
			if err != nil {
				errorCode = http.StatusNotFound //错误码， 200是正常，其它是错误
				errorMsg = fmt.Sprintf("UnKnown transation hash")
				goto COMPLETE
			}
			hashInfo, err = nc.ethService.QueryTxInfoByHash(req.TxHash)
			hashInfo.BlockNumber = withdrawInfo.BlockNumber
			hashInfo.To = ""

		case Global.TransactionType_DT_Transfer: //转账
			transferInfo, err := nc.Repository.GetTransferInfo(req.TxHash)
			if err != nil {
				errorCode = http.StatusNotFound //错误码， 200是正常，其它是错误
				errorMsg = fmt.Sprintf("UnKnown transation hash")
				goto COMPLETE
			}
			hashInfo, err = nc.ethService.QueryTxInfoByHash(req.TxHash)
			hashInfo.BlockNumber = transferInfo.BlockNumber
			hashInfo.To = transferInfo.ToUsername

		default:
			errorCode = http.StatusNotFound //错误码， 200是正常，其它是错误
			errorMsg = fmt.Sprintf("UnKnown transation type")
			goto COMPLETE
		}

	}

COMPLETE:
	msg.SetCode(int32(errorCode)) //状态码
	if errorCode == 200 {
		rsp := &Wallet.TxHashInfoRsp{
			//区块高度
			BlockNumber: hashInfo.BlockNumber,
			//燃气值
			Gas: hashInfo.Gas,
			//随机数
			Nonce: hashInfo.Nonce,
			//数据，hex格式
			Data: hashInfo.Data,
			//接收者账号
			To: hashInfo.To,
		}
		data, _ := proto.Marshal(rsp)
		msg.FillBody(data)

	} else {
		msg.SetErrorMsg([]byte(errorMsg)) //错误提示
		msg.FillBody(nil)
	}

	//处理完成，向dispatcher发送
	topic := msg.GetSource() + ".Frontend"
	rawData, _ := json.Marshal(msg)
	if err := nc.Producer.Public(topic, rawData); err == nil {
		nc.logger.Info("Succeed send message to ProduceChannel", zap.String("topic", topic))
	} else {
		nc.logger.Error("Failed to send  message to ProduceChannel", zap.Error(err))
	}
	_ = err
	return nil

}