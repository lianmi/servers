package repositories

import (
	// "encoding/json"
	"fmt"
	// "time"
	// "net/http"

	// "github.com/golang/protobuf/proto"
	"github.com/gomodule/redigo/redis"
	// "github.com/jinzhu/gorm"
	// Auth "github.com/lianmi/servers/api/proto/auth"
	Auth "github.com/lianmi/servers/api/proto/auth"
	// User "github.com/lianmi/servers/api/proto/user"
	// "github.com/lianmi/servers/internal/app/dispatcher/grpcclients"
	// "github.com/lianmi/servers/internal/app/dispatcher/nsqMq"
	LMCommon "github.com/lianmi/servers/internal/common"
	"github.com/lianmi/servers/internal/pkg/models"
	"github.com/lianmi/servers/util/dateutil"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

//会员付费成功后，需要新增3条佣金记录及BusinessCommission表记录
func (s *MysqlLianmiRepository) SaveToCommission(username, orderID, content string, blockNumber uint64, txHash string) error {
	var err error
	currYearMonth := dateutil.GetYearMonthString()

	redisConn := s.redisPool.Get()
	defer redisConn.Close()

	//TODO  修改用户状态 state
	userKey := fmt.Sprintf("userData:%s", username)
	state, _ := redis.Int(redisConn.Do("HGET", userKey, "State"))

	// 状态 0-预审核 1-付费用户(购买会员) 2-封号
	if state == 0 {
		redisConn.Do("HSET", userKey, "State", 1)
	} else {
		return errors.Wrapf(err, "SaveToCommission error: this  user state is not equal 0")
	}

	//从Distribution层级表查出所有需要分配佣金的用户账号
	d := new(models.Distribution)
	if err = s.db.Model(d).Where(&models.Distribution{
		Username: username,
	}).First(d).Error; err != nil {
		//记录找不到也会触发错误
		return errors.Wrapf(err, "SaveToCommission error or username not found")
	}

	if d.BusinessUsername == "" {
		return errors.Wrapf(err, "businessUsername is empty error")
	} else {
		b := &models.BusinessCommission{
			MembershipUsername: username,
			BusinessUsername:   d.BusinessUsername,
		}
		e := &models.BusinessCommission{}
		if err = s.db.Model(e).Where(b).First(e).Error; err == nil {
			s.logger.Error("已经存在此用户，不能新增佣金记录", zap.String("username", username), zap.String("BusinessUsername", d.BusinessUsername))
			//记录不存在才能添加
			return errors.Wrapf(err, "Can not Insert BusinessCommission, because this username had exists")
		}

		//当商户不为空时候，则需要增加此商户的佣金
		bc := &models.BusinessCommission{
			YearMonth:          currYearMonth,
			MembershipUsername: username,                    //One Two Three
			BusinessUsername:   d.BusinessUsername,          //归属的商户注册账号id
			Amount:             LMCommon.MEMBERSHIPPRICE,    //会员费用金额，单位是人民币
			OrderID:            orderID,                     //订单ID
			Content:            content,                     //附言 Text文本类型
			BlockNumber:        blockNumber,                 //交易成功打包的区块高度
			TxHash:             txHash,                      //交易成功打包的区块哈希
			Commission:         LMCommon.CommissionBusiness, //商户的佣金，11元
		}
		tx := s.base.GetTransaction()
		if err := tx.Save(bc).Error; err != nil {
			s.logger.Error("保存BusinessCommission失败", zap.Error(err))
			tx.Rollback()
			return err
		}

		//提交
		tx.Commit()

		ee := &models.BusinessUserCommissionStatistics{}
		if err = s.db.Model(ee).Where(&models.BusinessUserCommissionStatistics{
			BusinessUsername: d.BusinessUsername,
			YearMonth:        currYearMonth,
			IsRebate:         true,
		}).First(ee).Error; err == nil {
			s.logger.Error("BusinessUserCommissionStatistics表已经返现，不能新增记录 ", zap.String("YearMonth", currYearMonth), zap.String("BusinessUsername", d.BusinessUsername))
			//记录不存在才能添加
			return errors.Wrapf(err, "Can not Insert BusinessUserCommissionStatistics, because IsRebate is true")
		}

		//查询出total
		model := &models.BusinessCommission{
			BusinessUsername: d.BusinessUsername,
			YearMonth:        currYearMonth,
		}
		db := s.db.Model(model).Where(model)
		var totalCount *uint64
		err := db.Count(totalCount).Error
		if err != nil {
			s.logger.Error("查询BusinessCommission总数出错",
				zap.String("BusinessUsername", d.BusinessUsername),
				zap.String("YearMonth", currYearMonth),
				zap.Error(err))
			return err
		}

		totalCommission := *totalCount * LMCommon.CommissionBusiness

		bcs := &models.BusinessUserCommissionStatistics{
			BusinessUsername: d.BusinessUsername,
			YearMonth:        currYearMonth,
			Total:            int64(*totalCount),       //本月新增付费会员总数
			TotalCommission:  float64(totalCommission), //本月返佣总金额
			IsRebate:         false,
		}

		tx2 := s.base.GetTransaction()

		if err := tx2.Save(bcs).Error; err != nil {
			s.logger.Error("保存BusinessUserCommissionStatistics失败", zap.Error(err))
			tx2.Rollback()
			return err
		}

		//提交
		tx2.Commit()
	}

	if d.UsernameLevelOne != "" {
		//支付成功后，需要插入佣金表Commission -  第一级
		b := &models.Commission{
			UsernameLevel:    d.UsernameLevelOne,
			BusinessUsername: d.BusinessUsername,
		}
		e := &models.Commission{}
		if err = s.db.Model(e).Where(b).First(e).Error; err == nil {
			s.logger.Error("已经存在此用户，不能新增佣金记录", zap.String("UsernameLevel", d.UsernameLevelOne), zap.String("BusinessUsername", d.BusinessUsername))
			//记录不存在才能添加
			return errors.Wrapf(err, "Can not Insert Commission, because this UsernameLevelOne had exists")
		}

		commissionOne := &models.Commission{
			YearMonth:        currYearMonth,
			UsernameLevel:    d.UsernameLevelOne,       //One Two Three
			BusinessUsername: d.BusinessUsername,       //归属的商户注册账号id
			Amount:           LMCommon.MEMBERSHIPPRICE, //会员费用金额，单位是人民币
			OrderID:          orderID,                  //订单ID
			Content:          content,                  //附言 Text文本类型
			BlockNumber:      blockNumber,              //交易成功打包的区块高度
			TxHash:           txHash,                   //交易成功打包的区块哈希
			Commission:       LMCommon.CommissionOne,   //第一级佣金
		}
		tx := s.base.GetTransaction()
		if err := tx.Save(commissionOne).Error; err != nil {
			s.logger.Error("保存commissionOne失败", zap.Error(err))
			tx.Rollback()
			return err
		}
		//提交
		tx.Commit()

		//普通用户的佣金月统计  NormalUserCommissionStatistics
		nucs := &models.NormalUserCommissionStatistics{
			Username:  d.UsernameLevelOne,
			YearMonth: currYearMonth,
			IsRebate:  true,
		}
		ncs := &models.NormalUserCommissionStatistics{}
		if err = s.db.Model(ncs).Where(nucs).First(ncs).Error; err == nil {
			s.logger.Error("NormalUserCommissionStatistics表已经返现，不能新增记录 ", zap.String("YearMonth", currYearMonth), zap.String("Username", d.UsernameLevelOne))
		} else {

			//统计d.UsernameLevelOne对应的用户在当月的所有佣金总额
			model := &models.Commission{
				BusinessUsername: d.BusinessUsername,
				UsernameLevel:    d.UsernameLevelOne,
				YearMonth:        currYearMonth,
			}
			db := s.db.Model(model).Where(model)
			type Amount struct{ Total float64 }
			amount := Amount{}
			db.Select("SUM(commission) AS total").Scan(&amount)

			newnucs := &models.NormalUserCommissionStatistics{
				Username:        d.UsernameLevelOne,
				YearMonth:       currYearMonth,
				TotalCommission: amount.Total, //本月返佣总金额
				IsRebate:        false,
			}

			tx2 := s.base.GetTransaction()

			if err := tx2.Save(newnucs).Error; err != nil {
				s.logger.Error("保存NormalUserCommissionStatistics失败", zap.Error(err))
				tx2.Rollback()
				return err
			}

			//提交
			tx2.Commit()
		}

	}

	if d.UsernameLevelTwo != "" {
		//支付成功后，需要插入佣金表Commission -  第二级
		b := &models.Commission{
			UsernameLevel:    d.UsernameLevelTwo,
			BusinessUsername: d.BusinessUsername,
		}
		e := &models.Commission{}
		if err = s.db.Model(b).Where(b).First(e).Error; err == nil {
			s.logger.Error("已经存在此用户，不能新增佣金记录", zap.String("UsernameLevel", d.UsernameLevelTwo), zap.String("BusinessUsername", d.BusinessUsername))
			//记录不存在才能添加
			return errors.Wrapf(err, "Can not Insert Commission, because this UsernameLevelTwo had exists")
		}

		commissionTwo := &models.Commission{
			YearMonth:        currYearMonth,
			UsernameLevel:    d.UsernameLevelTwo,       //One Two Three
			BusinessUsername: d.BusinessUsername,       //归属的商户注册账号id
			Amount:           LMCommon.MEMBERSHIPPRICE, //会员费用金额，单位是人民币
			OrderID:          orderID,                  //订单ID
			Content:          content,                  //附言 Text文本类型
			BlockNumber:      blockNumber,              //交易成功打包的区块高度
			TxHash:           txHash,                   //交易成功打包的区块哈希
			Commission:       LMCommon.CommissionTwo,   //第二级佣金
		}
		tx := s.base.GetTransaction()
		if err := tx.Save(commissionTwo).Error; err != nil {
			s.logger.Error("保存commissionTwo失败", zap.Error(err))
			tx.Rollback()
			return err
		}
		//提交
		tx.Commit()

		//普通用户的佣金月统计  NormalUserCommissionStatistics
		nucs := &models.NormalUserCommissionStatistics{
			Username:  d.UsernameLevelTwo,
			YearMonth: currYearMonth,
			IsRebate:  true,
		}
		ncs := &models.NormalUserCommissionStatistics{}
		if err = s.db.Model(nucs).Where(nucs).First(ncs).Error; err == nil {
			s.logger.Error("NormalUserCommissionStatistics表已经返现，不能新增记录 ", zap.String("YearMonth", currYearMonth), zap.String("Username", d.UsernameLevelTwo))
		} else {
			//统计d.UsernameLevelTwo对应的用户在当月的所有佣金总额
			model := &models.Commission{
				BusinessUsername: d.BusinessUsername,
				UsernameLevel:    d.UsernameLevelTwo,
				YearMonth:        currYearMonth,
			}
			db := s.db.Model(model).Where(model)
			type Amount struct{ Total float64 }
			amount := Amount{}
			db.Select("SUM(commission) AS total").Scan(&amount)

			newnucs := &models.NormalUserCommissionStatistics{
				Username:        d.UsernameLevelTwo,
				YearMonth:       currYearMonth,
				TotalCommission: amount.Total, //本月返佣总金额
				IsRebate:        false,
			}

			tx2 := s.base.GetTransaction()

			if err := tx2.Save(newnucs).Error; err != nil {
				s.logger.Error("保存NormalUserCommissionStatistics失败", zap.Error(err))
				tx2.Rollback()
				return err
			}

			//提交
			tx2.Commit()
		}

	}

	if d.UsernameLevelThree != "" {
		//支付成功后，需要插入佣金表Commission -  第三级
		b := &models.Commission{
			UsernameLevel:    d.UsernameLevelThree,
			BusinessUsername: d.BusinessUsername,
		}
		e := &models.Commission{}
		if err = s.db.Model(b).Where(b).First(e).Error; err == nil {
			s.logger.Error("已经存在此用户，不能新增佣金记录", zap.String("UsernameLevel", d.UsernameLevelThree), zap.String("BusinessUsername", d.BusinessUsername))
			//记录不存在才能添加
			return errors.Wrapf(err, "Can not Insert Commission, because this UsernameLevelThree had exists")
		}
		commissionThree := &models.Commission{
			YearMonth:        currYearMonth,
			UsernameLevel:    d.UsernameLevelThree,     //One Two Three
			BusinessUsername: d.BusinessUsername,       //归属的商户注册账号id
			Amount:           LMCommon.MEMBERSHIPPRICE, //会员费用金额，单位是人民币
			OrderID:          orderID,                  //订单ID
			Content:          content,                  //附言 Text文本类型
			BlockNumber:      blockNumber,              //交易成功打包的区块高度
			TxHash:           txHash,                   //交易成功打包的区块哈希
			Commission:       LMCommon.CommissionThree, //第三级佣金
		}
		tx := s.base.GetTransaction()
		if err := tx.Save(commissionThree).Error; err != nil {
			s.logger.Error("保存commissionThree失败", zap.Error(err))
			tx.Rollback()
			return err
		}
		//提交
		tx.Commit()

		//普通用户的佣金月统计  NormalUserCommissionStatistics
		nucs := &models.NormalUserCommissionStatistics{
			Username:  d.UsernameLevelThree,
			YearMonth: currYearMonth,
			IsRebate:  true,
		}
		ncs := &models.NormalUserCommissionStatistics{}
		if err = s.db.Model(ncs).Where(nucs).First(ncs).Error; err == nil {
			s.logger.Error("NormalUserCommissionStatistics表已经返现，不能新增记录 ", zap.String("YearMonth", currYearMonth), zap.String("Username", d.UsernameLevelThree))
		} else {
			//统计d.UsernameLevelThree对应的用户在当月的所有佣金总额
			model := &models.Commission{
				BusinessUsername: d.BusinessUsername,
				UsernameLevel:    d.UsernameLevelThree,
				YearMonth:        currYearMonth,
			}
			db := s.db.Model(model).Where(model)
			type Amount struct{ Total float64 }
			amount := Amount{}
			db.Select("SUM(commission) AS total").Scan(&amount)

			newnucs := &models.NormalUserCommissionStatistics{
				Username:        d.UsernameLevelThree,
				YearMonth:       currYearMonth,
				TotalCommission: amount.Total, //本月返佣总金额
				IsRebate:        false,
			}

			tx2 := s.base.GetTransaction()

			if err := tx2.Save(newnucs).Error; err != nil {
				s.logger.Error("保存NormalUserCommissionStatistics失败", zap.Error(err))
				tx2.Rollback()
				return err
			}

			//提交
			tx2.Commit()
		}

	}

	return nil
}

//TODO
//商户查询当前名下用户总数，按月统计付费会员总数及返佣金额，是否已经返佣
func (s *MysqlLianmiRepository) GetBusinessMembership(businessUsername string) (*Auth.GetBusinessMembershipResp, error) {
	var err error
	currYearMonth := dateutil.GetYearMonthString()

	//查询出total
	model := &models.BusinessCommission{
		BusinessUsername: businessUsername,
		YearMonth:        currYearMonth,
	}
	db := s.db.Model(model).Where(model)
	var totalCount *uint64
	err = db.Count(totalCount).Error
	if err != nil {
		s.logger.Error("查询BusinessCommission总数出错",
			zap.String("BusinessUsername", businessUsername),
			zap.String("YearMonth", currYearMonth),
			zap.Error(err))
		return nil, err
	}
	rsp := &Auth.GetBusinessMembershipResp{
		Totalmembers: *totalCount,
	}

	total := new(uint64)
	var bucss []*models.BusinessUserCommissionStatistics
	where := &models.BusinessUserCommissionStatistics{BusinessUsername: businessUsername}
	orderStr := "year_month desc" //按照年月降序
	if err := s.base.GetPages(&models.BusinessUserCommissionStatistics{}, &bucss, 1, 100, total, where, orderStr); err != nil {
		s.logger.Error("获取BusinessUserCommissionStatistics信息失败", zap.Error(err))
	}

	for _, record := range bucss {
		rsp.Details = append(rsp.Details, &Auth.BusinessUserMonthDetail{
			BusinessUsername: businessUsername,
			YearMonth:        record.YearMonth,
			Total:            uint64(record.Total),
			TotalCommission:  record.TotalCommission,
			IsRebate:         record.IsRebate,
			RebateTime:       uint64(record.RebateTime),
		})
	}
	_ = total

	return rsp, err

}

//用户按月统计付费会员总数及返佣金额，是否已经返佣
func (s *MysqlLianmiRepository) GetNormalMembership(username string) (*Auth.GetMembershipResp, error) {
	var err error
	total := new(uint64)

	var bucss []*models.NormalUserCommissionStatistics
	where := &models.NormalUserCommissionStatistics{Username: username}
	orderStr := "year_month desc" //按照年月降序
	if err = s.base.GetPages(&models.NormalUserCommissionStatistics{}, &bucss, 1, 100, total, where, orderStr); err != nil {
		s.logger.Error("获取NormalUserCommissionStatistics信息失败", zap.Error(err))
	}
	rsp := &Auth.GetMembershipResp{}
	for _, record := range bucss {
		rsp.CommssionDetails = append(rsp.CommssionDetails, &Auth.UserMonthCommssionDetail{
			Username:        username,
			YearMonth:       record.YearMonth,
			TotalCommission: record.TotalCommission,
			IsRebate:        record.IsRebate,
			RebateTime:      uint64(record.RebateTime),
		})
	}
	_ = total

	return rsp, nil
}

//保存提交佣金提现申请(商户，用户)
func (s *MysqlLianmiRepository) SubmitCommssionWithdraw(username, yearMonth string) (*Auth.CommssionWithdrawResp, error) {
	var err error
	var withdrawCommission float64

	redisConn := s.redisPool.Get()
	defer redisConn.Close()

	//获取 用户UserType
	userKey := fmt.Sprintf("userData:%s", username)
	userType, _ := redis.Int(redisConn.Do("HGET", userKey, "UserType"))

	//获取 yearMonth对应的 withdrawCommission
	if userType == 1 { //用户类型 1-普通，
		nucs := new(models.NormalUserCommissionStatistics)
		if err = s.db.Model(nucs).Where(&models.NormalUserCommissionStatistics{
			Username:  username,
			YearMonth: yearMonth,
		}).First(nucs).Error; err != nil {
			//记录找不到也会触发错误
			return nil, errors.Wrapf(err, "SubmitCommssionWithdraw error or Username not found")
		}
		withdrawCommission = nucs.TotalCommission

	} else if userType == 2 { //用户类型 2-商户
		bucs := new(models.BusinessUserCommissionStatistics)
		if err = s.db.Model(bucs).Where(&models.BusinessUserCommissionStatistics{
			BusinessUsername: username,
			YearMonth:        yearMonth,
		}).First(bucs).Error; err != nil {
			//记录找不到也会触发错误
			return nil, errors.Wrapf(err, "SubmitCommssionWithdraw error or BusinessUsername not found")
		}
		withdrawCommission = bucs.TotalCommission
	} else {
		return nil, errors.Wrapf(err, "SubmitCommssionWithdraw error: usertype not found")
	}

	commissionWithdraw := &models.CommissionWithdraw{
		Username:           username,
		UserType:           userType,
		YearMonth:          yearMonth,
		WithdrawCommission: withdrawCommission,
	}
	tx := s.base.GetTransaction()

	if err := tx.Save(commissionWithdraw).Error; err != nil {
		s.logger.Error("保存CommissionWithdraw失败", zap.Error(err))
		tx.Rollback()
		return nil, err
	}

	//提交
	tx.Commit()

	rsp := &Auth.CommssionWithdrawResp{
		Username:        username,
		YearMonth:       yearMonth,
		CommssionAmount: withdrawCommission,
	}
	return rsp, nil
}