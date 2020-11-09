package repositories

import (
	// "encoding/json"
	// "fmt"
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

	//从Distribution层级表查出所有需要分配佣金的用户账号
	d := new(models.Distribution)
	if err = s.db.Model(d).Where("username = ?", username).First(d).Error; err != nil {
		//记录找不到也会触发错误
		return errors.Wrapf(err, "SaveToCommission error or username not found")
	}

	if d.BusinessUsername == "" {
		return errors.Wrapf(err, "businessUsername is empty error")
	} else {
		b := new(models.BusinessCommission{
			MembershipUsername: username,
			BusinessUsername:   d.BusinessUsername,
		})
		e := &models.BusinessCommission{}
		if err = s.db.Model(b).First(e).Error; err == nil {
			s.logger.Error("已经存在此用户，不能新增佣金记录", zap.String("username", username), zap.String("BusinessUsername", d.BusinessUsername))
			//记录不存在才能添加
			return errors.Wrapf(err, "Can not Insert BusinessCommission, because this username had exists")
		}

		//当商户不为空时候，则需要增加此商户的佣金
		bc := &models.BusinessCommission{
			MembershipUsername: username,                    //One Two Three
			BusinessUsername:   d.BusinessUsername,          //归属的商户注册账号id
			Amount:             LMCommon.MEMBERSHIPPRICE,    //会员费用金额，单位是人民币
			OrderID:            orderID,                     //订单ID
			Content:            content,                     //附言 Text文本类型
			BlockNumber:        blockNumber,                 //交易成功打包的区块高度
			TxHash:             txHash,                      //交易成功打包的区块哈希
			Commission:         LMCommon.CommissionBusiness, //商户的佣金，11元
		}
		currYearMonth := dateutil.GetYearMonthString()

		buss := &models.BusinessUserCommissionStatistics{
			BusinessUsername: d.BusinessUsername,
			YearMonth:        currYearMonth,
			IsRebate:         true,
		}
		ee := &models.BusinessUserCommissionStatistics{}
		if err = s.db.Model(buss).First(ee).Error; err == nil {
			s.logger.Error("BusinessUserCommissionStatistics表已经返现，不能新增记录 ", zap.String("YearMonth", currYearMonth), zap.String("BusinessUsername", d.BusinessUsername))
			//记录不存在才能添加
			return errors.Wrapf(err, "Can not Insert BusinessUserCommissionStatistics, because IsRebate is true")
		}
		tx := s.base.GetTransaction()
		if err := tx.Save(bc).Error; err != nil {
			s.logger.Error("保存BusinessCommission失败", zap.Error(err))
			tx.Rollback()
			return err
		}

		//提交
		tx.Commit()
	}

	if d.UsernameLevelOne != "" {
		//支付成功后，需要插入佣金表Commission -  第一级
		b := new(models.Commission{
			UsernameLevel:    d.UsernameLevelOne,
			BusinessUsername: d.BusinessUsername,
		})
		e := &models.Commission{}
		if err = s.db.Model(b).First(e).Error; err == nil {
			s.logger.Error("已经存在此用户，不能新增佣金记录", zap.String("UsernameLevel", d.UsernameLevelOne), zap.String("BusinessUsername", d.BusinessUsername))
			//记录不存在才能添加
			return errors.Wrapf(err, "Can not Insert Commission, because this UsernameLevelOne had exists")
		}

		commissionOne := &models.Commission{
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
	}

	if d.UsernameLevelTwo != "" {
		//支付成功后，需要插入佣金表Commission -  第二级
		b := new(models.Commission{
			UsernameLevel:    d.UsernameLevelTwo,
			BusinessUsername: d.BusinessUsername,
		})
		e := &models.Commission{}
		if err = s.db.Model(b).First(e).Error; err == nil {
			s.logger.Error("已经存在此用户，不能新增佣金记录", zap.String("UsernameLevel", d.UsernameLevelTwo), zap.String("BusinessUsername", d.BusinessUsername))
			//记录不存在才能添加
			return errors.Wrapf(err, "Can not Insert Commission, because this UsernameLevelTwo had exists")
		}

		commissionTwo := &models.Commission{
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
	}

	if d.UsernameLevelThree != "" {
		//支付成功后，需要插入佣金表Commission -  第三级
		b := new(models.Commission{
			UsernameLevel:    d.UsernameLevelThree,
			BusinessUsername: d.BusinessUsername,
		})
		e := &models.Commission{}
		if err = s.db.Model(b).First(e).Error; err == nil {
			s.logger.Error("已经存在此用户，不能新增佣金记录", zap.String("UsernameLevel", d.UsernameLevelThree), zap.String("BusinessUsername", d.BusinessUsername))
			//记录不存在才能添加
			return errors.Wrapf(err, "Can not Insert Commission, because this UsernameLevelThree had exists")
		}
		commissionThree := &models.Commission{
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
	}

	redisConn := s.redisPool.Get()
	defer redisConn.Close()

	//统计出NormalUserCommissionStatistics及

	return nil
}

//TODO
//商户查询当前名下用户总数，按月统计付费会员总数及返佣金额，是否已经返佣
func (s *MysqlLianmiRepository) GetBusinessMembership(isRebate bool) (*Auth.GetBusinessMembershipResp, error) {
	var err error

	redisConn := s.redisPool.Get()
	defer redisConn.Close()

	return nil, err

}
