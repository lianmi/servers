package repositories

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/gomodule/redigo/redis"
	"github.com/jinzhu/gorm"
	Auth "github.com/lianmi/servers/api/proto/auth"
	Service "github.com/lianmi/servers/api/proto/service"
	User "github.com/lianmi/servers/api/proto/user"
	"github.com/lianmi/servers/internal/app/authservice/nsqBackend"
	"github.com/lianmi/servers/internal/common"
	"github.com/lianmi/servers/internal/pkg/models"
	"github.com/pkg/errors"
	"go.uber.org/zap"

	"github.com/lianmi/servers/util/dateutil"
)

type LianmiRepository interface {
	GetUser(ID uint64) (p *models.User, err error)
	BlockUser(username string) (p *models.User, err error)
	DisBlockUser(username string) (p *models.User, err error)
	Register(user *models.User) (err error)
	Resetpwd(mobile, password string, user *models.User) error
	ChanPassword(username, oldPassword, newPassword string) error
	AddRole(role *models.Role) (err error)
	DeleteUser(id uint64) bool
	GetUserRoles(where interface{}) []*models.Role
	CheckUser(isMaster bool, smscode, username, password, deviceID, os string, clientType int) bool
	GetUserAvatar(where interface{}, sel string) string

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

	//检测校验码是否正确
	CheckSmsCode(mobile, smscode string) bool

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

	QueryCustomerServices() ([]*models.CustomerServiceInfo, error)

	AddCustomerService(sc *Service.CustomerServiceInfo) ([]*models.CustomerServiceInfo, error)

	DeleteCustomerService(sc *Service.CustomerServiceInfo) bool

	UpdateCustomerService(sc *Service.CustomerServiceInfo) ([]*models.CustomerServiceInfo, error)

	QueryGrades(req *Service.GradeReq, pageIndex int, pageSize int, total *uint64, where interface{}) ([]*models.Grade, error)

	//客服人员增加求助记录，以便发给用户评分
	AddGrade(req *Service.AddGradeReq) (string, error)

	SubmitGrade(req *Service.SubmitGradeReq) error

	GetMembershipCardSaleMode(businessUsername string) (int, error)

	SetMembershipCardSaleMode(businessUsername string, saleType int) error
}

type MysqlLianmiRepository struct {
	logger    *zap.Logger
	db        *gorm.DB
	redisPool *redis.Pool
	nsqClient *nsqBackend.NsqClient
	base      *BaseRepository
}

func NewMysqlLianmiRepository(logger *zap.Logger, db *gorm.DB, redisPool *redis.Pool, nc *nsqBackend.NsqClient) LianmiRepository {
	return &MysqlLianmiRepository{
		logger:    logger.With(zap.String("type", "LianmiRepository")),
		db:        db,
		redisPool: redisPool,
		nsqClient: nc,
		base:      NewBaseRepository(logger, db),
	}
}

func (s *MysqlLianmiRepository) GetUser(ID uint64) (p *models.User, err error) {
	p = new(models.User)
	if err = s.db.Model(p).Where("id = ?", ID).First(p).Error; err != nil {
		//记录找不到也会触发错误
		// fmt.Println("GetUser first error:", err.Error())
		return nil, errors.Wrapf(err, "Get user error[id=%d]", ID)
	}
	s.logger.Debug("GetUser run...")
	return
}

/*
封号
1. 将users表的用户记录的state设置为3
2. 踢出此用户的所有主从设备
*/
func (s *MysqlLianmiRepository) BlockUser(username string) (p *models.User, err error) {

	redisConn := s.redisPool.Get()
	defer redisConn.Close()

	p = new(models.User)
	if err = s.db.Model(p).Where("username = ?", username).First(p).Error; err != nil {
		return nil, errors.Wrapf(err, "Get user error[username=%s]", username)
	}

	p.State = 2 //1-正常， 2-封号

	tx := s.base.GetTransaction()

	if err := tx.Save(p).Error; err != nil {
		s.logger.Error("封号失败", zap.Error(err))
		tx.Rollback()

	}
	//提交
	tx.Commit()

	//将此用户所在在线设备全部踢出
	deviceListKey := fmt.Sprintf("devices:%s", username)

	//查询出所有主从设备
	deviceIDSlice, _ := redis.Strings(redisConn.Do("ZRANGEBYSCORE", deviceListKey, "-inf", "+inf"))
	for index, eDeviceID := range deviceIDSlice {
		s.logger.Debug("查询出所有主从设备", zap.Int("index", index), zap.String("eDeviceID", eDeviceID))
		deviceKey := fmt.Sprintf("DeviceJwtToken:%s", eDeviceID)
		jwtToken, _ := redis.String(redisConn.Do("GET", deviceKey))
		s.logger.Debug("Redis GET ", zap.String("deviceKey", deviceKey), zap.String("jwtToken", jwtToken))

		//向当前主设备及从设备发出踢下线
		if err := s.SendKickedMsgToDevice(jwtToken, username, eDeviceID); err != nil {
			s.logger.Error("Failed to Send Kicked Msg To Device to ProduceChannel", zap.Error(err))
		}

		_, err = redisConn.Do("DEL", deviceKey) //删除deviceKey

		deviceHashKey := fmt.Sprintf("devices:%s:%s", username, eDeviceID)
		_, err = redisConn.Do("DEL", deviceHashKey) //删除deviceHashKey

	}

	//删除所有与之相关的key
	_, err = redisConn.Do("DEL", deviceListKey) //删除deviceListKey

	s.logger.Debug("BlockUser run.")

	return
}

/*
解封
1. 将users表的用户记录的state设置为1
*/
func (s *MysqlLianmiRepository) DisBlockUser(username string) (p *models.User, err error) {
	p = new(models.User)
	if err = s.db.Model(p).Where("username = ?", username).First(p).Error; err != nil {
		return nil, errors.Wrapf(err, "Get user error[username=%s]", username)
	}

	p.State = 1 //1-正常， 2-封号

	tx := s.base.GetTransaction()

	if err := tx.Save(p).Error; err != nil {
		s.logger.Error("解封失败", zap.Error(err))
		tx.Rollback()

	}
	//提交
	tx.Commit()

	s.logger.Debug("DisBlockUser run.")
	return
}

