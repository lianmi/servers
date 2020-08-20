package repositories

import (
	"fmt"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/gomodule/redigo/redis"
	"github.com/jinzhu/gorm"
	Auth "github.com/lianmi/servers/api/proto/auth"
	pb "github.com/lianmi/servers/api/proto/user"
	"github.com/lianmi/servers/internal/app/authservice/kafkaBackend"
	"github.com/lianmi/servers/internal/common"
	"github.com/lianmi/servers/internal/pkg/models"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

type UsersRepository interface {
	GetUser(ID uint64) (p *models.User, err error)
	BlockUser(ID uint64) (p *models.User, err error)
	Register(user *models.User) (err error)
	Resetpwd(mobile, password string, user *models.User) error
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

	//获取用户信息
	GetUsers(PageNum int, PageSize int, total *uint64, where interface{}) []*models.User

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
}

type MysqlUsersRepository struct {
	logger    *zap.Logger
	db        *gorm.DB
	redisPool *redis.Pool
	kafka     *kafkaBackend.KafkaClient
	base      *BaseRepository
}

func NewMysqlUsersRepository(logger *zap.Logger, db *gorm.DB, redisPool *redis.Pool, kc *kafkaBackend.KafkaClient) UsersRepository {
	return &MysqlUsersRepository{
		logger:    logger.With(zap.String("type", "UsersRepository")),
		db:        db,
		redisPool: redisPool,
		kafka:     kc,
		base:      NewBaseRepository(logger, db),
	}
}

func (s *MysqlUsersRepository) GetUser(ID uint64) (p *models.User, err error) {
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
func (s *MysqlUsersRepository) BlockUser(ID uint64) (p *models.User, err error) {
	// p = new(models.User)
	// if err = s.db.Model(p).Where("id = ?", ID).First(p).Error; err != nil {
	// 	return nil, errors.Wrapf(err, "Get user error[id=%d]", ID)
	// }
	// s.logger.Debug("BlockUser run...")

	//TODO
	return
}

//注册用户，username需要唯一
func (s *MysqlUsersRepository) Register(user *models.User) (err error) {
	//获取redis里最新id， 生成唯一的username
	var newIndex uint64

	redisConn := s.redisPool.Get()
	defer redisConn.Close()

	if newIndex, err = redis.Uint64(redisConn.Do("INCR", "usernameindex")); err != nil {
		s.logger.Error("redisConn GET usernameindex Error", zap.Error(err))
		return err
	}

	if user.GetUserType() == pb.UserType_Ut_Operator { //10086
		user.Username = fmt.Sprintf("admin%d", newIndex)
	} else {
		user.Username = fmt.Sprintf("id%d", newIndex)
	}

	if err := s.base.Create(user); err != nil {
		s.logger.Error("注册用户失败")
		return err
	}

	//创建redis的sync:{用户账号} myInfoAt 时间戳
	myInfoAtKey := fmt.Sprintf("sync:%s", user.Username)
	redisConn.Do("HSET", myInfoAtKey, "myInfoAt", time.Now().Unix())

	s.logger.Debug("注册用户成功", zap.String("Username", user.Username))
	return nil

}

//重置密码
func (s *MysqlUsersRepository) Resetpwd(mobile, password string, user *models.User) error {

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

// 获取用户角色
func (s *MysqlUsersRepository) GetUserRoles(where interface{}) []*models.Role {
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
func (s *MysqlUsersRepository) CheckUser(isMaster bool, smscode, username, password, deviceID, os string, clientType int) bool {
	var err error
	redisConn := s.redisPool.Get()
	defer redisConn.Close()

	where := models.User{Username: username}
	var user models.User
	if err := s.base.First(where, &user); err != nil {
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
		s.logger.Error("校验码不匹配")
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

	//当前所有主从设备数量 ZCOUNT devices:lsj001 -inf +inf
	count, _ := redis.Int(redisConn.Do("ZCOUNT", deviceListKey, "-inf", "+inf"))
	s.logger.Debug("当前所有主从设备数量", zap.Int("count", count))

	return true
}

//向其它端发送此从设备MultiLoginEvent事件
func (s *MysqlUsersRepository) SendMultiLoginEventToOtherDevices(isOnline bool, username, deviceID, curOs string, curClientType int, curLogonAt uint64) (err error) {
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
		if err := s.kafka.Produce(topic, targetMsg); err == nil {
			s.logger.Info("message succeed send to ProduceChannel", zap.String("topic", topic))
		} else {
			s.logger.Error(" failed to send message to ProduceChannel", zap.Error(err))
		}
	}

	return nil
}

func (s *MysqlUsersRepository) AddRole(role *models.Role) (err error) {
	if err := s.db.Create(role).Error; err != nil {
		s.logger.Error("新建用户角色失败")
		return err
	} else {
		return nil
	}
}

func (s *MysqlUsersRepository) DeleteUser(id uint64) bool {
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

func (s *MysqlUsersRepository) GetUserAvatar(where interface{}, sel string) string {
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

func (s *MysqlUsersRepository) GetUserID(where interface{}) uint64 {
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
func (s *MysqlUsersRepository) GetTokenByUserId(where interface{}) string {
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

func (s *MysqlUsersRepository) GetUsers(PageNum int, PageSize int, total *uint64, where interface{}) []*models.User {
	var users []*models.User
	if err := s.base.GetPages(&models.User{}, &users, PageNum, PageSize, total, where); err != nil {
		s.logger.Error("获取用户信息失败", zap.Error(err))
	}
	return users
}

//判断用户名是否已存在
func (s *MysqlUsersRepository) ExistUserByName(username string) bool {
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
func (s *MysqlUsersRepository) ExistUserByMobile(mobile string) bool {
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
func (s *MysqlUsersRepository) UpdateUser(user *models.User, role *models.Role) bool {
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
func (s *MysqlUsersRepository) GetUserByID(id int) *models.User {
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
func (s *MysqlUsersRepository) SaveUserToken(username, deviceID string, token string, expire time.Time) bool {

	redisConn := s.redisPool.Get()
	defer redisConn.Close()

	deviceKey := fmt.Sprintf("DeviceJwtToken:%s", deviceID)
	_, err := redisConn.Do("SET", deviceKey, token) //deviceID关联到token, mqtt消息必须要验证这个

	_, err = redisConn.Do("EXPIRE", deviceKey, common.ExpireTime) //过期时间

	//重新读出token，看看是否可以读出

	if tokenInRedis, err := redis.String(redisConn.Do("GET", deviceKey)); err != nil {
		s.logger.Error("重新读出token失败", zap.Error(err))
		return false
	} else {
		isEqueal := tokenInRedis == token
		s.logger.Debug("重新读出token成功", zap.String("tokenInRedis", tokenInRedis), zap.Bool("isEqueal", isEqueal))
	}
	_ = err
	return true
}

/*
登出
1. 如果是主设备，则踢出此用户的所有主从设备， 如果仅仅是从设备，就删除自己的数据
2. 删除redis里的此用户的哈希记录
3. 如果是商户的登出，则需要删除数据库里其对应的所有OPK(下次登录需要重新上传)
*/
func (s *MysqlUsersRepository) SignOut(token, username, deviceID string) bool {
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

func (s *MysqlUsersRepository) SendKickedMsgToDevice(jwtToken, username, eDeviceID string) error {
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
	if err := s.kafka.Produce(topic, kickMsg); err == nil {
		s.logger.Info("message succeed send to ProduceChannel", zap.String("topic", topic))
		return err
	} else {
		s.logger.Error(" failed to send message to ProduceChannel", zap.Error(err))
	}
	return nil
}

func (s *MysqlUsersRepository) ExistsTokenInRedis(deviceID, token string) bool {
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
func (s *MysqlUsersRepository) GenerateSmsCode(mobile string) bool {
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
		err = redisConn.Send("DEL", key) //删除key
	}

	//TODO 调用短信接口发送  暂时固定为123456

	err = redisConn.Send("SET", key, "123456") //增加key

	err = redisConn.Send("EXPIRE", key, 300) //设置有效期为300秒

	_ = err

	//一次性写入到Redis
	if err := redisConn.Flush(); err != nil {
		s.logger.Error("写入redis失败", zap.Error(err))
		return false
	}
	s.logger.Debug("GenerateSmsCode, 写入redis成功")
	return true
}

//检测校验码是否正确
func (s *MysqlUsersRepository) CheckSmsCode(mobile, smscode string) bool {
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
			s.logger.Warn("smscode is expire")
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
