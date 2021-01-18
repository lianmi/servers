package services

import (
	"context"
	"encoding/json"
	"fmt"
	// Global "github.com/lianmi/servers/api/proto/global"
	Wallet "github.com/lianmi/servers/api/proto/wallet"
	"github.com/lianmi/servers/internal/app/walletservice/repositories"
	LMCommon "github.com/lianmi/servers/internal/common"
	"github.com/lianmi/servers/internal/pkg/models"
	"github.com/pkg/errors"
	"time"

	"github.com/gomodule/redigo/redis"
	"github.com/lianmi/servers/internal/pkg/blockchain"

	// uuid "github.com/satori/go.uuid"
	"go.uber.org/zap"
)

type WalletService interface {

	//订单完成或退款
	TransferByOrder(ctx context.Context, req *Wallet.TransferReq) (*Wallet.TransferResp, error)

	//获取用户钱包eth及LNMC代币余额
	GetUserBalance(ctx context.Context, req *Wallet.GetUserBalanceReq) (*Wallet.GetUserBalanceResp, error)

	//根据HD的索引号，获取对应的钱包地址
	GetWalletAddressbyBip32Index(ctx context.Context, req *Wallet.GetWalletAddressbyBip32IndexReq) (*Wallet.GetWalletAddressbyBip32IndexResp, error)

	//订单图片上链
	OrderImagesOnBlockchain(ctx context.Context, req *Wallet.OrderImagesOnBlockchainReq) (*Wallet.OrderImagesOnBlockchainResp, error)

	//获取某个订单的链上pending状态
	DoOrderPendingState(ctx context.Context, req *Wallet.OrderPendingStateReq) (*Wallet.OrderPendingStateResp, error)

	//支付宝预支付
	DoPreAlipay(ctx context.Context, req *Wallet.PreAlipayReq) (*Wallet.PreAlipayResp, error)

	//支付宝回调处理，用户充值
	DepositForPay(ctx context.Context, req *Wallet.DepositForPayReq) (*Wallet.DepositForPayResp, error)
}

type DefaultApisService struct {
	logger     *zap.Logger
	Repository repositories.WalletRepository
	redisPool  *redis.Pool
	ethService *blockchain.Service //连接以太坊geth的websocket
}

func NewApisService(logger *zap.Logger, Repository repositories.WalletRepository, redisPool *redis.Pool, ethService *blockchain.Service) WalletService {
	return &DefaultApisService{
		logger:     logger.With(zap.String("type", "WalletService")),
		Repository: Repository,
		redisPool:  redisPool,
		ethService: ethService,
	}
}