//注册用户，username需要唯一
func (s *MysqlLianmiRepository) Register(user *models.User) (err error) {
	//获取redis里最新id， 生成唯一的username
	var newIndex uint64

	redisConn := s.redisPool.Get()
	defer redisConn.Close()

	if newIndex, err = redis.Uint64(redisConn.Do("INCR", "usernameindex")); err != nil {
		s.logger.Error("redisConn GET usernameindex Error", zap.Error(err))
		return err
	}

	if user.GetUserType() == User.UserType_Ut_Operator { //10086
		user.Username = fmt.Sprintf("admin%d", newIndex)
	} else {
		user.Username = fmt.Sprintf("id%d", newIndex)
	}

	//将用户信息缓存到redis里
	userKey := fmt.Sprintf("userData:%s", user.Username)
	if _, err := redisConn.Do("HMSET", redis.Args{}.Add(userKey).AddFlat(user)...); err != nil {
		s.logger.Error("错误：HMSET", zap.Error(err))
	}

	if err := s.base.Create(user); err != nil {
		s.logger.Error("db写入错误，注册用户失败")
		return err
	}

	//创建redis的sync:{用户账号} myInfoAt 时间戳
	//myInfoAt, friendsAt, friendUsersAt, teamsAt, tagsAt, systemMsgAt, watchAt, productAt,  generalProductAt

	syncKey := fmt.Sprintf("sync:%s", user.Username)
	redisConn.Do("HSET", syncKey, "myInfoAt", time.Now().UnixNano()/1e6)
	redisConn.Do("HSET", syncKey, "friendsAt", time.Now().UnixNano()/1e6)
	redisConn.Do("HSET", syncKey, "friendUsersAt", time.Now().UnixNano()/1e6)
	redisConn.Do("HSET", syncKey, "teamsAt", time.Now().UnixNano()/1e6)
	redisConn.Do("HSET", syncKey, "tagsAt", time.Now().UnixNano()/1e6)
	redisConn.Do("HSET", syncKey, "systemMsgAt", time.Now().UnixNano()/1e6)
	redisConn.Do("HSET", syncKey, "watchAt", time.Now().UnixNano()/1e6)
	redisConn.Do("HSET", syncKey, "productAt", time.Now().UnixNano()/1e6)
	redisConn.Do("HSET", syncKey, "generalProductAt", time.Now().UnixNano()/1e6)

	//网点商户自动建群
	if user.GetUserType() == User.UserType_Ut_Business {

		var newTeamIndex uint64
		if newTeamIndex, err = redis.Uint64(redisConn.Do("INCR", "TeamIndex")); err != nil {
			s.logger.Error("redisConn GET TeamIndex Error", zap.Error(err))
			return err
		}
		pTeam := new(models.Team)
		pTeam.CreatedAt = time.Now().UnixNano() / 1e6
		pTeam.TeamID = fmt.Sprintf("team%d", newTeamIndex) //群id， 自动生成
		pTeam.Teamname = fmt.Sprintf("team%d", newTeamIndex)
		pTeam.Nick = fmt.Sprintf("%s的群", user.Nick)
		pTeam.Owner = user.Username
		pTeam.Type = 1
		pTeam.VerifyType = 1
		pTeam.InviteMode = 1

		//默认的设置
		pTeam.Status = 1 //Init(1) - 初始状态,审核中 Normal(2) - 正常状态 Blocked(3) - 封禁状态
		pTeam.MemberLimit = common.PerTeamMembersLimit
		pTeam.MemberNum = 1  //刚刚建群是只有群主1人
		pTeam.MuteType = 1   //None(1) - 所有人可发言
		pTeam.InviteMode = 1 //邀请模式,初始为1

		//使用事务同时更新创建群数据
		tx := s.base.GetTransaction()

		if err := tx.Save(pTeam).Error; err != nil {
			s.logger.Error("更新群team表失败", zap.Error(err))
			tx.Rollback()
			return err
		}

		//提交
		tx.Commit()
	}

	s.logger.Debug("注册用户成功", zap.String("Username", user.Username))
	return nil

}

//重置密码
func (s *MysqlLianmiRepository) Resetpwd(mobile, password string, user *models.User) error {

	redisConn := s.redisPool.Get()
	defer redisConn.Close()

	// var user models.User
	sel := "id"
	where := models.User{Mobile: mobile}
	err := s.base.First(&where, &user, sel)
	//记录不存在错误(RecordNotFound)，返回false
	if gorm.IsRecordNotFoundError(err) {
		return err
	}
	//其他类型的错误，写下日志，返回false
	if err != nil {
		s.logger.Error("根据手机号码获取用户信息失败", zap.Error(err))
		return err
	}

	//替换旧密码
	user.Password = password

	tx := s.base.GetTransaction()
	if err := tx.Save(user).Error; err != nil {
		s.logger.Error("更新用户失败", zap.Error(err))
		tx.Rollback()
		return err
	}

	tx.Commit()

	return nil
}

//修改密码
func (s *MysqlLianmiRepository) ChanPassword(username, oldPassword, newPassword string) error {
	var user models.User
	sel := "id"
	where := models.User{Username: username}
	err := s.base.First(&where, &user, sel)
	//记录不存在错误(RecordNotFound)，返回false
	if gorm.IsRecordNotFoundError(err) {
		return err
	}
	//其他类型的错误，写下日志，返回false
	if err != nil {
		s.logger.Error("获取用户信息失败", zap.Error(err))
		return err
	}

	//判断旧密码
	if oldPassword == user.Password {
		user.Password = newPassword
	}

	tx := s.base.GetTransaction()
	if err := tx.Save(user).Error; err != nil {
		s.logger.Error("修改密码失败", zap.Error(err))
		tx.Rollback()
		return err
	}

	tx.Commit()

	return nil

}

// 获取用户角色
func (s *MysqlLianmiRepository) GetUserRoles(where interface{}) []*models.Role {
	var roles []*models.Role
	if err := s.base.Find(where, &roles, ""); err != nil {
		s.logger.Error("获取用户角色错误")
	}
	return roles
}

