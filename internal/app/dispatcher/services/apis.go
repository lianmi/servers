package services

import (
	"context"
	// "fmt"
	"strings"
	"time"

	Auth "github.com/lianmi/servers/api/proto/auth"
	Order "github.com/lianmi/servers/api/proto/order"
	User "github.com/lianmi/servers/api/proto/user"

	// Wallet "github.com/lianmi/servers/api/proto/wallet"
	"github.com/lianmi/servers/internal/app/dispatcher/repositories"
	"github.com/lianmi/servers/internal/pkg/models"
	"github.com/pkg/errors"
	"go.uber.org/zap"

	pb "github.com/lianmi/servers/api/proto/user"
	LMCommon "github.com/lianmi/servers/internal/common"
)

type LianmiApisService interface {
	GetAppVersion(oldVersion string) (string, error)

	BlockUser(username string) error
	DisBlockUser(username string) error
	Register(user *models.User) (string, error)

	ResetPassword(mobile, password string, user *models.User) error
	GetUserRoles(username string) []*models.Role
	GetUser(username string) (*Auth.UserRsp, error)
	GetUserDb(objname string) (string, error)
	// 微信登录之后绑定手机
	UserBindmobile(username, mobile string) error

	GetIsBindWechat(username string) (bool, error)

	// 手机登录之后绑定微信
	UserBindWechat(username, openId string) error

	//保存消息推送设置
	SavePushSetting(username string, newRemindSwitch, detailSwitch, teamSwitch, soundSwitch bool) error

	//查询用户消息设置
	GetPushSetting(username string) (*models.PushSetting, error)

	UnBindmobile(username string) error

	GetSystemMsgs(systemMsgAt uint64) (systemMsgs []*models.SystemMsg, err error)

	//多条件不定参数批量分页获取用户列表
	QueryUsers(req *User.QueryUsersReq) (*User.QueryUsersResp, error)

	QueryAllUsernames() ([]string, error)

	//检测用户登录
	CheckUser(isMaster bool, username, password, deviceID, os string, userType int) (bool, string)

	//  使用手机及短信验证码登录
	LoginBySmscode(username, mobile, smscode, deviceID, os string, userType int) (bool, string)

	// 判断用户名是否已存在
	ExistUserByName(username string) bool
	// 判断手机号码是否已存在
	ExistUserByMobile(mobile string) bool

	SaveUserToken(username, deviceID string, token string, expire time.Time) bool

	//获取当前用户的主设备
	GetAllDevices(username string) (string, error)

	SignOut(token, username, deviceID string) bool
	ExistsTokenInRedis(deviceID, token string) bool

	//生成注册校验码
	GenerateSmsCode(mobile string) bool

	//根据手机号返回注册账号
	GetUsernameByMobile(mobile string) (string, error)

	//根据注册账号返回手机号
	GetMobileByUsername(username string) (string, error)

	//检测校验码是否正确
	CheckSmsCode(mobile, smscode string) bool

	//授权新创建的群组
	ApproveTeam(teamID string) error

	//封禁群组
	BlockTeam(teamID string) error

	//解封群组
	DisBlockTeam(teamID string) error

	//保存禁言的值，用于设置群禁言或解禁
	UpdateTeamMute(teamID string, muteType int) error

	GetProductInfo(product string) (*Order.Product, error)

	AddGeneralProduct(generalProductInfo *models.GeneralProductInfo) error

	GetGeneralProductByID(productID string) (p *models.GeneralProduct, err error)

	GetGeneralProductPage(req *Order.GetGeneralProductPageReq) (*Order.GetGeneralProductPageResp, error)

	UpdateGeneralProduct(generalProductInfo *models.GeneralProductInfo) error

	DeleteGeneralProduct(productID string) bool

	QueryCustomerServices(req *Auth.QueryCustomerServiceReq) ([]*models.CustomerServiceInfo, error)

	AddCustomerService(req *Auth.AddCustomerServiceReq) error

	DeleteCustomerService(req *Auth.DeleteCustomerServiceReq) bool

	//订单模块
	//商户端: 将完成订单拍照所有图片上链
	UploadOrderImages(ctx context.Context, req *Order.UploadOrderImagesReq) (*Order.UploadOrderImagesResp, error)

	//将订单body经过RSA加密后提交到服务端
	UploadOrderBody(ctx context.Context, req *models.UploadOrderBodyReq) error

	//用户端: 根据 OrderID 获取所有订单拍照图片
	DownloadOrderImage(orderID string) (*Order.DownloadOrderImagesResp, error)

	//修改在线客服资料
	UpdateCustomerService(req *Auth.UpdateCustomerServiceReq) error

	QueryGrades(req *Auth.GradeReq, pageIndex int, pageSize int, total *int64, where interface{}) ([]*models.Grade, error)

	AddGrade(req *Auth.AddGradeReq) (string, error)

	SubmitGrade(req *Auth.SubmitGradeReq) error

	//查询VIP会员价格表
	GetVipPriceList(payType int) (*Auth.GetVipPriceResp, error)

	//商户查询当前名下用户总数，按月统计付费会员总数及返佣金额，是否已经返佣
	GetBusinessMembership(businessUsername string) (*Auth.GetBusinessMembershipResp, error)

	//对某个用户的推广会员佣金进行统计
	CommissonSatistics(username string) (*Auth.CommissonSatisticsResp, error)

	//用户查询按月统计发展的付费会员总数及返佣金额，是否已经返佣
	GetCommissionStatistics(username string) (*Auth.GetCommssionsResp, error)

	//更新用户表
	UpdateUser(username string, user *models.User) error

	//更新商店表
	UpdateStore(username string, store *models.Store) error

	AddTag(tag *models.Tag) error

	//提交佣金提现申请(商户，用户)
	SubmitCommssionWithdraw(username, yearMonth string) (*Auth.CommssionWithdrawResp, error)

	// 增加群成员资料
	AddTeamUser(teamUserInfo *models.TeamUserInfo) error

	//设置群管理员s
	UpdateTeamUserManager(teamID, managerUsername string, isAdd bool) error

	// 修改群成员呢称、扩展
	UpdateTeamUserMyInfo(teamID, username, aliasName, ex string) error

	//修改群通知方式
	UpdateTeamUserNotifyType(teamID string, notifyType int) error

	//解除禁言
	SetMuteTeamUser(teamID, dissMuteUser string, isMute bool, mutedays int) error

	GetTeams() []string

	GetChargeProductID() (string, error)

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
	AddStore(req *models.Store) error

	GetStore(businessUsername string) (*User.Store, error)

	GetStores(req *Order.QueryStoresNearbyReq) (*Order.QueryStoresNearbyResp, error)

	//保存excel某一行的网点
	SaveExcelToDb(lotteryStore *models.LotteryStore) error

	//查询并分页获取采集的网点
	GetLotteryStores(req *models.LotteryStoreReq) ([]*models.LotteryStore, error)

	//批量增加网点
	BatchAddStores(req *models.LotteryStoreReq) error

	//批量网点opk
	AdminDefaultOPK() error

	//将店铺通过审核
	AuditStore(req *Auth.AuditStoreReq) error

	//获取某个商户的所有商品列表
	GetProductsList(req *Order.ProductsListReq) (*Order.ProductsListResp, error)

	//设置商品的子类型
	SetProductSubType(req *Order.ProductSetSubTypeReq) error

	//获取某个用户对所有店铺点赞情况, UI会保存在本地表里,  UI主动发起同步
	UserLikes(username string) (*User.UserLikesResp, error)

	//获取店铺的所有点赞的用户列表
	StoreLikes(businessUsername string) (*User.StoreLikesResp, error)

	//获取店铺的所有点赞总数
	StoreLikesCount(businessUsername string) (int, error)

	//对某个店铺点赞
	ClickLike(username, businessUsername string) (int64, error)

	//取消对某个店铺点赞
	DeleteClickLike(username, businessUsername string) (int64, error)

	//取消对某个店铺点赞
	GetIsLike(username, businessUsername string) (bool, error)

	//将点赞记录插入到UserLike表
	AddUserLike(username, businessUser string) error

	//支付宝预支付
	// PreAlipay(ctx context.Context, req *Wallet.PreAlipayReq) (*Wallet.PreAlipayResp, error)

	//支付宝付款成功
	// AlipayDone(ctx context.Context, outTradeNo string) error

	//微信预支付
	// PreWXpay(ctx context.Context, req *Wallet.PreWXpayReq) (*Wallet.PreWXpayResp, error)

	//设置当前商户默认OPK
	SetDefaultOPK(username, opk string) error
	// 获取指定商户的商品列表
	GetStoreProductLists(o *Order.ProductsListReq) (*[]models.StoreProductItems, error)
	// 向商户添加商品信息
	AddStoreProductItem(item *models.StoreProductItems) error
	// 从数据库获取通用商品信息
	GetGeneralProductFromDB(req *Order.GetGeneralProductPageReq) (*[]models.GeneralProduct, error)
	// 保存订单项目到数据库
	SavaOrderItemToDB(item *models.OrderItems) error
	// 通过用户id 获取他的订单列表
	GetOrderListByUser(username string, limit int, offset, status int) (*[]models.OrderItems, error)
	// 通过订单id 获取订单信息
	GetOrderListByID(orderID string) (*models.OrderItems, error)
	// 通过订单id 修改订单状态到指定的状态
	SetOrderStatusByOrderID(orderID string, status int) error
	// 更新特定的订单状态 , 并返回这个最新的订单状态信息
	UpdateOrderStatus(userid string, storeID string, orderID string, status int) (*models.OrderItems, error)
	// 通过用户名获取用户的类型
	GetUserType(username string) (int, error)
	// 在微信回调中修改订单状态 , 这里只能修改成 支付完成
	UpdateOrderStatusByWechatCallback(orderid string) error
	// 获取指定商户id 的opk 公钥
	GetStoreOpkByBusiness(businessId string) (string, error)

	//从redis里获取订单当前最新的数据及状态
	GetOrderInfo(orderID string) (*models.OrderInfo, error)
	// 处理订单兑奖业务 username 当前token 的用户 orderID 处理的订单 . prize 兑奖的金额 , 并返回购买的用户id
	OrderPushPrize(username string, orderID string, prize float64, prizedPhoto string) (string, error)
	// 通过用户名和订单id 删除这条订单
	OrderDeleteByUserAndOrderid(username string, orderid string) error
	// 删除该用户名的所有订单
	DeleteUserOrdersByUserID(username string) error
	// 通过关键字查询订单
	OrderSerachByKeyWord(username string, req *models.ReqKeyWordDataType) (*[]models.OrderItems, error)

	//管理员修改app版本号
	ManagerSetVersionLast(req *models.VersionInfo) error

	// 通过 微信 open id 获取userid
	GetUserByWechatOpenid(openid string) (string, error)
	// 绑定微信openid
	UpdateUserWxOpenID(username string, openid string) error

	ManagerAddSystemMsg(level int, title, content string) error

	ManagerDeleteSystemMsg(id uint) error
}