//PS 这里是本地调用，不是Grpc Client的调用，订单完成或退款
func (s *DefaultApisService) TransferByOrder(ctx context.Context, req *Wallet.TransferReq) (*Wallet.TransferResp, error) {
	s.logger.Debug("DefaultApisService, TransferByOrder", zap.String("OrderID", req.OrderID), zap.Int32("PayType", req.PayType))

	var err error

	var blockNumber uint64
	var txHash string

	redisConn := s.redisPool.Get()
	defer redisConn.Close()

	//查找出订单id是否存在，取出订单详情
	orderID := req.OrderID
	if orderID == "" {
		return &Wallet.TransferResp{
			ErrCode: 500,
			ErrMsg:  "OrderID is empty",
		}, errors.Wrap(err, "OrderID is empty")
	}

	if !(req.PayType == LMCommon.OrderTransferForDone || req.PayType == LMCommon.OrderTransferForCancel) {
		return &Wallet.TransferResp{
			ErrCode: 500,
			ErrMsg:  "PayType is not 1 or 2",
		}, errors.Wrap(err, "PayType is not 1 or 2")
	}
	//根据订单id获取buyUser及businessUser是谁
	orderIDKey := fmt.Sprintf("Order:%s", orderID)

	//获取订单的支付状态
	isPayed, err := redis.Bool(redisConn.Do("HGET", orderIDKey, "IsPayed"))
	productID, err := redis.String(redisConn.Do("HGET", orderIDKey, "ProductID"))
	buyUser, err := redis.String(redisConn.Do("HGET", orderIDKey, "BuyUser"))
	businessUser, err := redis.String(redisConn.Do("HGET", orderIDKey, "BusinessUser"))
	attachHash, err := redis.String(redisConn.Do("HGET", orderIDKey, "AttachHash"))
	orderTotalAmount, err := redis.Float64(redisConn.Do("HGET", orderIDKey, "OrderTotalAmount")) //人民币

	if isPayed == false {
		return &Wallet.TransferResp{
			ErrCode: 500,
			ErrMsg:  "Order is not Payed",
		}, errors.Wrap(err, "Order is not Payed")
	}

	if orderTotalAmount <= 0 {
		return &Wallet.TransferResp{
			ErrCode: 500,
			ErrMsg:  "OrderTotalAmount is less than 0",
		}, errors.Wrap(err, "OrderTotalAmount is less than 0")
	}

	//扣除手续费后，本次转账给商户的代币数量, 无小数点
	amountLNMC := uint64(orderTotalAmount * 100)

	buyUserWalletAddress, err := redis.String(redisConn.Do("HGET", fmt.Sprintf("userWallet:%s", buyUser), "WalletAddress"))
	if s.ethService.CheckIsvalidAddress(buyUserWalletAddress) == false {
		s.logger.Warn("buyUser非法钱包地址", zap.String("buyUserWalletAddress", buyUserWalletAddress))
		return &Wallet.TransferResp{
			ErrCode: 500,
			ErrMsg:  "BuyUser wallet address is not valid",
		}, errors.Wrap(err, "BuyUser wallet address is not valid")
	}

	businessUserWalletAddress, err := redis.String(redisConn.Do("HGET", fmt.Sprintf("userWallet:%s", businessUser), "WalletAddress"))
	if s.ethService.CheckIsvalidAddress(businessUserWalletAddress) == false {
		s.logger.Warn("BusinessUser非法钱包地址", zap.String("businessUserWalletAddress", businessUserWalletAddress))
		return &Wallet.TransferResp{
			ErrCode: 500,
			ErrMsg:  "BusinessUser wallet address is not valid",
		}, errors.Wrap(err, "BusinessUser wallet address is not valid")
	}

	//获取平台HD钱包利用bip32派生一个子私钥及子地址，作为证明人 - B签，也就是中转钱包地址

	bip32Index, err := redis.Uint64(redisConn.Do("HGET", fmt.Sprintf("userWallet:%s", buyUser), "Bip32Index"))
	newKeyPair := s.ethService.GetKeyPairsFromLeafIndex(bip32Index)
	bip32WalletAddress := newKeyPair.AddressHex //中转账号

	//中转账号的代币余额
	balanceLNMCBip32, err := s.ethService.GetLNMCTokenBalance(bip32WalletAddress)
	if err != nil {
		return &Wallet.TransferResp{
			ErrCode: 500,
			ErrMsg:  "Get LNMC Token Balance error",
		}, errors.Wrap(err, "Get LNMC Token Balance error")
	}

	if req.PayType == LMCommon.OrderTransferForDone { //向商户支付

		//商户钱包余额
		balanceLNMCBusinessUser, err := s.ethService.GetLNMCTokenBalance(businessUserWalletAddress)
		if err != nil {
			return &Wallet.TransferResp{
				ErrCode: 500,
				ErrMsg:  "Get LNMC Token Balance error",
			}, errors.Wrap(err, "Get LNMC Token Balance error")
		}

		//从中转账号转账代币到商户钱包地址

		if blockNumber, txHash, err = s.ethService.TransferLNMCTokenToAddress(newKeyPair.PrivateKeyHex, businessUserWalletAddress, int64(amountLNMC)); err != nil {
			return &Wallet.TransferResp{
				ErrCode: 500,
				ErrMsg:  "Transfer LNMC Token To Address error",
			}, errors.Wrap(err, "Transfer LNMC Token To Address error")
		}

		//转出之后，bip32的代币余额
		balanceLNMCBip32After, err := s.ethService.GetLNMCTokenBalance(bip32WalletAddress)
		if err != nil {
			s.logger.Error("balanceLNMCBip32After query error", zap.Error(err))
		}

		//转出之后，商户钱包余额
		balanceLNMCBusinessUserAfter, err := s.ethService.GetLNMCTokenBalance(businessUserWalletAddress)
		if err != nil {
			return &Wallet.TransferResp{
				ErrCode: 500,
				ErrMsg:  "Get LNMC Token Balance error",
			}, errors.Wrap(err, "Get LNMC Token Balance error")
		}
		s.logger.Debug("中转账号向商户转账详细",
			zap.String("中转账号地址", bip32WalletAddress),
			zap.Uint64("中转账号转账之前的代币余额", balanceLNMCBip32),
			zap.Uint64("中转账号转账代币数量", amountLNMC),
			zap.Uint64("中转账号转账之后的代币余额", balanceLNMCBip32After),
			zap.String("商户账号地址", businessUserWalletAddress),
			zap.Uint64("商户转账之前的代币余额", balanceLNMCBusinessUser),
			zap.Uint64("商户转账之后的代币余额", balanceLNMCBusinessUserAfter),
		)

		//增加商户到账记录到 MySQL
		lnmcOrderTransferHistory := &models.LnmcOrderTransferHistory{
			OrderID:                   orderID,                   //订单ID
			ProductID:                 productID,                 //商品ID
			PayType:                   1,                         //订单完成的到账
			BuyUser:                   buyUser,                   //买家注册号
			BusinessUser:              businessUser,              //商户注册号
			BuyUserWalletAddress:      buyUserWalletAddress,      //买家链上地址，默认是用户HD钱包的第0号索引，用于存储连米币
			BusinessUserWalletAddress: businessUserWalletAddress, //商户链上地址，默认是用户HD钱包的第0号索引，用于存储连米币
			AttachHash:                attachHash,                //订单内容哈希，上链
			Bip32Index:                bip32Index,                //买家对应平台HD钱包Bip32派生索引号
			BalanceLNMCBefore:         balanceLNMCBip32,          //平台HD钱包在转账时刻的连米币数量
			OrderTotalAmount:          orderTotalAmount,          //人民币格式的订单总金额
			AmountLNMC:                amountLNMC,                //本次转账的连米币数量, 无小数点
			BalanceLNMCAfter:          balanceLNMCBip32After,     //平台HD钱包在转账之后的连米币数量
			BlockNumber:               blockNumber,               //成功执行合约的所在区块高度
			TxHash:                    txHash,                    //交易哈希

		}

		if err := s.Repository.AddLnmcOrderTransferHistory(lnmcOrderTransferHistory); err != nil {
			s.logger.Error("到账 AddLnmcOrderTransferHistory  error", zap.Error(err))
		} else {
			s.logger.Debug("到账 AddLnmcOrderTransferHistory succeed")
		}

		//更新接收者的收款历史表
		lnmcCollectionHistory := &models.LnmcCollectionHistory{
			FromUsername:      buyUser,                      //发送者
			FromWalletAddress: buyUserWalletAddress,         //发送者钱包地址
			ToUsername:        businessUser,                 //接收者
			ToWalletAddress:   businessUserWalletAddress,    //中转钱包地址
			OrderID:           orderID,                      //订单ID
			BalanceLNMCBefore: balanceLNMCBip32,             //接收方用户在转账时刻的连米币数量
			AmountLNMC:        amountLNMC,                   //本次转账的用户连米币数量
			BalanceLNMCAfter:  balanceLNMCBusinessUserAfter, //接收方用户在转账之后的连米币数量
			Bip32Index:        bip32Index,                   //平台HD钱包Bip32派生索引号
			BlockNumber:       blockNumber,
			TxHash:            txHash,
		}
		if err := s.Repository.AddeCollectionHistory(lnmcCollectionHistory); err != nil {
			s.logger.Error("商户到账记录 AddeCollectionHistory  error", zap.Error(err))
		} else {
			s.logger.Debug("商户到账记录 AddeCollectionHistory succeed")
		}

	} else if req.PayType == LMCommon.OrderTransferForCancel { //退款
		//买家钱包余额
		balanceLNMCBuyUser, err := s.ethService.GetLNMCTokenBalance(buyUserWalletAddress)
		if err != nil {
			return &Wallet.TransferResp{
				ErrCode: 500,
				ErrMsg:  "Get LNMC Token Balance error",
			}, errors.Wrap(err, "Get LNMC Token Balance error")
		}

		//从中转账号转账代币到买家钱包地址
		if blockNumber, txHash, err = s.ethService.TransferLNMCTokenToAddress(newKeyPair.PrivateKeyHex, buyUserWalletAddress, int64(amountLNMC)); err != nil {
			return &Wallet.TransferResp{
				ErrCode: 500,
				ErrMsg:  "Transfer LNMC Token To buyer Address error",
			}, errors.Wrap(err, "Transfer LNMC Token To buyer Address error")
		}
		//转出之后，bip32的代币余额
		balanceLNMCBip32After, err := s.ethService.GetLNMCTokenBalance(bip32WalletAddress)
		if err != nil {
			s.logger.Error("balanceLNMCBip32After query error", zap.Error(err))
		}

		//转出之后，买家的代币余额
		balanceLNMCBuyUserAfter, err := s.ethService.GetLNMCTokenBalance(buyUserWalletAddress)
		if err != nil {
			s.logger.Error("balanceLNMCBuyerUserAfter query error", zap.Error(err))
		}
		s.logger.Debug("中转账号向买家退款详细",
			zap.String("中转账号地址", bip32WalletAddress),
			zap.Uint64("中转账号转账之前的代币余额", balanceLNMCBip32),
			zap.Uint64("中转账号转账代币数量", amountLNMC),
			zap.Uint64("中转账号转账之后的代币余额", balanceLNMCBip32After),
			zap.String("买家账号地址", businessUserWalletAddress),
			zap.Uint64("买家转账之前的代币余额", balanceLNMCBuyUser),
			zap.Uint64("买家转账之后的代币余额", balanceLNMCBuyUserAfter),
		)

		//增加退款记录到 MySQL
		lnmcOrderTransferHistory := &models.LnmcOrderTransferHistory{
			OrderID:                   orderID,                   //订单ID
			ProductID:                 productID,                 //商品ID
			PayType:                   2,                         //订单撤单或拒单的退款
			BuyUser:                   buyUser,                   //买家注册号
			BusinessUser:              businessUser,              //商户注册号
			BuyUserWalletAddress:      buyUserWalletAddress,      //买家链上地址，默认是用户HD钱包的第0号索引，用于存储连米币
			BusinessUserWalletAddress: businessUserWalletAddress, //商户链上地址，默认是用户HD钱包的第0号索引，用于存储连米币
			AttachHash:                attachHash,                //订单内容哈希，上链
			Bip32Index:                bip32Index,                //买家对应平台HD钱包Bip32派生索引号
			BalanceLNMCBefore:         balanceLNMCBip32,          //平台HD钱包在转账时刻的连米币数量
			OrderTotalAmount:          orderTotalAmount,          //人民币格式的订单总金额
			AmountLNMC:                amountLNMC,                //本次转账的连米币数量, 无小数点
			BalanceLNMCAfter:          balanceLNMCBip32After,     //平台HD钱包在转账之后的连米币数量
			BlockNumber:               blockNumber,               //成功执行合约的所在区块高度
			TxHash:                    txHash,                    //交易哈希

		}

		if err := s.Repository.AddLnmcOrderTransferHistory(lnmcOrderTransferHistory); err != nil {
			s.logger.Error("退款 AddLnmcOrderTransferHistory  error", zap.Error(err))
		} else {
			s.logger.Debug("退款 AddLnmcOrderTransferHistory succeed")
		}

	}

	resp := &Wallet.TransferResp{
		ErrCode: 200,
		ErrMsg:  "",
	}

	return resp, nil
}