/*
登录处理
1. 当主设备登录成功后：

#创建有序集合: devices:{username} ，序号从1开始，score是clientType, value是设备id
ZADD devices:lsj001 5 "959bb0ae-1c12-4b60-8741-173361ceba8a"

#列出当前设备列表: devices:{username}
ZRANGE devices:lsj001 0 -1
1) "959bb0ae-1c12-4b60-8741-173361ceba8a"

#创建哈希表: devices:{username}:{设备id}
HMSET devices:lsj001:959bb0ae-1c12-4b60-8741-173361ceba8a deviceid "959bb0ae-1c12-4b60-8741-173361ceba8a" ismaster 1 usertype 1 clienttype 5 os "Android"

2. 当从设备登录成功后:

#插入有序集合: devices:{username}, clientType =4
ZADD devices:lsj001 4 "11111111-2222-3333-3333-44444444444"

#列出当前设备列表
ZRANGE devices:lsj001 0 -1
1) "959bb0ae-1c12-4b60-8741-173361ceba8a"
2) "11111111-2222-3333-3333-44444444444"

#查询出clientType =4的成员
ZRANGEBYSCORE devices:lsj001 4 4
1) "11111111-2222-3333-3333-44444444444"

#创建哈希表: devices:{username}:{设备id}
HMSET devices:lsj001:11111111-2222-3333-3333-44444444444 deviceid "11111111-2222-3333-3333-44444444444" ismaster 0 usertype 1 clienttype 3 os "iOS" protocolversion "2.0" sdkversion "3.0"

*/
func (s *MysqlLianmiRepository) CheckUser(isMaster bool, smscode, username, password, deviceID, os string, clientType int) bool {
	var err error
	redisConn := s.redisPool.Get()
	defer redisConn.Close()

	where := models.User{Username: username}
	var user models.User
	if err := s.base.First(&where, &user); err != nil {
		s.logger.Error("用户账号不存在")
		return false
	}

	if user.State == 3 {
		s.logger.Error("用户被封号")
		return false
	}

	mobile := user.Mobile
	if mobile == "" {
		s.logger.Error("用户手机不存在")
		return false
	}

	//检测校验码
	if !s.CheckSmsCode(mobile, smscode) {
		s.logger.Error("校验码不匹配", zap.String("mobile", mobile), zap.String("smscode", smscode))
		return false
	}

	//用完就要删除
	smscodeKey := fmt.Sprintf("smscode:%s", mobile)
	_, err = redisConn.Do("DEL", smscodeKey) //删除smscode

	deviceListKey := fmt.Sprintf("devices:%s", username)

	if isMaster { //主设备

		//主设备需要核对密码，从设备则无须核对
		if user.Password != password {
			s.logger.Error("密码不匹配")
			return false
		}

		//查询出所有主从设备
		deviceIDSlice, _ := redis.Strings(redisConn.Do("ZRANGEBYSCORE", deviceListKey, "-inf", "+inf"))
		for index, eDeviceID := range deviceIDSlice {
			s.logger.Debug("查询出所有主从设备", zap.Int("index", index), zap.String("eDeviceID", eDeviceID))
			deviceKey := fmt.Sprintf("DeviceJwtToken:%s", eDeviceID)
			jwtToken, _ := redis.String(redisConn.Do("GET", deviceKey))
			s.logger.Debug("Redis GET ", zap.String("deviceKey", deviceKey), zap.String("jwtToken", jwtToken))

			//向当前主设备及从设备发出踢下线
			if err := s.SendKickedMsgToDevice(jwtToken, username, eDeviceID); err != nil {
				s.logger.Error("Failed to Send Kicked Msg To Device to ProduceChannel", zap.Error(err))
			}

			_, err = redisConn.Do("DEL", deviceKey) //删除deviceKey

			deviceHashKey := fmt.Sprintf("devices:%s:%s", username, eDeviceID)
			_, err = redisConn.Do("DEL", deviceHashKey) //删除deviceHashKey

		}

		//删除所有与之相关的key
		_, err = redisConn.Do("DEL", deviceListKey) //删除deviceListKey

		err = redisConn.Send("ZADD", deviceListKey, clientType, deviceID) //有序集合

		deviceHashKey := fmt.Sprintf("devices:%s:%s", username, deviceID) //创建主设备的哈希表, index为1

		err = redisConn.Send("HMSET",
			deviceHashKey,
			"username", username,
			"deviceid", deviceID,
			"ismaster", 1,
			"usertype", user.UserType,
			"clientType", clientType,
			"os", os,
			"logonAt", uint64(time.Now().UnixNano()/1e6))

		_ = err

		//一次性写入到Redis
		if err := redisConn.Flush(); err != nil {
			s.logger.Error("写入redis失败", zap.Error(err))
			return false
		} else {
			s.logger.Debug("写入redis成功",
				zap.String("deviceListKey", deviceListKey),
				zap.String("deviceHashKey", deviceHashKey))
		}

	} else {
		//从设备登录
		//查询出是否之前相同clientType的从设备在线，如果有则踢出
		curSlaceDeviceSlice, _ := redis.Strings(redisConn.Do("ZRANGEBYSCORE", deviceListKey, clientType, clientType))

		if len(curSlaceDeviceSlice) > 0 {
			index := 0
			oldDeviceID := curSlaceDeviceSlice[index]

			s.logger.Debug("查询出之前的旧的在线从设备", zap.Int("index", index), zap.String("oldDeviceID", oldDeviceID))
			deviceKey := fmt.Sprintf("DeviceJwtToken:%s", oldDeviceID)
			jwtToken, _ := redis.String(redisConn.Do("GET", deviceKey))
			s.logger.Debug("Redis GET ", zap.String("deviceKey", deviceKey), zap.String("jwtToken", jwtToken))

			//取出当前旧的设备的os， clientType， logonAt
			curDeviceHashKey := fmt.Sprintf("devices:%s:%s", username, oldDeviceID)
			curOs, _ := redis.String(redisConn.Do("HGET", curDeviceHashKey, "os"))
			curClientType, _ := redis.Int(redisConn.Do("HGET", curDeviceHashKey, "clientType"))
			curLogonAt, _ := redis.Uint64(redisConn.Do("HGET", curDeviceHashKey, "logonAt"))

			//当前从设备踢下线
			if err := s.SendKickedMsgToDevice(jwtToken, username, oldDeviceID); err != nil {
				s.logger.Error("Failed to Send Kicked Msg To Device to ProduceChannel", zap.Error(err))
			}

			//移除单个元素 ZREM deviceListKey {设备id}
			_, err = redisConn.Do("ZREM", deviceListKey, oldDeviceID)

			_, err = redisConn.Do("DEL", deviceKey) //删除deviceKey

			deviceHashKey := fmt.Sprintf("devices:%s:%s", username, oldDeviceID)
			_, err = redisConn.Do("DEL", deviceHashKey) //删除deviceHashKey

			//向其它端发送此从设备离线的事件
			if err := s.SendMultiLoginEventToOtherDevices(false, username, oldDeviceID, curOs, curClientType, curLogonAt); err != nil {
				s.logger.Error("Failed to Send MultiLoginEvent to Other Devices to ProduceChannel", zap.Error(err))
			}

		}

		//向其它端发送此从设备上线的事件
		logonAt := uint64(time.Now().UnixNano() / 1e6)
		if err := s.SendMultiLoginEventToOtherDevices(true, username, deviceID, os, clientType, logonAt); err != nil {
			s.logger.Error("Failed to Send MultiLoginEvent to Other Devices to ProduceChannel", zap.Error(err))
		}

		//增加新的从设备
		//ZADD devices:lsj001 4 "11111111-2222-3333-3333-44444444444"
		err = redisConn.Send("ZADD", deviceListKey, clientType, deviceID) //有序集合

		//HMSET
		deviceHashKey := fmt.Sprintf("devices:%s:%s", username, deviceID) //创建从设备的哈希表

		err = redisConn.Send("HMSET",
			deviceHashKey,
			"username", username,
			"deviceid", deviceID,
			"ismaster", 0,
			"usertype", user.UserType,
			"clientType", clientType,
			"os", os,
			"logonAt", uint64(time.Now().UnixNano()/1e6))

		//一次性写入到Redis
		if err := redisConn.Flush(); err != nil {
			s.logger.Error("写入redis失败", zap.Error(err))
			return false
		}

		_ = err
	}

	//当前所有主从设备数量
	count, _ := redis.Int(redisConn.Do("ZCARD", deviceListKey))
	s.logger.Debug("当前所有主从设备数量", zap.Int("count", count))

	return true
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
		// kickMsg.SetTaskID(uint32(taskId))
		targetMsg.SetBusinessTypeName("Auth")
		targetMsg.SetBusinessType(uint32(2))
		targetMsg.SetBusinessSubType(uint32(3)) //MultiLoginEvent = 3

		targetMsg.BuildHeader("AuthService", time.Now().UnixNano()/1e6)

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
		//构建数据完成，向dispatcher发送
		topic := "Auth.Frontend"
		rawData, _ := json.Marshal(targetMsg)
		if err := s.nsqClient.Producer.Public(topic, rawData); err == nil {
			s.logger.Info("message succeed send to ProduceChannel", zap.String("topic", topic))
		} else {
			s.logger.Error(" failed to send message to ProduceChannel", zap.Error(err))
		}
	}

	return nil
}

