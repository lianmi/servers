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
	"strings"
	"time"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"github.com/gomodule/redigo/redis"

	// Global "github.com/lianmi/servers/api/proto/global"
	User "github.com/lianmi/servers/api/proto/user"
	LMCommon "github.com/lianmi/servers/internal/common"
	"github.com/lianmi/servers/internal/pkg/models"
	"github.com/pkg/errors"

	// uuid "github.com/satori/go.uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func (s *MysqlLianmiRepository) GetUser(username string) (user *models.User, err error) {
	user = new(models.User)

	if err = s.db.Model(user).Where(&models.User{
		UserBase: models.UserBase{
			Username: username,
		},
	}).First(user).Error; err != nil {
		//记录找不到也会触发错误
		return nil, errors.Wrapf(err, "Get user error[Username=%s]", username)
	}
	s.logger.Debug("GetUser run...")
	return
}

func (s *MysqlLianmiRepository) GetUserDb(objname string) (string, error) {
	// 超级用户创建OSSClient实例。
	client, err := oss.New(LMCommon.Endpoint, LMCommon.SuperAccessID, LMCommon.SuperAccessKeySecret)

	if err != nil {
		return "", errors.Wrapf(err, "oss.New失败[objname=%s]", objname)

	}

	// 获取存储空间。
	bucket, err := client.Bucket(LMCommon.BucketName)
	if err != nil {
		return "", errors.Wrapf(err, "client.Bucket失败[objname=%s]", objname)

	}

	//生成签名URL下载链接， 300s后过期

	signedURL, err := bucket.SignURL(objname, oss.HTTPGet, 300)
	if err != nil {
		s.logger.Error("bucket.SignURL error", zap.Error(err))
		return "", errors.Wrapf(err, "bucket.SignURL失败[objname=%s]", objname)
	} else {
		s.logger.Debug("bucket.SignURL 生成成功")

	}
	return signedURL, nil
}

//根据注册用户id获取redis里此用户的缓存
func (s *MysqlLianmiRepository) GetUserDataFromRedis(username string) (p *models.UserBase, err error) {

	redisConn := s.redisPool.Get()
	defer redisConn.Close()

	userKey := fmt.Sprintf("userData:%s", username)

	userBaseData := new(models.UserBase)

	if result, err := redis.Values(redisConn.Do("HGETALL", userKey)); err == nil {
		if err := redis.ScanStruct(result, userBaseData); err != nil {
			return nil, err
		} else {
			return userBaseData, nil
		}
	}
	return nil, err
}

//多条件不定参数批量分页获取用户列表
func (s *MysqlLianmiRepository) QueryUsers(req *User.QueryUsersReq) (*User.QueryUsersResp, error) {
	var err error
	var total int64
	page := int(req.Page)
	pageSize := int(req.PageSize)

	var list []*models.User
	var where []interface{}

	where = append(where, []interface{}{"user_type", "=", int(req.UserType)})
	where = append(where, []interface{}{"user_type", "=", int(req.State)})

	if req.Mobile != "" {
		where = append(where, []interface{}{"mobile", "=", req.Mobile})
	}
	if req.ReferrerUsername != "" {
		where = append(where, []interface{}{"referrer_username", "=", req.ReferrerUsername})
	}
	if req.TrueName != "" {
		where = append(where, []interface{}{"true_name", "=", req.TrueName})
	}

	db2 := s.db
	db2, err = s.base.BuildWhere(db2, where)
	if err != nil {
		s.logger.Error("BuildWhere错误", zap.Error(err))
	}

	db2.Find(&list)

	total = int64(len(list))

	var users []models.User
	userModel := new(models.User)

	s.db.Model(&userModel).Find(&users, "user_type=?", 0)

	// db2.Model(&userModel).Scopes(IsNormalUser, Paginate(page, pageSize)).Find(&users)
	// db2.Model(&userModel).Scopes(IsBusinessUser, Paginate(page, pageSize)).Find(&users)
	// db2.Model(&userModel).Scopes(IsPreBusinessUser, LegalPerson([]string{"杜老板"}), Paginate(page, pageSize)).Find(&users)
	db2.Model(&userModel).Scopes(Paginate(page, pageSize)).Find(&users)

	// count =
	s.logger.Debug("分页显示users列表, len: ", zap.Int("len", len(users)))

	for _, user := range users {
		// log.Printf("idx=%d, username=%s, mobile=%d\n", idx, user.Username, user.Mobile)
		s.logger.Debug("分页显示users列表 ",
			zap.String("username", user.Username),
			zap.String("Mobile", user.Mobile),
		)
	}

	//测试分页
	// s.logger.Error("测试分页")
	// s.usersPage(1, 20)

	// return list, total, nil

	var avatar string

	resp := &User.QueryUsersResp{
		Total: uint64(total),
	}

	for _, userData := range users {
		if (userData.Avatar != "") && !strings.HasPrefix(userData.Avatar, "http") {

			avatar = LMCommon.OSSUploadPicPrefix + userData.Avatar + "?x-oss-process=image/resize,w_50/quality,q_50"
		}

		resp.Users = append(resp.Users, &User.User{
			Username: userData.Username,
			Gender:   User.Gender(userData.Gender),
			Nick:     userData.Nick,
			Avatar:   avatar,
			Label:    userData.Label,
			// Mobile:       userData.Mobile, 隐私
			// Email:        userData.Email,
			// UserType:     User.UserType(userData.UserType),
			// Extend:       userData.Extend,
			// TrueName:     userData.TrueName,
			// IdentityCard: userData.IdentityCard,
			// Province:     userData.Province,
			// City:         userData.City,
			// County:       userData.County,
			// Street:       userData.Street,
			// Address:      userData.Address,
		})

	}

	return resp, nil

}