//获取用户钱包eth及LNMC代币余额
func (s *DefaultApisService) GetUserBalance(ctx context.Context, req *Wallet.GetUserBalanceReq) (*Wallet.GetUserBalanceResp, error) {
	var err error
	redisConn := s.redisPool.Get()
	defer redisConn.Close()

	username := req.Username
	userWalletAddress, err := redis.String(redisConn.Do("HGET", fmt.Sprintf("userWallet:%s", username), "WalletAddress"))
	if s.ethService.CheckIsvalidAddress(userWalletAddress) == false {
		s.logger.Warn("user非法钱包地址", zap.String("username", username), zap.String("userWalletAddress", userWalletAddress))
		return nil, errors.Wrap(err, "username wallet address is not valid")
	}

	balanceEth, err := s.ethService.GetWeiBalance(userWalletAddress)
	if err != nil {
		return nil, err
	}

	balanceLNMC, err := s.ethService.GetLNMCTokenBalance(userWalletAddress)
	if err != nil {
		return nil, err
	}

	resp := &Wallet.GetUserBalanceResp{
		BalanceEth:    balanceEth,
		BalanceLNMC:   balanceLNMC,
		WalletAddress: userWalletAddress,
	}

	return resp, nil
}

//根据HD的索引号，获取对应的钱包地址
func (s *DefaultApisService) GetWalletAddressbyBip32Index(ctx context.Context, req *Wallet.GetWalletAddressbyBip32IndexReq) (*Wallet.GetWalletAddressbyBip32IndexResp, error) {
	newKeyPair := s.ethService.GetKeyPairsFromLeafIndex(req.Bip32Index)
	bip32WalletAddress := newKeyPair.AddressHex //中转账号
	return &Wallet.GetWalletAddressbyBip32IndexResp{
		WalletAddress: bip32WalletAddress,
	}, nil
}

