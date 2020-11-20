package repositories

import (
	"fmt"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/gomodule/redigo/redis"
	"github.com/jinzhu/gorm"
	Auth "github.com/lianmi/servers/api/proto/auth"
	User "github.com/lianmi/servers/api/proto/user"
	"github.com/lianmi/servers/internal/app/dispatcher/multichannel"
	"github.com/lianmi/servers/internal/pkg/models"
	"go.uber.org/zap"
)

type LianmiRepository interface {

	//根据注册用户idd获取用户的资料
	GetUser(username string) (p *models.User, err error)

	QueryUsers(req *User.QueryUsersReq) ([]*User.User, int64, error)

	//注册(用户及商户)
	Register(user *models.User) (err error)

	//重置密码
	ResetPassword(mobile, password string, user *models.User) error

	AddRole(role *models.Role) (err error)

	DeleteUser(id uint64) bool

	GetUserRoles(where interface{}) []*models.Role

	CheckUser(isMaster bool, smscode, username, password, deviceID, os string, clientType int) bool

	GetUserAvatar(where interface{}, sel string) string

	SaveUser(user *models.User) error

	SaveTag(tag *models.Tag) error

	//获取用户ID
	GetUserID(where interface{}) uint64

	//根据用户id获取token
	GetTokenByUserId(where interface{}) string

	//保存用户token
	SaveUserToken(username, deviceID string, token string, expire time.Time) bool

	//获取所有用户
	GetAllUsers(pageIndex int, pageSize int, total *uint64, where interface{}) []*models.User

	//判断用户名是否已存在
	ExistUserByName(username string) bool

	// 判断手机号码是否已存在
	ExistUserByMobile(mobile string) bool

	//更新用户
	UpdateUser(user *models.User, role *models.Role) bool

	//获取用户
	GetUserByID(id int) *models.User

	//登出
	SignOut(token, username, deviceID string) bool

	//令牌是否存在
	ExistsTokenInRedis(deviceID, token string) bool

	//生成注册校验码
	GenerateSmsCode(mobile string) bool

	GetUsernameByMobile(mobile string) (string, error)

	//检测校验码是否正确
	CheckSmsCode(mobile, smscode string) bool

	//授权新创建的群组
	ApproveTeam(teamID string) error

	//封禁群组
	BlockTeam(teamID string) error

	//解封群组
	DisBlockTeam(teamID string) error

	//======后台相关======/
	BlockUser(username string) (err error)

	DisBlockUser(username string) (p *models.User, err error)

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

	//客服人员增加求助记录，以便发给用户评分
	AddGrade(req *Auth.AddGradeReq) (string, error)

	SubmitGrade(req *Auth.SubmitGradeReq) error

	GetBusinessMembership(businessUsername string) (*Auth.GetBusinessMembershipResp, error)

	GetNormalMembership(username string) (*Auth.GetMembershipResp, error)

	//会员付费成功后，需要新增4条佣金记录
	SaveToCommission(username, orderID, content string, blockNumber uint64, txHash string) error

	//提交佣金提现申请(商户，用户)
	SubmitCommssionWithdraw(username, yearMonth string) (*Auth.CommssionWithdrawResp, error)

	SaveTeamUser(pTeamUser *models.TeamUser) error

	GetTeams() []string

	SaveTeam(pTeam *models.Team) error

	DeleteTeamUser(teamID, username string) error

	SetTeamManager(teamID, username string) error

	GetPages(model interface{}, out interface{}, pageIndex, pageSize int, totalCount *uint64, where interface{}, orders ...string) error

	GetTeamUsers(teamID string, PageNum int, PageSize int, total *uint64, where interface{}) []*models.TeamUser

	SaveFriend(pFriend *models.Friend) error

	DeleteFriend(userID, friendUserID uint64) error

	//修改或增加店铺资料
	SaveStore(req *User.Store) error

	//根据商户账号id获取店铺资料
	GetStore(businessUsername string) (*User.Store, error)

	//根据gps位置获取一定范围内的店铺列表
	GetStores(req *User.QueryStoresNearbyReq) (*User.QueryStoresNearbyResp, error)
}

type MysqlLianmiRepository struct {
	logger    *zap.Logger
	db        *gorm.DB
	redisPool *redis.Pool
	multiChan *multichannel.NsqChannel
	base      *BaseRepository
}

func NewMysqlLianmiRepository(logger *zap.Logger, db *gorm.DB, redisPool *redis.Pool, multiChan *multichannel.NsqChannel) LianmiRepository { //, walletSvc Wallet.LianmiWalletClient
	return &MysqlLianmiRepository{
		logger:    logger.With(zap.String("type", "LianmiRepository")),
		db:        db,
		redisPool: redisPool,
		multiChan: multiChan,
		base:      NewBaseRepository(logger, db),
	}
}

//向其它端发送此从设备MultiLoginEvent事件
func (s *MysqlLianmiRepository) SendMultiLoginEventToOtherDevices(isOnline bool, username, deviceID, curOs string, curClientType int, curLogonAt uint64) (err error) {
	redisConn := s.redisPool.Get()
	defer redisConn.Close()

	deviceListKey := fmt.Sprintf("devices:%s", username)

	deviceIDSliceNew, _ := redis.Strings(redisConn.Do("ZRANGEBYSCORE", deviceListKey, "-inf", "+inf"))
	//查询出当前在线所有主从设备
	for _, eDeviceID := range deviceIDSliceNew {
		targetMsg := &models.Message{}
		curDeviceKey := fmt.Sprintf("DeviceJwtToken:%s", eDeviceID)
		curJwtToken, _ := redis.String(redisConn.Do("GET", curDeviceKey))
		s.logger.Debug("Redis GET ", zap.String("curDeviceKey", curDeviceKey), zap.String("curJwtToken", curJwtToken))

		targetMsg.UpdateID()
		//构建消息路由, 第一个参数是要处理的业务类型，后端服务器处理完成后，需要用此来拼接topic: {businessTypeName.Frontend}
		targetMsg.BuildRouter("Auth", "", "Auth.Frontend")

		targetMsg.SetJwtToken(curJwtToken)
		targetMsg.SetUserName(username)
		targetMsg.SetDeviceID(eDeviceID)
		targetMsg.SetBusinessTypeName("Auth")
		targetMsg.SetBusinessType(uint32(2))
		targetMsg.SetBusinessSubType(uint32(3)) //MultiLoginEvent = 3

		targetMsg.BuildHeader("Dispatcher", time.Now().UnixNano()/1e6)

		//构造负载数据
		clients := make([]*Auth.DeviceInfo, 0)
		deviceInfo := &Auth.DeviceInfo{
			Username:     username,
			ConnectionId: "",
			DeviceId:     deviceID,
			DeviceIndex:  0,
			IsMaster:     isOnline,
			Os:           curOs,
			ClientType:   Auth.ClientType(curClientType),
			LogonAt:      curLogonAt,
		}

		clients = append(clients, deviceInfo)

		resp := &Auth.MultiLoginEventRsp{
			State:   false,
			Clients: clients,
		}

		data, _ := proto.Marshal(resp)
		targetMsg.FillBody(data) //网络包的body，承载真正的业务数据

		targetMsg.SetCode(200) //成功的状态码

		//构建数据完成，向NsqChan发送
		s.multiChan.NsqChan <- targetMsg

	}

	return nil
}
