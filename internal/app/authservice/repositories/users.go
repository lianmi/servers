package repositories

import (
	"fmt"
	"time"

	"github.com/gomodule/redigo/redis"
	"github.com/jinzhu/gorm"
	"github.com/lianmi/servers/internal/common"
	"github.com/lianmi/servers/internal/pkg/models"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

type UsersRepository interface {
	GetUser(ID uint64) (p *models.User, err error)
	Register(user *models.User) (err error)
	AddRole(role *models.Role) (err error)
	DeleteUser(id uint64) bool
	GetUserRoles(where interface{}) []*models.Role
	CheckUser(where interface{}) bool
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

	//更新用户
	UpdateUser(user *models.User, role *models.Role) bool

	//获取用户
	GetUserByID(id int) *models.User

	//登出
	SignOut(token, username, deviceID string) bool

	//令牌是否存在
	ExistsTokenInRedis(token string) bool
}

type MysqlUsersRepository struct {
	logger    *zap.Logger
	db        *gorm.DB
	redisPool *redis.Pool
	base      *BaseRepository
}

func NewMysqlUsersRepository(logger *zap.Logger, db *gorm.DB, redisPool *redis.Pool) UsersRepository {
	return &MysqlUsersRepository{
		logger:    logger.With(zap.String("type", "UsersRepository")),
		db:        db,
		redisPool: redisPool,
		base:      NewBaseRepository(logger, db),
	}
}

func (s *MysqlUsersRepository) GetUser(ID uint64) (p *models.User, err error) {
	p = new(models.User)
	if err = s.db.Model(p).Where("id = ?", ID).First(p).Error; err != nil {
		return nil, errors.Wrapf(err, "Get user error[id=%d]", ID)
	}
	s.logger.Debug("GetUser run...")
	return
}

func (s *MysqlUsersRepository) Register(user *models.User) (err error) {

	if err := s.base.Create(user); err != nil {
		s.logger.Error("新建用户失败")
		return err
	} else {
		return nil
	}
}

// 获取用户角色
func (s *MysqlUsersRepository) GetUserRoles(where interface{}) []*models.Role {
	var roles []*models.Role
	if err := s.base.Find(where, &roles, ""); err != nil {
		s.logger.Error("获取用户角色错误")
	}
	return roles
}

func (s *MysqlUsersRepository) CheckUser(where interface{}) bool {
	var user models.User
	if err := s.base.First(where, &user); err != nil {
		s.logger.Error("手机号或密码错误")
		return false
	}
	return true
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
	err := redisConn.Send("SET", token, 1) //增加key
	deviceKey := fmt.Sprintf("%s:%s", token, deviceID)
	err = redisConn.Send("SET", deviceKey, 1)                    //增加deviceKey， mqtt消息必须要验证这个
	err = redisConn.Send("EXPIRE", token, common.ExpireTime)     //过期时间
	err = redisConn.Send("EXPIRE", deviceKey, common.ExpireTime) //过期时间
	_ = err

	//一次性写入到Redis
	if err := redisConn.Flush(); err != nil {
		s.logger.Error("写入redis失败", zap.Error(err))
		return false
	}

	return true
}

/*
登出
1. 如果是主设备，则踢出此用户的所有主从设备， 如果仅仅是从设备，就删除自己的token
2. 删除redis里的此用户的哈希记录
3. 如果是商户的登出，则需要删除数据库里其对应的所有OPK(下次登录需要重新上传)
*/
func (s *MysqlUsersRepository) SignOut(token, username, deviceID string) bool {
	redisConn := s.redisPool.Get()
	defer redisConn.Close()
	err := redisConn.Send("DEL", token) //删除key
	deviceKey := fmt.Sprintf("%s:%s", token, deviceID)
	err = redisConn.Send("DEL", deviceKey) //删除deviceKey
	_ = err

	//一次性写入到Redis
	if err := redisConn.Flush(); err != nil {
		s.logger.Error("写入redis失败", zap.Error(err))
		return false
	}
	s.logger.Debug("写入redis成功")
	return true
}

func (s *MysqlUsersRepository) ExistsTokenInRedis(token string) bool {
	redisConn := s.redisPool.Get()
	defer redisConn.Close()

	if isExists, err := redis.Bool(redisConn.Do("GET", token)); err != nil {
		s.logger.Error("redisConn GET token Error", zap.Error(err))
		return false
	} else {
		s.logger.Info("redisConn GET token ok ", zap.String("token", token))
		return isExists
	}

}
