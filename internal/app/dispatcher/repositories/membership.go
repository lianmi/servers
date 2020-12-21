package repositories

import (
	"fmt"
	// "time"

	"github.com/gomodule/redigo/redis"
	Auth "github.com/lianmi/servers/api/proto/auth"
	Global "github.com/lianmi/servers/api/proto/global"
	LMCommon "github.com/lianmi/servers/internal/common"
	"github.com/lianmi/servers/internal/pkg/models"
	"github.com/lianmi/servers/util/dateutil"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

//会员付费成功后，按系统设定的比例进行佣金计算及写库， 需要新增3条佣金amount记录
func (s *MysqlLianmiRepository) AddCommission(orderTotalAmount float64, username, orderID, content string, blockNumber uint64, txHash string) error {
	var err error
	currYearMonth := dateutil.GetYearMonthString()

	redisConn := s.redisPool.Get()
	defer redisConn.Close()

	//修改用户状态 state
	userKey := fmt.Sprintf("userData:%s", username)
	state, _ := redis.Int(redisConn.Do("HGET", userKey, "State"))

	// 状态 0-预审核 1-付费用户(购买会员) 2-封号
	if state == 0 {
		redisConn.Do("HSET", userKey, "State", 1)
	} else {
		return errors.Wrapf(err, "AddCommission error: this  user state is not equal 0")
	}

	//从Distribution层级表查出所有需要分配佣金的用户账号
	distribution := new(models.Distribution)
	if err = s.db.Model(distribution).Where(&models.Distribution{
		Username: username,
	}).First(distribution).Error; err != nil {
		//记录找不到也会触发错误
		return errors.Wrapf(err, "AddCommission error or username not found")
	}

	if distribution.BusinessUsername == "" {
		return errors.Wrapf(err, "businessUsername is empty error")
	} else {
		e := &models.BusinessUnderling{}
		if err = s.db.Model(e).Where(&models.BusinessUnderling{
			MembershipUsername: username,
			BusinessUsername:   distribution.BusinessUsername,
		}).First(e).Error; err == nil {
			s.logger.Error("已经存在此用户，不能新增记录", zap.String("username", username), zap.String("BusinessUsername", distribution.BusinessUsername))
			//记录不存在才能添加
			return errors.Wrapf(err, "Can not Insert BusinessUnderling, because this username had exists")
		}

		//当商户不为空时候，则需要增加记录
		bc := &models.BusinessUnderling{
			MembershipUsername: username,                      //One Two Three
			BusinessUsername:   distribution.BusinessUsername, //归属的商户注册账号id
		}

		//如果没有记录，则增加，如果有记录，则更新全部字段
		if err := s.db.Clauses(clause.OnConflict{DoNothing: true}).Create(&bc).Error; err != nil {
			s.logger.Error("增加BusinessUnderling失败, failed to upsert BusinessUnderling", zap.Error(err))
			return err
		} else {
			s.logger.Debug("增加BusinessUnderling成功, upsert BusinessUnderling succeed")
		}

		//增加到店铺下属用户列表 redis SADD, SMEMBERS可以获取该商户的全部下属用户总数
		storeUsersKey := fmt.Sprintf("StoreUsers:%s", distribution.BusinessUsername)
		if _, err = redisConn.Do("SADD", storeUsersKey, username); err != nil {
			s.logger.Error("SADD storelikeKey Error", zap.Error(err))
			return err
		}

		ee := &models.BusinessUserStatistics{}
		where := models.BusinessUserStatistics{
			BusinessUsername: distribution.BusinessUsername,
			YearMonth:        currYearMonth,
		}
		//查询出该商户的全部下属用户总数
		model := &models.BusinessUnderling{
			BusinessUsername: distribution.BusinessUsername,
		}
		db := s.db.Model(model).Where(model)
		var totalCount *int64
		err := db.Count(totalCount).Error
		if err != nil {
			s.logger.Error("查询BusinessUnderling总数出错",
				zap.String("BusinessUsername", distribution.BusinessUsername),
				zap.String("YearMonth", currYearMonth),
				zap.Error(err))
			return err
		}

		if err = s.db.Model(ee).Where(&where).First(ee).Error; err != nil {
			//记录不存在, 需要添加
			if errors.Is(err, gorm.ErrRecordNotFound) {

				bcs := &models.BusinessUserStatistics{
					BusinessUsername: distribution.BusinessUsername,
					YearMonth:        currYearMonth,
					UnderlingTotal:   int64(*totalCount), //本月新增会员总数
				}

				tx2 := s.base.GetTransaction()

				if err := tx2.Create(bcs).Error; err != nil {
					s.logger.Error("增加BusinessUserStatistics失败", zap.Error(err))
					tx2.Rollback()
					return err
				}

				//提交
				tx2.Commit()
			} else {

				return errors.Wrapf(err, "Db errp")

			}
		} else {
			//记录存在, Update

			result := s.db.Model(ee).Where(&where).Update("underling_total", int64(*totalCount))
			s.logger.Debug("Update BusinessUserStatistics result: ", zap.Int64("RowsAffected", result.RowsAffected), zap.Error(result.Error))

			if result.Error != nil {
				s.logger.Error("Update BusinessUserStatistics失败", zap.Error(result.Error))
				return result.Error
			} else {
				mtxt := fmt.Sprintf("Update BusinessUserStatistics成功:  本月新增会员总数: %distribution", int64(*totalCount))
				s.logger.Debug(mtxt)
			}
		}

	}

	if distribution.UsernameLevelOne != "" {
		//支付成功后，需要插入佣金表Commission -  第一级
		e := &models.Commission{}
		if err = s.db.Model(e).Where(&models.Commission{
			UsernameLevel:    distribution.UsernameLevelOne,
			BusinessUsername: distribution.BusinessUsername,
		}).First(e).Error; err == nil {
			s.logger.Error("已经存在此用户佣金记录，不能新增", zap.String("UsernameLevel", distribution.UsernameLevelOne), zap.String("BusinessUsername", distribution.BusinessUsername))
			//记录不存在才能添加
			return errors.Wrapf(err, "Can not Insert Commission, because this UsernameLevelOne had exists")
		}

		commissionOne := &models.Commission{
			YearMonth:        currYearMonth,
			UsernameLevel:    distribution.UsernameLevelOne,             //One Two Three
			BusinessUsername: distribution.BusinessUsername,             //归属的商户注册账号id
			Amount:           orderTotalAmount,                          //会员费用金额，单位是人民币
			OrderID:          orderID,                                   //订单ID
			Content:          content,                                   //附言 Text文本类型
			BlockNumber:      blockNumber,                               //交易成功打包的区块高度
			TxHash:           txHash,                                    //交易成功打包的区块哈希
			Commission:       LMCommon.CommissionOne * orderTotalAmount, //TODO 第一级佣金， 按比例
		}

		//如果没有记录，则增加，如果有记录，则更新全部字段
		if err := s.db.Clauses(clause.OnConflict{DoNothing: true}).Create(&commissionOne).Error; err != nil {
			s.logger.Error("增加commissionOne失败, failed to upsert Commission", zap.Error(err))
			return err
		} else {
			s.logger.Debug("增加commissionOne成功, upsert Commission succeed")
		}

		//普通用户的佣金月统计  CommissionStatistics
		nucs := &models.CommissionStatistics{
			Username:  distribution.UsernameLevelOne,
			YearMonth: currYearMonth,
			IsRebate:  true,
		}
		ncs := &models.CommissionStatistics{}
		if err = s.db.Model(ncs).Where(nucs).First(ncs).Error; err == nil {
			s.logger.Error("NormalUserCommissionStatistics表已经返现，不能新增记录 ", zap.String("YearMonth", currYearMonth), zap.String("Username", distribution.UsernameLevelOne))
		} else {

			//统计d.UsernameLevelOne对应的用户在当月的所有佣金总额
			model := &models.Commission{
				BusinessUsername: distribution.BusinessUsername,
				UsernameLevel:    distribution.UsernameLevelOne,
				YearMonth:        currYearMonth,
			}
			db := s.db.Model(model).Where(model)
			type Amount struct{ Total float64 }
			amount := Amount{}
			db.Select("SUM(commission) AS total").Scan(&amount)

			newnucs := &models.CommissionStatistics{
				Username:        distribution.UsernameLevelOne,
				YearMonth:       currYearMonth,
				TotalCommission: amount.Total, //本月返佣总金额
				IsRebate:        false,
			}

			tx2 := s.base.GetTransaction()

			if err := tx2.Create(newnucs).Error; err != nil {
				s.logger.Error("增加NormalUserCommissionStatistics失败", zap.Error(err))
				tx2.Rollback()
				return err
			}

			//提交
			tx2.Commit()
		}

	}

	if distribution.UsernameLevelTwo != "" {
		//支付成功后，需要插入佣金表Commission -  第二级
		e := &models.Commission{}
		if err = s.db.Model(e).Where(&models.Commission{
			UsernameLevel:    distribution.UsernameLevelTwo,
			BusinessUsername: distribution.BusinessUsername,
		}).First(e).Error; err == nil {
			s.logger.Error("已经存在此用户，不能新增佣金记录", zap.String("UsernameLevel", distribution.UsernameLevelTwo), zap.String("BusinessUsername", distribution.BusinessUsername))
			//记录不存在才能添加
			return errors.Wrapf(err, "Can not Insert Commission, because this UsernameLevelTwo had exists")
		}

		commissionTwo := &models.Commission{
			YearMonth:        currYearMonth,
			UsernameLevel:    distribution.UsernameLevelTwo,             //One Two Three
			BusinessUsername: distribution.BusinessUsername,             //归属的商户注册账号id
			Amount:           orderTotalAmount,                          //会员费用金额，单位是人民币
			OrderID:          orderID,                                   //订单ID
			Content:          content,                                   //附言 Text文本类型
			BlockNumber:      blockNumber,                               //交易成功打包的区块高度
			TxHash:           txHash,                                    //交易成功打包的区块哈希
			Commission:       LMCommon.CommissionTwo * orderTotalAmount, //TODO 第二级佣金
		}

		//如果没有记录，则增加，如果有记录，则更新全部字段
		if err := s.db.Clauses(clause.OnConflict{DoNothing: true}).Create(&commissionTwo).Error; err != nil {
			s.logger.Error("增加commissionTwo失败, failed to upsert Commission", zap.Error(err))
			return err
		} else {
			s.logger.Debug("增加commissionTwo成功, upsert Commission succeed")
		}

		//普通用户的佣金月统计  CommissionStatistics
		nucs := &models.CommissionStatistics{
			Username:  distribution.UsernameLevelTwo,
			YearMonth: currYearMonth,
			IsRebate:  true,
		}
		ncs := &models.CommissionStatistics{}
		if err = s.db.Model(nucs).Where(nucs).First(ncs).Error; err == nil {
			s.logger.Error("NormalUserCommissionStatistics表已经返现，不能新增记录 ", zap.String("YearMonth", currYearMonth), zap.String("Username", distribution.UsernameLevelTwo))
		} else {
			//统计d.UsernameLevelTwo对应的用户在当月的所有佣金总额
			model := &models.Commission{
				BusinessUsername: distribution.BusinessUsername,
				UsernameLevel:    distribution.UsernameLevelTwo,
				YearMonth:        currYearMonth,
			}
			db := s.db.Model(model).Where(model)
			type Amount struct{ Total float64 }
			amount := Amount{}
			db.Select("SUM(commission) AS total").Scan(&amount)

			newnucs := &models.CommissionStatistics{
				Username:        distribution.UsernameLevelTwo,
				YearMonth:       currYearMonth,
				TotalCommission: amount.Total, //本月返佣总金额
				IsRebate:        false,
			}

			tx2 := s.base.GetTransaction()

			if err := tx2.Create(newnucs).Error; err != nil {
				s.logger.Error("增加NormalUserCommissionStatistics失败", zap.Error(err))
				tx2.Rollback()
				return err
			}

			//提交
			tx2.Commit()
		}

	}

	if distribution.UsernameLevelThree != "" {
		//支付成功后，需要插入佣金表Commission -  第三级
		e := &models.Commission{}
		if err = s.db.Model(e).Where(&models.Commission{
			UsernameLevel:    distribution.UsernameLevelThree,
			BusinessUsername: distribution.BusinessUsername,
		}).First(e).Error; err == nil {
			s.logger.Error("已经存在此用户，不能新增佣金记录", zap.String("UsernameLevel", distribution.UsernameLevelThree), zap.String("BusinessUsername", distribution.BusinessUsername))
			//记录不存在才能添加
			return errors.Wrapf(err, "Can not Insert Commission, because this UsernameLevelThree had exists")
		}
		commissionThree := &models.Commission{
			YearMonth:        currYearMonth,
			UsernameLevel:    distribution.UsernameLevelThree,             //One Two Three
			BusinessUsername: distribution.BusinessUsername,               //归属的商户注册账号id
			Amount:           orderTotalAmount,                            //会员费用金额，单位是人民币
			OrderID:          orderID,                                     //订单ID
			Content:          content,                                     //附言 Text文本类型
			BlockNumber:      blockNumber,                                 //交易成功打包的区块高度
			TxHash:           txHash,                                      //交易成功打包的区块哈希
			Commission:       LMCommon.CommissionThree * orderTotalAmount, //TODO 第三级佣金
		}

		//如果没有记录，则增加，如果有记录，则更新全部字段
		if err := s.db.Clauses(clause.OnConflict{DoNothing: true}).Create(&commissionThree).Error; err != nil {
			s.logger.Error("增加commissionThree失败, failed to upsert Commission", zap.Error(err))
			return err
		} else {
			s.logger.Debug("增加commissionThree成功, upsert Commission succeed")
		}

		//普通用户的佣金月统计  CommissionStatistics
		nucs := &models.CommissionStatistics{
			Username:  distribution.UsernameLevelThree,
			YearMonth: currYearMonth,
			IsRebate:  true,
		}
		ncs := &models.CommissionStatistics{}
		if err = s.db.Model(ncs).Where(nucs).First(ncs).Error; err == nil {
			s.logger.Error("NormalUserCommissionStatistics表已经返现，不能新增记录 ", zap.String("YearMonth", currYearMonth), zap.String("Username", distribution.UsernameLevelThree))
		} else {
			//统计d.UsernameLevelThree对应的用户在当月的所有佣金总额
			model := &models.Commission{
				BusinessUsername: distribution.BusinessUsername,
				UsernameLevel:    distribution.UsernameLevelThree,
				YearMonth:        currYearMonth,
			}
			db := s.db.Model(model).Where(model)
			type Amount struct{ Total float64 }
			amount := Amount{}
			db.Select("SUM(commission) AS total").Scan(&amount)

			newnucs := &models.CommissionStatistics{
				Username:        distribution.UsernameLevelThree,
				YearMonth:       currYearMonth,
				TotalCommission: amount.Total, //本月返佣总金额
				IsRebate:        false,
			}

			tx2 := s.base.GetTransaction()

			if err := tx2.Create(newnucs).Error; err != nil {
				s.logger.Error("增加NormalUserCommissionStatistics失败", zap.Error(err))
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
//商户查询当前名下用户总数
func (s *MysqlLianmiRepository) GetBusinessMembership(businessUsername string) (*Auth.GetBusinessMembershipResp, error) {
	var err error
	currYearMonth := dateutil.GetYearMonthString()

	//查询出total
	model := &models.BusinessUnderling{
		BusinessUsername: businessUsername,
	}
	db := s.db.Model(model).Where(model)
	var totalCount *int64
	err = db.Count(totalCount).Error
	if err != nil {
		s.logger.Error("查询BusinessUnderling总数出错",
			zap.String("BusinessUsername", businessUsername),
			zap.String("YearMonth", currYearMonth),
			zap.Error(err))
		return nil, err
	}
	rsp := &Auth.GetBusinessMembershipResp{
		Totalmembers: uint64(*totalCount),
	}

	total := new(int64)
	var bucss []*models.BusinessUserStatistics
	where := models.BusinessUserStatistics{BusinessUsername: businessUsername}
	orderStr := "year_month desc" //按照年月降序
	if err := s.base.GetPages(&models.BusinessUserStatistics{}, &bucss, 1, 100, total, &where, orderStr); err != nil {
		s.logger.Error("获取BusinessUserStatistics信息失败", zap.Error(err))
	}

	for _, v := range bucss {
		rsp.Details = append(rsp.Details, &Auth.BusinessUserMonthDetail{
			BusinessUsername: businessUsername,
			YearMonth:        v.YearMonth,
			Total:            uint64(v.UnderlingTotal),
		})
	}
	_ = total

	return rsp, err

}

//用户按月统计付费会员总数及返佣金额，是否已经返佣
func (s *MysqlLianmiRepository) GetNormalMembership(username string) (*Auth.GetMembershipResp, error) {
	var err error
	total := new(int64)

	var bucss []*models.CommissionStatistics
	where := models.CommissionStatistics{Username: username}
	orderStr := "year_month desc" //按照年月降序
	if err = s.base.GetPages(&models.CommissionStatistics{}, &bucss, 1, 100, total, &where, orderStr); err != nil {
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

//保存提交佣金提现申请(商户，用户通用)
func (s *MysqlLianmiRepository) SubmitCommssionWithdraw(username, yearMonth string) (*Auth.CommssionWithdrawResp, error) {
	var err error
	var withdrawCommission float64

	redisConn := s.redisPool.Get()
	defer redisConn.Close()

	//获取 用户UserType
	userKey := fmt.Sprintf("userData:%s", username)
	userType, _ := redis.Int(redisConn.Do("HGET", userKey, "UserType"))

	//获取 yearMonth对应的 withdrawCommission

		nucs := new(models.CommissionStatistics)
		if err = s.db.Model(nucs).Where(&models.CommissionStatistics{
			Username:  username,
			YearMonth: yearMonth,
		}).First(nucs).Error; err != nil {
			//记录找不到也会触发错误
			return nil, errors.Wrapf(err, "SubmitCommssionWithdraw error or Username not found")
		}
		withdrawCommission = nucs.TotalCommission



	commissionWithdraw := &models.CommissionWithdraw{
		Username:           username,
		UserType:           userType,
		YearMonth:          yearMonth,
		WithdrawCommission: withdrawCommission,
	}
	//如果没有记录，则增加，如果有记录，则更新全部字段
	if err := s.db.Clauses(clause.OnConflict{DoNothing: true}).Create(&commissionWithdraw).Error; err != nil {
		s.logger.Error("增加commissionWithdraw失败, failed to upsert CommissionWithdraw", zap.Error(err))
		return nil, err
	} else {
		s.logger.Debug("增加commissionWithdraw成功, upsert CommissionWithdraw succeed")
	}

	rsp := &Auth.CommssionWithdrawResp{
		Username:        username,
		YearMonth:       yearMonth,
		CommssionAmount: withdrawCommission,
	}
	return rsp, nil
}

func (s *MysqlLianmiRepository) GetVipPriceList(payType int) (*Auth.GetVipPriceResp, error) {
	var vipPriceList []models.VipPrice
	var where models.VipPrice
	if payType > 0 {
		where = models.VipPrice{
			PayType: payType,
		}

	}
	if err := s.db.Model(vipPriceList).Where(&where).Find(&vipPriceList).Error; err != nil {
		return nil, errors.Wrapf(err, "PayType not found[payType=%distribution]", payType)
	}
	var resp Auth.GetVipPriceResp
	for _, vipPrice := range vipPriceList {

		//此价格是否激活，true的状态才可用
		if vipPrice.IsActive {

			resp.Pricelist = append(resp.Pricelist, &Auth.VipPrice{

				PayType: Global.VipUserPayType(vipPrice.PayType), //VIP类型，1-包年，2-包季， 3-包月

				Title: vipPrice.Title, //价格标题说明

				Price: vipPrice.Price, //价格, 单位: 元

				Days: int32(vipPrice.Days), //开通时长 本记录对应的天数，例如包年增加365天，包季是90天，包月是30天

			})
		}
	}
	return &resp, nil
}

//根据PayType获取到VIP价格
func (s *MysqlLianmiRepository) GetVipUserPrice(payType int) (*models.VipPrice, error) {
	p := new(models.VipPrice)
	where := models.VipPrice{
		PayType: payType,
	}
	if err := s.db.Model(p).Where(&where).First(p).Error; err != nil {
		return nil, errors.Wrapf(err, "PayType not found[payType=%distribution]", payType)
	}
	return p, nil
}
