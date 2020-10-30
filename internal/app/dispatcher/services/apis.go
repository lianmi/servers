package services

import (
	"context"
	Service "github.com/lianmi/servers/api/proto/service"
	// User "github.com/lianmi/servers/api/proto/user"
	"github.com/lianmi/servers/internal/app/dispatcher/repositories"
	"github.com/lianmi/servers/internal/pkg/models"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"time"

	pb "github.com/lianmi/servers/api/proto/user"
)

type LianmiApisService interface {
	BlockUser(username string) error
	DisBlockUser(username string) (*models.User, error)
	Register(user *models.User) (string, error)
	ResetPassword(mobile, password string, user *models.User) error
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

	// GetUser(ID uint64) (*models.User, error)

	//检测校验码是否正确
	CheckSmsCode(mobile, smscode string) bool

	//修改密码
	ChanPassword(username string, req *Service.ChanPasswordReq) error

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

	QueryCustomerServices(req *Service.QueryCustomerServiceReq) ([]*models.CustomerServiceInfo, error)

	AddCustomerService(req *Service.AddCustomerServiceReq) error

	DeleteCustomerService(req *Service.DeleteCustomerServiceReq) bool

	UpdateCustomerService(req *Service.UpdateCustomerServiceReq) error

	QueryGrades(req *Service.GradeReq, pageIndex int, pageSize int, total *uint64, where interface{}) ([]*models.Grade, error)

	AddGrade(req *Service.AddGradeReq) (string, error)

	SubmitGrade(req *Service.SubmitGradeReq) error

	GetMembershipCardSaleMode(businessUsername string) (int, error)

	SetMembershipCardSaleMode(businessUsername string, saleType int) error

	GetBusinessMembership(isRebate bool) (*Service.GetBusinessMembershipResp, error)

	PayForMembership(payForUsername string) error

	//Grpc 获取用户信息
	GetUser(ctx context.Context, in *Service.UserReq) (*Service.UserRsp, error)
}

type DefaultLianmiApisService struct {
	logger        *zap.Logger
	Repository    repositories.LianmiRepository
	grpcClientSvc Service.LianmiApisClient
}

func NewLianmiApisService(logger *zap.Logger, repository repositories.LianmiRepository, lc Service.LianmiApisClient) LianmiApisService {
	return &DefaultLianmiApisService{
		logger:        logger.With(zap.String("type", "DefaultLianmiApisService")),
		Repository:    repository,
		grpcClientSvc: lc,
	}
}

func (s *DefaultLianmiApisService) GetUser(ctx context.Context, in *Service.UserReq) (*Service.UserRsp, error) {
	s.logger.Debug("GetUser", zap.Uint64("ID", in.Id))
	/*
		if p, err := s.Repository.GetUser(in.Id); err != nil {
			return nil, errors.Wrap(err, "Get user error")
		} else {
			return &Service.UserRsp{
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
	*/

	//从微服务获取用户数据
	resp, err := s.grpcClientSvc.GetUser(ctx, in)

	return resp, err
}

