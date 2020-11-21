package services

import (
	"context"
	Auth "github.com/lianmi/servers/api/proto/auth"
	Order "github.com/lianmi/servers/api/proto/order"
	User "github.com/lianmi/servers/api/proto/user"
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
	DisBlockUser(username string) error
	Register(user *models.User) (string, error)

	ResetPassword(mobile, password string, user *models.User) error
	GetUserRoles(username string) []*models.Role
	GetUser(username string) (*Auth.UserRsp, error)

	//多条件不定参数批量分页获取用户列表
	QueryUsers(req *User.QueryUsersReq) (*User.QueryUsersResp, error)

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

	//根据手机号返回注册账号
	GetUsernameByMobile(mobile string) (string, error)

	//检测校验码是否正确
	CheckSmsCode(mobile, smscode string) bool

	//授权新创建的群组
	ApproveTeam(teamID string) error

	//封禁群组
	BlockTeam(teamID string) error

	//解封群组
	DisBlockTeam(teamID string) error

	//保存禁言的值，用于设置群禁言或解禁
	SaveTeamMute(teamID string, muteType int) error

	//======后台相关======/
	AddGeneralProduct(generalProduct *models.GeneralProduct) error

	GetGeneralProductByID(productID string) (p *models.GeneralProduct, err error)

	GetGeneralProductPage(pageIndex, pageSize int, total *int64, where interface{}) ([]*models.GeneralProduct, error)

	UpdateGeneralProduct(generalProduct *models.GeneralProduct) error

	DeleteGeneralProduct(productID string) bool

	QueryCustomerServices(req *Auth.QueryCustomerServiceReq) ([]*models.CustomerServiceInfo, error)

	AddCustomerService(req *Auth.AddCustomerServiceReq) error

	DeleteCustomerService(req *Auth.DeleteCustomerServiceReq) bool

	//修改在线客服资料
	UpdateCustomerService(req *Auth.UpdateCustomerServiceReq) error

	QueryGrades(req *Auth.GradeReq, pageIndex int, pageSize int, total *int64, where interface{}) ([]*models.Grade, error)

	AddGrade(req *Auth.AddGradeReq) (string, error)

	SubmitGrade(req *Auth.SubmitGradeReq) error

	//商户查询当前名下用户总数，按月统计付费会员总数及返佣金额，是否已经返佣
	GetBusinessMembership(businessUsername string) (*Auth.GetBusinessMembershipResp, error)

	//普通用户佣金返佣统计
	GetNormalMembership(username string) (*Auth.GetMembershipResp, error)

	PreOrderForPayMembership(ctx context.Context, username, deviceID, payForUsername string) (*Auth.PreOrderForPayMembershipResp, error)

	ConfirmPayForMembership(ctx context.Context, username string, req *Auth.ConfirmPayForMembershipReq) (*Auth.ConfirmPayForMembershipResp, error)

	UpdateUser(username string, user *models.User) error

	SaveTag(tag *models.Tag) error

	//提交佣金提现申请(商户，用户)
	SubmitCommssionWithdraw(username, yearMonth string) (*Auth.CommssionWithdrawResp, error)

	// 增加群成员资料
	AddTeamUser(pTeamUser *models.TeamUser) error

	// 修改群成员资料
	SaveTeamUser(pTeamUser *models.TeamUser) error

	//解除禁言
	SetMuteTeamUser(teamID, dissMuteUser string, isMute bool, mutedays int) error

	GetTeams() []string

	DeleteTeamUser(teamID, username string) error

	//创建群
	CreateTeam(pTeam *models.Team) error

	// 更新群数据
	UpdateTeam(teamID string, pTeam *models.Team) error

	GetTeamUsers(teamID string, PageNum int, PageSize int, total *int64, where interface{}) []*models.TeamUser

	//添加好友
	AddFriend(pFriend *models.Friend) error

	//修改好友资料
	UpdateFriend(pFriend *models.Friend) error

	DeleteFriend(userID, friendUserID uint64) error

	// 增加或修改店铺资料
	SaveStore(req *User.Store) error

	GetStore(businessUsername string) (*User.Store, error)

	GetStores(req *User.QueryStoresNearbyReq) (*User.QueryStoresNearbyResp, error)

	AuditStore(req *Auth.AuditStoreReq) error
}