/*
封号
1. 将users表的用户记录的state设置为3
2. 踢出此用户的所有主从设备
*/
func (s *MysqlLianmiRepository) BlockUser(username string) (err error) {

	redisConn := s.redisPool.Get()
	defer redisConn.Close()

	result := s.db.Model(&models.User{}).Where(&models.User{
		UserBase: models.UserBase{
			Username: username,
		},
	}).Update("status", 2) //将Status变为 2, 1-正常， 2-封号

	//updated records count
	s.logger.Debug("BlockUser result: ", zap.Int64("RowsAffected", result.RowsAffected), zap.Error(result.Error))

	if result.Error != nil {
		s.logger.Error("封号失败", zap.Error(result.Error))
		return result.Error
	}

	//将此用户所在在线设备全部踢出
	deviceListKey := fmt.Sprintf("devices:%s", username)

	//删除所有与之相关的key
	_, err = redisConn.Do("DEL", deviceListKey) //删除deviceListKey

	s.logger.Debug("BlockUser run.")

	return nil
}

/*
解封
1. 将users表的用户记录的state设置为1
*/
func (s *MysqlLianmiRepository) DisBlockUser(username string) (err error) {

	result := s.db.Model(&models.User{}).Where(&models.User{
		UserBase: models.UserBase{
			Username: username,
		},
	}).Update("status", 1) //将Status变为 1, 1-正常， 2-封号

	//updated records count
	s.logger.Debug("DisBlockUser result: ", zap.Int64("RowsAffected", result.RowsAffected), zap.Error(result.Error))

	if result.Error != nil {
		s.logger.Error("解封失败", zap.Error(result.Error))
		return result.Error
	}

	return
}