// //确认购买会员的支付交易
// func (s *DefaultApisService) SendConfirmPayForMembership(ctx context.Context, req *Wallet.SendConfirmPayForMembershipReq) (*Wallet.SendConfirmPayForMembershipResp, error) {
// 	//发起购买会员的账号
// 	username := req.Username
// 	orderID := req.OrderID
// 	content := req.Content

// 	redisConn := s.redisPool.Get()
// 	defer redisConn.Close()

// 	//约定 凡是购买会员的接收钱包账户是叶子3
// 	bip32Index := uint64(LMCommon.MEMBERSHIPINDEX)
// 	newKeyPair := s.ethService.GetKeyPairsFromLeafIndex(bip32Index)
// 	toWalletAddress := newKeyPair.AddressHex //中转账号

// 	prePayForMembershipKey := fmt.Sprintf("PrePayForMembership:%s", orderID)
// 	amountLNMC, err := redis.Uint64(redisConn.Do("HGET", prePayForMembershipKey, "AmountLNMC"))
// 	orderTotalAmount := float64(amountLNMC / 100) //实际花费，人民币格式

// 	//获得付费类型- 包月，包季，包年
// 	payType, err := redis.Int(redisConn.Do("HGET", prePayForMembershipKey, "PayType"))
// 	if err != nil {
// 		return nil, err
// 	}

