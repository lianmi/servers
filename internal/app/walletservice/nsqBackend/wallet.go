package nsqBackend

import (
	"encoding/hex"
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

		//给用户钱包发送6000000个gas
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
		nc.logger.Error("Failed to send message to ProduceChannel", zap.Error(err))
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
		nc.logger.Error("Failed to send message to ProduceChannel", zap.Error(err))
	}
	_ = err
	return nil
}

/*
10-3 发起转账
1. 用户下单需要支付或者手工转账时，向服务端发起一个转账申请, 接收者也必须开通钱包 。
2. 服务端收到请求后，判断发起方的余额是否足够支付，如果足够，则动态部署一个多签合约。
3. 证明人为一个平台派生地址，3方分别是发起方、接收方、平台方(要记录BIP44对应的index, 因为每个转账的index都自动递增，不能相同，用来区分每笔转账 )
4. 服务端向发起方返回合约地址及Tx裸交易二进制序列

*/
func (nc *NsqClient) HandlePreTransfer(msg *models.Message) error {
	var err error
	errorCode := 200
	var errorMsg string
	var data []byte

	var walletAddress string   //用户钱包地址
	var toWalletAddress string //接收者钱包地址

	var blockNumber uint64
	var contractAddress string //多签智能合约地址
	var hash string
	var newBip32Index uint64 //自增的平台HD钱包派生索引号

	var amountLNMCBefore uint64 //用户当前代币数量
	var amountLNMC uint64       //本次转账的代币数量,  等于amount * 100

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
			zap.String("targetUserName", req.GetTargetUserName()),
			zap.Float64("amount", req.GetAmount()), //人民币格式 ，有小数点
		)

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

		//检测接收者钱包是否注册, 如果没注册， 则不能转账
		if isExists, err := redis.Bool(redisConn.Do("HEXISTS", fmt.Sprintf("userWallet:%s", req.GetTargetUserName()), "WalletAddress")); err != nil {
			errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
			errorMsg = fmt.Sprintf("HEXISTS error")
			goto COMPLETE
		} else {
			if !isExists {
				nc.logger.Warn("钱包没注册，不能转账", zap.String("TargetUserName", req.GetTargetUserName()))
				errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
				errorMsg = fmt.Sprintf("Target Wallet had not registered")
				goto COMPLETE
			}
		}

		walletAddress, err = redis.String(redisConn.Do("HGET", fmt.Sprintf("userWallet:%s", username), "WalletAddress"))
		if err != nil {
			errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
			errorMsg = fmt.Sprintf("HGET error")
			goto COMPLETE
		}

		toWalletAddress, err = redis.String(redisConn.Do("HGET", fmt.Sprintf("userWallet:%s", req.GetTargetUserName()), "WalletAddress"))
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
		amountLNMCBefore, err = redis.Uint64(redisConn.Do("HGET", fmt.Sprintf("userWallet:%s", username), "LNMCAmount"))
		if err != nil {
			errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
			errorMsg = fmt.Sprintf("HGET error")
			goto COMPLETE
		}
		nc.logger.Info("当前用户的钱包信息",
			zap.String("username", username),
			zap.String("walletAddress", walletAddress),
			zap.Uint64("当前余额 amountLNMCBefore", amountLNMCBefore),
		)

		amountLNMC = uint64(req.GetAmount() * 100)

		//amount是人民币格式（单位是 元），要转为int64
		if amountLNMCBefore < amountLNMC {
			nc.logger.Warn("余额不足")
			errorCode = http.StatusBadRequest              //错误码， 400
			errorMsg = fmt.Sprintf("Not sufficient funds") //  余额不足
			goto COMPLETE
		}

		//平台HD钱包利用bip32派生一个子私钥及子地址，作为证明人 - B签
		newBip32Index, err = redis.Uint64(redisConn.Do("INCR", "Bip32Index"))
		newKeyPair := nc.ethService.GetKeyPairsFromLeafIndex(newBip32Index)

		nc.logger.Info("平台HD钱包利用bip32派生一个子私钥及子地址",
			zap.String("username", username),
			zap.Uint64("newBip32Index", newBip32Index),
			zap.String("PrivateKeyHex", newKeyPair.PrivateKeyHex),
			zap.String("AddressHex", newKeyPair.AddressHex),
		)

		//向此新地址转入若干gas用来运行合约
		blockNumber, hash, err = nc.ethService.TransferEthToOtherAccount(newKeyPair.AddressHex, LMCommon.GASLIMIT)
		if err != nil {
			nc.logger.Error("向此新地址转入若干gas用来运行合约 失败", zap.String("AddressHex", newKeyPair.AddressHex), zap.Error(err))
			errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
			errorMsg = fmt.Sprintf("Transfer Eth error")
			goto COMPLETE
		} else {
			nc.logger.Info("向此新地址转入若干gas用来运行合约成功",
				zap.String("AddressHex", newKeyPair.AddressHex),
				zap.Uint64("blockNumber", blockNumber),
				zap.String("hash", hash),
			)

		}

		//调用eth接口， 部署多签合约, A-发起者， B-证明人
		contractAddress, blockNumber, hash, err = nc.ethService.DeployMultiSig(walletAddress, newKeyPair.AddressHex)
		if err != nil {
			nc.logger.Error("部署多签合约 失败", zap.String("walletAddress", walletAddress), zap.String("AddressHex", newKeyPair.AddressHex), zap.Error(err))
			errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
			errorMsg = fmt.Sprintf("Deploy MultiSig Contract error")
			goto COMPLETE
		} else {
			nc.logger.Info("部署多签合约成功",
				zap.String("AddressHex", newKeyPair.AddressHex),
				zap.String("contractAddress", contractAddress),
				zap.Uint64("blockNumber", blockNumber),
				zap.String("hash", hash),
			)

		}

		//保存预审核转账记录到 MySQL
		lnmcTransferHistory := &models.LnmcTransferHistory{
			Username:            username,                //发起支付
			ToUsername:          req.GetTargetUserName(), //接收者
			WalletAddress:       walletAddress,           // 发起方钱包账户
			ToWalletAddress:     toWalletAddress,         //接收者钱包账户
			AmountLNMC:          amountLNMC,              //本次转账的用户连米币数量
			AmountLNMCBefore:    amountLNMCBefore,        //发送方用户在转账时刻的连米币数量
			Bip32Index:          newBip32Index,           //平台HD钱包Bip32派生索引号
			ContractAddress:     contractAddress,         //多签合约地址
			ContractBlockNumber: blockNumber,             //多签合约所在区块高度
			ContractHash:        hash,                    //多签合约的哈希
			State:               0,                       //多签合约执行状态，0-默认未执行，1-A签，2-全部完成

		}
		nc.SaveLnmcTransferHistory(lnmcTransferHistory)

		//发起者钱包账户向多签合约账户转账，由于服务端没有发起者的私钥，所以只能生成裸交易，让发起者签名后才能向多签合约账户转账
		tokens := fmt.Sprintf("%d", amountLNMC)
		rawTxToMulsig, err := nc.ethService.GenerateTransferLNMCTokenTx(walletAddress, contractAddress, tokens)
		if err != nil {
			nc.logger.Error("构造发起者向多签合约转账的交易 失败", zap.String("walletAddress", walletAddress), zap.String("contractAddress", contractAddress), zap.Error(err))
			errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
			errorMsg = fmt.Sprintf("GenerateTransferLNMCTokenTx error")
			goto COMPLETE
		}

		//从多签合约账号转到目标接收者账号
		rawTxToTarget, err := nc.ethService.GenerateRawTx(contractAddress, walletAddress, toWalletAddress, tokens)
		if err != nil {
			nc.logger.Error("GenerateRawTx 失败", zap.String("walletAddress", walletAddress), zap.String("AddressHex", newKeyPair.AddressHex), zap.Error(err))
			errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
			errorMsg = fmt.Sprintf("Generate RawDesc error")
			goto COMPLETE
		}

		rsp := &Wallet.PreTransferRsp{
			RawTxToMulsig: &Wallet.RawDesc{
				Nonce:           rawTxToMulsig.Nonce,
				GasPrice:        rawTxToMulsig.GasPrice,
				GasLimit:        rawTxToMulsig.GasLimit,
				ChainID:         rawTxToMulsig.ChainID,
				Txdata:          rawTxToMulsig.Txdata,
				ContractAddress: contractAddress, //多签合约
				Value:           0,
			},
			RawTxToTarget: &Wallet.RawDesc{
				Nonce:           rawTxToTarget.Nonce,
				GasPrice:        rawTxToTarget.GasPrice,
				GasLimit:        rawTxToTarget.GasLimit,
				ChainID:         rawTxToTarget.ChainID,
				Txdata:          rawTxToTarget.Txdata,
				ContractAddress: contractAddress, //多签合约
				Value:           0,
			},
			Time: uint64(time.Now().UnixNano() / 1e6), // 当前时间
		}
		data, _ = proto.Marshal(rsp)

		//保存预审核转账记录到 redis
		_, err = redisConn.Do("HMSET",
			fmt.Sprintf("PreTransfer:%s:%s", username, contractAddress),
			"Username", username,
			"ToUsername", req.GetTargetUserName(),
			"WalletAddress", walletAddress,
			"ToWalletAddress", toWalletAddress,
			"AmountLNMC", amountLNMC,
			"AmountLNMCBefore", amountLNMCBefore,
			"Bip32Index", newBip32Index,
			"ContractAddress", contractAddress,
			"ContractBlockNumber", blockNumber,
			"ContractHash", hash,
			"State", 0,
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

/*
10-4 确认转账
1. 发起方收到服务端的预审核10-3回包后 ，需要对返回的裸交易哈希进行签名(A签)
2. 服务端收到后， 如果是普通转账，则进行B签, 并广播到链上，完成转账， 接收方将收到代币
3. 与9-11 的区别是请求参数没有携带 订单id
*/
func (nc *NsqClient) HandleConfirmTransfer(msg *models.Message) error {
	var err error
	errorCode := 200
	var errorMsg string
	var data []byte

	var walletAddress string //用户钱包地址
	// var toWalletAddress string //接收者钱包地址

	// var blockNumber uint64
	var contractAddress string //多签智能合约地址
	// var hash string
	var newBip32Index uint64 //自增的平台HD钱包派生索引号

	var amountLNMCBefore uint64 //用户当前代币数量
	var amountLNMC uint64       //本次转账的代币数量, 无小数点

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
			zap.String("contractAddress", req.GetContractAddress()),
			zap.ByteString("SignedTxToMultisig", req.GetSignedTxToMultisig()),
			zap.ByteString("SignedTxToTarget", req.GetSignedTxToTarget()), //签名后的Tx(A签)
		)

		if req.GetContractAddress() == "" {

			nc.logger.Warn("合约地址不能为空", zap.String("contractAddress", req.GetContractAddress()))
			errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
			errorMsg = fmt.Sprintf("contractAddress is empty")
			goto COMPLETE
		}
		contractAddress = req.GetContractAddress()

		if len(req.SignedTxToMultisig) == 0 {

			nc.logger.Warn("SignedTxToMultisig不能为空")
			errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
			errorMsg = fmt.Sprintf("SignedTxToMultisig is empty")
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

		if nc.ethService.CheckIsvalidAddress(walletAddress) == false {
			nc.logger.Warn("非法钱包地址", zap.String("WalletAddress", walletAddress))
			errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
			errorMsg = fmt.Sprintf("WalletAddress is not valid")
			goto COMPLETE
		}

		//当前用户的代币余额
		amountLNMCBefore, err = redis.Uint64(redisConn.Do("HGET", fmt.Sprintf("userWallet:%s", username), "LNMCAmount"))
		if err != nil {
			errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
			errorMsg = fmt.Sprintf("HGET error")
			goto COMPLETE
		}
		nc.logger.Info("当前用户的钱包信息",
			zap.String("username", username),
			zap.String("walletAddress", walletAddress),
			zap.Uint64("当前余额 amountLNMCBefore", amountLNMCBefore),
		)

		amountLNMC, err = redis.Uint64(redisConn.Do("HGET", fmt.Sprintf("PreTransfer:%s:%s", username, contractAddress), "AmountLNMC"))
		if err != nil {
			errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
			errorMsg = fmt.Sprintf("HGET AmountLNMC error")
			goto COMPLETE
		}

		//平台HD钱包利用bip32派生一个子私钥及子地址，作为证明人 - B签
		newBip32Index, err = redis.Uint64(redisConn.Do("HGET", fmt.Sprintf("PreTransfer:%s:%s", username, contractAddress), "Bip32Index"))
		if err != nil {
			errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
			errorMsg = fmt.Sprintf("HGET Bip32Index error")
			goto COMPLETE
		}
		newKeyPair := nc.ethService.GetKeyPairsFromLeafIndex(newBip32Index)

		nc.logger.Info("平台HD钱包利用bip32派生一个子私钥及子地址",
			zap.String("username", username),
			zap.Uint64("newBip32Index", newBip32Index),
			zap.String("PrivateKeyHex", newKeyPair.PrivateKeyHex),
			zap.String("AddressHex", newKeyPair.AddressHex),
		)

		//调用eth接口，将发起方签名的转账给多签的交易数据广播到链上
		nc.ethService.SendSignedTxToGeth(hex.EncodeToString(req.GetSignedTxToMultisig()))

		//调用eth接口，将发起方签名的从多签合约转到目标接收者的交易数据广播到链上- A签
		blockNumber, hash, err := nc.ethService.SendSignedTxToGeth(hex.EncodeToString(req.GetSignedTxToTarget()))
		if err != nil {
			errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
			errorMsg = fmt.Sprintf("SendSignedTxToGeth error")
			goto COMPLETE
		}

		//更新转账记录到 MySQL
		lnmcTransferHistory := &models.LnmcTransferHistory{
			Username: username, //发起支付
			// ToUsername:          req.GetTargetUserName(), //接收者
			WalletAddress: walletAddress, // 发起方钱包账户
			// ToWalletAddress:     toWalletAddress,         //接收者钱包账户
			AmountLNMC:          amountLNMC,                                    //本次转账的用户连米币数量
			AmountLNMCBefore:    amountLNMCBefore,                              //发送方用户在转账时刻的连米币数量
			Bip32Index:          newBip32Index,                                 //平台HD钱包Bip32派生索引号
			ContractAddress:     contractAddress,                               //多签合约地址
			ContractBlockNumber: blockNumber,                                   //多签合约所在区块高度
			ContractHash:        hash,                                          //多签合约的哈希
			State:               1,                                             //多签合约执行状态，0-默认未执行，1-A签，2-全部完成
			SignedTx:            hex.EncodeToString(req.GetSignedTxToTarget()), //hex格式
			SucceedBlockNumber:  blockNumber,
			SucceedHash:         hash,
		}
		nc.UpdateLnmcTransferHistory(lnmcTransferHistory)

		//更新转账记录到 redis  HSET
		_, err = redisConn.Do("HSET",
			fmt.Sprintf("PreTransfer:%s:%s", username, contractAddress),
			"State", 1,
		)

		_, err = redisConn.Do("HSET",
			fmt.Sprintf("PreTransfer:%s:%s", username, contractAddress),
			"SignedTx", hex.EncodeToString(req.GetSignedTxToTarget()),
		)
		_, err = redisConn.Do("HSET",
			fmt.Sprintf("PreTransfer:%s:%s", username, contractAddress),
			"SucceedBlockNumber", blockNumber,
		)
		_, err = redisConn.Do("HSET",
			fmt.Sprintf("PreTransfer:%s:%s", username, contractAddress),
			"SucceedHash", hash,
		)

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