//注册用户，username需要唯一
func (s *MysqlLianmiRepository) Register(user *models.User) (err error) {
	//获取redis里最新id， 生成唯一的username
	var newIndex uint64
	var userType int
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

	// 判断推荐人是否为空
	if user.ReferrerUsername != "" {
		//查询出推荐人的类型
		referrerKey := fmt.Sprintf("userData:%s", user.ReferrerUsername)
		userType, _ = redis.Int(redisConn.Do("HGET", referrerKey, "UserType"))

		if userType == int(User.UserType_Ut_Normal) { //如果是用户，则查出对应的商户
			belongBusinessUser, _ = redis.String(redisConn.Do("HGET", referrerKey, "BelongBusinessUser"))

		} else if userType == int(User.UserType_Ut_Business) { //商户
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
		s.logger.Debug("Distribution表",
			zap.String("用户注册账号id, Username", user.Username),
			zap.String("归属的商户注册账号id, BusinessUsername", belongBusinessUser),
			zap.String("向后的一级, 即推荐人, UsernameLevelOne", user.ReferrerUsername),
			zap.String("向后的二级, 即推荐人的推荐人, UsernameLevelTwo", userLevelTwo),
			zap.String("向后的三级, 即推荐人的推荐人的推荐人,   UsernameLevelThree", userLevelThree),
		)

		//增加记录
		if err := s.db.Clauses(clause.OnConflict{DoNothing: true}).Create(distribution).Error; err != nil {
			s.logger.Error("增加Distribution表失败", zap.Error(err))
			return err
		} else {
			s.logger.Debug("增加Distribution表成功")
		}

	}

	//如果是商户注册
	if user.GetUserType() == User.UserType_Ut_Business {

		//商户的下属是自己
		user.BelongBusinessUser = user.Username

		//网点商户自动建群
		var newTeamIndex uint64
		if newTeamIndex, err = redis.Uint64(redisConn.Do("INCR", "TeamIndex")); err != nil {
			s.logger.Error("redisConn GET TeamIndex Error", zap.Error(err))
			return err
		}
		pTeam := new(models.Team)
		pTeam.TeamID = fmt.Sprintf("team%d", newTeamIndex) //群id， 自动生成
		pTeam.Teamname = fmt.Sprintf("team%d", newTeamIndex)
		pTeam.Nick = fmt.Sprintf("%s的群", user.Nick)
		pTeam.Owner = user.Username
		pTeam.Type = 1
		pTeam.VerifyType = 1
		pTeam.InviteMode = 1

		//默认的设置
		pTeam.Status = 1 //Init(1) - 初始状态,审核中 Normal(2) - 正常状态 Blocked(3) - 封禁状态
		pTeam.MemberLimit = LMCommon.PerTeamMembersLimit
		pTeam.MemberNum = 1  //刚刚建群是只有群主1人
		pTeam.MuteType = 1   //None(1) - 所有人可发言
		pTeam.InviteMode = 1 //邀请模式,初始为1

		//创建群数据 增加记录
		if err := s.db.Clauses(clause.OnConflict{DoNothing: true}).Create(&pTeam).Error; err != nil {
			s.logger.Error("Register, failed to upsert team", zap.Error(err))
			return err
		} else {
			s.logger.Debug("CreateTeam, upsert team succeed")
		}

	}

	//将用户信息缓存到redis里
	userKey := fmt.Sprintf("userData:%s", user.Username)
	if _, err := redisConn.Do("HMSET", redis.Args{}.Add(userKey).AddFlat(user.UserBase)...); err != nil {
		s.logger.Error("错误：HMSET", zap.Error(err))
	}

	if err := s.base.Create(user); err != nil {
		s.logger.Error("db写入错误，注册用户失败")
		return err
	}

	//创建redis的sync:{用户账号} myInfoAt 时间戳
	//myInfoAt, friendsAt, friendUsersAt, teamsAt, tagsAt, watchAt, productAt

	syncKey := fmt.Sprintf("sync:%s", user.Username)
	redisConn.Do("HSET", syncKey, "myInfoAt", time.Now().UnixNano()/1e6)
	redisConn.Do("HSET", syncKey, "friendsAt", time.Now().UnixNano()/1e6)
	redisConn.Do("HSET", syncKey, "friendUsersAt", time.Now().UnixNano()/1e6)
	redisConn.Do("HSET", syncKey, "teamsAt", time.Now().UnixNano()/1e6)
	redisConn.Do("HSET", syncKey, "tagsAt", time.Now().UnixNano()/1e6)
	redisConn.Do("HSET", syncKey, "watchAt", time.Now().UnixNano()/1e6)
	redisConn.Do("HSET", syncKey, "productAt", time.Now().UnixNano()/1e6)

	s.logger.Debug("注册用户成功", zap.String("Username", user.Username))
	return nil

}

//重置密码
func (s *MysqlLianmiRepository) ResetPassword(mobile, password string, user *models.User) (err error) {

	redisConn := s.redisPool.Get()
	defer redisConn.Close()

	if err = s.db.Model(user).Where(&models.User{
		UserBase: models.UserBase{
			Mobile: mobile,
		},
	}).First(user).Error; err != nil {
		//记录不存在错误(RecordNotFound)，返回false
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
		return errors.Wrapf(err, "Query user error[mobile=%s]", mobile)
	}

	result := s.db.Model(&models.User{}).Where(&models.User{
		UserBase: models.UserBase{
			Mobile: mobile,
		},
	}).Update("password", password) //替换旧密码

	//updated records count
	s.logger.Debug("ResetPassword result: ",
		zap.Int64("RowsAffected", result.RowsAffected),
		zap.Error(result.Error))

	if result.Error != nil {
		s.logger.Error("重置密码失败", zap.Error(result.Error))
		return result.Error
	}

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

//  使用手机及短信验证码登录
func (s *MysqlLianmiRepository) LoginBySmscode(username, mobile, smscode, deviceID, os string, clientType int) (bool, string) {
	s.logger.Debug("LoginBySmsode start...")
	var err error
	redisConn := s.redisPool.Get()
	defer redisConn.Close()

	where := models.User{
		UserBase: models.UserBase{
			Mobile: mobile,
		},
	}
	var user models.User

	err = s.db.Model(&models.User{}).Where(&where).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			s.logger.Error("错误：用户手机不存在")
			return false, ""
		} else {

			s.logger.Error("db err", zap.Error(err))
			return false, ""

		}
	}

	if user.State == 2 {
		s.logger.Error("用户被封号")
		return false, ""
	}

	//查询当前在线的主设备
	deviceOnlineKey := fmt.Sprintf("devices:%s", username)
	curOnlineDevieID, _ := redis.String(redisConn.Do("GET", deviceOnlineKey))

	if curOnlineDevieID != "" {
		if curOnlineDevieID == deviceID {
			s.logger.Debug("当前设备id与即将登录的设备相同")

		} else {

			//取出当前设备的os， clientType， logonAt
			curDeviceHashKey := fmt.Sprintf("devices:%s:%s", username, curOnlineDevieID)
			isMaster, _ := redis.Bool(redisConn.Do("HGET", curDeviceHashKey, "ismaster"))
			curOs, _ := redis.String(redisConn.Do("HGET", curDeviceHashKey, "os"))
			curClientType, _ := redis.Int(redisConn.Do("HGET", curDeviceHashKey, "clientType"))
			curLogonAt, _ := redis.Uint64(redisConn.Do("HGET", curDeviceHashKey, "logonAt"))
			curJwtToken, _ := redis.String(redisConn.Do("GET", fmt.Sprintf("DeviceJwtToken:%s", curOnlineDevieID)))
			s.logger.Debug("当前在线设备id与即将登录的设备不同",
				zap.Bool("isMaster", isMaster),
				zap.String("username", username),
				zap.String("当前在线curOnlineDevieID", curOnlineDevieID),
				zap.String("即将登录deviceID", deviceID),
				zap.String("curJwtToken", curJwtToken),
				zap.String("curOs", curOs),
				zap.Int("curClientType", curClientType),
				zap.Uint64("curLogonAt", curLogonAt))

			//删除当前主设备的redis缓存
			_, err = redisConn.Do("DEL", curDeviceHashKey)

		}
	}
	_, err = redisConn.Do("SET", deviceOnlineKey, deviceID)

	deviceHashKey := fmt.Sprintf("devices:%s:%s", username, deviceID) //创建主设备的哈希表, index为1

	_, err = redisConn.Do("HMSET",
		deviceHashKey,
		"username", username,
		"deviceid", deviceID,
		"ismaster", 1,
		"usertype", user.UserType,
		"clientType", clientType,
		"os", os,
		"logonAt", uint64(time.Now().UnixNano()/1e6),
	)

	if err != nil {
		s.logger.Error("HMSET deviceListKey err", zap.Error(err))
		return false, ""

	} else {
		s.logger.Debug("HMSET deviceListKey success")
	}

	s.logger.Debug("LoginBySmscode end")
	if deviceID == curOnlineDevieID {
		return true, ""
	} else {
		return true, curOnlineDevieID
	}
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
func (s *MysqlLianmiRepository) CheckUser(isMaster bool, username, password, deviceID, os string, clientType int) (bool, string) {
	s.logger.Debug("CheckUser start...")
	var err error
	redisConn := s.redisPool.Get()
	defer redisConn.Close()

	where := models.User{
		UserBase: models.UserBase{
			Username: username,
		},
	}
	var user models.User

	err = s.db.Model(&models.User{}).Where(&where).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			s.logger.Error("错误：用户账号不存在")
			return false, ""
		} else {

			s.logger.Error("db err", zap.Error(err))
			return false, ""

		}
	}

	if user.State == 2 {
		s.logger.Error("用户被封号")
		return false, ""
	}

	//主设备登录，需要检测是否有另外一台主设备未登出，如果未登出，则向其发出踢下线消息

	//主设备需要核对密码，从设备则无须核对
	if user.Password != password {
		s.logger.Error("密码不匹配")
		return false, ""
	}

	//查询当前在线的主设备
	deviceOnlineKey := fmt.Sprintf("devices:%s", username)
	curOnlineDevieID, _ := redis.String(redisConn.Do("GET", deviceOnlineKey))

	if curOnlineDevieID != "" {
		if curOnlineDevieID == deviceID {
			s.logger.Debug("当前设备id与即将登录的设备相同")

		} else {

			//取出当前设备的os， clientType， logonAt
			curDeviceHashKey := fmt.Sprintf("devices:%s:%s", username, curOnlineDevieID)
			isMaster, _ := redis.Bool(redisConn.Do("HGET", curDeviceHashKey, "ismaster"))
			curOs, _ := redis.String(redisConn.Do("HGET", curDeviceHashKey, "os"))
			curClientType, _ := redis.Int(redisConn.Do("HGET", curDeviceHashKey, "clientType"))
			curLogonAt, _ := redis.Uint64(redisConn.Do("HGET", curDeviceHashKey, "logonAt"))
			curJwtToken, _ := redis.String(redisConn.Do("GET", fmt.Sprintf("DeviceJwtToken:%s", curOnlineDevieID)))
			s.logger.Debug("当前在线设备id与即将登录的设备不同",
				zap.Bool("isMaster", isMaster),
				zap.String("username", username),
				zap.String("当前在线curOnlineDevieID", curOnlineDevieID),
				zap.String("即将登录deviceID", deviceID),
				zap.String("curJwtToken", curJwtToken),
				zap.String("curOs", curOs),
				zap.Int("curClientType", curClientType),
				zap.Uint64("curLogonAt", curLogonAt))

			//删除当前主设备的redis缓存
			_, err = redisConn.Do("DEL", curDeviceHashKey)

		}
	}
	_, err = redisConn.Do("SET", deviceOnlineKey, deviceID)

	deviceHashKey := fmt.Sprintf("devices:%s:%s", username, deviceID) //创建主设备的哈希表, index为1

	_, err = redisConn.Do("HMSET",
		deviceHashKey,
		"username", username,
		"deviceid", deviceID,
		"ismaster", 1,
		"usertype", user.UserType,
		"clientType", clientType,
		"os", os,
		"logonAt", uint64(time.Now().UnixNano()/1e6),
	)

	if err != nil {
		s.logger.Error("HMSET deviceListKey err", zap.Error(err))
		return false, ""

	} else {
		s.logger.Debug("HMSET deviceListKey success")
	}

	s.logger.Debug("CheckUser end")
	if deviceID == curOnlineDevieID {
		return true, ""
	} else {
		return true, curOnlineDevieID
	}
}