type DefaultLianmiApisService struct {
	logger             *zap.Logger
	Repository         repositories.LianmiRepository
	orderGrpcClientSvc Order.LianmiOrderClient //order的grpc client
	// walletGrpcClientSvc Wallet.LianmiWalletClient //wallet的grpc client
}

func NewLianmiApisService(logger *zap.Logger, repository repositories.LianmiRepository, oc Order.LianmiOrderClient) LianmiApisService {
	return &DefaultLianmiApisService{
		logger:             logger.With(zap.String("type", "DefaultLianmiApisService")),
		Repository:         repository,
		orderGrpcClientSvc: oc,
		// walletGrpcClientSvc: wc,
	}
}

func (s *DefaultLianmiApisService) GetUser(username string) (*Auth.UserRsp, error) {
	s.logger.Debug("GetUser", zap.String("username", username))
	var avatar string

	fUserData, err := s.Repository.GetUser(username)
	if err != nil {
		return nil, err
	}
	s.logger.Debug("GetUser Success",
		zap.Int("Gender", fUserData.Gender),
		zap.String("Nick", fUserData.Nick),
		zap.String("Avatar", fUserData.Avatar),
		zap.String("Label", fUserData.Label),
		zap.String("Mobile", fUserData.Mobile),
		zap.String("Email", fUserData.Email),
		zap.String("Extend", fUserData.Extend),
		zap.Int("AllowType", fUserData.AllowType),
		zap.Int("UserType", fUserData.UserType),
		zap.Int("State", fUserData.State),
	)
	if fUserData.Avatar != "" {
		if strings.HasPrefix(fUserData.Avatar, "https") {
			avatar = fUserData.Avatar + "?x-oss-process=image/resize,w_50/quality,q_50"
		} else {

			avatar = LMCommon.OSSUploadPicPrefix + fUserData.Avatar + "?x-oss-process=image/resize,w_50/quality,q_50"
		}

	}

	return &Auth.UserRsp{
		User: &User.User{
			Username:     fUserData.Username,
			Gender:       User.Gender(fUserData.Gender),
			Nick:         fUserData.Nick,
			Avatar:       avatar,
			Label:        fUserData.Label,
			Mobile:       fUserData.Mobile,
			Email:        fUserData.Email,
			Extend:       fUserData.Extend,
			AllowType:    pb.AllowType(fUserData.AllowType),
			UserType:     User.UserType(fUserData.UserType),
			State:        User.UserState(fUserData.State),
			TrueName:     fUserData.TrueName,
			IdentityCard: fUserData.IdentityCard,
			Province:     fUserData.Province,
			City:         fUserData.City,
			Area:         fUserData.Area,
			// Street:             fUserData.Street,
			Address:            fUserData.Address,
			ReferrerUsername:   fUserData.ReferrerUsername,
			BelongBusinessUser: fUserData.BelongBusinessUser,
			VipEndDate:         uint64(fUserData.VipEndDate),
			CreatedAt:          uint64(fUserData.CreatedAt),
			UpdatedAt:          uint64(fUserData.UpdatedAt),
		},
	}, nil
}

