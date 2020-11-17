package services

import (
	"context"
	Auth "github.com/lianmi/servers/api/proto/auth"
	User "github.com/lianmi/servers/api/proto/user"
	"github.com/lianmi/servers/internal/app/authservice/repositories"
	"github.com/lianmi/servers/internal/pkg/models"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"time"

	pb "github.com/lianmi/servers/api/proto/user"
)

type AuthService interface {
	BlockUser(username string) error
	DisBlockUser(username string) (*models.User, error)
	Register(user *models.User) (string, error)
	GetUserRoles(username string) []*models.Role

	//检测用户登录
	// CheckUser(isMaster bool, smscode, username, password, deviceID, os string, clientType int) bool

	// 判断用户名是否已存在
	ExistUserByName(username string) bool
	// 判断手机号码是否已存在
	ExistUserByMobile(mobile string) bool
	SaveUserToken(username, deviceID string, token string, expire time.Time) bool
	SignOut(token, username, deviceID string) bool
	ExistsTokenInRedis(deviceID, token string) bool

	//生成注册校验码
	GenerateSmsCode(mobile string) bool

	// GetUser(ID uint64) (*models.User, error)

	//检测校验码是否正确
	// CheckSmsCode(mobile, smscode string) bool

	//修改密码
	// ChanPassword(username string, req *Auth.ChanPasswordReq) error

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

	QueryCustomerServices(req *Auth.QueryCustomerServiceReq) ([]*models.CustomerServiceInfo, error)

	AddCustomerService(req *Auth.AddCustomerServiceReq) error

	DeleteCustomerService(req *Auth.DeleteCustomerServiceReq) bool

	UpdateCustomerService(req *Auth.UpdateCustomerServiceReq) error

	QueryGrades(req *Auth.GradeReq, pageIndex int, pageSize int, total *uint64, where interface{}) ([]*models.Grade, error)

	AddGrade(req *Auth.AddGradeReq) (string, error)

	SubmitGrade(req *Auth.SubmitGradeReq) error

	//Grpc 获取用户信息
	GetUser(ctx context.Context, in *Auth.UserReq) (*Auth.UserRsp, error)
}

type DefaultLianmiAuthService struct {
	logger     *zap.Logger
	Repository repositories.LianmiRepository
}

func NewLianmiAuthService(logger *zap.Logger, repository repositories.LianmiRepository) AuthService {
	return &DefaultLianmiAuthService{
		logger:     logger.With(zap.String("type", "authservice.services")),
		Repository: repository,
	}
}

func (s *DefaultLianmiAuthService) GetUser(ctx context.Context, in *Auth.UserReq) (*Auth.UserRsp, error) {
	s.logger.Debug("GrpcServer: GetUser", zap.Uint64("ID", in.Id))
	if p, err := s.Repository.GetUser(in.Id); err != nil {
		return nil, errors.Wrap(err, "Get user error")
	} else {
		return &Auth.UserRsp{
			User: &User.User{
				Username:          p.Username,
				Gender:            User.Gender(p.Gender),
				Nick:              p.Nick,
				Avatar:            p.Avatar,
				Label:             p.Label,
				Mobile:            p.Mobile,
				Email:             p.Email,
				UserType:          User.UserType(p.UserType),
				Extend:            p.Extend,
				ContactPerson:     p.ContactPerson,
				Introductory:      p.Introductory,
				Province:          p.Province,
				City:              p.City,
				County:            p.County,
				Street:            p.Street,
				Address:           p.Address,
				Branchesname:      p.Branchesname,
				LegalPerson:       p.LegalPerson,
				LegalIdentityCard: p.LegalIdentityCard,
			},
		}, nil
	}

}

func (s *DefaultLianmiAuthService) BlockUser(username string) (err error) {
	s.logger.Debug("BlockUser", zap.String("username", username))
	if err = s.Repository.BlockUser(username); err != nil {
		return errors.Wrap(err, "Block user error")
	}

	return nil
}
func (s *DefaultLianmiAuthService) DisBlockUser(username string) (p *models.User, err error) {
	s.logger.Debug("DisBlockUser", zap.String("username", username))
	if p, err = s.Repository.DisBlockUser(username); err != nil {
		return nil, errors.Wrap(err, "DisBlockUser user error")
	}

	return
}

//生成短信校验码
func (s *DefaultLianmiAuthService) GenerateSmsCode(mobile string) bool {
	return s.Repository.GenerateSmsCode(mobile)
}

//检测校验码是否正确
// func (s *DefaultLianmiAuthService) CheckSmsCode(mobile, smscode string) bool {
// 	return s.Repository.CheckSmsCode(mobile, smscode)
// }

