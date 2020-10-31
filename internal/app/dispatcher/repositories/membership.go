package repositories

import (
	// "encoding/json"
	"fmt"
	// "time"
	// "net/http"

	// "github.com/golang/protobuf/proto"
	"github.com/gomodule/redigo/redis"
	// "github.com/jinzhu/gorm"
	// Auth "github.com/lianmi/servers/api/proto/auth"
	Service "github.com/lianmi/servers/api/proto/service"
	// User "github.com/lianmi/servers/api/proto/user"
	// "github.com/lianmi/servers/internal/app/dispatcher/grpcclients"
	// "github.com/lianmi/servers/internal/app/dispatcher/nsqMq"
	// LMCommon "github.com/lianmi/servers/internal/common"
	// "github.com/lianmi/servers/internal/pkg/models"
	// "github.com/pkg/errors"
	"go.uber.org/zap"
	// "github.com/lianmi/servers/util/dateutil"
)

//TODO
//商户查询当前名下用户总数，按月统计付费会员总数及返佣金额，是否已经返佣
func (s *MysqlLianmiRepository) GetBusinessMembership(isRebate bool) (*Service.GetBusinessMembershipResp, error) {
	var err error

	redisConn := s.redisPool.Get()
	defer redisConn.Close()

	return nil, err

}

//TODO
func (s *MysqlLianmiRepository) PayForMembership(payForUsername string) error {

	//支付完成后，需要向商户，支付者，会员获得者推送系统通知
	//构建数据完成，向dispatcher发送
	return nil
}

//预生成一个购买会员的订单， 返回OrderID及预转账裸交易数据
func (s *MysqlLianmiRepository) PreOrderForPayMembership(username, deviceID string) error {

	var err error
	/*
		errorCode := 200
		var errorMsg string
		var data []byte

		var walletAddress string   //用户钱包地址
		var toWalletAddress string //接收者钱包地址,
		var bip32Index uint64      //购买会员的钱包是第2号叶子

		var balanceLNMC uint64 //用户当前代币数量
		var blockNumber uint64
		var hash string

		var amountLNMC uint64 //本次转账的代币数量,  等于amount * 100
		var balanceETH uint64 //当前用户的Eth余额
	*/
	redisConn := s.redisPool.Get()
	defer redisConn.Close()

	//取出当前设备的os， clientType， logonAt
	curDeviceHashKey := fmt.Sprintf("devices:%s:%s", username, deviceID)
	isMaster, _ := redis.Bool(redisConn.Do("HGET", curDeviceHashKey, "ismaster"))
	curOs, _ := redis.String(redisConn.Do("HGET", curDeviceHashKey, "os"))
	curClientType, _ := redis.Int(redisConn.Do("HGET", curDeviceHashKey, "clientType"))
	curLogonAt, _ := redis.Uint64(redisConn.Do("HGET", curDeviceHashKey, "logonAt"))

	s.logger.Debug("HandlePreTransfer",
		zap.Bool("isMaster", isMaster),
		zap.String("username", username),
		zap.String("deviceID", deviceID),
		zap.String("curOs", curOs),
		zap.Int("curClientType", curClientType),
		zap.Uint64("curLogonAt", curLogonAt))

	/*
			//系统HD钱包的第1号叶子
			bip32Index = LMCommon.MEMBERSHIPINDEX
			keyPair := s.ethService.GetKeyPairsFromLeafIndex(bip32Index)

			s.logger.Info("用户对应的叶子编号、子私钥及子地址",
				zap.String("username", username),
				zap.Uint64("newBip32Index", bip32Index),
				zap.String("PrivateKeyHex", keyPair.PrivateKeyHex),
				zap.String("AddressHex", keyPair.AddressHex),
			)

			//当前用户的代币余额
			balanceLNMC, err = s.ethService.GetLNMCTokenBalance(walletAddress)
			if err != nil {
				errorCode = http.StatusInternalServerError //错误码， 200是正常，其它是错误
				errorMsg = fmt.Sprintf("HGET error")
				goto COMPLETE
			}

			//由于会员价格是99元，是人民币，以元为单位，因此，需要乘以100
			amountLNMC = uint64(LMCommon.MEMBERSHIPPRICE * 100)

			//当前用户的Eth余额
			balanceETH, err = s.ethService.GetWeiBalance(walletAddress)

			s.logger.Info("当前用户的钱包信息",
				zap.String("username", username),
				zap.String("walletAddress", walletAddress),
				zap.Uint64("当前代币余额 balanceLNMC", balanceLNMC),
				zap.Uint64("当前ETH余额 balanceETH", balanceETH),
				zap.Uint64("转账代币数量  amountLNMC", amountLNMC),
			)
			if balanceETH < LMCommon.GASLIMIT {
				s.logger.Warn("gas余额不足")
				errorCode = http.StatusPaymentRequired       //错误码， 402
				errorMsg = fmt.Sprintf("Not sufficient gas") //  余额不足
				goto COMPLETE
			}

			//判断是否有足够代币数量
			if balanceLNMC < amountLNMC {
				s.logger.Warn("余额不足",
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
			s.Repository.SaveLnmcTransferHistory(lnmcTransferHistory)

			//发起者钱包账户向接收者账户转账，由于服务端没有发起者的私钥，所以只能生成裸交易，让发起者签名后才能向接收者账户转账
			tokens := int64(amountLNMC)
			rawDescToTarget, err := s.ethService.GenerateTransferLNMCTokenTx(walletAddress, toWalletAddress, tokens)
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

	*/
	_ = err
	return nil
}
