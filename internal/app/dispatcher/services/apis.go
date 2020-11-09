package services

import (
	"context"
	Auth "github.com/lianmi/servers/api/proto/auth"
	Order "github.com/lianmi/servers/api/proto/order"
	Wallet "github.com/lianmi/servers/api/proto/wallet"
	LMCommon "github.com/lianmi/servers/internal/common"

	"github.com/lianmi/servers/internal/app/dispatcher/repositories"
	"github.com/lianmi/servers/internal/pkg/models"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"time"

	pb "github.com/lianmi/servers/api/proto/user"
)

type LianmiApisService interface {
	BlockUser(username string) error
	DisBlockUser(username string) (*models.User, error)
	Register(user *models.User) (string, error)
	ResetPassword(mobile, password string, user *models.User) error
	GetUserRoles(username string) []*models.Role

	//检测用户登录
	CheckUser(isMaster bool, smscode, username, password, deviceID, os string, clientType int) bool

	// 判断用户名是否已存在
	ExistUserByName(username string) bool
	// 判断手机号码是否已存在
	ExistUserByMobile(mobile string) bool
	SaveUserToken(username, deviceID string, token string, expire time.Time) bool
	SignOut(token, username, deviceID string) bool
	ExistsTokenInRedis(deviceID, token string) bool

	//生成注册校验码
	GenerateSmsCode(mobile string) bool

	// GetUser(ID uint64) (*models.User, error)

	//检测校验码是否正确
	CheckSmsCode(mobile, smscode string) bool

	//修改密码
	ChanPassword(username string, req *Auth.ChanPasswordReq) error

	//授权新创建的群组
	ApproveTeam(teamID string) error

	//封禁群组
	BlockTeam(teamID string) error

	//解封群组
	DisBlockTeam(teamID string) error

	//======后台相关======/
	AddGeneralProduct(generalProduct *models.GeneralProduct) error

	GetGeneralProductByID(productID string) (p *models.GeneralProduct, err error)

	GetGeneralProductPage(pageIndex, pageSize int, total *uint64, where interface{}) ([]*models.GeneralProduct, error)

	UpdateGeneralProduct(generalProduct *models.GeneralProduct) error

	DeleteGeneralProduct(productID string) bool

	QueryCustomerServices(req *Auth.QueryCustomerServiceReq) ([]*models.CustomerServiceInfo, error)

	AddCustomerService(req *Auth.AddCustomerServiceReq) error

	DeleteCustomerService(req *Auth.DeleteCustomerServiceReq) bool

	UpdateCustomerService(req *Auth.UpdateCustomerServiceReq) error

	QueryGrades(req *Auth.GradeReq, pageIndex int, pageSize int, total *uint64, where interface{}) ([]*models.Grade, error)

	AddGrade(req *Auth.AddGradeReq) (string, error)

	SubmitGrade(req *Auth.SubmitGradeReq) error

	//商户查询当前名下用户总数，按月统计付费会员总数及返佣金额，是否已经返佣
	GetBusinessMembership(isRebate bool) (*Auth.GetBusinessMembershipResp, error)

	PreOrderForPayMembership(ctx context.Context, username, deviceID, payForUsername string) (*Auth.PreOrderForPayMembershipResp, error)

	ConfirmPayForMembership(ctx context.Context, username string, req *Auth.ConfirmPayForMembershipReq) (*Auth.ConfirmPayForMembershipResp, error)

	//Grpc 获取用户信息
	GetUser(ctx context.Context, in *Auth.UserReq) (*Auth.UserRsp, error)
}

type DefaultLianmiApisService struct {
	logger              *zap.Logger
	Repository          repositories.LianmiRepository
	authGrpcClientSvc   Auth.LianmiAuthClient     //auth的grpc client
	orderGrpcClientSvc  Order.LianmiOrderClient   //order的grpc client
	walletGrpcClientSvc Wallet.LianmiWalletClient //wallet的grpc client
}

