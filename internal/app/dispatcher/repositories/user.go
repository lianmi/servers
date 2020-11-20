/*
多条件查询

gorm 封装map 查询条件
https://blog.csdn.net/qq_28053177/article/details/82187821

go语言对gorm不固定条件查询封装
https://blog.csdn.net/cqims21/article/details/103604914?utm_medium=distribute.pc_relevant_t0.none-task-blog-BlogCommendFromMachineLearnPai2-1.control&depth_1-utm_source=distribute.pc_relevant_t0.none-task-blog-BlogCommendFromMachineLearnPai2-1.control


GORM最佳实践之不定参数的用法
https://jingwei.link/2018/11/10/golang-variadic-with-gorm-2.html#buildcondition-%E5%87%BD%E6%95%B0
*/

package repositories

import (
	"fmt"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/gomodule/redigo/redis"
	"github.com/jinzhu/gorm"
	Auth "github.com/lianmi/servers/api/proto/auth"
	Global "github.com/lianmi/servers/api/proto/global"
	User "github.com/lianmi/servers/api/proto/user"
	"github.com/lianmi/servers/internal/common"
	LMCommon "github.com/lianmi/servers/internal/common"
	"github.com/lianmi/servers/internal/pkg/models"
	"github.com/pkg/errors"
	uuid "github.com/satori/go.uuid"
	"go.uber.org/zap"
)

func (s *MysqlLianmiRepository) GetUser(username string) (p *models.User, err error) {
	p = new(models.User)

	if err = s.db.Model(p).Where(&models.User{
		Username: username,
	}).First(p).Error; err != nil {
		//记录找不到也会触发错误
		return nil, errors.Wrapf(err, "Get user error[Username=%s]", username)
	}
	s.logger.Debug("GetUser run...")
	return
}

func (s *MysqlLianmiRepository) QueryUsers(req *User.QueryUsersReq) ([]*User.User, int64, error) {
	var err error
	var total int64

	var list []*User.User
	where := []interface{}{
		[]interface{}{"user_type", "=", int(req.UserType)},
	}

	db := s.db
	db, err = s.base.BuildWhere(db, where)
	if err != nil {
		s.logger.Error("BuildWhere错误", zap.Error(err))
	}

	db.Find(&list)

	total = int64(len(list))

	return list, total, nil
}