// 	//获得会员的账号
// 	payForUsername, err := redis.String(redisConn.Do("HGET", prePayForMembershipKey, "PayForUsername"))
// 	if err != nil {
// 		return nil, err
// 	}

// 	//发起支付的用户钱包地址
// 	userWalletAddress, err := redis.String(redisConn.Do("HGET", prePayForMembershipKey, "WalletAddress"))
// 	if err != nil {
// 		return nil, err
// 	}

// 	//附言
// 	content, err = redis.String(redisConn.Do("HGET", prePayForMembershipKey, "Content"))
// 	if err != nil {
// 		return nil, err
// 	}

// 	balanceLNMC, err := s.ethService.GetLNMCTokenBalance(userWalletAddress)

// 	//调用eth接口，将发起方签名的转到目标接收者的交易数据广播到链上- A签
// 	blockNumber, hash, err := s.ethService.SendSignedTxToGeth(req.GetSignedTxToTarget())
// 	if err != nil {
// 		return nil, err
// 	}

// 	s.logger.Info("发起方转到目标接收者的交易数据广播到链上 A签成功 ",
// 		zap.String("username", username),
// 		zap.String("toWalletAddress", toWalletAddress),
// 		zap.Uint64("blockNumber", blockNumber),
// 		zap.String("hash", hash),
// 	)

