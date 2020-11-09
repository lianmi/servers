package repositories

import (
	// "encoding/json"
	// "fmt"
	// "time"
	// "net/http"

	// "github.com/golang/protobuf/proto"
	// "github.com/gomodule/redigo/redis"
	// "github.com/jinzhu/gorm"
	// Auth "github.com/lianmi/servers/api/proto/auth"
	Auth "github.com/lianmi/servers/api/proto/auth"
	// User "github.com/lianmi/servers/api/proto/user"
	// "github.com/lianmi/servers/internal/app/dispatcher/grpcclients"
	// "github.com/lianmi/servers/internal/app/dispatcher/nsqMq"
	LMCommon "github.com/lianmi/servers/internal/common"
	"github.com/lianmi/servers/internal/pkg/models"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	// "github.com/lianmi/servers/util/dateutil"
)

//会员付费成功后，需要新增4条佣金记录
func (s *MysqlLianmiRepository) SaveToCommission(username, orderID, content string, blockNumber uint64, txHash string) error {
	var err error

	//从Distribution层级表查出所有需要分配佣金的用户账号
	d := new(models.Distribution)
	if err = s.db.Model(d).Where("username = ?", username).First(d).Error; err != nil {
		//记录找不到也会触发错误
		// fmt.Println("GetUser first error:", err.Error())
		return errors.Wrapf(err, "SaveToCommission error")
	}

	if d.BusinessUsername == "" {
		return errors.Wrapf(err, "businessUsername is empty error")
	}

	if d.UsernameLevelOne != "" {
		//支付成功后，需要插入佣金表Commission -  第一级
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
		commissionThree := &models.Commission{
			UsernameLevel:    d.UsernameLevelTwo,       //One Two Three
			BusinessUsername: d.BusinessUsername,       //归属的商户注册账号id
			Amount:           LMCommon.MEMBERSHIPPRICE, //会员费用金额，单位是人民币
			OrderID:          orderID,                  //订单ID
			Content:          content,                  //附言 Text文本类型
			BlockNumber:      blockNumber,              //交易成功打包的区块高度
			TxHash:           txHash,                   //交易成功打包的区块哈希
			Commission:       LMCommon.CommissionTwo,   //第三级佣金
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