/*
封号
1. 将users表的用户记录的state设置为3
2. 踢出此用户的所有主从设备
*/
func (s *MysqlLianmiRepository) BlockUser(username string) (err error) {

	redisConn := s.redisPool.Get()
	defer redisConn.Close()

	p := new(models.User)
	if err = s.db.Model(p).Where(&models.User{
		Username: username,
	}).First(p).Error; err != nil {
		return errors.Wrapf(err, "Get user error[username=%s]", username)
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

	return nil
}

/*
解封
1. 将users表的用户记录的state设置为1
*/
func (s *MysqlLianmiRepository) DisBlockUser(username string) (p *models.User, err error) {
	p = new(models.User)
	if err = s.db.Model(p).Where(&models.User{
		Username: username,
	}).First(p).Error; err != nil {
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
	var businessUserType int
	var belongBusinessUser string

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

	//如果是普通用户注册，需要查找推荐人及种子商户
	if user.GetUserType() == User.UserType_Ut_Normal {
		if user.ReferrerUsername != "" {
			//查询出邀请人的类型
			referrerKey := fmt.Sprintf("userData:%s", user.ReferrerUsername)
			businessUserType, _ = redis.Int(redisConn.Do("HGET", referrerKey, "UserType"))

			if businessUserType == int(User.UserType_Ut_Normal) { //如果是用户，则查出对应的商户
				belongBusinessUser, _ = redis.String(redisConn.Do("HGET", referrerKey, "BelongBusinessUser"))

			} else if businessUserType == int(User.UserType_Ut_Business) { //商户
				belongBusinessUser, _ = redis.String(redisConn.Do("HGET", referrerKey, "Username"))

			}

			if belongBusinessUser != "" {
				user.BelongBusinessUser = belongBusinessUser
			}

			//查出user.ReferrerUsername对应的 ReferrerUsername, 即推荐人的推荐人
			userLevelTwo, _ := redis.String(redisConn.Do("HGET", referrerKey, "ReferrerUsername"))

			//查出userLevelTwo对应的 ReferrerUsername, 即推荐人的推荐人的推荐人
			userLevelThree, _ := redis.String(redisConn.Do("HGET", fmt.Sprintf("userData:%s", userLevelTwo), "ReferrerUsername"))

			//会员层级表
			distribution := &models.Distribution{
				Username:           user.Username,         //用户注册账号id
				BusinessUsername:   belongBusinessUser,    //归属的商户注册账号id
				UsernameLevelOne:   user.ReferrerUsername, //向后的一级, 即推荐人
				UsernameLevelTwo:   userLevelTwo,          //向后的二级, 即推荐人的推荐人
				UsernameLevelThree: userLevelThree,        //向后的三级, 即推荐人的推荐人的推荐人
			}
			s.logger.Debug("distribution的值",
				zap.String("Username", user.Username),
				zap.String("BusinessUsername", belongBusinessUser),
				zap.String("UsernameLevelOne", user.ReferrerUsername),
				zap.String("UsernameLevelTwo", userLevelTwo),
				zap.String("UsernameLevelThree", userLevelThree),
			)

			//使用事务同时增加Distribution数据
			tx := s.base.GetTransaction()
			if err := tx.Save(distribution).Error; err != nil {
				s.logger.Error("增加Distribution表失败", zap.Error(err))
				tx.Rollback()
				return err
			}

			//提交
			tx.Commit()

		}
	} else if user.GetUserType() == User.UserType_Ut_Business {

		//如果是商户注册，则无须记录推荐人，种子商户是自己
		user.BelongBusinessUser = user.Username

		//网点商户自动建群
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

	s.logger.Debug("注册用户成功", zap.String("Username", user.Username))
	return nil

}

//重置密码
func (s *MysqlLianmiRepository) ResetPassword(mobile, password string, user *models.User) (err error) {

	redisConn := s.redisPool.Get()
	defer redisConn.Close()

	if err = s.db.Model(user).Where(&models.User{
		Mobile: mobile,
	}).First(user).Error; err != nil {
		return errors.Wrapf(err, "Query user error[mobile=%s]", mobile)
	}
	//记录不存在错误(RecordNotFound)，返回false
	if gorm.IsRecordNotFoundError(err) {
		return err
	}

	//替换旧密码
	user.Password = password

	tx := s.base.GetTransaction()
	if err := tx.Save(user).Error; err != nil {
		s.logger.Error("更新用户密码失败", zap.Error(err))
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
	if err = s.base.First(&where, &user); err != nil {
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

		// _ = err

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

		// _ = err
	}

	//当前所有主从设备数量
	count, _ := redis.Int(redisConn.Do("ZCARD", deviceListKey))
	s.logger.Debug("当前所有主从设备数量", zap.Int("count", count))

	return true
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
	kickMsg.SetBusinessTypeName("Auth")
	kickMsg.SetBusinessType(uint32(businessType))
	kickMsg.SetBusinessSubType(uint32(businessSubType))

	kickMsg.BuildHeader("Dispatcher", time.Now().UnixNano()/1e6)

	//构造负载数据
	resp := &Auth.KickedEventRsp{
		ClientType: 0,
		Reason:     Auth.KickReason_SamePlatformKick,
		TimeTag:    uint64(time.Now().UnixNano() / 1e6),
	}
	data, _ := proto.Marshal(resp)
	kickMsg.FillBody(data) //网络包的body，承载真正的业务数据

	kickMsg.SetCode(200) //成功的状态码

	//构建数据完成，向NsqChan发送
	s.multiChan.NsqChan <- kickMsg
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
		s.logger.Debug("redisConn GET token ok ", zap.String("token", token))
		return isExists
	}

}

//生成注册校验码, 一个手机号只能一天获取5次
func (s *MysqlLianmiRepository) GenerateSmsCode(mobile string) bool {
	var err error
	var isExists bool
	var count uint64
	redisConn := s.redisPool.Get()
	defer redisConn.Close()
	key := fmt.Sprintf("smscode:%s", mobile)
	keyCount := fmt.Sprintf("smscode_count:%s", mobile)

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

	_, err = redisConn.Do("INCR", keyCount) //增加次数
	if err != nil {
		s.logger.Error("INCR keyCount 失败", zap.Error(err))
		return false
	}

	if count, err = redis.Uint64(redisConn.Do("INCR", keyCount)); err != nil {
		s.logger.Error("INCR keyCount 失败", zap.Error(err))
		return false
	}

	if count > LMCommon.SMSCOUNT {
		_, err = redisConn.Do("EXPIRE", keyCount, 12*3600) //设置失效时间为12小时
		if err != nil {
			s.logger.Error("EXPIRE keyCount 失败", zap.Error(err))
			return false
		}
		s.logger.Warn("此手机已经超过了上限")
		return false
	}

	_, err = redisConn.Do("EXPIRE", key, 70) //设置有效期为70秒
	if err != nil {
		s.logger.Error("EXPIRE key 失败", zap.Error(err))
		return false
	}
	if isExists, err = redis.Bool(redisConn.Do("EXISTS", key)); err != nil {
		s.logger.Error("redisConn GET smscode Error", zap.Error(err))
		return false
	}

	if isExists {

		s.logger.Debug("GenerateSmsCode, 生成注册校验码, 写入redis成功", zap.String("key", key))

	} else {
		s.logger.Warn("GenerateSmsCode, 生成注册校验码失败", zap.String("key", key))
	}

	_ = err

	return true
}

func (s *MysqlLianmiRepository) GetUsernameByMobile(mobile string) (string, error) {
	var err error
	p := new(models.User)

	if err = s.db.Model(p).Where(&models.User{
		Mobile: mobile,
	}).First(p).Error; err != nil {
		s.logger.Error("MySQL里读取错误或记录不存在", zap.Error(err))
		return "", errors.Wrapf(err, "Get username error[mobile=%s]", mobile)
	} else {
		return p.Username, nil
	}
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
				_, err = redisConn.Do("EXPIRE", key, 600) //设置有效期为600秒, 以便注册或登陆的时候校验
				if err != nil {
					s.logger.Error("EXPIRE key 失败", zap.Error(err))
					return false
				}
				return smscodeInRedis == smscode
			}
		}
	}
	return false

}