// 	// 获取发送者链上代币余额
// 	balanceAfter, err := s.ethService.GetLNMCTokenBalance(userWalletAddress)
// 	if err != nil {
// 		return nil, err
// 	}
// 	s.logger.Info("获取发送者链上代币余额",
// 		zap.String("username", username),
// 		zap.String("userWalletAddress", userWalletAddress),
// 		zap.Uint64("balanceAfter", balanceAfter),
// 	)

// 	//更新Redis里用户钱包的代币数量
// 	redisConn.Do("HSET",
// 		fmt.Sprintf("userWallet:%s", username),
// 		"LNMCAmount",
// 		balanceAfter)

// 	//更新转账记录到 MySQL
// 	lnmcTransferHistory := &models.LnmcTransferHistory{
// 		Username:          username,          //发起支付
// 		OrderID:           orderID,           //如果是订单支付 ，非空
// 		WalletAddress:     userWalletAddress, //发起方钱包账户
// 		BalanceLNMCBefore: balanceLNMC,       //发送方用户在转账时刻的连米币数量
// 		AmountLNMC:        amountLNMC,        //本次转账的用户连米币数量
// 		BalanceLNMCAfter:  balanceAfter,      //发送方用户在转账之后的连米币数量
// 		Bip32Index:        bip32Index,        //平台HD钱包Bip32派生索引号
// 		State:             1,                 //执行状态，0-默认未执行，1-A签，2-全部完成
// 		BlockNumber:       blockNumber,
// 		TxHash:            hash,
// 		Content:           content,
// 	}
// 	s.Repository.UpdateLnmcTransferHistory(lnmcTransferHistory)

// 	//更新State到 redis  HSET
// 	_, err = redisConn.Do("HSET",
// 		prePayForMembershipKey,
// 		"State", 1,
// 	)

// 	_, err = redisConn.Do("HSET",
// 		prePayForMembershipKey,
// 		"SignedTx", req.GetSignedTxToTarget(),
// 	)
// 	_, err = redisConn.Do("HSET",
// 		prePayForMembershipKey,
// 		"BlockNumber", blockNumber,
// 	)
// 	_, err = redisConn.Do("HSET",
// 		prePayForMembershipKey,
// 		"Hash", hash,
// 	)

// 	//到期时间, ms
// 	curVipEndDate, err := redis.Int64(redisConn.Do("HGET", fmt.Sprintf("userData:%s", username), "VipEndDate"))

// 	if curVipEndDate == 0 || curVipEndDate < time.Now().UnixNano()/1e6 {
// 		curVipEndDate = time.Now().UnixNano() / 1e6
// 	}
// 	curTime := time.Unix(curVipEndDate/1e3, 0) //将秒转换为time类型

// 	//增加到期时间
// 	var endTime int64

// 	switch Global.VipUserPayType(payType) {
// 	case Global.VipUserPayType_VIP_Year: //包年
// 		endTime = curTime.AddDate(0, 0, 365).UnixNano() / 1e6
// 	case Global.VipUserPayType_VIP_Season: //包季
// 		endTime = curTime.AddDate(0, 0, 90).UnixNano() / 1e6
// 	case Global.VipUserPayType_VIP_Month: //包月
// 		endTime = curTime.AddDate(0, 0, 30).UnixNano() / 1e6
// 		// case Global.VipUserPayType_VIP_Week: //包周，体验卡

// 	}
// 	_, err = redisConn.Do("HSET", fmt.Sprintf("userData:%s", username), "VipEndDate", endTime)

// 	//如果替他人支付，通知对方
// 	if username != payForUsername {

// 		//TODO

// 	}

// 	//确认支付成功后，就需要分配佣金
// 	s.Repository.AddCommission(orderTotalAmount, username, req.OrderID)

