package services

import (
	"time"

	"github.com/lianmi/servers/internal/app/authservice/repositories"
	"github.com/lianmi/servers/internal/pkg/models"
	"github.com/pkg/errors"
	"go.uber.org/zap"

	pb "github.com/lianmi/servers/api/proto/user"
)

type UsersService interface {
	GetUser(ID uint64) (*models.User, error)
	BlockUser(ID uint64) (*models.User, error)
	Register(user *models.User) (string, error)
	ChanPassword(oldpassword, smsCode, password string) (string, error)
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

	//检测校验码是否正确
	CheckSmsCode(mobile, smscode string) bool
}

type DefaultUsersService struct {
	logger     *zap.Logger
	Repository repositories.UsersRepository
}

func NewUserService(logger *zap.Logger, Repository repositories.UsersRepository) UsersService {
	return &DefaultUsersService{
		logger:     logger.With(zap.String("type", "DefaultUsersService")),
		Repository: Repository,
	}
}

func (s *DefaultUsersService) GetUser(ID uint64) (p *models.User, err error) {
	s.logger.Debug("GetUser", zap.Uint64("ID", ID))
	if p, err = s.Repository.GetUser(ID); err != nil {
		return nil, errors.Wrap(err, "Get user error")
	}

	return
}

func (s *DefaultUsersService) BlockUser(ID uint64) (p *models.User, err error) {
	s.logger.Debug("BlockUser", zap.Uint64("ID", ID))
	if p, err = s.Repository.BlockUser(ID); err != nil {
		return nil, errors.Wrap(err, "Block user error")
	}

	return
}

//生成短信校验码
func (s *DefaultUsersService) GenerateSmsCode(mobile string) bool {

	return s.Repository.GenerateSmsCode(mobile)

}

//检测校验码是否正确
func (s *DefaultUsersService) CheckSmsCode(mobile, smscode string) bool {
	return s.Repository.CheckSmsCode(mobile, smscode)
}

func (s *DefaultUsersService) Register(user *models.User) (string, error) {
    // var username string
    var err error
	if err = s.Repository.Register(user); err != nil {
		return "",  errors.Wrap(err, "Register user error")
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

func (s *DefaultUsersService) ChanPassword(oldpassword, smsCode, password string) (string, error) {
	return "", nil
}

func (s *DefaultUsersService) GetUserRoles(username string) []*models.Role {
	where := models.Role{UserName: username}
	return s.Repository.GetUserRoles(&where)
}

//CheckUser 身份验证
func (s *DefaultUsersService) CheckUser(isMaster bool, smscode, username, password, deviceID, os string, clientType int) bool {
	
	return s.Repository.CheckUser(isMaster, smscode, username, password, deviceID, os, clientType)
}

func (s *DefaultUsersService) ExistUserByName(username string) bool {

	return s.Repository.ExistUserByName(username)
}

// 判断手机号码是否已存在
func (s *DefaultUsersService) ExistUserByMobile(mobile string) bool {
	return s.Repository.ExistUserByMobile(mobile)
}

func (s *DefaultUsersService) SaveUserToken(username, deviceID string, token string, expire time.Time) bool {
	return s.Repository.SaveUserToken(username, deviceID, token, expire)
}

func (s *DefaultUsersService) SignOut(token, username, deviceID string) bool {
	return s.Repository.SignOut(token, username, deviceID)
}

func (s *DefaultUsersService) ExistsTokenInRedis(deviceID, token string) bool {
	return s.Repository.ExistsTokenInRedis(deviceID, token)
}