func NewLianmiApisService(logger *zap.Logger, repository repositories.LianmiRepository, lc Auth.LianmiAuthClient, oc Order.LianmiOrderClient, wc Wallet.LianmiWalletClient) LianmiApisService {
	return &DefaultLianmiApisService{
		logger:              logger.With(zap.String("type", "DefaultLianmiApisService")),
		Repository:          repository,
		authGrpcClientSvc:   lc,
		orderGrpcClientSvc:  oc,
		walletGrpcClientSvc: wc,
	}
}

func (s *DefaultLianmiApisService) GetUser(ctx context.Context, in *Auth.UserReq) (*Auth.UserRsp, error) {
	s.logger.Debug("GetUser", zap.Uint64("ID", in.Id))

	//从微服务获取用户数据
	resp, err := s.authGrpcClientSvc.GetUser(ctx, in)

	return resp, err
}

func (s *DefaultLianmiApisService) BlockUser(username string) (err error) {
	s.logger.Debug("BlockUser", zap.String("username", username))
	// if err = s.Repository.BlockUser(username); err != nil {
	// 	return errors.Wrap(err, "Block user error")
	// }

	return nil
}
func (s *DefaultLianmiApisService) DisBlockUser(username string) (p *models.User, err error) {
	s.logger.Debug("DisBlockUser", zap.String("username", username))
	if p, err = s.Repository.DisBlockUser(username); err != nil {
		return nil, errors.Wrap(err, "DisBlockUser user error")
	}

	return
}

//生成短信校验码
func (s *DefaultLianmiApisService) GenerateSmsCode(mobile string) bool {
	return s.Repository.GenerateSmsCode(mobile)
}

//检测校验码是否正确
func (s *DefaultLianmiApisService) CheckSmsCode(mobile, smscode string) bool {
	return s.Repository.CheckSmsCode(mobile, smscode)
}

func (s *DefaultLianmiApisService) Register(user *models.User) (string, error) {
	// var username string
	var err error
	if err = s.Repository.Register(user); err != nil {
		return "", errors.Wrap(err, "Register user error")
	}

	//当成功插入User数据后，user为指针地址，可以获取到ID的值。省去了查数据库拿ID的值步骤
	var role models.Role
	role.UserID = user.ID
	role.UserName = user.Username
	role.Value = ""
	if user.GetUserType() == pb.UserType_Ut_Operator { //10086
		role.Value = "admin"
	}
	//同时增加用户类型角色
	if err = s.Repository.AddRole(&role); err != nil {
		//增加角色失败，需要删除users表的对应用户
		if s.Repository.DeleteUser(user.ID) == false {

			return "", errors.Wrap(err, "Register role error")
		}
	}

	return user.Username, nil
}

func (s *DefaultLianmiApisService) ResetPassword(mobile, password string, user *models.User) error {
	// if err := s.Repository.ResetPassword(mobile, password, user); err != nil {
	// 	return errors.Wrap(err, "ResetPassword error")
	// }

	return nil
}

//修改密码
func (s *DefaultLianmiApisService) ChanPassword(username string, req *Auth.ChanPasswordReq) error {

	// return s.Repository.ChanPassword(username, req)
	return nil
}

func (s *DefaultLianmiApisService) GetUserRoles(username string) []*models.Role {
	where := models.Role{UserName: username}
	return s.Repository.GetUserRoles(&where)
}

//CheckUser 身份验证
func (s *DefaultLianmiApisService) CheckUser(isMaster bool, smscode, username, password, deviceID, os string, clientType int) bool {

	return s.Repository.CheckUser(isMaster, smscode, username, password, deviceID, os, clientType)
}

func (s *DefaultLianmiApisService) ExistUserByName(username string) bool {

	return s.Repository.ExistUserByName(username)
}

// 判断手机号码是否已存在
func (s *DefaultLianmiApisService) ExistUserByMobile(mobile string) bool {
	return s.Repository.ExistUserByMobile(mobile)
}

func (s *DefaultLianmiApisService) SaveUserToken(username, deviceID string, token string, expire time.Time) bool {
	return s.Repository.SaveUserToken(username, deviceID, token, expire)
}