// 	return &Wallet.SendConfirmPayForMembershipResp{
// 		OrderTotalAmount: orderTotalAmount,
// 		PayForUsername:   payForUsername,
// 		BlockNumber:      blockNumber,
// 		Hash:             hash,
// 		Time:             uint64(time.Now().UnixNano() / 1e6),
// 	}, nil
// }

//订单图片上链
func (s *DefaultApisService) OrderImagesOnBlockchain(ctx context.Context, req *Wallet.OrderImagesOnBlockchainReq) (*Wallet.OrderImagesOnBlockchainResp, error) {

	var data []byte

	redisConn := s.redisPool.Get()
	defer redisConn.Close()

	buyUserWalletAddress, err := redis.String(redisConn.Do("HGET", fmt.Sprintf("userWallet:%s", req.BuyUsername), "WalletAddress"))
	if s.ethService.CheckIsvalidAddress(buyUserWalletAddress) == false {
		s.logger.Warn("buyUser非法钱包地址", zap.String("buyUserWalletAddress", buyUserWalletAddress))
		return nil, errors.Wrap(err, "BuyUser wallet address is not valid")
	}

	businessUserWalletAddress, err := redis.String(redisConn.Do("HGET", fmt.Sprintf("userWallet:%s", req.BusinessUsername), "WalletAddress"))
	if s.ethService.CheckIsvalidAddress(businessUserWalletAddress) == false {
		s.logger.Warn("BusinessUser非法钱包地址", zap.String("businessUserWalletAddress", businessUserWalletAddress))
		return nil, errors.Wrap(err, "BusinessUser wallet address is not valid")
	}
	//订单详情
	orderIDKey := fmt.Sprintf("Order:%s", req.OrderID)
	//获取订单的具体信息
	productID, _ := redis.String(redisConn.Do("HGET", orderIDKey, "ProductID"))
	buyUser, _ := redis.String(redisConn.Do("HGET", orderIDKey, "BuyUser"))
	businessUser, _ := redis.String(redisConn.Do("HGET", orderIDKey, "BusinessUser"))
	orderTotalAmount, _ := redis.Float64(redisConn.Do("HGET", orderIDKey, "OrderTotalAmount"))
	attachHash, _ := redis.String(redisConn.Do("HGET", orderIDKey, "AttachHash"))

	orderImagesOnBlockChain := &models.OrderImagesOnBlockChainHistory{
		OrderID:          req.OrderID,      //订单IDs
		ProductID:        productID,        //商品ID
		AttachHash:       attachHash,       //订单内容hash
		BuyUsername:      buyUser,          //买家注册号
		BusinessUsername: businessUser,     //商户注册号
		Cost:             orderTotalAmount, //本订单的总金额
		BusinessOssImage: req.OrderImage,   //订单图片在商户的oss objectID
	}
	data, err = json.Marshal(orderImagesOnBlockChain)
	if err != nil {
		return nil, err
	}
	amount := int64(orderTotalAmount * 100)

	//调用ETH接口
	blockNumber, hash, err := s.ethService.TransferEthToOtherAccount(buyUserWalletAddress, amount, data)
	if err != nil {
		return nil, err
	}
	return &Wallet.OrderImagesOnBlockchainResp{
		OrderID:     req.OrderID,
		BlockNumber: blockNumber,
		Hash:        hash,
		Time:        uint64(time.Now().UnixNano() / 1e6),
	}, nil
}

//支付宝预支付
func (s *DefaultApisService) DoPreAlipay(ctx context.Context, req *Wallet.PreAlipayReq) (*Wallet.PreAlipayResp, error) {
	return s.Repository.DoPreAlipay(ctx, req)
}