func (s *MysqlLianmiRepository) AddRole(role *models.Role) (err error) {
	if err := s.db.Create(role).Error; err != nil {
		s.logger.Error("新建用户角色失败")
		return err
	} else {
		return nil
	}
}

func (s *MysqlLianmiRepository) DeleteUser(id uint64) bool {
	//采用事务同时删除用户和相应的用户角色
	var (
		userWhere = models.User{ID: id}
		user      models.User
		roleWhere = models.Role{UserID: id}
		role      models.Role
	)
	tx := s.base.GetTransaction()
	tx.Where(&roleWhere).Delete(&role)
	if err := tx.Where(&userWhere).Delete(&user).Error; err != nil {
		s.logger.Error("删除用户失败", zap.Error(err))
		tx.Rollback()
		return false
	}
	tx.Commit()
	return true
}

func (s *MysqlLianmiRepository) GetUserAvatar(where interface{}, sel string) string {
	var user models.User
	// conditionString, values, _ := s.base.BuildCondition(map[string]interface{}{
	//     "id":       id,
	//     // "itemName like": "%22220",
	//     // "id in":         []int{20, 19, 30},
	//     // "num !=" : 20,
	// })
	// err := s.base.First(conditionString, values, &user, sel)
	err := s.base.First(&where, &user, sel)
	//记录不存在错误(RecordNotFound)，返回false
	if gorm.IsRecordNotFoundError(err) {
		s.logger.Error("获取用户头像失败", zap.Error(err))
		return "" //TODO 默认
	}
	return user.Avatar
}

func (s *MysqlLianmiRepository) GetUserID(where interface{}) uint64 {
	var user models.User
	// conditionString, values, _ := s.base.BuildCondition(map[string]interface{}{
	//     "username":       username,
	// })
	// where := models.User{Username: username}
	err := s.base.First(&where, &user, "id")
	//记录不存在错误(RecordNotFound)，返回false
	if gorm.IsRecordNotFoundError(err) {
		s.logger.Error("获取用户id失败", zap.Error(err))
		return 0 //TODO 默认
	}

	return user.ID
}

//根据用户id获取token
func (s *MysqlLianmiRepository) GetTokenByUserId(where interface{}) string {
	var tbToken models.Token
	// where := models.User{Username: username}
	err := s.base.First(&where, &tbToken, "token")
	//记录不存在错误(RecordNotFound)，返回false
	if gorm.IsRecordNotFoundError(err) {
		s.logger.Error("获取Token失败", zap.Error(err))
		return "" //TODO 默认
	}

	return tbToken.Token
}

func (s *MysqlLianmiRepository) GetAllUsers(pageIndex int, pageSize int, total *uint64, where interface{}) []*models.User {
	var users []*models.User
	if err := s.base.GetPages(&models.User{}, &users, pageIndex, pageSize, total, where); err != nil {
		s.logger.Error("获取用户信息失败", zap.Error(err))
	}
	return users
}

