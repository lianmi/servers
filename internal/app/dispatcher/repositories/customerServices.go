package repositories

import (
	"fmt"
	"time"

	"github.com/gomodule/redigo/redis"
	Auth "github.com/lianmi/servers/api/proto/auth"
	"github.com/lianmi/servers/internal/pkg/models"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"gorm.io/gorm/clause"

	"github.com/lianmi/servers/util/dateutil"
)

//获取空闲的在线客服id数组
func (s *MysqlLianmiRepository) QueryCustomerServices(req *Auth.QueryCustomerServiceReq) ([]*models.CustomerServiceInfo, error) {

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

func (s *MysqlLianmiRepository) AddCustomerService(req *Auth.AddCustomerServiceReq) error {
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

	//如果没有记录，则增加，如果有记录，则更新全部字段
	if err := s.db.Clauses(clause.OnConflict{DoNothing: true}).Create(&c).Error; err != nil {
		s.logger.Error("增加CustomerServiceInfo失败, failed to upsert CustomerServiceInfo", zap.Error(err))
		return err
	} else {
		s.logger.Debug("增加CustomerServiceInfo成功, upsert CustomerServiceInfo succeed")
	}

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

func (s *MysqlLianmiRepository) DeleteCustomerService(req *Auth.DeleteCustomerServiceReq) bool {

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

//修改在线客服资料
func (s *MysqlLianmiRepository) UpdateCustomerService(req *Auth.UpdateCustomerServiceReq) error {
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

	c.JobNumber = req.JobNumber
	c.Evaluation = req.Evaluation
	c.NickName = req.NickName
	c.Type = int(req.Type)

	where := models.CustomerServiceInfo{
		Username: username,
	}
	// 同时更新多个字段
	result := s.db.Model(&models.CustomerServiceInfo{}).Where(&where).Select("job_number", "evaluation", "nick_name", "type").Updates(c)

	//updated records count
	s.logger.Debug("UpdateCustomerService result: ",
		zap.Int64("RowsAffected", result.RowsAffected),
		zap.Error(result.Error))

	if result.Error != nil {
		s.logger.Error("修改在线客服资料失败", zap.Error(result.Error))
		return result.Error
	} else {
		s.logger.Debug("修改在线客服资料成功")
	}

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

func (s *MysqlLianmiRepository) QueryGrades(req *Auth.GradeReq, pageIndex int, pageSize int, total *int64, where interface{}) ([]*models.Grade, error) {
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
func (s *MysqlLianmiRepository) AddGrade(req *Auth.AddGradeReq) (string, error) {
	var err error
	var index uint64

	redisConn := s.redisPool.Get()
	defer redisConn.Close()

	key := fmt.Sprintf("CustomerServiceInfo:%s", req.CustomerServiceUsername)

	cstype, err := redis.Int(redisConn.Do("HGET", key, "Type"))              //账号类型，1-客服，2-技术
	jobNumber, err := redis.String(redisConn.Do("HGET", key, "JobNumber"))   //工号
	evaluation, err := redis.String(redisConn.Do("HGET", key, "Evaluation")) //职称
	nickName, err := redis.String(redisConn.Do("HGET", key, "NickName"))     //呢称
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

	//如果没有记录，则增加，如果有记录，则更新全部字段
	if err := s.db.Clauses(clause.OnConflict{DoNothing: true}).Create(&c).Error; err != nil {
		s.logger.Error("增加Grade失败, failed to upsert Grade", zap.Error(err))
		return "", err
	} else {
		s.logger.Debug("增加Grade成功, upsert Grade succeed")
	}

	return title, nil

}

//用户提交评分，修改对应的字段
func (s *MysqlLianmiRepository) SubmitGrade(req *Auth.SubmitGradeReq) error {

	c := new(models.Grade)
	c.AppUsername = req.AppUsername
	c.GradeNum = int(req.GradeNum)

	where := models.Grade{
		Title: req.Title,
	}
	// 同时更新多个字段
	result := s.db.Model(&models.Grade{}).Where(&where).Select("app_username", "grade_num").Updates(c)

	//updated records count
	s.logger.Debug("SubmitGrade result: ",
		zap.Int64("RowsAffected", result.RowsAffected),
		zap.Error(result.Error))

	if result.Error != nil {
		s.logger.Error("用户提交评分Grade失败", zap.Error(result.Error))
		return result.Error
	} else {
		s.logger.Debug("用户提交评分Grade成功")
	}

	return nil

}