type DefaultLianmiApisService struct {
	logger              *zap.Logger
	Repository          repositories.LianmiRepository
	orderGrpcClientSvc  Order.LianmiOrderClient   //order的grpc client
	walletGrpcClientSvc Wallet.LianmiWalletClient //wallet的grpc client
}

func NewLianmiApisService(logger *zap.Logger, repository repositories.LianmiRepository, oc Order.LianmiOrderClient, wc Wallet.LianmiWalletClient) LianmiApisService {
	return &DefaultLianmiApisService{
		logger:              logger.With(zap.String("type", "DefaultLianmiApisService")),
		Repository:          repository,
		orderGrpcClientSvc:  oc,
		walletGrpcClientSvc: wc,
	}
}

func (s *DefaultLianmiApisService) GetUser(username string) (*Auth.UserRsp, error) {
	s.logger.Debug("GetUser", zap.String("username", username))

	fUserData, err := s.Repository.GetUser(username)
	if err != nil {
		return nil, err
	}
	return &Auth.UserRsp{
		User: &User.User{
			Username:         fUserData.Username,
			Gender:           User.Gender(fUserData.Gender),
			Nick:             fUserData.Nick,
			Avatar:           fUserData.Avatar,
			Label:            fUserData.Label,
			Mobile:           fUserData.Mobile,
			Email:            fUserData.Email,
			UserType:         User.UserType(fUserData.UserType),
			State:            User.UserState(fUserData.State),
			Extend:           fUserData.Extend,
			ContactPerson:    fUserData.ContactPerson,
			ReferrerUsername: fUserData.ReferrerUsername,
		},
	}, nil
}

//多条件不定参数批量分页获取用户列表
func (s *DefaultLianmiApisService) QueryUsers(req *User.QueryUsersReq) (*User.QueryUsersResp, error) {
	if users, total, err := s.Repository.QueryUsers(req); err != nil {
		return nil, err
	} else {
		resp := &User.QueryUsersResp{
			Total: uint64(total),
		}

		for _, userData := range users {
			resp.Users = append(resp.Users, &User.User{
				Username:      userData.Username,
				Gender:        User.Gender(userData.Gender),
				Nick:          userData.Nick,
				Avatar:        userData.Avatar,
				Label:         userData.Label,
				Mobile:        userData.Mobile,
				Email:         userData.Email,
				UserType:      User.UserType(userData.UserType),
				Extend:        userData.Extend,
				ContactPerson: userData.ContactPerson,
			})

		}

		return resp, nil
	}
}

func (s *DefaultLianmiApisService) BlockUser(username string) (err error) {

	return s.Repository.BlockUser(username)

}
func (s *DefaultLianmiApisService) DisBlockUser(username string) error {

	return s.Repository.DisBlockUser(username)
}

//生成短信校验码
func (s *DefaultLianmiApisService) GenerateSmsCode(mobile string) bool {
	return s.Repository.GenerateSmsCode(mobile)
}

//根据手机号获取注册账号id
func (s *DefaultLianmiApisService) GetUsernameByMobile(mobile string) (string, error) {
	return s.Repository.GetUsernameByMobile(mobile)
}

//检测校验码是否正确
func (s *DefaultLianmiApisService) CheckSmsCode(mobile, smscode string) bool {
	return s.Repository.CheckSmsCode(mobile, smscode)
}

func (s *DefaultLianmiApisService) Register(user *models.User) (string, error) {
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
	if err := s.Repository.ResetPassword(mobile, password, user); err != nil {
		return errors.Wrap(err, "ResetPassword error")
	}

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
	return s.Repository.ApproveTeam(teamID)
}

//封禁群组
func (s *DefaultLianmiApisService) BlockTeam(teamID string) error {
	return s.Repository.BlockTeam(teamID)

}