//判断用户名是否已存在
func (s *MysqlLianmiRepository) ExistUserByName(username string) bool {
	var user models.User
	sel := "id"
	// conditionString, values, _ := s.base.BuildCondition(map[string]interface{}{
	//     "username":       username,
	//     // "itemName like": "%22220",
	//     // "id in":         []int{20, 19, 30},
	//     // "num !=" : 20,
	// })
	where := models.User{Username: username}
	err := s.base.First(&where, &user, sel)
	//记录不存在错误(RecordNotFound)，返回false
	if gorm.IsRecordNotFoundError(err) {
		return false
	}
	//其他类型的错误，写下日志，返回false
	if err != nil {
		s.logger.Error("根据用户名获取用户信息失败", zap.Error(err))
		return false
	}
	return true
}

// 判断手机号码是否已存在
func (s *MysqlLianmiRepository) ExistUserByMobile(mobile string) bool {
	var user models.User
	sel := "id"
	where := models.User{Mobile: mobile}
	err := s.base.First(&where, &user, sel)
	//记录不存在错误(RecordNotFound)，返回false
	if gorm.IsRecordNotFoundError(err) {
		return false
	}
	//其他类型的错误，写下日志，返回false
	if err != nil {
		s.logger.Error("根据手机号码获取用户信息失败", zap.Error(err))
		return false
	}
	return true
}

//更新用户
func (s *MysqlLianmiRepository) UpdateUser(user *models.User, role *models.Role) bool {
	//使用事务同时更新用户数据和角色数据
	tx := s.base.GetTransaction()
	if err := tx.Save(user).Error; err != nil {
		s.logger.Error("更新用户失败", zap.Error(err))
		tx.Rollback()
		return false
	}
	if err := tx.Save(&role).Error; err != nil {
		s.logger.Error("更新用户角色失败", zap.Error(err))
		tx.Rollback()
		return false
	}
	tx.Commit()
	return true
}

//获取用户
func (s *MysqlLianmiRepository) GetUserByID(id int) *models.User {
	var user models.User
	if err := s.base.FirstByID(&user, id); err != nil {
		s.logger.Error("获取用户失败", zap.Error(err))
	}
	return &user
}

/*
保存用户token到redis里
登出的处理需要删除redis里的key
*/
func (s *MysqlLianmiRepository) SaveUserToken(username, deviceID string, token string, expire time.Time) bool {

	redisConn := s.redisPool.Get()
	defer redisConn.Close()

	deviceKey := fmt.Sprintf("DeviceJwtToken:%s", deviceID)
	_, err := redisConn.Do("SET", deviceKey, token) //deviceID关联到token, mqtt消息必须要验证这个

	_, err = redisConn.Do("EXPIRE", deviceKey, common.ExpireTime) //过期时间

	//重新读出token，看看是否可以读出

	// if tokenInRedis, err := redis.String(redisConn.Do("GET", deviceKey)); err != nil {
	// 	s.logger.Error("重新读出token失败", zap.Error(err))
	// 	return false
	// } else {
	// 	isEqueal := tokenInRedis == token
	// 	s.logger.Debug("重新读出token成功", zap.String("tokenInRedis", tokenInRedis), zap.Bool("isEqueal", isEqueal))
	// }
	_ = err
	return true
}

/*
登出
1. 如果是主设备，则踢出此用户的所有主从设备， 如果仅仅是从设备，就删除自己的数据
2. 删除redis里的此用户的哈希记录
3. 如果是商户的登出，则需要删除数据库里其对应的所有OPK(下次登录需要重新上传)
*/
func (s *MysqlLianmiRepository) SignOut(token, username, deviceID string) bool {
	redisConn := s.redisPool.Get()
	defer redisConn.Close()
	var err error

	//取出当前旧的设备的os， clientType， logonAt
	curDeviceHashKey := fmt.Sprintf("devices:%s:%s", username, deviceID)
	isMaster, _ := redis.Bool(redisConn.Do("HGET", curDeviceHashKey))
	curOs, _ := redis.String(redisConn.Do("HGET", curDeviceHashKey, "os"))
	curClientType, _ := redis.Int(redisConn.Do("HGET", curDeviceHashKey, "clientType"))
	curLogonAt, _ := redis.Uint64(redisConn.Do("HGET", curDeviceHashKey, "logonAt"))

	s.logger.Debug("SignOut", zap.Bool("isMaster", isMaster),
		zap.String("username", username),
		zap.String("deviceID", deviceID),
		zap.String("curOs", curOs),
		zap.Int("curClientType", curClientType),
		zap.Uint64("curLogonAt", curLogonAt))

	deviceListKey := fmt.Sprintf("devices:%s", username)

	if isMaster { //如果是主设备
		//查询出所有主从设备
		deviceIDSlice, _ := redis.Strings(redisConn.Do("ZRANGEBYSCORE", deviceListKey, "-inf", "+inf"))
		for index, eDeviceID := range deviceIDSlice {
			s.logger.Debug("查询出所有主从设备", zap.Int("index", index), zap.String("eDeviceID", eDeviceID))
			deviceKey := fmt.Sprintf("DeviceJwtToken:%s", eDeviceID)
			jwtToken, _ := redis.String(redisConn.Do("GET", deviceKey))
			s.logger.Debug("Redis GET ", zap.String("deviceKey", deviceKey), zap.String("jwtToken", jwtToken))

			//向当前主设备及从设备发出踢下线
			if err := s.SendKickedMsgToDevice(jwtToken, username, eDeviceID); err != nil {
				s.logger.Error("Failed to Send Kicked Msg To Device to ProduceChannel", zap.Error(err))
			}

			_, err = redisConn.Do("DEL", deviceKey) //删除deviceKey

			deviceHashKey := fmt.Sprintf("devices:%s:%s", username, eDeviceID)
			_, err = redisConn.Do("DEL", deviceHashKey) //删除deviceHashKey

		}

		//删除所有与之相关的key
		_, err = redisConn.Do("DEL", deviceListKey) //删除deviceListKey

	} else { //如果是从设备

		//删除token
		deviceKey := fmt.Sprintf("DeviceJwtToken:%s", deviceID)
		_, err = redisConn.Do("DEL", deviceKey)

		//删除有序集合里的元素
		//移除单个元素 ZREM deviceListKey {设备id}
		_, err = redisConn.Do("ZREM", deviceListKey, deviceID)

		//删除哈希
		deviceHashKey := fmt.Sprintf("devices:%s:%s", username, deviceID)
		_, err = redisConn.Do("DEL", deviceHashKey)

		//多端登录状态变化事件
		//向其它端发送此从设备离线的事件
		if err := s.SendMultiLoginEventToOtherDevices(false, username, deviceID, curOs, curClientType, curLogonAt); err != nil {
			s.logger.Error("Failed to Send MultiLoginEvent to Other Devices to ProduceChannel", zap.Error(err))
		}
	}

	s.logger.Debug("SignOut end")
	_ = err
	return true
}

