package services

import (
	"time"

	"github.com/lianmi/servers/internal/app/authservice/repositories"
	"github.com/lianmi/servers/internal/pkg/models"
	"github.com/pkg/errors"
	"go.uber.org/zap"

	pb "github.com/lianmi/servers/api/proto/user"
)

type LianmiApisService interface {
	GetUser(ID uint64) (*models.User, error)
	BlockUser(username string) (*models.User, error)
	DisBlockUser(username string) (*models.User, error)
	Register(user *models.User) (string, error)
	Resetpwd(mobile, password string, user *models.User) error
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

	//修改密码
	ChanPassword(username, oldPassword, newPassword string) error

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
}

type DefaultLianmiApisService struct {
	logger     *zap.Logger
	Repository repositories.LianmiRepository
}

func NewLianmiApisService(logger *zap.Logger, Repository repositories.LianmiRepository) LianmiApisService {
	return &DefaultLianmiApisService{
		logger:     logger.With(zap.String("type", "DefaultLianmiApisService")),
		Repository: Repository,
	}
}

func (s *DefaultLianmiApisService) GetUser(ID uint64) (p *models.User, err error) {
	s.logger.Debug("GetUser", zap.Uint64("ID", ID))
	if p, err = s.Repository.GetUser(ID); err != nil {
		return nil, errors.Wrap(err, "Get user error")
	}

	return
}

func (s *DefaultLianmiApisService) BlockUser(username string) (p *models.User, err error) {
	s.logger.Debug("BlockUser", zap.String("username", username))
	if p, err = s.Repository.BlockUser(username); err != nil {
		return nil, errors.Wrap(err, "Block user error")
	}

	return
}
func (s *DefaultLianmiApisService) DisBlockUser(username string) (p *models.User, err error) {
	s.logger.Debug("DisBlockUser", zap.String("username", username))
	if p, err = s.Repository.DisBlockUser(username); err != nil {
		return nil, errors.Wrap(err, "DisBlockUser user error")
	}

	return
}

//生成短信校验码
func (s *DefaultLianmiApisService) GenerateSmsCode(mobile string) bool {

	return s.Repository.GenerateSmsCode(mobile)

}

//检测校验码是否正确
func (s *DefaultLianmiApisService) CheckSmsCode(mobile, smscode string) bool {
	return s.Repository.CheckSmsCode(mobile, smscode)
}

func (s *DefaultLianmiApisService) Register(user *models.User) (string, error) {
	// var username string
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

func (s *DefaultLianmiApisService) Resetpwd(mobile, password string, user *models.User) error {
	if err := s.Repository.Resetpwd(mobile, password, user); err != nil {
		return errors.Wrap(err, "Resetpwd error")
	}

	return nil
}

//修改密码
func (s *DefaultLianmiApisService) ChanPassword(username, oldPassword, newPassword string) error {

	return s.Repository.ChanPassword(username, oldPassword, newPassword)
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

//======后台相关======/
func (s *DefaultLianmiApisService) AddGeneralProduct(generalProduct *models.GeneralProduct) error {
	return s.Repository.AddGeneralProduct(generalProduct)
}

//查询通用商品(id) - Read
func (s *DefaultLianmiApisService) GetGeneralProductByID(productID string) (p *models.GeneralProduct, err error) {

	return s.Repository.GetGeneralProductByID(productID)
}

//查询通用商品分页 - Page
func (s *DefaultLianmiApisService) GetGeneralProductPage(pageIndex, pageSize int, total *uint64, where interface{}) ([]*models.GeneralProduct, error) {

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