//同时增加用户类型角色
func (s *MysqlLianmiRepository) AddRole(role *models.Role) (err error) {
	if err := s.db.Create(role).Error; err != nil {
		s.logger.Error("新建用户角色失败")
		return err
	} else {
		return nil
	}
}

func (s *MysqlLianmiRepository) DeleteUser(id uint) bool {
	//采用事务同时删除用户和相应的用户角色
	var (
		userWhere = models.User{}
		user      models.User
		roleWhere = models.Role{}
		role      models.Role
	)
	userWhere.ID = id
	roleWhere.UserID = id

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

//判断用户名是否已存在
func (s *MysqlLianmiRepository) ExistUserByName(username string) bool {
	var user models.User

	where := models.User{
		UserBase: models.UserBase{
			Username: username,
		},
	}

	err := s.db.Model(&models.User{}).Where(&where).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			s.logger.Error("错误：用户账号不存在")
			return false
		} else {

			s.logger.Error("根据用户名获取用户信息失败 db err", zap.Error(err))
			return false

		}
	}
	return true
}

// 判断手机号码是否已存在
func (s *MysqlLianmiRepository) ExistUserByMobile(mobile string) bool {
	var user models.User
	where := models.User{
		UserBase: models.UserBase{
			Mobile: mobile,
		},
	}

	err := s.db.Model(&models.User{}).Where(&where).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			s.logger.Debug("记录不存在")
			return false
		} else {

			s.logger.Error("根据手机号码获取用户信息失败 db err", zap.Error(err))
			return false

		}
	}
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

	_, err = redisConn.Do("EXPIRE", deviceKey, LMCommon.ExpireTime) //过期时间

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