func (s *DefaultLianmiApisService) GetAppVersion(oldVersion string) (string, error) {
	return s.Repository.GetAppVersion(oldVersion)
}

func (s *DefaultLianmiApisService) GetUserDb(objname string) (string, error) {
	return s.Repository.GetUserDb(objname)

}

//微信登录之后绑定手机
func (s *DefaultLianmiApisService) UserBindmobile(username, mobile string) error {
	return s.Repository.UserBindmobile(username, mobile)
}

func (s *DefaultLianmiApisService) GetIsBindWechat(username string) (bool, error) {
	return s.Repository.GetIsBindWechat(username)
}

func (s *DefaultLianmiApisService) UserBindWechat(username, openId string) error {
	return s.Repository.UserBindWechat(username, openId)
}

func (s *DefaultLianmiApisService) SavePushSetting(username string, newRemindSwitch, detailSwitch, teamSwitch, soundSwitch bool) error {
	return s.Repository.SavePushSetting(username, newRemindSwitch, detailSwitch, teamSwitch, soundSwitch)
}

//查询用户消息设置
func (s *DefaultLianmiApisService) GetPushSetting(username string) (*models.PushSetting, error) {
	return s.Repository.GetPushSetting(username)
}

func (s *DefaultLianmiApisService) UnBindmobile(username string) error {
	return s.Repository.UnBindmobile(username)
}