func (s *DefaultLianmiApisService) SignOut(token, username, deviceID string) bool {
	return s.Repository.SignOut(token, username, deviceID)
}

func (s *DefaultLianmiApisService) ExistsTokenInRedis(deviceID, token string) bool {
	return s.Repository.ExistsTokenInRedis(deviceID, token)
}

func (s *DefaultLianmiApisService) ApproveTeam(teamID string) error {
	// return s.Repository.ApproveTeam(teamID)

	//TODO
	return nil
}

//封禁群组
func (s *DefaultLianmiApisService) BlockTeam(teamID string) error {
	// return s.Repository.BlockTeam(teamID)

	//TODO
	return nil
}

//解封群组
func (s *DefaultLianmiApisService) DisBlockTeam(teamID string) error {
	// return s.Repository.DisBlockTeam(teamID)
	//TODO
	return nil
}

//======后台相关======/
func (s *DefaultLianmiApisService) AddGeneralProduct(generalProduct *models.GeneralProduct) error {
	// return s.Repository.AddGeneralProduct(generalProduct)

	//TODO
	return nil
}

//查询通用商品(id) - Read
func (s *DefaultLianmiApisService) GetGeneralProductByID(productID string) (p *models.GeneralProduct, err error) {

	// return s.Repository.GetGeneralProductByID(productID)
	//TODO
	return nil, nil
}

//查询通用商品分页 - Page
func (s *DefaultLianmiApisService) GetGeneralProductPage(pageIndex, pageSize int, total *uint64, where interface{}) ([]*models.GeneralProduct, error) {

	// return s.Repository.GetGeneralProductPage(pageIndex, pageSize, total, where)

	//TODO
	return nil, nil
}

// 修改通用商品 - Update
func (s *DefaultLianmiApisService) UpdateGeneralProduct(generalProduct *models.GeneralProduct) error {

	// return s.Repository.UpdateGeneralProduct(generalProduct)
	//TODO
	return nil

}

// 删除通用商品 - Delete
func (s *DefaultLianmiApisService) DeleteGeneralProduct(productID string) bool {

	// return s.Repository.DeleteGeneralProduct(productID)
	//TODO
	return false

}

//获取在线客服id数组
func (s *DefaultLianmiApisService) QueryCustomerServices(req *Auth.QueryCustomerServiceReq) ([]*models.CustomerServiceInfo, error) {

	// return s.Repository.QueryCustomerServices(req)
	//TODO
	return nil, nil

}

//增加在线客服id
func (s *DefaultLianmiApisService) AddCustomerService(req *Auth.AddCustomerServiceReq) error {

	// return s.Repository.AddCustomerService(req)
	//TODO
	return nil

}

func (s *DefaultLianmiApisService) DeleteCustomerService(req *Auth.DeleteCustomerServiceReq) bool {
	// return s.Repository.DeleteCustomerService(req)
	//TODO
	return false

}

func (s *DefaultLianmiApisService) UpdateCustomerService(req *Auth.UpdateCustomerServiceReq) error {

	// return s.Repository.UpdateCustomerService(req)
	//TODO
	return nil

}

func (s *DefaultLianmiApisService) QueryGrades(req *Auth.GradeReq, pageIndex int, pageSize int, total *uint64, where interface{}) ([]*models.Grade, error) {
	// return s.Repository.QueryGrades(req, pageIndex, pageSize, total, where)
	//TODO
	return nil, nil
}

func (s *DefaultLianmiApisService) AddGrade(req *Auth.AddGradeReq) (string, error) {
	// return s.Repository.AddGrade(req)
	//TODO
	return "", nil
}

func (s *DefaultLianmiApisService) SubmitGrade(req *Auth.SubmitGradeReq) error {
	// return s.Repository.SubmitGrade(req)
	//TODO
	return nil
}

//商户查询当前名下用户总数，按月统计付费会员总数及返佣金额，是否已经返佣
func (s *DefaultLianmiApisService) GetBusinessMembership(isRebate bool) (*Auth.GetBusinessMembershipResp, error) {
	// return s.Repository.GetBusinessMembership(isRebate)

	//TODO
	return nil, nil
}