func (s *MysqlLianmiRepository) SendKickedMsgToDevice(jwtToken, username, eDeviceID string) error {
	businessType := 2
	businessSubType := 5 //KickedEvent

	kickMsg := &models.Message{}
	kickMsg.UpdateID()
	//构建消息路由, 第一个参数是要处理的业务类型，后端服务器处理完成后，需要用此来拼接topic: {businessTypeName.Frontend}
	kickMsg.BuildRouter("Auth", "", "Auth.Frontend")

	kickMsg.SetJwtToken(jwtToken)
	kickMsg.SetUserName(username)
	kickMsg.SetDeviceID(string(eDeviceID))
	// kickMsg.SetTaskID(uint32(taskId))
	kickMsg.SetBusinessTypeName("Auth")
	kickMsg.SetBusinessType(uint32(businessType))
	kickMsg.SetBusinessSubType(uint32(businessSubType))

	kickMsg.BuildHeader("AuthService", time.Now().UnixNano()/1e6)

	//构造负载数据
	resp := &Auth.KickedEventRsp{
		ClientType: 0,
		Reason:     Auth.KickReason_SamePlatformKick,
		TimeTag:    uint64(time.Now().UnixNano() / 1e6),
	}
	data, _ := proto.Marshal(resp)
	kickMsg.FillBody(data) //网络包的body，承载真正的业务数据

	kickMsg.SetCode(200) //成功的状态码

	//构建数据完成，向dispatcher发送
	topic := "Auth.Frontend"
	rawData, _ := json.Marshal(kickMsg)
	if err := s.nsqClient.Producer.Public(topic, rawData); err == nil {
		s.logger.Info("message succeed send to ProduceChannel", zap.String("topic", topic))
		return err
	} else {
		s.logger.Error(" failed to send message to ProduceChannel", zap.Error(err))
	}
	return nil
}

func (s *MysqlLianmiRepository) ExistsTokenInRedis(deviceID, token string) bool {
	redisConn := s.redisPool.Get()
	defer redisConn.Close()
	deviceKey := fmt.Sprintf("DeviceJwtToken:%s", deviceID)
	if isExists, err := redis.Bool(redisConn.Do("EXISTS", deviceKey)); err != nil {
		s.logger.Error("redisConn GET token Error", zap.Error(err))
		return false
	} else {
		s.logger.Info("redisConn GET token ok ", zap.String("token", token))
		return isExists
	}

}

//生成注册校验码
func (s *MysqlLianmiRepository) GenerateSmsCode(mobile string) bool {
	var err error
	var isExists bool
	redisConn := s.redisPool.Get()
	defer redisConn.Close()
	key := fmt.Sprintf("smscode:%s", mobile)

	if isExists, err = redis.Bool(redisConn.Do("EXISTS", key)); err != nil {
		s.logger.Error("redisConn GET smscode Error", zap.Error(err))
		return false
	}

	if isExists {
		_, err = redisConn.Do("DEL", key) //删除key
	}

	//TODO 调用短信接口发送  暂时固定为123456

	_, err = redisConn.Do("SET", key, "123456") //增加key
	if err != nil {
		s.logger.Error("SET key失败", zap.Error(err))
		return false
	}

	_, err = redisConn.Do("EXPIRE", key, 600) //设置有效期为600秒
	if err != nil {
		s.logger.Error("EXPIRE key 失败", zap.Error(err))
		return false
	}

	s.logger.Debug("GenerateSmsCode, 写入redis成功")

	_ = err

	return true
}

//检测校验码是否正确
func (s *MysqlLianmiRepository) CheckSmsCode(mobile, smscode string) bool {
	var err error
	var isExists bool

	redisConn := s.redisPool.Get()
	defer redisConn.Close()
	key := fmt.Sprintf("smscode:%s", mobile)

	if isExists, err = redis.Bool(redisConn.Do("EXISTS", key)); err != nil {
		s.logger.Error("redisConn GET smscode Error", zap.Error(err))
		return false
	} else {
		if !isExists {
			s.logger.Warn("isExists=false, smscode is expire", zap.String("key", key))
			return false
		} else {
			if smscodeInRedis, err := redis.String(redisConn.Do("GET", key)); err != nil {
				s.logger.Error("redisConn GET smscode Error", zap.Error(err))
				return false
			} else {
				s.logger.Info("redisConn GET smscode ok ", zap.String("smscodeInRedis", smscodeInRedis))
				return smscodeInRedis == smscode
			}
		}
	}
	return false

}

//获取空闲的在线客服id数组
func (s *MysqlLianmiRepository) QueryCustomerServices() ([]*models.CustomerServiceInfo, error) {

	var err error

	redisConn := s.redisPool.Get()
	defer redisConn.Close()

	csList := make([]*models.CustomerServiceInfo, 0)

	csUsernameList, err := redis.Strings(redisConn.Do("ZRANGE", "CustomerServiceList", 0, -1))
	if err != nil {
		return nil, err
	}
	for _, csUsername := range csUsernameList {
		key := fmt.Sprintf("CustomerServiceInfo:%s", csUsername)

		isIdle, _ := redis.Bool(redisConn.Do("HGET", key, "IsIdle"))           //是否空闲
		cstype, _ := redis.Int(redisConn.Do("HGET", key, "Type"))              //账号类型，1-客服，2-技术
		jobNumber, _ := redis.String(redisConn.Do("HGET", key, "JobNumber"))   //工号
		evaluation, _ := redis.String(redisConn.Do("HGET", key, "Evaluation")) //职称
		nickName, _ := redis.String(redisConn.Do("HGET", key, "NickName"))     //呢称
		if isIdle && cstype == 1 {
			csList = append(csList, &models.CustomerServiceInfo{
				Username:   csUsername, //客服或技术人员的注册账号id
				JobNumber:  jobNumber,  //客服或技术人员的工号
				Type:       cstype,     //客服或技术人员的类型， 1-客服，2-技术
				Evaluation: evaluation, //职称, 技术工程师，技术员等
				NickName:   nickName,   //呢称,
			})
		}

	}
	return csList, nil
}