func (s *DefaultLianmiApisService) GetSystemMsgs(systemMsgAt uint64) (systemMsgs []*models.SystemMsg, err error) {
	return s.Repository.GetSystemMsgs(systemMsgAt)
}

//多条件不定参数批量分页获取用户列表
func (s *DefaultLianmiApisService) QueryUsers(req *User.QueryUsersReq) (*User.QueryUsersResp, error) {
	return s.Repository.QueryUsers(req)
}

func (s *DefaultLianmiApisService) QueryAllUsernames() ([]string, error) {
	return s.Repository.QueryAllUsernames()
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

//根据注册账号返回手机号
func (s *DefaultLianmiApisService) GetMobileByUsername(username string) (string, error) {
	return s.Repository.GetMobileByUsername(username)
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
func (s *DefaultLianmiApisService) CheckUser(isMaster bool, username, password, deviceID, os string, userType int) (bool, string) {

	return s.Repository.CheckUser(isMaster, username, password, deviceID, os, userType)
}

//  使用手机及短信验证码登录
func (s *DefaultLianmiApisService) LoginBySmscode(username, mobile, smscode, deviceID, os string, userType int) (bool, string) {

	return s.Repository.LoginBySmscode(username, mobile, smscode, deviceID, os, userType)
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

func (s *DefaultLianmiApisService) GetAllDevices(username string) (string, error) {
	return s.Repository.GetDeviceFromRedis(username)
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
func (s *DefaultLianmiApisService) UpdateTeamMute(teamID string, muteType int) error {
	return s.Repository.UpdateTeamMute(teamID, muteType)

}

//======后台相关======/
func (s *DefaultLianmiApisService) AddGeneralProduct(generalProductInfo *models.GeneralProductInfo) error {
	return s.Repository.AddGeneralProduct(generalProductInfo)

}

func (s *DefaultLianmiApisService) GetProductInfo(product string) (*Order.Product, error) {
	return s.Repository.GetProductInfo(product)
}

//查询通用商品(id) - Read
func (s *DefaultLianmiApisService) GetGeneralProductByID(productID string) (p *models.GeneralProduct, err error) {

	return s.Repository.GetGeneralProductByID(productID)

}

//查询通用商品分页 - Page
func (s *DefaultLianmiApisService) GetGeneralProductPage(req *Order.GetGeneralProductPageReq) (*Order.GetGeneralProductPageResp, error) {

	return s.Repository.GetGeneralProductPage(req)

}

// 修改通用商品 - Update
func (s *DefaultLianmiApisService) UpdateGeneralProduct(generalProductInfo *models.GeneralProductInfo) error {

	return s.Repository.UpdateGeneralProduct(generalProductInfo)

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

//对某个用户的推广会员佣金进行统计
func (s *DefaultLianmiApisService) CommissonSatistics(username string) (*Auth.CommissonSatisticsResp, error) {
	return s.Repository.CommissonSatistics(username)
}

//用户查询按月统计发展的付费会员总数及返佣金额，是否已经返佣
func (s *DefaultLianmiApisService) GetCommissionStatistics(username string) (*Auth.GetCommssionsResp, error) {
	return s.Repository.GetCommissionStatistics(username)
}

//提交佣金提现申请(商户，用户)
func (s *DefaultLianmiApisService) SubmitCommssionWithdraw(username, yearMonth string) (*Auth.CommssionWithdrawResp, error) {
	return s.Repository.SubmitCommssionWithdraw(username, yearMonth)
}

// 修改群成员资料
func (s *DefaultLianmiApisService) UpdateTeamUserManager(teamID, managerUsername string, isAdd bool) error {
	return s.Repository.UpdateTeamUserManager(teamID, managerUsername, isAdd)
}

// 修改群成员呢称、扩展
func (s *DefaultLianmiApisService) UpdateTeamUserMyInfo(teamID, username, aliasName, ex string) error {
	return s.Repository.UpdateTeamUserMyInfo(teamID, username, aliasName, ex)
}

//修改群通知方式
func (s *DefaultLianmiApisService) UpdateTeamUserNotifyType(teamID string, notifyType int) error {
	return s.Repository.UpdateTeamUserNotifyType(teamID, notifyType)
}

// 增加群成员资料
func (s *DefaultLianmiApisService) AddTeamUser(teamUserInfo *models.TeamUserInfo) error {
	return s.Repository.AddTeamUser(teamUserInfo)
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

func (s *DefaultLianmiApisService) GetChargeProductID() (string, error) {
	return s.Repository.GetChargeProductID()
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

//更新商店表
func (s *DefaultLianmiApisService) UpdateStore(username string, store *models.Store) error {
	return s.Repository.UpdateStore(username, store)
}

func (s *DefaultLianmiApisService) AddTag(tag *models.Tag) error {
	return s.Repository.AddTag(tag)
}

//修改或增加店铺资料
func (s *DefaultLianmiApisService) AddStore(req *models.Store) error {
	return s.Repository.AddStore(req)
}

func (s *DefaultLianmiApisService) GetStore(businessUsername string) (*User.Store, error) {
	return s.Repository.GetStore(businessUsername)
}

func (s *DefaultLianmiApisService) GetStores(req *Order.QueryStoresNearbyReq) (*Order.QueryStoresNearbyResp, error) {
	return s.Repository.GetStores(req)
}

func (s *DefaultLianmiApisService) AuditStore(req *Auth.AuditStoreReq) error {
	return s.Repository.AuditStore(req)
}

//保存excel某一行的网点
func (s *DefaultLianmiApisService) SaveExcelToDb(lotteryStore *models.LotteryStore) error {
	return s.Repository.SaveExcelToDb(lotteryStore)
}

//查询并分页获取采集的网点
func (s *DefaultLianmiApisService) GetLotteryStores(req *models.LotteryStoreReq) ([]*models.LotteryStore, error) {
	return s.Repository.GetLotteryStores(req)
}

//批量增加网点
func (s *DefaultLianmiApisService) BatchAddStores(req *models.LotteryStoreReq) error {
	return s.Repository.BatchAddStores(req)
}

//批量网点opk
func (s *DefaultLianmiApisService) AdminDefaultOPK() error {
	return s.Repository.AdminDefaultOPK()
}

//获取某个商户的所有商品列表
func (s *DefaultLianmiApisService) GetProductsList(req *Order.ProductsListReq) (*Order.ProductsListResp, error) {
	return s.Repository.GetProductsList(req)
}

//设置商品的子类型
func (s *DefaultLianmiApisService) SetProductSubType(req *Order.ProductSetSubTypeReq) error {
	return s.Repository.SetProductSubType(req)
}

//获取某个用户对所有店铺点赞情况, UI会保存在本地表里,  UI主动发起同步
func (s *DefaultLianmiApisService) UserLikes(username string) (*User.UserLikesResp, error) {
	return s.Repository.UserLikes(username)
}

//获取店铺的所有点赞的用户列表
func (s *DefaultLianmiApisService) StoreLikes(businessUsername string) (*User.StoreLikesResp, error) {
	return s.Repository.StoreLikes(businessUsername)
}

//获取店铺的所有点赞总数
func (s *DefaultLianmiApisService) StoreLikesCount(businessUsername string) (int, error) {
	return s.Repository.StoreLikesCount(businessUsername)
}

//对某个店铺点赞
func (s *DefaultLianmiApisService) ClickLike(username, businessUsername string) (int64, error) {
	return s.Repository.ClickLike(username, businessUsername)
}

//取消对某个店铺点赞
func (s *DefaultLianmiApisService) DeleteClickLike(username, businessUsername string) (int64, error) {
	return s.Repository.DeleteClickLike(username, businessUsername)
}

//取消对某个店铺点赞
func (s *DefaultLianmiApisService) GetIsLike(username, businessUsername string) (bool, error) {
	return s.Repository.GetIsLike(username, businessUsername)
}

//将点赞记录插入到UserLike表
func (s *DefaultLianmiApisService) AddUserLike(username, businessUser string) error {
	return s.Repository.AddUserLike(username, businessUser)
}

//商户端: 将完成订单拍照所有图片上链
func (s *DefaultLianmiApisService) UploadOrderImages(ctx context.Context, req *Order.UploadOrderImagesReq) (*Order.UploadOrderImagesResp, error) {

	orderInfo, err := s.Repository.GetOrderInfo(req.OrderID)
	if err != nil {
		s.logger.Error("从Redis里取出此Order数据 Error")
	}

	if orderInfo.ProductID == "" {

		s.logger.Error("ProductID is empty")

		return nil, errors.Wrapf(err, "ProductID is empty[OrderID=%s]", req.OrderID)
	}

	if orderInfo.BuyerUsername == "" {
		s.logger.Error("BuyeerUsername is empty")
		return nil, errors.Wrapf(err, "BuyerUsername is empty[OrderID=%s]", req.OrderID)
	}

	if orderInfo.BusinessUsername == "" {
		s.logger.Error("BusinessUsername is empty")
		return nil, errors.Wrapf(err, "BusinessUsername is empty[OrderID=%s]", req.OrderID)
	}

	s.logger.Debug("UploadOrderImages",
		zap.Int("State", orderInfo.State), //状态
		zap.String("OrderID", req.OrderID),
		zap.String("ProductID", orderInfo.ProductID),
		zap.String("BuyUser", orderInfo.BuyerUsername),
		zap.String("BusinessUser", orderInfo.BusinessUsername),
		zap.String("AttachHash", orderInfo.AttachHash), //订单内容hash
		zap.Float64("OrderTotalAmount", orderInfo.Cost),
		zap.String("OrderImageFile", req.Image),
	)

	//增加订单拍照图片上链历史表
	err = s.Repository.SaveOrderImagesBlockchain(
		req,
		orderInfo.Cost,
		0,
		orderInfo.BuyerUsername,
		orderInfo.BusinessUsername,
		"")

	if err != nil {
		return nil, err
	}

	resp := &Order.UploadOrderImagesResp{
		OrderID: req.OrderID,
		// 区块高度
		BlockNumber: 0,
		// 交易哈希hex
		Hash: "000000",
		//时间
		Time: uint64(time.Now().UnixNano() / 1e6),
	}

	return resp, nil
}

//买家将订单body经过RSA加密后提交到彩票中心或第三方公证, mqtt客户端来接收
func (s *DefaultLianmiApisService) UploadOrderBody(ctx context.Context, req *models.UploadOrderBodyReq) error {

	//根据订单ID获取详细信息
	orderInfo, err := s.Repository.GetOrderInfo(req.OrderID)
	if err != nil {
		s.logger.Error("从Redis里取出此Order数据 Error")
	}

	if orderInfo.ProductID == "" {
		s.logger.Error("ProductID is empty")

		return errors.Wrapf(err, "ProductID is empty[OrderID=%s]", req.OrderID)
	}

	if orderInfo.BuyerUsername == "" {
		s.logger.Error("BuyeerUsername is empty")
		return errors.Wrapf(err, "BuyerUsername is empty[OrderID=%s]", req.OrderID)
	}

	if orderInfo.BusinessUsername == "" {
		s.logger.Error("BusinessUsername is empty")
		return errors.Wrapf(err, "BusinessUsername is empty[OrderID=%s]", req.OrderID)
	}

	s.logger.Debug("UploadOrderBody",
		zap.Int("State", orderInfo.State), //状态
		zap.String("OrderID", req.OrderID),
		zap.String("ProductID", orderInfo.ProductID),
		zap.String("BuyUser", orderInfo.BuyerUsername),
		zap.String("BusinessUser", orderInfo.BusinessUsername),
		zap.String("AttachHash", orderInfo.AttachHash), //订单内容hash
		zap.Float64("OrderTotalAmount", orderInfo.Cost),
		zap.String("BodyObjFile", req.BodyObjFile),
	)

	//将数据保存到MySQL
	err = s.Repository.SaveOrderBody(req)

	if err != nil {
		return err
	}

	return nil
}

//用户端: 根据 OrderID 获取所有订单拍照图片
func (s *DefaultLianmiApisService) DownloadOrderImage(orderID string) (*Order.DownloadOrderImagesResp, error) {
	return s.Repository.DownloadOrderImage(orderID)
}

//查询VIP会员价格表
func (s *DefaultLianmiApisService) GetVipPriceList(payType int) (*Auth.GetVipPriceResp, error) {
	return s.Repository.GetVipPriceList(payType)
}

//设置当前商户默认OPK
func (s *DefaultLianmiApisService) SetDefaultOPK(username, opk string) error {
	return s.Repository.SetDefaultOPK(username, opk)

}

func (s *DefaultLianmiApisService) GetStoreProductLists(o *Order.ProductsListReq) (*[]models.StoreProductItems, error) {
	// panic("implement me")
	return s.Repository.GetStoreProductLists(o)
}

func (s *DefaultLianmiApisService) AddStoreProductItem(item *models.StoreProductItems) error {
	return s.Repository.AddStoreProductItem(item)
}

func (s *DefaultLianmiApisService) GetGeneralProductFromDB(req *Order.GetGeneralProductPageReq) (*[]models.GeneralProduct, error) {
	// panic("implement me")
	return s.Repository.GetGeneralProductFromDB(req)

}

func (s *DefaultLianmiApisService) SavaOrderItemToDB(item *models.OrderItems) error {
	// panic("implement me")
	return s.Repository.SavaOrderItemToDB(item)
}

func (s *DefaultLianmiApisService) GetOrderListByUser(username string, limit int, offset, status int) (*[]models.OrderItems, error) {
	return s.Repository.GetOrderListByUser(username, limit, offset, status)
}

func (s *DefaultLianmiApisService) GetOrderListByID(orderID string) (*models.OrderItems, error) {
	// panic("implement me")
	return s.Repository.GetOrderListByID(orderID)

}

func (s *DefaultLianmiApisService) SetOrderStatusByOrderID(orderID string, status int) error {
	return s.Repository.SetOrderStatusByOrderID(orderID, status)
}

// 修改订单状态接口
// 仅能处理 拒单,接单,确认收获这三种状态
// 其他状态均不可以想这个接口处理
func (s *DefaultLianmiApisService) UpdateOrderStatus(userid string, storeID string, orderID string, status int) (*models.OrderItems, error) {
	//panic("implement me")

	return s.Repository.UpdateOrderStatus(userid, storeID, orderID, status)

}

func (s *DefaultLianmiApisService) GetUserType(username string) (int, error) {
	//panic("implement me")
	return s.Repository.GetUserType(username)
}

func (s *DefaultLianmiApisService) UpdateOrderStatusByWechatCallback(orderid string) error {
	return s.Repository.UpdateOrderStatusByWechatCallback(orderid)
}

func (s *DefaultLianmiApisService) GetStoreOpkByBusiness(businessId string) (string, error) {
	//panic("implement me")
	return s.Repository.GetStoreOpkByBusiness(businessId)

}

//从redis里获取订单当前最新的数据及状态
func (s *DefaultLianmiApisService) GetOrderInfo(orderID string) (*models.OrderInfo, error) {
	return s.Repository.GetOrderInfo(orderID)

}

func (s *DefaultLianmiApisService) OrderPushPrize(username string, orderID string, prize float64, prizedPhoto string) (string, error) {
	// 更新数据库
	return s.Repository.OrderPushPrize(username, orderID, prize, prizedPhoto)
}

func (s *DefaultLianmiApisService) OrderDeleteByUserAndOrderid(username string, orderid string) error {
	//panic("implement me")
	return s.Repository.OrderDeleteByUserAndOrderid(username, orderid)

}

func (s *DefaultLianmiApisService) DeleteUserOrdersByUserID(username string) error {
	return s.Repository.DeleteUserOrdersByUserID(username)
}

func (s *DefaultLianmiApisService) OrderSerachByKeyWord(username string, req *models.ReqKeyWordDataType) (*[]models.OrderItems, error) {
	return s.Repository.OrderSerachByKeyWord(username, req)
}

//管理员修改app版本号
func (s *DefaultLianmiApisService) ManagerSetVersionLast(req *models.VersionInfo) error {
	return s.Repository.ManagerSetVersionLast(req)
}

func (s *DefaultLianmiApisService) GetUserByWechatOpenid(openid string) (string, error) {
	return s.Repository.GetUserByWechatOpenid(openid)
}

func (s *DefaultLianmiApisService) UpdateUserWxOpenID(username string, openid string) error {
	return s.Repository.UpdateUserWxOpenID(username, openid)
}

func (s *DefaultLianmiApisService) ManagerAddSystemMsg(level int, title, content string) error {
	return s.Repository.ManagerAddSystemMsg(level, title, content)
}

func (s *DefaultLianmiApisService) ManagerDeleteSystemMsg(id uint) error {

	return s.Repository.ManagerDeleteSystemMsg(id)
}
