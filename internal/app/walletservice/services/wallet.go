package services

import (
	"context"
	"encoding/json"
	"fmt"
	Wallet "github.com/lianmi/servers/api/proto/wallet"
	"github.com/lianmi/servers/internal/app/walletservice/repositories"
	LMCommon "github.com/lianmi/servers/internal/common"
	"github.com/lianmi/servers/internal/pkg/models"
	"github.com/pkg/errors"
	"time"

	"github.com/gomodule/redigo/redis"
	"github.com/lianmi/servers/internal/pkg/blockchain"

	uuid "github.com/satori/go.uuid"
	"go.uber.org/zap"
)

type WalletService interface {

	//订单完成或退款
	TransferByOrder(ctx context.Context, req *Wallet.TransferReq) (*Wallet.TransferResp, error)

	//获取用户钱包eth及LNMC代币余额
	GetUserBalance(ctx context.Context, req *Wallet.GetUserBalanceReq) (*Wallet.GetUserBalanceResp, error)

	//根据HD的索引号，获取对应的钱包地址
	GetWalletAddressbyBip32Index(ctx context.Context, req *Wallet.GetWalletAddressbyBip32IndexReq) (*Wallet.GetWalletAddressbyBip32IndexResp, error)

	//发起一个购买会员的预支付，返回裸交易
	SendPrePayForMembership(ctx context.Context, req *Wallet.SendPrePayForMembershipReq) (*Wallet.SendPrePayForMembershipResp, error)

	//确认购买会员的支付交易
	SendConfirmPayForMembership(ctx context.Context, req *Wallet.SendConfirmPayForMembershipReq) (*Wallet.SendConfirmPayForMembershipResp, error)

	//订单图片上链
	OrderImagesOnBlockchain(ctx context.Context, req *Wallet.OrderImagesOnBlockchainReq) (*Wallet.OrderImagesOnBlockchainResp, error)

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

//发起一个购买会员的预支付，返回裸交易
func (s *DefaultApisService) SendPrePayForMembership(ctx context.Context, req *Wallet.SendPrePayForMembershipReq) (*Wallet.SendPrePayForMembershipResp, error) {
	var err error
	redisConn := s.redisPool.Get()
	defer redisConn.Close()

	username := req.Username
	userWalletAddress, err := redis.String(redisConn.Do("HGET", fmt.Sprintf("userWallet:%s", username), "WalletAddress"))
	if s.ethService.CheckIsvalidAddress(userWalletAddress) == false {
		s.logger.Warn("user非法钱包地址", zap.String("username", username), zap.String("userWalletAddress", userWalletAddress))
		return nil, errors.Wrap(err, "username wallet address is not valid")
	}

	balanceLNMC, err := s.ethService.GetLNMCTokenBalance(userWalletAddress)
	if err != nil {
		return nil, err
	}

	//根据PayType获取到VIP价格
	vipPrice, err := s.Repository.GetVipUserPrice(int(req.PayType))
	if err != nil {
		return nil, err
	}

	//TODO 暂时不用理会优惠券, 人民币
	orderTotalAmount := float64(vipPrice.Price)

	//转换为连米币
	amountLNMC := uint64(orderTotalAmount * 100)

	//生成一个OrderID, 发起一个预支付
	orderID := uuid.NewV4().String()

	//约定 凡是购买会员的接收钱包账户是叶子3
	bip32Index := uint64(LMCommon.MEMBERSHIPINDEX)
	newKeyPair := s.ethService.GetKeyPairsFromLeafIndex(bip32Index)
	toWalletAddress := newKeyPair.AddressHex //中转账号

	//保存预审核转账记录到 MySQL
	lnmcTransferHistory := &models.LnmcTransferHistory{
		Username:          req.Username,      //发起支付
		ToUsername:        "",                //空
		OrderID:           orderID,           //非空
		WalletAddress:     userWalletAddress, //发起方钱包账户
		ToWalletAddress:   toWalletAddress,   //接收者钱包账户
		BalanceLNMCBefore: balanceLNMC,       //发送方用户在转账时刻的连米币数量
		AmountLNMC:        amountLNMC,        //本次转账的用户连米币数量
		Bip32Index:        bip32Index,        //平台HD钱包Bip32派生索引号
		State:             0,                 //执行状态，0-默认未执行，1-A签，2-全部完成
	}
	s.Repository.AddLnmcTransferHistory(lnmcTransferHistory)

	//发起者钱包账户向接收者账户转账，由于服务端没有发起者的私钥，所以只能生成裸交易，让发起者签名后才能向接收者账户转账
	tokens := int64(amountLNMC)
	rawDescToTarget, err := s.ethService.GenerateTransferLNMCTokenTx(userWalletAddress, toWalletAddress, tokens)
	if err != nil {
		s.logger.Error("构造购买会员的交易 失败", zap.String("userWalletAddress", userWalletAddress), zap.String("toWalletAddress", toWalletAddress), zap.Error(err))
	}

	//保存预审核转账记录到 redis
	_, err = redisConn.Do("HMSET",
		fmt.Sprintf("PrePayForMembership:%s", orderID),
		"Username", username,
		"OrderID", orderID,
		"PayForUsername", req.PayForUsername,
		"WalletAddress", userWalletAddress,
		"ToWalletAddress", toWalletAddress,
		"AmountLNMC", amountLNMC, //真实的花费, 连米币
		"BalanceLNMCBefore", balanceLNMC,
		"Bip32Index", bip32Index,
		"State", 0,
		"CreateAt", uint64(time.Now().UnixNano()/1e6),
	)

	return &Wallet.SendPrePayForMembershipResp{
		//向收款方转账的裸交易结构体
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
		OrderID:          orderID,
		OrderTotalAmount: orderTotalAmount,
		Time:             uint64(time.Now().UnixNano() / 1e6),
	}, nil
}

//确认购买会员的支付交易
func (s *DefaultApisService) SendConfirmPayForMembership(ctx context.Context, req *Wallet.SendConfirmPayForMembershipReq) (*Wallet.SendConfirmPayForMembershipResp, error) {
	//发起购买会员的账号
	username := req.Username
	orderID := req.OrderID
	content := req.Content

	redisConn := s.redisPool.Get()
	defer redisConn.Close()

	//约定 凡是购买会员的接收钱包账户是叶子3
	bip32Index := uint64(LMCommon.MEMBERSHIPINDEX)
	newKeyPair := s.ethService.GetKeyPairsFromLeafIndex(bip32Index)
	toWalletAddress := newKeyPair.AddressHex //中转账号

	amountLNMC, err := redis.Uint64(redisConn.Do("HGET", fmt.Sprintf("PrePayForMembership:%s", orderID), "AmountLNMC"))
	orderTotalAmount := float64(amountLNMC / 100)

	//获得会员的账号
	payForUsername, err := redis.String(redisConn.Do("HGET", fmt.Sprintf("PrePayForMembership:%s", orderID), "PayForUsername"))
	if err != nil {
		return nil, err
	}

	//发起支付的用户钱包地址
	userWalletAddress, err := redis.String(redisConn.Do("HGET", fmt.Sprintf("PrePayForMembership:%s", orderID), "WalletAddress"))
	if err != nil {
		return nil, err
	}

	//附言
	content, err = redis.String(redisConn.Do("HGET", fmt.Sprintf("PrePayForMembership:%s", orderID), "Content"))
	if err != nil {
		return nil, err
	}

	balanceLNMC, err := s.ethService.GetLNMCTokenBalance(userWalletAddress)
	// toBalanceLNMC, err := s.ethService.GetLNMCTokenBalance(toWalletAddress)

	//调用eth接口，将发起方签名的转到目标接收者的交易数据广播到链上- A签
	blockNumber, hash, err := s.ethService.SendSignedTxToGeth(req.GetSignedTxToTarget())
	if err != nil {
		return nil, err
	}

	s.logger.Info("发起方转到目标接收者的交易数据广播到链上 A签成功 ",
		zap.String("username", username),
		zap.String("toWalletAddress", toWalletAddress),
		zap.Uint64("blockNumber", blockNumber),
		zap.String("hash", hash),
	)

	// 获取发送者链上代币余额
	balanceAfter, err := s.ethService.GetLNMCTokenBalance(userWalletAddress)
	if err != nil {
		return nil, err
	}
	s.logger.Info("获取发送者链上代币余额",
		zap.String("username", username),
		zap.String("userWalletAddress", userWalletAddress),
		zap.Uint64("balanceAfter", balanceAfter),
	)

	//更新Redis里用户钱包的代币数量
	redisConn.Do("HSET",
		fmt.Sprintf("userWallet:%s", username),
		"LNMCAmount",
		balanceAfter)

	//更新转账记录到 MySQL
	lnmcTransferHistory := &models.LnmcTransferHistory{
		Username:          username,          //发起支付
		OrderID:           orderID,           //如果是订单支付 ，非空
		WalletAddress:     userWalletAddress, //发起方钱包账户
		BalanceLNMCBefore: balanceLNMC,       //发送方用户在转账时刻的连米币数量
		AmountLNMC:        amountLNMC,        //本次转账的用户连米币数量
		BalanceLNMCAfter:  balanceAfter,      //发送方用户在转账之后的连米币数量
		Bip32Index:        bip32Index,        //平台HD钱包Bip32派生索引号
		State:             1,                 //执行状态，0-默认未执行，1-A签，2-全部完成
		BlockNumber:       blockNumber,
		TxHash:            hash,
		Content:           content,
	}
	s.Repository.UpdateLnmcTransferHistory(lnmcTransferHistory)

	//更新转账记录到 redis  HSET
	_, err = redisConn.Do("HSET",
		fmt.Sprintf("PrePayForMembership:%s", orderID),
		"State", 1,
	)

	_, err = redisConn.Do("HSET",
		fmt.Sprintf("PrePayForMembership:%s", orderID),
		"SignedTx", req.GetSignedTxToTarget(),
	)
	_, err = redisConn.Do("HSET",
		fmt.Sprintf("PrePayForMembership:%s", orderID),
		"BlockNumber", blockNumber,
	)
	_, err = redisConn.Do("HSET",
		fmt.Sprintf("PrePayForMembership:%s", orderID),
		"Hash", hash,
	)

	//如果替他人支付，通知对方
	if username != payForUsername {

		//TODO

	}

	return &Wallet.SendConfirmPayForMembershipResp{
		OrderTotalAmount: orderTotalAmount,
		PayForUsername:   payForUsername,
		BlockNumber:      blockNumber,
		Hash:             hash,
		Time:             uint64(time.Now().UnixNano() / 1e6),
	}, nil
}

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