//修改用户资料
func (s *MysqlLianmiRepository) SaveUser(user *models.User) error {
	//使用事务同时更新用户数据
	tx := s.base.GetTransaction()

	if err := tx.Save(user).Error; err != nil {
		s.logger.Error("更新用户表失败", zap.Error(err))
		tx.Rollback()

	}
	//提交
	tx.Commit()

	return nil
}

//修改用户标签
func (s *MysqlLianmiRepository) SaveTag(tag *models.Tag) error {
	//使用事务同时更新Tag数据
	tx := s.base.GetTransaction()

	if err := tx.Save(tag).Error; err != nil {
		s.logger.Error("更新tag表失败", zap.Error(err))
		tx.Rollback()

	}
	//提交
	tx.Commit()

	return nil
}

//修改或增加店铺资料
func (s *MysqlLianmiRepository) SaveStore(req *User.Store) error {
	var err error
	var storeUUID string
	var state int
	sel := "id"
	p := new(models.Store)

	redisConn := s.redisPool.Get()
	defer redisConn.Close()

	//判断商户的注册id的合法性以及是否封禁等
	userData := new(models.User)

	userKey := fmt.Sprintf("userData:%s", req.BusinessUsername)
	if result, err := redis.Values(redisConn.Do("HGETALL", userKey)); err == nil {
		if err := redis.ScanStruct(result, userData); err != nil {

			s.logger.Error("错误：ScanStruct", zap.Error(err))
			return errors.Wrapf(err, "查询redis出错[Businessusername=%s]", req.BusinessUsername)

		}
	}
	// 判断是否是商户类型
	if userData.UserType != 2 {
		s.logger.Error("错误：此注册账号id不是商户类型")
		return errors.Wrapf(err, "此注册账号id不是商户类型[Businessusername=%s]", req.BusinessUsername)
	}

	//判断是否被封禁
	if userData.State == LMCommon.UserBlocked {
		s.logger.Debug("User is blocked", zap.String("Businessusername", req.BusinessUsername))
		return errors.Wrapf(err, "User is blocked[Businessusername=%s]", req.BusinessUsername)
	}

	//先查询对应的记录是否存在
	err = s.base.First(&models.Store{
		BusinessUsername: req.BusinessUsername,
	}, &p, sel)

	//记录不存在错误(RecordNotFound)，返回false
	if gorm.IsRecordNotFoundError(err) {
		storeUUID = uuid.NewV4().String()
		state = 0
	} else {
		s.db.Model(p).Where(&models.Store{
			BusinessUsername: req.BusinessUsername,
		}).First(p)
		storeUUID = p.StoreUUID
		state = p.State
	}
	if state == 1 {
		return errors.Wrapf(err, "已经审核通过的不能修改资料[Businessusername=%s]", req.BusinessUsername)
	}

	store := &models.Store{
		StoreUUID:         storeUUID,              //店铺的uuid
		StoreType:         int(req.StoreType),     //店铺类型,对应Global.proto里的StoreType枚举
		BusinessUsername:  req.BusinessUsername,   //商户注册号
		Introductory:      req.Introductory,       //商店简介 Text文本类型
		Province:          req.Province,           //省份, 如广东省
		City:              req.City,               //城市，如广州市
		County:            req.County,             //区，如天河区
		Street:            req.Street,             //街道
		Address:           req.Address,            //地址
		Branchesname:      req.Branchesname,       //网点名称
		LegalPerson:       req.LegalPerson,        //法人姓名
		LegalIdentityCard: req.LegalIdentityCard,  //法人身份证
		Longitude:         req.Longitude,          //商户地址的经度
		Latitude:          req.Latitude,           //商户地址的纬度
		WeChat:            req.Wechat,             //商户联系人微信号
		Keys:              req.Keys,               //商户经营范围搜索关键字
		LicenseURL:        req.BusinessLicenseUrl, //商户营业执照阿里云url
	}

	//使用事务同时更新Tag数据
	tx := s.base.GetTransaction()

	if err := tx.Save(store).Error; err != nil {
		s.logger.Error("更新Store表失败", zap.Error(err))
		tx.Rollback()

	}
	//提交
	tx.Commit()

	return nil

}