func (s *DefaultLianmiApisService) BlockUser(username string) (err error) {
	s.logger.Debug("BlockUser", zap.String("username", username))
	// if err = s.Repository.BlockUser(username); err != nil {
	// 	return errors.Wrap(err, "Block user error")
	// }

	return nil
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

func (s *DefaultLianmiApisService) ResetPassword(mobile, password string, user *models.User) error {
	// if err := s.Repository.ResetPassword(mobile, password, user); err != nil {
	// 	return errors.Wrap(err, "ResetPassword error")
	// }

	return nil
}

//修改密码
func (s *DefaultLianmiApisService) ChanPassword(username string, req *Service.ChanPasswordReq) error {

	// return s.Repository.ChanPassword(username, req)
	return nil
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
	// return s.Repository.ApproveTeam(teamID)

	//TODO
	return nil
}

//封禁群组
func (s *DefaultLianmiApisService) BlockTeam(teamID string) error {
	// return s.Repository.BlockTeam(teamID)

	//TODO
	return nil
}

//解封群组
func (s *DefaultLianmiApisService) DisBlockTeam(teamID string) error {
	// return s.Repository.DisBlockTeam(teamID)
	//TODO
	return nil
}

//======后台相关======/
func (s *DefaultLianmiApisService) AddGeneralProduct(generalProduct *models.GeneralProduct) error {
	// return s.Repository.AddGeneralProduct(generalProduct)

	//TODO
	return nil
}

//查询通用商品(id) - Read
func (s *DefaultLianmiApisService) GetGeneralProductByID(productID string) (p *models.GeneralProduct, err error) {

	// return s.Repository.GetGeneralProductByID(productID)
	//TODO
	return nil, nil
}

//查询通用商品分页 - Page
func (s *DefaultLianmiApisService) GetGeneralProductPage(pageIndex, pageSize int, total *uint64, where interface{}) ([]*models.GeneralProduct, error) {

	// return s.Repository.GetGeneralProductPage(pageIndex, pageSize, total, where)

	//TODO
	return nil, nil
}

// 修改通用商品 - Update
func (s *DefaultLianmiApisService) UpdateGeneralProduct(generalProduct *models.GeneralProduct) error {

	// return s.Repository.UpdateGeneralProduct(generalProduct)
	//TODO
	return nil

}

// 删除通用商品 - Delete
func (s *DefaultLianmiApisService) DeleteGeneralProduct(productID string) bool {

	// return s.Repository.DeleteGeneralProduct(productID)
	//TODO
	return false

}

//获取在线客服id数组
func (s *DefaultLianmiApisService) QueryCustomerServices(req *Service.QueryCustomerServiceReq) ([]*models.CustomerServiceInfo, error) {

	// return s.Repository.QueryCustomerServices(req)
	//TODO
	return nil, nil

}

//增加在线客服id
func (s *DefaultLianmiApisService) AddCustomerService(req *Service.AddCustomerServiceReq) error {

	// return s.Repository.AddCustomerService(req)
	//TODO
	return nil

}

func (s *DefaultLianmiApisService) DeleteCustomerService(req *Service.DeleteCustomerServiceReq) bool {
	// return s.Repository.DeleteCustomerService(req)
	//TODO
	return false

}

func (s *DefaultLianmiApisService) UpdateCustomerService(req *Service.UpdateCustomerServiceReq) error {

	// return s.Repository.UpdateCustomerService(req)
	//TODO
	return nil

}

func (s *DefaultLianmiApisService) QueryGrades(req *Service.GradeReq, pageIndex int, pageSize int, total *uint64, where interface{}) ([]*models.Grade, error) {
	// return s.Repository.QueryGrades(req, pageIndex, pageSize, total, where)
	//TODO
	return nil, nil
}

func (s *DefaultLianmiApisService) AddGrade(req *Service.AddGradeReq) (string, error) {
	// return s.Repository.AddGrade(req)
	//TODO
	return "", nil
}

func (s *DefaultLianmiApisService) SubmitGrade(req *Service.SubmitGradeReq) error {
	// return s.Repository.SubmitGrade(req)
	//TODO
	return nil
}

func (s *DefaultLianmiApisService) GetMembershipCardSaleMode(businessUsername string) (int, error) {
	// return s.Repository.GetMembershipCardSaleMode(businessUsername)
	//TODO
	return 0, nil
}

func (s *DefaultLianmiApisService) SetMembershipCardSaleMode(businessUsername string, saleType int) error {
	// return s.Repository.SetMembershipCardSaleMode(businessUsername, saleType)
	//TODO
	return nil
}

func (s *DefaultLianmiApisService) GetBusinessMembership(isRebate bool) (*Service.GetBusinessMembershipResp, error) {
	// return s.Repository.GetBusinessMembership(isRebate)

	//TODO
	return nil, nil
}

func (s *DefaultLianmiApisService) PayForMembership(payForUsername string) error {
	// return s.Repository.PayForMembership(payForUsername)

	//TODO
	return nil
}
