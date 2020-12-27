package repositories

import (
	"fmt"
	"time"

	"github.com/gomodule/redigo/redis"
	Auth "github.com/lianmi/servers/api/proto/auth"
	Global "github.com/lianmi/servers/api/proto/global"
	LMCommon "github.com/lianmi/servers/internal/common"
	"github.com/lianmi/servers/internal/pkg/models"
	"github.com/lianmi/servers/util/dateutil"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	// "gorm.io/gorm"
	"gorm.io/gorm/clause"
)

//TODO
//商户查询当前名下用户总数
func (s *MysqlLianmiRepository) GetBusinessMembership(businessUsername string) (*Auth.GetBusinessMembershipResp, error) {
	var err error
	currYearMonth := dateutil.GetYearMonthString()

	//查询出total
	model := &models.BusinessUnderling{
		BusinessUsername: businessUsername,
	}
	db2 := s.db.Model(model).Where(model)
	var totalCount *int64
	err = db2.Count(totalCount).Error
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
	orderStr := "`year_month` desc" //按照年月降序
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

//对某个用户的推广会员佣金进行统计
func (s *MysqlLianmiRepository) CommissonSatistics(username string) (*Auth.CommissonSatisticsResp, error) {
	var err error

	type Amount struct{ Total float64 }
	amount := Amount{}

	currYearMonth := dateutil.GetYearMonthString()

	//用户的佣金月统计  CommissionStatistics
	nucsWhere := &models.CommissionStatistics{
		Username:  username,
		YearMonth: currYearMonth,
		IsRebate:  true, //判断是否返现
	}
	ncs := &models.CommissionStatistics{}
	if err = s.db.Model(ncs).Where(nucsWhere).First(ncs).Error; err == nil {
		s.logger.Error("CommissionStatistics表已经返现，不能新增记录 ", zap.String("YearMonth", currYearMonth), zap.String("Username", username))
	} else {

		//统计d.UsernameLevelOne对应的用户在当月的所有佣金总额
		where := models.Commission{
			UsernameLevel: username,
			YearMonth:     currYearMonth,
		}
		db2 := s.db.Model(&models.Commission{}).Where(&where)

		db2.Select("SUM(commission) AS total").Scan(&amount)
		s.logger.Debug("SUM统计出当月的总佣金金额",
			zap.String("username", username),
			zap.String("currYearMonth", currYearMonth),
			zap.Float64("total", amount.Total),
		)

		newnucs := models.CommissionStatistics{
			Username:        username,
			YearMonth:       currYearMonth,
			TotalCommission: amount.Total, //本月返佣总金额
			IsRebate:        false,        //默认返现的值是false
		}

		//Save
		s.db.Save(&newnucs)

	}

	//TODO
	resp := &Auth.CommissonSatisticsResp{
		TotalPage: 1,
	}
	resp.Summary = append(resp.Summary, &Auth.PerLevelSummary{
		Yearmonth:       currYearMonth,
		TotalCommission: amount.Total,
		IsRebate:        false,
		RebateTime:      uint64(time.Now().UnixNano() / 1e6),
	})
	return resp, nil
}

//用户查询按月统计发展的付费会员总数及返佣金额，是否已经返佣
func (s *MysqlLianmiRepository) GetCommissionStatistics(username string) (*Auth.GetCommssionsResp, error) {
	var err error
	total := new(int64)

	var bucss []*models.CommissionStatistics
	where := models.CommissionStatistics{Username: username}
	orderStr := "updated_at desc" //按照年月降序
	if err = s.base.GetPages(&models.CommissionStatistics{}, &bucss, 1, 100, total, &where, orderStr); err != nil {
		s.logger.Error("获取CommissionStatistics信息失败", zap.Error(err))
	}
	rsp := &Auth.GetCommssionsResp{}
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
				BusinessUsername: LMCommon.VipBusinessUsername,

				ProductID: vipPrice.ProductID,

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