//根据注册用户id获取redis里此用户的设备id
func (s *MysqlLianmiRepository) GetDeviceFromRedis(username string) (string, error) {
	redisConn := s.redisPool.Get()
	defer redisConn.Close()

	deviceListKey := fmt.Sprintf("devices:%s", username)
	return redis.String(redisConn.Do("GET", deviceListKey))
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
	//删除所有与之相关的key
	_, err = redisConn.Do("DEL", deviceListKey)    //删除deviceListKey
	_, err = redisConn.Do("DEL", curDeviceHashKey) //删除当前旧的设备

	deviceKey := fmt.Sprintf("DeviceJwtToken:%s", deviceID)
	_, err = redisConn.Do("DEL", deviceKey) //删除DeviceJwtToken

	s.logger.Debug("SignOut end")
	_ = err
	return true
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

	_, err = redisConn.Do("EXPIRE", key, LMCommon.SMSEXPIRE) //设置有效期为
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
	user := new(models.User)

	if err = s.db.Model(user).Where(&models.User{
		UserBase: models.UserBase{
			Mobile: mobile,
		},
	}).First(user).Error; err != nil {
		s.logger.Error("MySQL里读取错误或记录不存在", zap.Error(err))
		return "", errors.Wrapf(err, "Get username error[mobile=%s]", mobile)
	} else {
		return user.Username, nil
	}
}