//解封群组
func (s *DefaultLianmiApisService) DisBlockTeam(teamID string) error {
	return s.Repository.DisBlockTeam(teamID)

}

//保存禁言的值，用于设置群禁言或解禁
func (s *DefaultLianmiApisService) SaveTeamMute(teamID string, muteType int) error {
	return s.Repository.SaveTeamMute(teamID, muteType)

}

//======后台相关======/
func (s *DefaultLianmiApisService) AddGeneralProduct(generalProduct *models.GeneralProduct) error {
	return s.Repository.AddGeneralProduct(generalProduct)

}

//查询通用商品(id) - Read
func (s *DefaultLianmiApisService) GetGeneralProductByID(productID string) (p *models.GeneralProduct, err error) {

	return s.Repository.GetGeneralProductByID(productID)

}

//查询通用商品分页 - Page
func (s *DefaultLianmiApisService) GetGeneralProductPage(pageIndex, pageSize int, total *int64, where interface{}) ([]*models.GeneralProduct, error) {

	return s.Repository.GetGeneralProductPage(pageIndex, pageSize, total, where)

}

// 修改通用商品 - Update
func (s *DefaultLianmiApisService) UpdateGeneralProduct(generalProduct *models.GeneralProduct) error {

	return s.Repository.UpdateGeneralProduct(generalProduct)

}

// 删除通用商品 - Delete
func (s *DefaultLianmiApisService) DeleteGeneralProduct(productID string) bool {

	return s.Repository.DeleteGeneralProduct(productID)

}

//获取在线客服id数组
func (s *DefaultLianmiApisService) QueryCustomerServices(req *Auth.QueryCustomerServiceReq) ([]*models.CustomerServiceInfo, error) {
	return s.Repository.QueryCustomerServices(req)
}

//增加在线客服id
func (s *DefaultLianmiApisService) AddCustomerService(req *Auth.AddCustomerServiceReq) error {
	return s.Repository.AddCustomerService(req)
}

func (s *DefaultLianmiApisService) DeleteCustomerService(req *Auth.DeleteCustomerServiceReq) bool {
	return s.Repository.DeleteCustomerService(req)
}

//修改在线客服资料
func (s *DefaultLianmiApisService) UpdateCustomerService(req *Auth.UpdateCustomerServiceReq) error {
	return s.Repository.UpdateCustomerService(req)
}

func (s *DefaultLianmiApisService) QueryGrades(req *Auth.GradeReq, pageIndex int, pageSize int, total *int64, where interface{}) ([]*models.Grade, error) {
	return s.Repository.QueryGrades(req, pageIndex, pageSize, total, where)
}

func (s *DefaultLianmiApisService) AddGrade(req *Auth.AddGradeReq) (string, error) {
	return s.Repository.AddGrade(req)
}

func (s *DefaultLianmiApisService) SubmitGrade(req *Auth.SubmitGradeReq) error {
	return s.Repository.SubmitGrade(req)
}

//商户查询当前名下用户总数，按月统计付费会员总数及返佣金额，是否已经返佣
func (s *DefaultLianmiApisService) GetBusinessMembership(businessUsername string) (*Auth.GetBusinessMembershipResp, error) {
	return s.Repository.GetBusinessMembership(businessUsername)
}

//普通用户佣金返佣统计
func (s *DefaultLianmiApisService) GetNormalMembership(username string) (*Auth.GetMembershipResp, error) {
	return s.Repository.GetNormalMembership(username)
}