//支付宝回调处理，用户充值
func (s *DefaultApisService) DepositForPay(ctx context.Context, req *Wallet.DepositForPayReq) (*Wallet.DepositForPayResp, error) {
	//TODO 检测TradeNo是否已经充值过了，避免刷币
	redisConn := s.redisPool.Get()
	defer redisConn.Close()

	//通过grpc获取发起购买者用户的余额
	//当前用户的代币余额
	getUserBalanceResp, err := s.GetUserBalance(ctx, &Wallet.GetUserBalanceReq{
		Username: req.Username,
	})
	if err != nil {
		s.logger.Error("GetUserBalance 错误", zap.Error(err))
		return nil, err
	}

	//调用eth接口， 给用户钱包充值连米币
	amount := int64(req.TotalAmount * 100)
	blockNumber, hash, balanceAfter, err := s.ethService.TransferLNMCFromLeaf1ToNormalAddress(getUserBalanceResp.WalletAddress, amount)
	if err != nil {
		s.logger.Error("TransferLNMCFromLeaf1ToNormalAddress 错误", zap.Error(err))
		return nil, err
	}

	s.logger.Info("充值前后的钱包信息",
		zap.String("username", req.Username),
		zap.Float64("支付成功的人民币", req.TotalAmount),
		zap.Int64("充值连米币", amount),
		zap.String("walletAddress", getUserBalanceResp.WalletAddress),
		zap.Uint64("充值前余额", getUserBalanceResp.BalanceLNMC),
		zap.Uint64("充值后余额", balanceAfter),
	)

	if err := s.Repository.SaveDepositForPay(req.TradeNo, hash, blockNumber, balanceAfter); err != nil {
		return nil, err
	}

	return &Wallet.DepositForPayResp{
		//充值之后的连米币余额
		BalanceLNMC: balanceAfter,
		// 区块高度
		BlockNumber: blockNumber,
		// 交易哈希hex
		Hash: hash,
		//时间
		Time: uint64(time.Now().UnixNano() / 1e6),
	}, nil
}

//获取某个订单的链上pending状态
func (s *DefaultApisService) DoOrderPendingState(ctx context.Context, req *Wallet.OrderPendingStateReq) (*Wallet.OrderPendingStateResp, error) {
	// var pending bool
	var err error

	redisConn := s.redisPool.Get()
	defer redisConn.Close()

	//根据OrderID查询出原始数据
	orderIDKey := fmt.Sprintf("Order:%s", req.OrderID)
	uuidStr, _ := redis.String(redisConn.Do("HGET", orderIDKey, "TransferUUID"))
	if uuidStr == "" {
		s.logger.Error("uuidStr is empty")
		return nil, errors.Wrap(err, "uuid is empty")
	}

	preTransferKey := fmt.Sprintf("PreTransfer:%s", uuidStr)

	signedTxHash, _ := redis.String(redisConn.Do("HGET", preTransferKey, "SignedTxHash"))
	pendingNonce, _ := redis.Uint64(redisConn.Do("HGET", preTransferKey, "PendingNonce"))

	s.logger.Debug("DoOrderPendingState start...",
		zap.String("OrderID", req.OrderID),
		zap.String("signedTxHash", signedTxHash),
		zap.Uint64("pendingNonce", pendingNonce),
	)

	if signedTxHash == "" {
		s.logger.Error("signedTxHash is empty")
		return nil, errors.Wrap(err, "TxHash is empty")
	}
	/*
		receipt, err := s.ethService.CheckTransactionReceipt(signedTxHash)
		if err != nil {
			s.logger.Error("CheckTransactionReceipt failed ", zap.Error(err))
			return nil, err
		}
		if receipt.Status == 0 {
			pending = false //打包完成
		} else if receipt.Status == 1 {
			pending = true //打包中
		}

		rsp := &Wallet.OrderPendingStateResp{
			Pending: pending,
			//燃气值
			CumulativeGasUsed: receipt.CumulativeGasUsed,
			//实际燃气值
			GasUsed: receipt.GasUsed,
			//当前交易的nonce
			Nonce: pendingNonce,
			// 交易哈希hex
			TxHash: receipt.TxHash.Hex(),
			// 交易区块哈希，如果打包成功就有值
			BlockHash: receipt.BlockHash.Hex(),
			// 交易区块高度，如果打包成功就有值
			BlockNumber: receipt.BlockNumber.Uint64(),
			// 交易index，如果打包成功就有值
			TransactionIndex: uint32(receipt.TransactionIndex),
		}
		return rsp, nil
	*/

	return nil, errors.Wrap(err, "TxHash is empty")

}