func (s *DefaultLianmiAuthService) Register(user *models.User) (string, error) {
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

//修改密码
// func (s *DefaultLianmiAuthService) ChanPassword(username string, req *Auth.ChanPasswordReq) error {

// 	return s.Repository.ChanPassword(username, req)
// }

func (s *DefaultLianmiAuthService) GetUserRoles(username string) []*models.Role {
	where := models.Role{UserName: username}
	return s.Repository.GetUserRoles(&where)
}

//CheckUser 身份验证
// func (s *DefaultLianmiAuthService) CheckUser(isMaster bool, smscode, username, password, deviceID, os string, clientType int) bool {

// 	return s.Repository.CheckUser(isMaster, smscode, username, password, deviceID, os, clientType)
// }

func (s *DefaultLianmiAuthService) ExistUserByName(username string) bool {

	return s.Repository.ExistUserByName(username)
}

// 判断手机号码是否已存在
func (s *DefaultLianmiAuthService) ExistUserByMobile(mobile string) bool {
	return s.Repository.ExistUserByMobile(mobile)
}

func (s *DefaultLianmiAuthService) SaveUserToken(username, deviceID string, token string, expire time.Time) bool {
	return s.Repository.SaveUserToken(username, deviceID, token, expire)
}

func (s *DefaultLianmiAuthService) SignOut(token, username, deviceID string) bool {
	return s.Repository.SignOut(token, username, deviceID)
}

func (s *DefaultLianmiAuthService) ExistsTokenInRedis(deviceID, token string) bool {
	return s.Repository.ExistsTokenInRedis(deviceID, token)
}

func (s *DefaultLianmiAuthService) ApproveTeam(teamID string) error {
	return s.Repository.ApproveTeam(teamID)
}

//封禁群组
func (s *DefaultLianmiAuthService) BlockTeam(teamID string) error {
	return s.Repository.BlockTeam(teamID)
}

//解封群组
func (s *DefaultLianmiAuthService) DisBlockTeam(teamID string) error {
	return s.Repository.DisBlockTeam(teamID)
}

//======后台相关======/
func (s *DefaultLianmiAuthService) AddGeneralProduct(generalProduct *models.GeneralProduct) error {
	return s.Repository.AddGeneralProduct(generalProduct)
}

//查询通用商品(id) - Read
func (s *DefaultLianmiAuthService) GetGeneralProductByID(productID string) (p *models.GeneralProduct, err error) {

	return s.Repository.GetGeneralProductByID(productID)
}

//查询通用商品分页 - Page
func (s *DefaultLianmiAuthService) GetGeneralProductPage(pageIndex, pageSize int, total *uint64, where interface{}) ([]*models.GeneralProduct, error) {

	return s.Repository.GetGeneralProductPage(pageIndex, pageSize, total, where)
}

// 修改通用商品 - Update
func (s *DefaultLianmiAuthService) UpdateGeneralProduct(generalProduct *models.GeneralProduct) error {

	return s.Repository.UpdateGeneralProduct(generalProduct)

}

// 删除通用商品 - Delete
func (s *DefaultLianmiAuthService) DeleteGeneralProduct(productID string) bool {

	return s.Repository.DeleteGeneralProduct(productID)

}

//获取在线客服id数组
func (s *DefaultLianmiAuthService) QueryCustomerServices(req *Auth.QueryCustomerServiceReq) ([]*models.CustomerServiceInfo, error) {

	return s.Repository.QueryCustomerServices(req)

}

//增加在线客服id
func (s *DefaultLianmiAuthService) AddCustomerService(req *Auth.AddCustomerServiceReq) error {

	return s.Repository.AddCustomerService(req)

}

func (s *DefaultLianmiAuthService) DeleteCustomerService(req *Auth.DeleteCustomerServiceReq) bool {
	return s.Repository.DeleteCustomerService(req)

}

func (s *DefaultLianmiAuthService) UpdateCustomerService(req *Auth.UpdateCustomerServiceReq) error {

	return s.Repository.UpdateCustomerService(req)

}

func (s *DefaultLianmiAuthService) QueryGrades(req *Auth.GradeReq, pageIndex int, pageSize int, total *uint64, where interface{}) ([]*models.Grade, error) {
	return s.Repository.QueryGrades(req, pageIndex, pageSize, total, where)
}

func (s *DefaultLianmiAuthService) AddGrade(req *Auth.AddGradeReq) (string, error) {
	return s.Repository.AddGrade(req)
}

func (s *DefaultLianmiAuthService) SubmitGrade(req *Auth.SubmitGradeReq) error {
	return s.Repository.SubmitGrade(req)
}
