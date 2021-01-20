/*
本文件实现 api/proto/wallet/grpc.proto 钱包的Grpc接口, rpc接口方法必须全部实现
*/
package grpcservers

import (
	"context"
	Wallet "github.com/lianmi/servers/api/proto/wallet"
	"github.com/lianmi/servers/internal/app/walletservice/services"
	"go.uber.org/zap"
)

type WalletServer struct {
	logger *zap.Logger

	service services.WalletService
}

func NewWalletServer(logger *zap.Logger, ps services.WalletService) (*WalletServer, error) {
	return &WalletServer{
		logger:  logger,
		service: ps,
	}, nil
}

//订单完成或退款
func (s *WalletServer) TransferByOrder(ctx context.Context, req *Wallet.TransferReq) (*Wallet.TransferResp, error) {
	return s.service.TransferByOrder(ctx, req)
}

//获取余额
func (s *WalletServer) GetUserBalance(ctx context.Context, req *Wallet.GetUserBalanceReq) (*Wallet.GetUserBalanceResp, error) {
	return s.service.GetUserBalance(ctx, req)
}

//根据bip32索引号获取地址
func (s *WalletServer) GetWalletAddressbyBip32Index(ctx context.Context, req *Wallet.GetWalletAddressbyBip32IndexReq) (*Wallet.GetWalletAddressbyBip32IndexResp, error) {
	return s.service.GetWalletAddressbyBip32Index(ctx, req)
}

//订单上链
func (s *WalletServer) OrderImagesOnBlockchain(ctx context.Context, req *Wallet.OrderImagesOnBlockchainReq) (*Wallet.OrderImagesOnBlockchainResp, error) {
	return s.service.OrderImagesOnBlockchain(ctx, req)
}

//支付宝发起预支付
func (s *WalletServer) DoPreAlipay(ctx context.Context, req *Wallet.PreAlipayReq) (*Wallet.PreAlipayResp, error) {
	return s.service.DoPreAlipay(ctx, req)
}

//微信发起预支付
func (s *WalletServer) DoPreWXpay(ctx context.Context, req *Wallet.PreWXpayReq) (*Wallet.PreWXpayResp, error) {
	return s.service.DoPreWXpay(ctx, req)
}

//支付宝支付成功的回调处理
func (s *WalletServer) DepositForPay(ctx context.Context, req *Wallet.DepositForPayReq) (*Wallet.DepositForPayResp, error) {
	return s.service.DepositForPay(ctx, req)
}

// 用户端: 根据 OrderID 获取此订单在链上的pending状态
func (s *WalletServer) DoOrderPendingState(ctx context.Context, req *Wallet.OrderPendingStateReq) (*Wallet.OrderPendingStateResp, error) {
	return s.service.DoOrderPendingState(ctx, req)
}