func (s *MysqlLianmiRepository) GetStore(businessUsername string) (*User.Store, error) {
	var err error
	p := new(models.Store)
	if err = s.db.Model(p).Where(&models.Store{
		BusinessUsername: businessUsername,
	}).First(p).Error; err != nil {
		s.logger.Error("MySQL里读取错误或记录不存在", zap.Error(err))
		return nil, errors.Wrapf(err, "Query error[BusinessUsername=%s]", businessUsername)
	}

	redisConn := s.redisPool.Get()
	defer redisConn.Close()

	//判断商户的注册id的合法性以及是否封禁等
	avatar, err := redis.String(redisConn.Do("HGET", fmt.Sprintf("userData:%s", businessUsername), "Avatar"))
	if err != nil {
		s.logger.Error("HGET Avatar error", zap.Error(err))
		return nil, errors.Wrapf(err, "Query error[BusinessUsername=%s]", businessUsername)
	}

	return &User.Store{
		StoreUUID:          p.StoreUUID,                   //店铺的uuid
		StoreType:          Global.StoreType(p.StoreType), //店铺类型,对应Global.proto里的StoreType枚举
		BusinessUsername:   p.BusinessUsername,            //商户注册号
		Avatar:             avatar,                        //头像
		Introductory:       p.Introductory,                //商店简介 Text文本类型
		Province:           p.Province,                    //省份, 如广东省
		City:               p.City,                        //城市，如广州市
		County:             p.County,                      //区，如天河区
		Street:             p.Street,                      //街道
		Address:            p.Address,                     //地址
		Branchesname:       p.Branchesname,                //网点名称
		LegalPerson:        p.LegalPerson,                 //法人姓名
		LegalIdentityCard:  p.LegalIdentityCard,           //法人身份证
		Longitude:          p.Longitude,                   //商户地址的经度
		Latitude:           p.Latitude,                    //商户地址的纬度
		Wechat:             p.WeChat,                      //商户联系人微信号
		Keys:               p.Keys,                        //商户经营范围搜索关键字
		BusinessLicenseUrl: p.LicenseURL,                  //商户营业执照阿里云url
		CreatedAt:          uint64(p.CreatedAt),
		UpdatedAt:          uint64(p.UpdatedAt),
	}, nil

}

//根据gps位置获取一定范围内的店铺列表
func (s *MysqlLianmiRepository) GetStores(req *User.QueryStoresNearbyReq) ([]*User.Store, error) {

	var err error
	// var pageIndex, pageSize int
	// total := new(uint64)
	// var maps string
	// if req.Province != "" {
	// 	maps = fmt.Sprintf("Province= %s and created_at <= %d", req.Province)
	// }

	//
	// var store models.Store
	// maps := make(map[string]interface{})

	// if req.StoreType > 0 {
	// 	maps["store_type"] = int(req.StoreType)

	// }
	// maps["state in"] = []int{1, 2}

	// conditionString, values, _ := s.base.BuildCondition(maps)
	// // err = s.base.First(conditionString, values, &store, sel)
	// var stores []*models.Store
	// if err := s.base.GetPages(&models.Store{}, &stores, pageIndex, pageSize, total, conditionString); err != nil {
	// 	s.logger.Error("获取店铺列表失败", zap.Error(err))
	// }

	var list []*User.Store
	where := make([]interface{}, 0)
	if req.StoreType > 0 {
		where = append(where, []interface{}{"store_type", "=", int(req.StoreType)})

	}
	// where := []interface{}{
	// 	// []interface{}{"state", "in", []int{1, 2}},
	// 	[]interface{}{"store_type", "=", int(req.StoreType)},
	// }

	db := s.db
	db, err = s.base.BuildWhere(db, where)
	if err != nil {
		s.logger.Error("BuildWhere错误", zap.Error(err))
	}

	db.Find(&list)
	return list, nil
}