func (s *MysqlLianmiRepository) AddCustomerService(sc *Service.CustomerServiceInfo) ([]*models.CustomerServiceInfo, error) {
	var err error

	redisConn := s.redisPool.Get()
	defer redisConn.Close()

	username := sc.Username

	userKey := fmt.Sprintf("userData:%s", username)
	username2, _ := redis.String(redisConn.Do("HGET", userKey, "Username"))
	if username2 != username {
		return nil, errors.Wrapf(err, "Username is not exists error[username=%s]", username)
	}
	if reply, err := redisConn.Do("ZRANK", "CustomerServiceList", username); err == nil {
		if reply != nil {

			//已经存在，不能重复增加
			return nil, errors.Wrapf(err, "Username is exists error[username=%s]", username)
		}

	}
	csList := make([]*models.CustomerServiceInfo, 0)

	c := new(models.CustomerServiceInfo)

	tx := s.base.GetTransaction()

	if err := tx.Save(c).Error; err != nil {
		s.logger.Error("增加客户技术失败", zap.Error(err))
		tx.Rollback()

	}
	//提交
	tx.Commit()

	if _, err = redisConn.Do("ZADD", "CustomerServiceList", time.Now().UnixNano()/1e6, username); err != nil {
		s.logger.Error("ZADD Error", zap.Error(err))
	}

	_, err = redisConn.Do("HMSET",
		fmt.Sprintf("CustomerServiceInfo:%s", username),
		"Username", sc.Username,
		"IsIdle", false,
		"Type", sc.Type,
		"JobNumber", sc.JobNumber,
		"NickName", sc.NickName,
	)

	csUsernameList, err := redis.Strings(redisConn.Do("ZRANGE", "CustomerServiceList", 0, -1))
	if err != nil {
		return nil, err
	}
	for _, csUsername := range csUsernameList {
		key := fmt.Sprintf("CustomerServiceInfo:%s", csUsername)

		isIdle, _ := redis.Bool(redisConn.Do("HGET", key, "IsIdle"))           //是否空闲
		cstype, _ := redis.Int(redisConn.Do("HGET", key, "Type"))              //账号类型，1-客服，2-技术
		jobNumber, _ := redis.String(redisConn.Do("HGET", key, "JobNumber"))   //工号
		evaluation, _ := redis.String(redisConn.Do("HGET", key, "Evaluation")) //职称
		nickName, _ := redis.String(redisConn.Do("HGET", key, "NickName"))     //呢称
		if isIdle && cstype == 1 {
			csList = append(csList, &models.CustomerServiceInfo{
				Username:   csUsername, //客服或技术人员的注册账号id
				JobNumber:  jobNumber,  //客服或技术人员的工号
				Type:       cstype,     //客服或技术人员的类型， 1-客服，2-技术
				Evaluation: evaluation, //职称, 技术工程师，技术员等
				NickName:   nickName,   //呢称,
			})
		}

	}
	return csList, nil

}

func (s *MysqlLianmiRepository) DeleteCustomerService(sc *Service.CustomerServiceInfo) bool {

	redisConn := s.redisPool.Get()
	defer redisConn.Close()

	username := sc.Username

	userKey := fmt.Sprintf("userData:%s", username)
	username2, _ := redis.String(redisConn.Do("HGET", userKey, "Username"))
	if username2 != username {
		return false
	}
	var (
		gpWhere             = models.CustomerServiceInfo{Username: username}
		customerServiceInfo models.CustomerServiceInfo
	)
	tx := s.base.GetTransaction()
	if err := tx.Where(&gpWhere).Delete(&customerServiceInfo).Error; err != nil {
		s.logger.Error("删除在线客服表失败", zap.Error(err))
		tx.Rollback()
		return false
	}
	tx.Commit()
	return true

}

func (s *MysqlLianmiRepository) UpdateCustomerService(sc *Service.CustomerServiceInfo) ([]*models.CustomerServiceInfo, error) {
	var err error

	redisConn := s.redisPool.Get()
	defer redisConn.Close()

	username := sc.Username

	userKey := fmt.Sprintf("userData:%s", username)
	username2, _ := redis.String(redisConn.Do("HGET", userKey, "Username"))
	if username2 != username {
		return nil, errors.Wrapf(err, "Username is not exists error[username=%s]", username)
	}
	if reply, err := redisConn.Do("ZRANK", "CustomerServiceList", username); err == nil {
		if reply == nil {

			//不存在，必须先增加
			return nil, errors.Wrapf(err, "Username is not exists in list error[username=%s]", username)
		}

	}
	csList := make([]*models.CustomerServiceInfo, 0)

	c := new(models.CustomerServiceInfo)
	if err = s.db.Model(c).Where("username = ?", username).First(c).Error; err != nil {
		return nil, errors.Wrapf(err, "Get customerServiceInfo error[username=%s]", username)
	}

	c.JobNumber = sc.JobNumber
	c.Evaluation = sc.Evaluation
	c.NickName = sc.NickName
	c.Type = int(sc.Type)

	tx := s.base.GetTransaction()

	if err := tx.Save(c).Error; err != nil {
		s.logger.Error("修改客户技术失败", zap.Error(err))
		tx.Rollback()

	}
	//提交
	tx.Commit()

	_, err = redisConn.Do("HMSET",
		fmt.Sprintf("CustomerServiceInfo:%s", username),
		"Username", sc.Username,
		"IsIdle", false,
		"Type", sc.Type,
		"JobNumber", sc.JobNumber,
		"NickName", sc.NickName,
	)

	csUsernameList, err := redis.Strings(redisConn.Do("ZRANGE", "CustomerServiceList", 0, -1))
	if err != nil {
		return nil, err
	}
	for _, csUsername := range csUsernameList {
		key := fmt.Sprintf("CustomerServiceInfo:%s", csUsername)

		isIdle, _ := redis.Bool(redisConn.Do("HGET", key, "IsIdle"))           //是否空闲
		cstype, _ := redis.Int(redisConn.Do("HGET", key, "Type"))              //账号类型，1-客服，2-技术
		jobNumber, _ := redis.String(redisConn.Do("HGET", key, "JobNumber"))   //工号
		evaluation, _ := redis.String(redisConn.Do("HGET", key, "Evaluation")) //职称
		nickName, _ := redis.String(redisConn.Do("HGET", key, "NickName"))     //呢称
		if isIdle && cstype == 1 {
			csList = append(csList, &models.CustomerServiceInfo{
				Username:   csUsername, //客服或技术人员的注册账号id
				JobNumber:  jobNumber,  //客服或技术人员的工号
				Type:       cstype,     //客服或技术人员的类型， 1-客服，2-技术
				Evaluation: evaluation, //职称, 技术工程师，技术员等
				NickName:   nickName,   //呢称,
			})
		}

	}
	return csList, nil

}

