package services

import (
	"context"
	"fmt"
	// "time"

	Wallet "github.com/lianmi/servers/api/proto/wallet"
	"github.com/lianmi/servers/internal/app/walletservice/repositories"
	LMCommon "github.com/lianmi/servers/internal/common"
	"github.com/lianmi/servers/internal/pkg/models"
	"github.com/pkg/errors"

	"github.com/gomodule/redigo/redis"
	"github.com/lianmi/servers/internal/pkg/blockchain"
	"go.uber.org/zap"
)

type WalletService interface {

	//订单完成或退款
	TransferByOrder(ctx context.Context, req *Wallet.TransferReq) (*Wallet.TransferResp, error)
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

	// foo, err := redis.String(redisConn.Do("GET", "bar"))
	// if err != nil {
	// 	s.logger.Error("GET error", zap.Error(err))
	// } else {
	// 	s.logger.Debug("bar", zap.String("Foo", foo))
	// }

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
	orderTotalAmount, err := redis.Float64(redisConn.Do("HGET", orderIDKey, "OrderTotalAmount"))

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

	//本次转账的代币数量, 无小数点
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

	if req.PayType == LMCommon.OrderTransferForDone {

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

		if err := s.Repository.SaveLnmcOrderTransferHistory(lnmcOrderTransferHistory); err != nil {
			s.logger.Error("到账 SaveLnmcOrderTransferHistory  error", zap.Error(err))
		} else {
			s.logger.Debug("到账 SaveLnmcOrderTransferHistory succeed")
		}

	} else if req.PayType == LMCommon.OrderTransferForCancel {
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

		if err := s.Repository.SaveLnmcOrderTransferHistory(lnmcOrderTransferHistory); err != nil {
			s.logger.Error("退款 SaveLnmcOrderTransferHistory  error", zap.Error(err))
		} else {
			s.logger.Debug("退款 SaveLnmcOrderTransferHistory succeed")
		}

	}

	resp := &Wallet.TransferResp{
		ErrCode: 200,
		ErrMsg:  "",
	}

	return resp, nil
}
