package repositories

import (
	// "encoding/json"
	"fmt"
	"time"

	// "github.com/golang/protobuf/proto"
	"github.com/gomodule/redigo/redis"
	// "github.com/jinzhu/gorm"
	// Auth "github.com/lianmi/servers/api/proto/auth"
	Service "github.com/lianmi/servers/api/proto/service"
	// User "github.com/lianmi/servers/api/proto/user"
	// "github.com/lianmi/servers/internal/app/dispatcher/grpcclients"
	// "github.com/lianmi/servers/internal/app/dispatcher/nsqMq"
	// "github.com/lianmi/servers/internal/common"
	"github.com/lianmi/servers/internal/pkg/models"
	"github.com/pkg/errors"
	"go.uber.org/zap"

	"github.com/lianmi/servers/util/dateutil"
)

//获取空闲的在线客服id数组
func (s *MysqlLianmiRepository) QueryCustomerServices(req *Service.QueryCustomerServiceReq) ([]*models.CustomerServiceInfo, error) {

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
		if int(req.Type) == 0 {
			if isIdle == req.IsIdle {
				csList = append(csList, &models.CustomerServiceInfo{
					Username:   csUsername, //客服或技术人员的注册账号id
					JobNumber:  jobNumber,  //客服或技术人员的工号
					Type:       cstype,     //客服或技术人员的类型， 1-客服，2-技术
					Evaluation: evaluation, //职称, 技术工程师，技术员等
					NickName:   nickName,   //呢称,
				})
			}
		} else {
			if isIdle == req.IsIdle && cstype == int(req.Type) {
				csList = append(csList, &models.CustomerServiceInfo{
					Username:   csUsername, //客服或技术人员的注册账号id
					JobNumber:  jobNumber,  //客服或技术人员的工号
					Type:       cstype,     //客服或技术人员的类型， 1-客服，2-技术
					Evaluation: evaluation, //职称, 技术工程师，技术员等
					NickName:   nickName,   //呢称,
				})
			}
		}

	}
	return csList, nil
}

func (s *MysqlLianmiRepository) AddCustomerService(req *Service.AddCustomerServiceReq) error {
	var err error

	redisConn := s.redisPool.Get()
	defer redisConn.Close()

	username := req.Username

	userKey := fmt.Sprintf("userData:%s", username)
	username2, _ := redis.String(redisConn.Do("HGET", userKey, "Username"))
	if username2 != username {
		return errors.Wrapf(err, "Username is not exists error[username=%s]", username)
	}
	if reply, err := redisConn.Do("ZRANK", "CustomerServiceList", username); err == nil {
		if reply != nil {

			//已经存在，不能重复增加
			return errors.Wrapf(err, "Username is exists error[username=%s]", username)
		}

	}

	c := &models.CustomerServiceInfo{
		Username:   req.Username,
		JobNumber:  req.JobNumber,
		Type:       int(req.Type),
		Evaluation: req.Evaluation,
		NickName:   req.NickName,
	}

	tx := s.base.GetTransaction()

	if err := tx.Save(c).Error; err != nil {
		s.logger.Error("增加客户技术人员失败", zap.Error(err))
		tx.Rollback()
		return errors.Wrapf(err, "Save error")

	}
	//提交
	tx.Commit()

	if _, err = redisConn.Do("ZADD", "CustomerServiceList", time.Now().UnixNano()/1e6, username); err != nil {
		s.logger.Error("ZADD Error", zap.Error(err))
	}

	_, err = redisConn.Do("HMSET",
		fmt.Sprintf("CustomerServiceInfo:%s", username),
		"IsIdle", false,
		"Username", req.Username,
		"Type", req.Type,
		"JobNumber", req.JobNumber,
		"Evaluation", req.Evaluation,
		"NickName", req.NickName,
	)

	return nil

}

func (s *MysqlLianmiRepository) DeleteCustomerService(req *Service.DeleteCustomerServiceReq) bool {

	redisConn := s.redisPool.Get()
	defer redisConn.Close()

	username := req.Username

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
		s.logger.Error("删除在线客服人员失败", zap.Error(err))
		tx.Rollback()
		return false
	}
	tx.Commit()
	return true

}

func (s *MysqlLianmiRepository) UpdateCustomerService(req *Service.UpdateCustomerServiceReq) error {
	var err error

	redisConn := s.redisPool.Get()
	defer redisConn.Close()

	username := req.Username

	userKey := fmt.Sprintf("userData:%s", username)
	username2, _ := redis.String(redisConn.Do("HGET", userKey, "Username"))
	if username2 != username {
		return errors.Wrapf(err, "Username is not exists error[username=%s]", username)
	}
	if reply, err := redisConn.Do("ZRANK", "CustomerServiceList", username); err == nil {
		if reply == nil {

			//不存在，必须先增加
			return errors.Wrapf(err, "Username is not exists in list error[username=%s]", username)
		}

	}

	c := new(models.CustomerServiceInfo)
	if err = s.db.Model(c).Where("username = ?", username).First(c).Error; err != nil {
		return errors.Wrapf(err, "Get customerServiceInfo error[username=%s]", username)
	}

	c.JobNumber = req.JobNumber
	c.Evaluation = req.Evaluation
	c.NickName = req.NickName
	c.Type = int(req.Type)

	tx := s.base.GetTransaction()

	if err := tx.Save(c).Error; err != nil {
		s.logger.Error("修改客户技术失败", zap.Error(err))
		tx.Rollback()
		return errors.Wrapf(err, "Save error")
	}
	//提交
	tx.Commit()

	_, err = redisConn.Do("HMSET",
		fmt.Sprintf("CustomerServiceInfo:%s", username),
		"Username", req.Username,
		"IsIdle", false,
		"Type", req.Type,
		"JobNumber", req.JobNumber,
		"Evaluation", req.Evaluation,
		"NickName", req.NickName,
	)

	return nil

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