func (s *MysqlLianmiRepository) QueryGrades(req *Service.GradeReq, pageIndex int, pageSize int, total *uint64, where interface{}) ([]*models.Grade, error) {
	var grades []*models.Grade

	//构造查询条件

	if req.AppUsername != "" && req.CustomerServiceUsername == "" {
		if err := s.base.GetPages(&models.Grade{AppUsername: req.AppUsername}, &grades, pageIndex, pageSize, total, where); err != nil {
			s.logger.Error("获取客服评分历史失败", zap.Error(err))
		}
	}
	if req.AppUsername == "" && req.CustomerServiceUsername != "" {
		if err := s.base.GetPages(&models.Grade{CustomerServiceUsername: req.CustomerServiceUsername}, &grades, pageIndex, pageSize, total, where); err != nil {
			s.logger.Error("获取客服评分历史失败", zap.Error(err))
		}
	}

	if req.AppUsername != "" && req.CustomerServiceUsername != "" {
		if err := s.base.GetPages(&models.Grade{AppUsername: req.AppUsername, CustomerServiceUsername: req.CustomerServiceUsername}, &grades, pageIndex, pageSize, total, where); err != nil {
			s.logger.Error("获取客服评分历史失败", zap.Error(err))
		}
	}

	return grades, nil
}

//客服人员增加求助记录，以便发给用户评分
func (s *MysqlLianmiRepository) AddGrade(req *Service.AddGradeReq) (string, error) {
	var err error
	var index uint64

	redisConn := s.redisPool.Get()
	defer redisConn.Close()

	key := fmt.Sprintf("CustomerServiceInfo:%s", req.CustomerServiceUsername)

	cstype, err := redis.Int(redisConn.Do("HGET", key, "Type"))              //账号类型，1-客服，2-技术
	jobNumber, err := redis.String(redisConn.Do("HGET", key, "JobNumber"))   //工号
	evaluation, err := redis.String(redisConn.Do("HGET", key, "Evaluation")) //职称
	nickName, err := redis.String(redisConn.Do("HGET", key, "NickName"))     //呢称
	// catalog, err := redis.String(redisConn.Do("HGET", key, "Catalog"))       //分类
	// desc, err := redis.String(redisConn.Do("HGET", key, "Desc"))             //详情
	if err != nil {
		s.logger.Error("HGET失败", zap.Error(err))
		return "", err
	}
	if index, err = redis.Uint64(redisConn.Do("INCR", "CustomerServiceSeq")); err != nil {
		s.logger.Error("INCR失败", zap.Error(err))
		return "", err
	}
	title := fmt.Sprintf("consult-%s-%d", dateutil.GetDateString(), index)
	c := &models.Grade{
		Title:                   title,
		CustomerServiceUsername: req.CustomerServiceUsername,
		JobNumber:               jobNumber,
		Type:                    cstype,
		Evaluation:              evaluation,
		NickName:                nickName,
		Catalog:                 req.Catalog,
		Desc:                    req.Desc,
	}

	tx := s.base.GetTransaction()

	if err := tx.Save(c).Error; err != nil {
		s.logger.Error("增加客户评分失败", zap.Error(err))
		tx.Rollback()

	}
	//提交
	tx.Commit()
	return title, nil

}

func (s *MysqlLianmiRepository) SubmitGrade(req *Service.SubmitGradeReq) error {

	var err error

	c := new(models.Grade)
	if err = s.db.Model(c).Where("title = ?", req.Title).First(c).Error; err != nil {
		return errors.Wrapf(err, "SubmitGrade error[title=%s]", req.Title)
	}

	c.AppUsername = req.AppUsername
	c.GradeNum = int(req.GradeNum)

	tx := s.base.GetTransaction()

	if err := tx.Save(c).Error; err != nil {
		s.logger.Error("用户提交评分保存失败", zap.Error(err))
		tx.Rollback()
		return errors.Wrapf(err, "Submit Grade error[title=%s]", req.Title)
	}
	//提交
	tx.Commit()

	return nil

}

func (s *MysqlLianmiRepository) GetMembershipCardSaleMode(businessUsername string) (int, error) {
	var err error

	c := new(models.User)
	if err = s.db.Model(c).Where("username = ?", businessUsername).First(c).Error; err != nil {
		return 0, errors.Wrapf(err, "GetMembershipCardSaleMode error[businessUsername=%s]", businessUsername)
	}

	return c.MembershipCardSaleMode, nil
}

func (s *MysqlLianmiRepository) SetMembershipCardSaleMode(businessUsername string, saleType int) error {
	var err error

	c := new(models.User)
	if err = s.db.Model(c).Where("username = ?", businessUsername).First(c).Error; err != nil {
		return errors.Wrapf(err, "GetMembershipCardSaleMode error[businessUsername=%s]", businessUsername)
	}

	c.MembershipCardSaleMode = int(saleType)

	tx := s.base.GetTransaction()

	if err := tx.Save(c).Error; err != nil {
		s.logger.Error("用户提交评分保存失败", zap.Error(err))
		tx.Rollback()
		return errors.Wrapf(err, "Submit Grade error[businessUsername=%s]", businessUsername)
	}
	//提交
	tx.Commit()
	return nil
}