//根据注册账号返回手机号
func (s *MysqlLianmiRepository) GetMobileByUsername(username string) (string, error) {
	var err error
	user := new(models.User)

	if err = s.db.Model(user).Where(&models.User{
		UserBase: models.UserBase{
			Username: username,
		},
	}).First(user).Error; err != nil {
		s.logger.Error("MySQL里读取错误或记录不存在", zap.Error(err))
		return "", errors.Wrapf(err, "Get username error[username=%s]", username)
	} else {
		return user.Mobile, nil
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
				_, err = redisConn.Do("DEL", key) //删除smscode
				return smscodeInRedis == smscode
			}
		}
	}

}

//修改用户资料
func (s *MysqlLianmiRepository) UpdateUser(username string, user *models.User) error {
	where := models.User{
		UserBase: models.UserBase{
			Username: username,
		},
	}
	// 同时更新多个字段
	result := s.db.Model(&models.User{}).Where(&where).Updates(user)

	//updated records count
	s.logger.Debug("UpdateUser result: ",
		zap.Int64("RowsAffected", result.RowsAffected),
		zap.Error(result.Error))

	if result.Error != nil {
		s.logger.Error("UpdateUser, 修改用户资料数据失败", zap.Error(result.Error))
		return result.Error
	} else {
		s.logger.Error("UpdateUser, 修改用户资料数据成功")
	}
	return nil
}

//增加或修改用户标签 tags表
func (s *MysqlLianmiRepository) AddTag(tag *models.Tag) error {

	// 在冲突时，更新除主键以外的所有列到新值。
	if results := s.db.Clauses(clause.OnConflict{
		UpdateAll: true,
	}).Create(&tag); results != nil {
		s.logger.Error("AddTag, failed to Create tag", zap.Error(results.Error))
		return results.Error
	} else {
		s.logger.Debug("AddTag, Create tag succeed", zap.Int64("RowsAffected", results.RowsAffected))
	}

	// `username` 冲突时，将map里的指定字段值更新
	// s.db.Clauses(clause.OnConflict{
	// 	Columns:   []clause.Column{{Name: "username"}},
	// 	DoUpdates: clause.Assignments(map[string]interface{}{"target_username": tag.TargetUsername, "tag_type": tag.TagType}),
	// }).Create(&tag)

	return nil
}

//获取users表的所有用户账号
func (s *MysqlLianmiRepository) QueryAllUsernames() ([]string, error) {
	usernames := make([]string, 0)
	var users []models.User
	if err := s.db.Model(&models.User{}).Select("username").Find(&users).Error; err != nil {
		s.logger.Error("Failed to query users", zap.Error(err))
		return nil, err
	}

	for _, user := range users {
		usernames = append(usernames, user.Username)
	}

	return usernames, nil

}