//预生成一个购买会员的订单， 返回OrderID及预转账裸交易数据
//payForUsername  - 要给谁付费
func (s *DefaultLianmiApisService) PreOrderForPayMembership(ctx context.Context, username, deviceID, payForUsername string) (*Auth.PreOrderForPayMembershipResp, error) {

	//查询当前用户是否已经付费
	userInfo, err := s.Repository.GetUser(username)
	if err != nil {
		s.logger.Error("s.Repository.GetUser 错误", zap.Error(err))
		return nil, err
	}
	//用户不是付费会员
	if userInfo.State == 0 || userInfo.State == 2 {
		return nil, errors.Wrap(err, "用户不是付费会员")
	}

	//查询payForUsername是否已经付费
	payForUsernameInfo, err := s.Repository.GetUser(payForUsername)
	if err != nil {
		s.logger.Error("s.Repository.GetUser 错误", zap.Error(err))
		return nil, err
	}

	//接受支付的用户已经是付费会员， 不能重复支付
	if payForUsernameInfo.State == 1 {
		return nil, errors.Wrapf(err, "接受支付的用户已经是付费会员[%s]", payForUsername)
	}

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

	//确认支付成功后，就需要分配佣金
	s.Repository.SaveToCommission(username, req.OrderID, req.Content, resp.BlockNumber, resp.Hash)

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

//提交佣金提现申请(商户，用户)
func (s *DefaultLianmiApisService) SubmitCommssionWithdraw(username, yearMonth string) (*Auth.CommssionWithdrawResp, error) {
	return s.Repository.SubmitCommssionWithdraw(username, yearMonth)
}

// 修改群成员资料
func (s *DefaultLianmiApisService) SaveTeamUser(pTeamUser *models.TeamUser) error {
	return s.Repository.SaveTeamUser(pTeamUser)
}

// 增加群成员资料
func (s *DefaultLianmiApisService) AddTeamUser(pTeamUser *models.TeamUser) error {
	return s.Repository.AddTeamUser(pTeamUser)
}

//解除群成员的禁言
func (s *DefaultLianmiApisService) SetMuteTeamUser(teamID, dissMuteUser string, isMute bool, mutedays int) error {
	return s.Repository.SetMuteTeamUser(teamID, dissMuteUser, isMute, mutedays)
}

func (s *DefaultLianmiApisService) AddFriend(pFriend *models.Friend) error {
	return s.Repository.AddFriend(pFriend)
}

func (s *DefaultLianmiApisService) UpdateFriend(pFriend *models.Friend) error {
	return s.Repository.UpdateFriend(pFriend)
}

func (s *DefaultLianmiApisService) DeleteFriend(userID, friendUserID uint64) error {
	return s.Repository.DeleteFriend(userID, friendUserID)
}

func (s *DefaultLianmiApisService) GetTeams() []string {
	return s.Repository.GetTeams()
}

//创建群
func (s *DefaultLianmiApisService) CreateTeam(pTeam *models.Team) error {
	return s.Repository.CreateTeam(pTeam)
}

// 更新群数据
func (s *DefaultLianmiApisService) UpdateTeam(teamID string, pTeam *models.Team) error {
	return s.Repository.UpdateTeam(teamID, pTeam)
}

func (s *DefaultLianmiApisService) DeleteTeamUser(teamID, username string) error {
	return s.Repository.DeleteTeamUser(teamID, username)
}

func (s *DefaultLianmiApisService) GetTeamUsers(teamID string, PageNum int, PageSize int, total *int64, where interface{}) []*models.TeamUser {
	return s.Repository.GetTeamUsers(teamID, PageNum, PageSize, total, where)
}

func (s *DefaultLianmiApisService) UpdateUser(username string, user *models.User) error {
	return s.Repository.UpdateUser(username, user)
}

func (s *DefaultLianmiApisService) SaveTag(tag *models.Tag) error {
	return s.Repository.SaveTag(tag)
}

//修改或增加店铺资料
func (s *DefaultLianmiApisService) SaveStore(req *User.Store) error {
	return s.Repository.SaveStore(req)
}

func (s *DefaultLianmiApisService) GetStore(businessUsername string) (*User.Store, error) {
	return s.Repository.GetStore(businessUsername)
}

func (s *DefaultLianmiApisService) GetStores(req *User.QueryStoresNearbyReq) (*User.QueryStoresNearbyResp, error) {
	return s.Repository.GetStores(req)
}

func (s *DefaultLianmiApisService) AuditStore(req *Auth.AuditStoreReq) error {
	return s.Repository.AuditStore(req)
}
