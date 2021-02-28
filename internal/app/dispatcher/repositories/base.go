package repositories

import (
	"time"

	"github.com/gomodule/redigo/redis"
	Auth "github.com/lianmi/servers/api/proto/auth"
	Order "github.com/lianmi/servers/api/proto/order"
	User "github.com/lianmi/servers/api/proto/user"

	// Wallet "github.com/lianmi/servers/api/proto/wallet"
	"github.com/lianmi/servers/internal/app/dispatcher/multichannel"

	"github.com/lianmi/servers/internal/pkg/models"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type LianmiRepository interface {

	//根据注册用户id获取用户的资料
	GetUser(username string) (p *models.User, err error)

	//根据注册用户id获取redis里此用户的缓存
	GetUserDataFromRedis(username string) (p *models.UserBase, err error)

	//根据注册用户id获取redis里此用户的设备id
	GetDeviceFromRedis(username string) (string, error)

	//查询出用户表所有用户账号
	QueryAllUsernames() ([]string, error)

	//多条件不定参数批量分页获取用户列表
	QueryUsers(req *User.QueryUsersReq) (*User.QueryUsersResp, error)

	//注册(用户及商户)
	Register(user *models.User) (err error)

	//重置密码
	ResetPassword(mobile, password string, user *models.User) error

	//同时增加用户类型角色
	AddRole(role *models.Role) (err error)

	//删除用户
	DeleteUser(id uint) bool

	// 根据UserName获取用户角色
	GetUserRoles(where interface{}) []*models.Role

	// 检测用户是否可以登陆
	CheckUser(isMaster bool, username, password, deviceID, os string, clientType int) (bool, string)

	//  使用手机及短信验证码登录
	LoginBySmscode(username, mobile, smscode, deviceID, os string, clientType int) (bool, string)

	//更新用户
	UpdateUser(username string, user *models.User) error

	//更新商店
	UpdateStore(username string, store *models.Store) error

	//保存标签MarkTag
	AddTag(tag *models.Tag) error

	//保存用户token
	SaveUserToken(username, deviceID string, token string, expire time.Time) bool

	//判断用户名是否已存在
	ExistUserByName(username string) bool

	// 判断手机号码是否已存在
	ExistUserByMobile(mobile string) bool

	//获取用户
	GetUserByID(id int) *models.User

	//登出
	SignOut(token, username, deviceID string) bool

	//令牌是否存在
	ExistsTokenInRedis(deviceID, token string) bool

	//生成注册校验码
	GenerateSmsCode(mobile string) bool

	//根据手机号获取注册账号id
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

	//======后台相关======/
	BlockUser(username string) (err error)

	DisBlockUser(username string) (err error)

	GetProductInfo(productID string) (*Order.Product, error)

	AddGeneralProduct(generalProductInfo *models.GeneralProductInfo) error

	GetGeneralProductByID(productID string) (p *models.GeneralProduct, err error)

	GetGeneralProductPage(req *Order.GetGeneralProductPageReq) (*Order.GetGeneralProductPageResp, error)

	UpdateGeneralProduct(generalProductInfo *models.GeneralProductInfo) error

	DeleteGeneralProduct(productID string) bool

	QueryCustomerServices(req *Auth.QueryCustomerServiceReq) ([]*models.CustomerServiceInfo, error)

	AddCustomerService(req *Auth.AddCustomerServiceReq) error

	DeleteCustomerService(req *Auth.DeleteCustomerServiceReq) bool

	//修改在线客服资料
	UpdateCustomerService(req *Auth.UpdateCustomerServiceReq) error

	QueryGrades(req *Auth.GradeReq, pageIndex int, pageSize int, total *int64, where interface{}) ([]*models.Grade, error)

	//客服人员增加求助记录，以便发给用户评分
	AddGrade(req *Auth.AddGradeReq) (string, error)

	SubmitGrade(req *Auth.SubmitGradeReq) error

	GetBusinessMembership(businessUsername string) (*Auth.GetBusinessMembershipResp, error)

	//对某个用户的推广会员佣金进行统计
	CommissonSatistics(username string) (*Auth.CommissonSatisticsResp, error)

	//用户查询按月统计发展的付费会员总数及返佣金额，是否已经返佣
	GetCommissionStatistics(username string) (*Auth.GetCommssionsResp, error)

	//根据PayType获取到VIP价格
	GetVipUserPrice(payType int) (*models.VipPrice, error)

	//提交佣金提现申请(商户，用户)
	SubmitCommssionWithdraw(username, yearMonth string) (*Auth.CommssionWithdrawResp, error)

	//增加群成员
	AddTeamUser(pTeamUser *models.TeamUser) error

	//设置获取取消群管理员
	UpdateTeamUserManager(teamID, managerUsername string, isAdd bool) error

	// 修改群成员呢称、扩展
	UpdateTeamUserMyInfo(teamID, username, aliasName, ex string) error

	//修改群通知方式
	UpdateTeamUserNotifyType(teamID string, notifyType int) error

	//解除群成员的禁言
	SetMuteTeamUser(teamID, dissMuteUser string, isMute bool, mutedays int) error

	GetChargeProductID() (string, error)

	GetTeams() []string

	//创建群
	CreateTeam(pTeam *models.Team) error

	//更新群数据
	UpdateTeam(teamID string, pTeam *models.Team) error

	DeleteTeamUser(teamID, username string) error

	GetPages(model interface{}, out interface{}, pageIndex, pageSize int, totalCount *int64, where interface{}, orders ...string) error

	GetTeamUsers(teamID string, PageNum int, PageSize int, total *int64, where interface{}) []*models.TeamUser

	AddFriend(pFriend *models.Friend) error

	UpdateFriend(pFriend *models.Friend) error

	DeleteFriend(userID, friendUserID uint64) error

	//修改或增加店铺资料
	AddStore(req *User.Store) error

	//根据商户账号id获取店铺资料
	GetStore(businessUsername string) (*User.Store, error)

	//根据gps位置获取一定范围内的店铺列表
	GetStores(req *Order.QueryStoresNearbyReq) (*Order.QueryStoresNearbyResp, error)

	//后台管理员将店铺审核通过
	AuditStore(req *Auth.AuditStoreReq) error

	//获取某个商户的所有商品列表
	GetProductsList(req *Order.ProductsListReq) (*Order.ProductsListResp, error)

	//设置商品的子类型
	SetProductSubType(req *Order.ProductSetSubTypeReq) error

	//获取当前用户对所有店铺点赞情况
	UserLikes(username string) (*User.UserLikesResp, error)

	//获取店铺的所有点赞的用户列表
	StoreLikes(username string) (*User.StoreLikesResp, error)

	ClickLike(username, businessUsername string) (int64, error)

	DeleteClickLike(username, businessUsername string) (int64, error)

	//将点赞记录插入到UserLike表
	AddUserLike(username, businessUser string) error

	//从redis里获取订单当前最新的数据及状态
	GetOrderInfo(orderID string) (*models.OrderInfo, error)

	//增加订单拍照图片上链历史表
	SaveOrderImagesBlockchain(req *Order.UploadOrderImagesReq, orderTotalAmount float64, blcokNumber uint64, buyUser, businessUser, hash string) error

	//修改订单的body类型及body加密阿里云文件上链历史表
	SaveOrderBody(req *Order.UploadOrderBodyReq) error

	//用户端: 根据 OrderID 获取所有订单拍照图片
	DownloadOrderImage(orderID string) (*Order.DownloadOrderImagesResp, error)

	//根据订单号获取支付用户及金额
	GetAlipayInfoByTradeNo(outTradeNo string) (string, float64, bool, error)

	//查询VIP会员价格表
	GetVipPriceList(payType int) (*Auth.GetVipPriceResp, error)

	//获取各种彩票的开售及停售时刻
	QueryLotterySaleTimes() (*Order.QueryLotterySaleTimesRsp, error)

	//清除所有OPK
	ClearAllOPK(username string) error

	//获取当前商户的所有OPK
	GetAllOPKS(username string) (*Order.GetAllOPKSRsp, error)

	//删除当前商户的指定OPK
	EraseOPK(username string, req *Order.EraseOPKSReq) error

	//设置当前商户默认OPK
	SetDefaultOPK(username, opk string) error
}

type MysqlLianmiRepository struct {
	logger    *zap.Logger
	db        *gorm.DB
	redisPool *redis.Pool
	multiChan *multichannel.NsqChannel
	base      *BaseRepository
}

func NewMysqlLianmiRepository(logger *zap.Logger, db *gorm.DB, redisPool *redis.Pool, multiChan *multichannel.NsqChannel) LianmiRepository {
	return &MysqlLianmiRepository{
		logger:    logger.With(zap.String("type", "LianmiRepository")),
		db:        db,
		redisPool: redisPool,
		multiChan: multiChan,
		base:      NewBaseRepository(logger, db),
	}
}