func (s *DefaultLianmiApisService) PreOrderForPayMembership(ctx context.Context, username, deviceID, payForUsername string) (*Auth.PreOrderForPayMembershipResp, error) {
	//通过grpc获取发起购买者用户的余额
	//当前用户的代币余额
	getUserBalanceResp, err := s.walletGrpcClientSvc.GetUserBalance(ctx, &Wallet.GetUserBalanceReq{
		Username: username,
	})
	if err != nil {
		s.logger.Error("walletGrpcClientSvc.GetUserBalance 错误", zap.Error(err))
		return nil, err
	}

	//由于会员价格是99元，是人民币，以元为单位，因此，需要乘以100
	amountLNMC := uint64(LMCommon.MEMBERSHIPPRICE * 100)

	s.logger.Info("当前用户的钱包信息",
		zap.String("username", username),
		zap.Uint64("当前代币余额 balanceLNMC", getUserBalanceResp.BalanceLNMC),
		zap.Uint64("当前ETH余额 balanceETH", getUserBalanceResp.BalanceEth),
	)
	if getUserBalanceResp.BalanceEth < LMCommon.GASLIMIT {
		return nil, errors.Wrap(err, "gas余额不足")
	}

	//判断是否有足够代币数量
	if getUserBalanceResp.BalanceLNMC < amountLNMC {
		return nil, errors.Wrap(err, "LNMC余额不足")
	}

	//调用钱包Grpcserver，生成一个类似 10-3 的预支付裸交易
	sendPrePayForMembershipResp, err := s.walletGrpcClientSvc.SendPrePayForMembership(ctx, &Wallet.SendPrePayForMembershipReq{
		Username:       username,
		PayForUsername: payForUsername,
	})
	if err != nil {
		s.logger.Error("walletGrpcClientSvc.SendPrePayForMembership 错误", zap.Error(err))
		return nil, err
	}

	return &Auth.PreOrderForPayMembershipResp{
		//订单的总金额, 支付的时候以这个金额计算, 人民币格式，带小数点 99.00
		OrderTotalAmount: LMCommon.MEMBERSHIPPRICE,
		//服务端生成的订单id
		OrderID: sendPrePayForMembershipResp.OrderID,
		//向收款方转账的裸交易结构体
		RawDescToTarget: sendPrePayForMembershipResp.RawDescToTarget,
		//时间
		Time: sendPrePayForMembershipResp.Time,
	}, nil
}

func (s *DefaultLianmiApisService) ConfirmPayForMembership(ctx context.Context, username string, req *Auth.ConfirmPayForMembershipReq) (*Auth.ConfirmPayForMembershipResp, error) {

	//调用钱包的GrpcServer接口，进行类似 10-4 的确认交易
	resp, err := s.walletGrpcClientSvc.SendConfirmPayForMembership(ctx, &Wallet.SendConfirmPayForMembershipReq{
		Username: username,
		//订单ID（ 非空的时候，targetUserName 必须是空
		OrderID: req.OrderID,
		//签名后的转给目标接收者的Tx(A签) hex格式
		SignedTxToTarget: req.SignedTxToTarget,
		//附言
		Content: req.Content,
	})
	if err != nil {
		s.logger.Error("walletGrpcClientSvc.SendConfirmPayForMembership 错误", zap.Error(err))
		return nil, err
	}


	s.Repository.SaveToCommission(username, req.OrderID,req.Content,resp.BlockNumber,resp.Hash)

	return &Auth.ConfirmPayForMembershipResp{
		//要给谁付费
		PayForUsername: resp.PayForUsername,
		//订单的总金额, 支付的时候以这个金额计算, 人民币格式，带小数点 99.00
		OrderTotalAmount: LMCommon.MEMBERSHIPPRICE,
		// 区块高度
		BlockNumber: resp.BlockNumber,
		// 交易哈希hex
		Hash: resp.Hash,
		//交易时间
		Time: resp.Time,
	}, nil
}
